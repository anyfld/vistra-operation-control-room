package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/gen/proto/v1/protov1connect"
)

const (
	readHeaderTimeout = 10 * time.Second
	shutdownTimeout   = 30 * time.Second
)

type ExampleServiceHandler struct{}

func (h *ExampleServiceHandler) Ping(
	ctx context.Context,
	req *connect.Request[protov1.PingRequest],
) (*connect.Response[protov1.PingResponse], error) {
	return connect.NewResponse(&protov1.PingResponse{
		Message: "pong",
	}), nil
}

func main() {
	handler := &ExampleServiceHandler{}
	path, httpHandler := protov1connect.NewExampleServiceHandler(handler)

	mux := http.NewServeMux()
	mux.Handle(path, httpHandler)

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	server := &http.Server{ //nolint:exhaustruct
		Addr:              addr,
		Handler:           h2c.NewHandler(mux, &http2.Server{}), //nolint:exhaustruct
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		log.Println("Shutting down server...")

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Starting server on %s", addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
