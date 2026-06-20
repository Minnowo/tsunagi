package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"tsunagi/src/api"
	"tsunagi/src/api/grpcapi"
	"tsunagi/src/api/httpapi"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
)

func CmdServerMain(ctx context.Context, c *cli.Command) error {

	host := c.Value("host").(string)
	httpPort := c.Value("http-port").(uint32)
	grpcPort := c.Value("grpc-port").(uint32)

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 2)

	base := &api.TsunagiBase{}
	base.Init(nil)

	httpServer, err := buildHttpServer(base, host, int(httpPort))

	if err != nil {
		return err
	}

	grpcServer, grpcLis, err := buildGrpcServer(base, host, int(grpcPort))

	if err != nil {
		return err
	}

	go func() {
		log.Info().Str("addr", httpServer.Addr).Msg("http listening")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("http: %w", err)
		}
		errCh <- nil
	}()

	go func() {
		log.Info().Str("addr", grpcLis.Addr().String()).Msg("grpc listening")

		if err := grpcServer.Serve(grpcLis); err != nil {
			errCh <- fmt.Errorf("grpc: %w", err)
		}
		errCh <- nil
	}()

	// wait for signal or first server error
	select {
	case <-ctx.Done():
		log.Info().Msg("shutting down")
	case err := <-errCh:
		if err != nil {
			log.Error().Err(err).Msg("server error")
		}
	}

	// grpcServer.GracefulStop()
	grpcServer.Stop()

	if err := httpServer.Shutdown(context.Background()); err != nil {
		log.Error().Err(err).Msg("http shutdown error")
	}

	// drain second goroutine
	<-errCh

	return nil
}

func corsAllowAll(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func buildHttpServer(base *api.TsunagiBase, host string, port int) (*http.Server, error) {

	r := chi.NewRouter()
	r.Use(corsAllowAll)

	server := httpapi.HttpRelayApi{
		TsunagiBase: base,
	}
	server.Init(nil)
	server.Register(r)

	return &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: r,
	}, nil
}

func buildGrpcServer(base *api.TsunagiBase, host string, port int) (*grpc.Server, net.Listener, error) {

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, nil, err
	}

	s := grpc.NewServer()

	server := grpcapi.RelayApi{
		TsunagiBase: base,
	}
	server.Init(nil)
	server.Register(s)

	return s, lis, nil
}
