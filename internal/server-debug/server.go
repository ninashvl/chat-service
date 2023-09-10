package serverdebug

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/ninashvl/chat-service/internal/buildinfo"
	"github.com/ninashvl/chat-service/internal/logger"
)

const (
	readHeaderTimeout = time.Second
	shutdownTimeout   = 3 * time.Second
)

//go:generate options-gen -out-filename=server_options.gen.go -from-struct=Options
type Options struct {
	addr string `option:"mandatory" validate:"required,hostname_port"`
}

type Server struct {
	lg  *zap.Logger
	srv *http.Server
}

func New(opts Options) (*Server, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validation debug server options error: %v", err)
	}

	lg := zap.L().Named("server-debug")

	e := echo.New()
	e.Use(middleware.Recover())

	s := &Server{
		lg: lg,
		srv: &http.Server{
			Addr:              opts.addr,
			Handler:           e,
			ReadHeaderTimeout: readHeaderTimeout,
		},
	}

	index := newIndexPage()
	index.addPage("/version", "Get build information")
	index.addPage("/debug/pprof/", "Go std profiler")
	index.addPage("/debug/pprof/profile?seconds=30", "Take half min profile")

	e.GET("/", index.handler)
	e.GET("/version", s.Version)
	e.GET("/log/level", s.getLogLevel)
	e.PUT("/log/level", s.setLogLevel)
	pprof.Register(e, "/debug/pprof")

	return s, nil
}

func (s *Server) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		return s.srv.Shutdown(ctx)
	})

	eg.Go(func() error {
		s.lg.Info("listen and serve", zap.String("addr", s.srv.Addr))

		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %v", err)
		}
		return nil
	})

	return eg.Wait()
}

func (s *Server) Version(eCtx echo.Context) error {
	data, err := json.Marshal(buildinfo.BuildInfo)
	if err != nil {
		return eCtx.NoContent(http.StatusInternalServerError)
	}
	return eCtx.JSONBlob(http.StatusOK, data)
}

func (s *Server) getLogLevel(eCtx echo.Context) error {
	s.lg.Debug("getting log level")
	return eCtx.String(http.StatusOK, logger.LogLevel())
}

func (s *Server) setLogLevel(eCtx echo.Context) error {
	lvl := eCtx.FormValue("level")

	err := logger.SetLogLevel(logger.NewOptions(strings.ToLower(lvl)))
	s.lg.Debug("setting log level", zap.String("level", lvl), zap.Error(err))
	if err != nil {
		return eCtx.NoContent(http.StatusBadRequest)
	}

	return eCtx.String(http.StatusOK, s.lg.Level().String())
}
