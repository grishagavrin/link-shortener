package handlersgrpc

import (
	"context"
	"errors"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/errs"
	ls "github.com/grishagavrin/link-shortener/internal/proto"
	"github.com/grishagavrin/link-shortener/internal/storage/models"
	"github.com/grishagavrin/link-shortener/internal/utils/db"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Repository interface for working with global storage
type Repository interface {
	GetLinkDB(context.Context, models.ShortURL) (models.Origin, error)
}

// GRPCHandlers поддерживает все необходимые методы сервера.
type GRPCHandler struct {
	ls.UnimplementedApiServiceServer
	l    *zap.Logger
	stor Repository
}

func New(stor Repository, l *zap.Logger) *GRPCHandler {
	return &GRPCHandler{
		l:    l,
		stor: stor,
	}
}

func (s *GRPCHandler) GetLink(ctx context.Context, url *ls.GetLinkReq) (*ls.GetLinkRes, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var response ls.GetLinkRes

	if len(url.Id) != config.LENHASH {
		return nil, status.Errorf(codes.InvalidArgument, errs.ErrCorrectURL.Error())
	}

	s.l.Info("Get ID:", zap.String("id", url.Id))

	s.l.Info("Get ID:", zap.String("id", url.Id))
	foundedURL, err := s.stor.GetLinkDB(ctx, models.ShortURL(url.Id))

	if err != nil {
		if errors.Is(err, errs.ErrURLIsGone) {
			s.l.Info(errs.ErrURLIsGone.Error(), zap.Error(err))
			return nil, status.Errorf(codes.AlreadyExists, errs.ErrCorrectURL.Error())
		}

		s.l.Info(errs.ErrBadRequest.Error(), zap.Error(err))
		return nil, status.Errorf(codes.NotFound, errs.ErrCorrectURL.Error())

	}

	header := metadata.Pairs("Location", string(foundedURL))
	grpc.SendHeader(ctx, header)

	return &response, nil
}

func (s *GRPCHandler) GetPing(ctx context.Context, empt *emptypb.Empty) (*ls.GetPingRes, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var response ls.GetPingRes

	conn, err := db.SQLDBConnection(s.l)
	if err == nil {
		err = conn.Ping(ctx)
		if err != nil {
			return nil, status.Error(codes.Internal, errs.ErrInternalSrv.Error())
		}

		return &response, nil
	} else {
		s.l.Info("not connect to db", zap.Error(err))
		return nil, status.Error(codes.Internal, errs.ErrInternalSrv.Error())
	}
}
