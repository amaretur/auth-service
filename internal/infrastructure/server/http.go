package server 

import (
	"fmt"
	"time"
	"context"
	"net/http"

	"github.com/amaretur/auth-service/pkg/log"
)

type HttpConfig struct {
	Port			int
	Handler			http.Handler
	MaxHeaderBytes	int
	ReadTimeout		time.Duration
	WriteTimeout	time.Duration
}

type Http struct {
	logger		log.Logger
	httpServer	*http.Server
}

func NewHttp(logger log.Logger) *Http {
	return &Http{
		logger: logger,
	}
}

func (s *Http) Run(errChan chan error, conf *HttpConfig) {

	s.httpServer = &http.Server {
		Addr:			fmt.Sprintf(":%d", conf.Port),
		Handler:		conf.Handler,
		MaxHeaderBytes:	1 << conf.MaxHeaderBytes,
		ReadTimeout:	conf.ReadTimeout * time.Second,
		WriteTimeout:	conf.WriteTimeout * time.Second,
	}

	s.logger.Infof("Starting HTTP server on port %d...\n", conf.Port)

	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		errChan <- fmt.Errorf("start http server error: %s\n", err)
	}

	errChan <- nil
}

func (s *Http) Shutdown() error {

	ctx, cancel := context.WithTimeout(
		context.Background(), 10 * time.Second,
	)

	defer cancel()

	return s.httpServer.Shutdown(ctx)
}
