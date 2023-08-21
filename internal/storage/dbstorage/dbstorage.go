package dbstorage

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/grishagavrin/link-shortener/internal/errs"
	istorage "github.com/grishagavrin/link-shortener/internal/storage/iStorage"
	"github.com/grishagavrin/link-shortener/internal/user"
	"github.com/grishagavrin/link-shortener/internal/utils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// PostgreSQLStorage storage
type PostgreSQLStorage struct {
	dbi *pgxpool.Pool
	l   *zap.Logger
}

func New(dbi *pgxpool.Pool, l *zap.Logger) (*PostgreSQLStorage, error) {
	// Check if scheme exist
	sql := `
	CREATE TABLE IF NOT EXISTS public.short_links(
		id serial,
		user_id varchar(50),
		origin  varchar(255) not null,
		short   varchar(50)  not null,
		correlation_id varchar(100),
		is_deleted boolean default false
	);

	CREATE UNIQUE INDEX IF NOT EXISTS short_links_origin_uindex
    on public.short_links(origin);
	`

	if _, err := dbi.Exec(context.Background(), sql); err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrDatabaseExec, err)
	}

	return &PostgreSQLStorage{
		dbi: dbi,
		l:   l,
	}, nil
}

func (s *PostgreSQLStorage) GetLinkDB(ctx context.Context, shortKey istorage.ShortURL) (istorage.Origin, error) {
	var origin istorage.Origin
	var gone bool

	query := "SELECT origin, is_deleted FROM public.short_links WHERE short=$1"
	err := s.dbi.QueryRow(ctx, query, string(shortKey)).Scan(&origin, &gone)

	if gone {
		return "", errs.ErrURLIsGone
	}

	if err != nil {
		return "", errs.ErrURLNotFound
	}

	return origin, nil
}

func (s *PostgreSQLStorage) LinksByUser(ctx context.Context, userID user.UniqUser) (istorage.ShortLinks, error) {
	query := "SELECT origin, short FROM public.short_links WHERE user_id=$1"

	origins := istorage.ShortLinks{}
	rows, err := s.dbi.Query(ctx, query, string(userID))
	if err != nil {
		return origins, err
	}

	for rows.Next() {
		var origin istorage.Origin
		var short istorage.ShortURL

		err = rows.Scan(&short, &origin)
		if err != nil {
			return origins, err
		}
		origins[short] = origin
	}

	return origins, nil
}

// Save url
func (s *PostgreSQLStorage) SaveLinkDB(ctx context.Context, userID user.UniqUser, url istorage.Origin) (istorage.ShortURL, error) {

	shortKey, err := utils.RandStringBytes()
	if err != nil {
		return "", err
	}

	queryInsert := `
	INSERT INTO public.short_links (user_id, origin, short) 
	VALUES (@user_id, @origin, @short);
	`

	queryGet := `
	SELECT short FROM public.short_links where origin=$1
	`

	args := pgx.NamedArgs{
		"user_id": userID,
		"origin":  url,
		"short":   shortKey,
	}

	pgErr := &pgconn.PgError{}

	if _, err := s.dbi.Exec(ctx, queryInsert, args); err != nil {
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				var short istorage.ShortURL
				_ = s.dbi.QueryRow(ctx, queryGet, string(url)).Scan(&short)

				return short, errs.ErrAlreadyHasShort
			}
		}
		return shortKey, nil
	}

	return shortKey, nil
}

// Save url batch
func (s *PostgreSQLStorage) SaveBatch(ctx context.Context, userID user.UniqUser, urls []istorage.BatchReqURL) ([]istorage.BatchResURL, error) {
	type temp struct{ CorrID, Origin, Short string }

	var buffer []temp
	for _, v := range urls {
		shortKey, _ := utils.RandStringBytes()

		var t = temp{
			CorrID: v.CorrID,
			Origin: v.Origin,
			Short:  string(shortKey),
		}
		buffer = append(buffer, t)
	}

	var shorts []istorage.BatchResURL

	query := `
		INSERT INTO public.short_links (user_id, origin, short) 
		VALUES (@user_id, @origin, @short)
		ON CONFLICT (origin) DO NOTHING;
		`

	// Start transaction
	tx, err := s.dbi.Begin(ctx)
	defer tx.Rollback(ctx)
	if err != nil {
		return shorts, err
	}

	for _, v := range buffer {
		// Add record to transaction
		args := pgx.NamedArgs{
			"user_id":        userID,
			"origin":         v.Origin,
			"short":          v.Short,
			"correlation_id": v.CorrID,
		}

		if _, err = tx.Exec(ctx, query, args); err == nil {
			shorts = append(shorts, istorage.BatchResURL{
				Short:  v.Short,
				CorrID: v.CorrID,
			})
		} else {
			s.l.Info("Save bunch error", zap.Error(err))
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return shorts, nil
}

func (s *PostgreSQLStorage) BunchUpdateAsDeleted(ctx context.Context, shortIds []string, userID string) error {
	if len(shortIds) == 0 {
		return errs.ErrCorrelation
	}
	qtyCPU := runtime.NumCPU()

	inputCh := make(chan string)

	go func() {
		for _, v := range shortIds {
			inputCh <- v
		}
		close(inputCh)
	}()

	// fanOut
	chs := make([]chan string, 0, qtyCPU)
	for i := 0; i < qtyCPU; i++ {
		ch := make(chan string)
		chs = append(chs, ch)
	}

	go func() {
		defer func(chs []chan string) {
			for _, ch := range chs {
				close(ch)
			}
		}(chs)

		for i := 0; ; i++ {
			if i == len(chs) {
				i = 0
			}

			num, ok := <-inputCh
			if !ok {
				return
			}
			ch := chs[i]
			ch <- num
		}
	}()

	// worker
	workerChs := make([]chan *pgx.Batch, 0, qtyCPU)
	for _, fanOutCh := range chs {
		workerCh := make(chan *pgx.Batch)
		newWorker(ctx, fanOutCh, workerCh, userID)
		workerChs = append(workerChs, workerCh)
	}

	// fanIn
	wg := &sync.WaitGroup{}
	for _, inputCh := range workerChs {
		wg.Add(1)
		go func(inputCh chan *pgx.Batch) {
			defer wg.Done()
			for batch := range inputCh {
				results := s.dbi.SendBatch(ctx, batch)
				_, err := results.Exec()
				if err != nil {
					var pgErr *pgconn.PgError
					if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
						s.l.Info("update error on batch string")
					}
					s.l.Info("unable to insert batch row", zap.Error(err))
				}
				results.Close()
			}
		}(inputCh)
	}

	wg.Wait()

	return nil
}

func newWorker(ctx context.Context, inputCh chan string, out chan *pgx.Batch, userID string) {
	go func() {
		query := `
		UPDATE public.short_links 
		SET is_deleted=true 
		WHERE user_id=$1 AND short=$2;
		`

		batch := &pgx.Batch{}

		for num := range inputCh {
			batch.Queue(query, userID, num)
		}

		out <- batch
		close(out)
	}()
}
