package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi"
	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/handlers/middlewares"
	"github.com/grishagavrin/link-shortener/internal/logger"
	"github.com/grishagavrin/link-shortener/internal/storage"
	"github.com/grishagavrin/link-shortener/internal/storage/dbstorage"
	"github.com/grishagavrin/link-shortener/internal/storage/ramstorage"
	"github.com/grishagavrin/link-shortener/internal/user"
	"github.com/grishagavrin/link-shortener/internal/utils/db"
	"go.uber.org/zap"
)

type Handler struct {
	s storage.Repository
}

var errEmptyBody = errors.New("body is empty")
var errFieldsJSON = errors.New("invalid fields in json")
var errInternalSrv = errors.New("internal error on server")
var errCorrectURL = fmt.Errorf("enter correct url parameter")
var errNoContent = errors.New("no content")
var errBadRequest = errors.New("bad request")

var myCook string = "default"

func New() (*Handler, error) {
	_, err := db.Instance()
	if err == nil {
		logger.Info("Set db handler")
		s, err := dbstorage.New()
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		return &Handler{
			s: s,
		}, nil
	} else {
		r, err := ramstorage.New()
		if err != nil {
			return nil, err
		}
		return &Handler{s: r}, nil
	}
}

func (h *Handler) GetLink(res http.ResponseWriter, req *http.Request) {
	q := chi.URLParam(req, "id")
	if len(q) != config.LENHASH {
		http.Error(res, errCorrectURL.Error(), http.StatusBadRequest)
		return
	}

	foundedURL, err := h.s.GetLinkDB(user.UniqUser("all"), storage.URLKey(q))
	if err == nil {
		http.Redirect(res, req, string(foundedURL), http.StatusTemporaryRedirect)
		return
	} else {
		if errors.Is(err, dbstorage.ErrURLIsGone) {
			http.Error(res, dbstorage.ErrURLIsGone.Error(), http.StatusGone)
			return
		}
		logger.Info("Get error", zap.Error(err))
	}
	http.Error(res, err.Error(), http.StatusBadRequest)
}

func (h *Handler) SaveBatch(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)

	if err != nil {
		setBadResponse(res, errInternalSrv)
		return
	}
	var urls []storage.BatchURL
	err = json.Unmarshal(body, &urls)
	if err != nil {
		setBadResponse(res, errCorrectURL)
		return
	}

	shorts, err := h.s.SaveBatch(urls)
	if err != nil {
		setBadResponse(res, errInternalSrv)
		return
	}

	baseURL, err := config.Instance().GetCfgValue(config.BaseURL)
	if err != nil {
		setBadResponse(res, errBadRequest)
	}

	// Prepare results
	for k := range shorts {
		shorts[k].Short = fmt.Sprintf("%s/%s", baseURL, shorts[k].Short)
	}
	body, err = json.Marshal(shorts)
	if err == nil {
		// Prepare response
		res.Header().Add("Content-Type", "application/json; charset=utf-8")
		res.WriteHeader(http.StatusCreated)
		_, err = res.Write(body)
		if err == nil {
			return
		}
	}

	setBadResponse(res, errInternalSrv)
}

func (h *Handler) SaveTXT(res http.ResponseWriter, req *http.Request) {
	baseURL, err := config.Instance().GetCfgValue(config.BaseURL)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	b, _ := io.ReadAll(req.Body)
	body := string(b)

	if body == "" {
		http.Error(res, errEmptyBody.Error(), http.StatusBadRequest)
		return
	}

	userIDCtx := req.Context().Value(middlewares.UserIDCtxName)
	userID := "default"
	if userIDCtx != nil {
		// Convert interface type to user.UniqUser
		userID = userIDCtx.(string)
	}

	urlKey, err := h.s.SaveLinkDB(user.UniqUser(userID), storage.ShortURL(body))
	status := http.StatusCreated

	if errors.Is(err, dbstorage.ErrAlreadyHasShort) {
		status = http.StatusConflict
	}

	response := fmt.Sprintf("%s/%s", baseURL, urlKey)
	res.WriteHeader(status)
	res.Write([]byte(response))
}

func (h *Handler) SaveJSON(res http.ResponseWriter, req *http.Request) {
	baseURL, err := config.Instance().GetCfgValue(config.BaseURL)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	body, _ := io.ReadAll(req.Body)
	reqBody := struct {
		URL string `json:"url"`
	}{}

	decJSON := json.NewDecoder(strings.NewReader(string(body)))
	decJSON.DisallowUnknownFields()

	if err := decJSON.Decode(&reqBody); err != nil {
		http.Error(res, errFieldsJSON.Error(), http.StatusBadRequest)
		return
	}

	userIDCtx := req.Context().Value(middlewares.UserIDCtxName)
	userID := "default"
	if userIDCtx != nil {
		// Convert interface type to user.UniqUser
		userID = userIDCtx.(string)
	}
	dbURL, err := h.s.SaveLinkDB(user.UniqUser(userID), storage.ShortURL(reqBody.URL))
	status := http.StatusCreated
	if errors.Is(err, dbstorage.ErrAlreadyHasShort) {
		status = http.StatusConflict
	}
	// }

	resBody := struct {
		Result string `json:"result"`
	}{
		Result: fmt.Sprintf("%s/%s", baseURL, dbURL),
	}

	js, err := json.Marshal(resBody)
	if err != nil {
		http.Error(res, errInternalSrv.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("content-type", "application/json")
	res.WriteHeader(status)
	res.Write(js)
}

func (h *Handler) GetPing(res http.ResponseWriter, req *http.Request) {
	conn, err := db.Instance()
	if err == nil {
		if err := conn.Ping(req.Context()); err == nil {
			res.WriteHeader(http.StatusOK)
		}
	} else {
		logger.Info("not connect to db", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Handler) GetLinks(res http.ResponseWriter, req *http.Request) {
	userIDCtx := req.Context().Value(middlewares.UserIDCtxName)
	// Convert interface type to user.UniqUser
	userID := userIDCtx.(string)

	links, err := h.s.LinksByUser(user.UniqUser(userID))
	if err != nil {
		http.Error(res, errNoContent.Error(), http.StatusNoContent)
		return
	}

	type coupleLinks struct {
		Short  string `json:"short_url"`
		Origin string `json:"original_url"`
	}
	var lks []coupleLinks
	baseURL, _ := config.Instance().GetCfgValue(config.BaseURL)

	// Get all links
	for k, v := range links {
		lks = append(lks, coupleLinks{
			Short:  fmt.Sprintf("%s/%s", baseURL, string(k)),
			Origin: string(v),
		})
	}

	body, err := json.Marshal(lks)
	if err == nil {
		// Prepare response
		res.Header().Add("Content-Type", "application/json; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		_, err = res.Write(body)
		if err == nil {
			return
		}
	}
}

func (h *Handler) DeleteBatch(res http.ResponseWriter, req *http.Request) {
	const workersCount = 10
	body, err := io.ReadAll(req.Body)

	if err != nil {
		setBadResponse(res, errInternalSrv)
		return
	}

	var correlationIDs []string
	err = json.Unmarshal(body, &correlationIDs)
	if err != nil {
		setBadResponse(res, errCorrectURL)
		return
	}
	// Validate count
	if len(correlationIDs) == 0 {
		setBadResponse(res, errCorrectURL)
		return
	}

	userIDCtx := req.Context().Value(middlewares.UserIDCtxName)
	userID := "default"
	if userIDCtx != nil {
		// Convert interface type to user.UniqUser
		userID = userIDCtx.(string)
	}
	inputCh := make(chan string)
	go func() {
		for _, id := range correlationIDs {
			inputCh <- id
		}
		close(inputCh)
	}()

	// здесь fanOut
	fanOutChs := fanOut(inputCh, workersCount)
	// fanOutChs range all slices
	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	for _, fanOutCh := range fanOutChs {
		// To bunch saving
		fanInSave(ctx, fanOutCh, errCh, wg, userID)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	if err := <-errCh; err != nil {
		fmt.Println("Handler error")
		setBadResponse(res, errInternalSrv)
		cancel()
		return
	}

	res.WriteHeader(http.StatusAccepted)
	cancel()
}

// setBadRequest set bad response
func setBadResponse(w http.ResponseWriter, e error) {
	http.Error(w, e.Error(), http.StatusBadRequest)
}

func fanInSave(ctx context.Context, input <-chan string, errCh chan<- error, wg *sync.WaitGroup, userID string) {
	wg.Add(1)
	go func() {
		var IDs []string
		var defErr error

		defer func() {
			if defErr != nil {
				select {
				case errCh <- defErr:
				case <-ctx.Done():
					fmt.Println("Aborting")
				}
			}
			wg.Done()
		}()

		for ID := range input {
			IDs = append(IDs, ID)
		}
		err := dbstorage.BunchUpdateAsDeleted(ctx, IDs, userID)

		if err != nil {
			defErr = err
			return
		}
	}()
}

func fanOut(inputCh chan string, n int) []chan string {
	chs := make([]chan string, 0, n)
	for i := 0; i < n; i++ {
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

			url, ok := <-inputCh
			if !ok {
				return
			}

			ch := chs[i]
			ch <- url
		}
	}()

	return chs
}

func newWorker(input, out chan string) {
	go func() {
		for num := range input {
			// out <- num / 2
			out <- num
		}

		close(out)
	}()
}
