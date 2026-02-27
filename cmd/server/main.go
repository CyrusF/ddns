package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bytedance/ddns/internal/config"
	"github.com/bytedance/ddns/internal/server"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	store := server.NewIPStore()
	proxy := server.NewTCPProxy(store, cfg.Server.TargetPort)
	apiHandler := server.NewAPIHandler(store, cfg.Server.Token)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Start HTTP API server on port A2
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.APIPort),
		Handler: apiHandler,
	}

	go func() {
		slog.Info("api server starting", "port", cfg.Server.APIPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("api server error", "error", err)
			cancel()
		}
	}()

	// Start TCP proxy listener on port A1
	proxyAddr := fmt.Sprintf(":%d", cfg.Server.ProxyPort)
	listener, err := net.Listen("tcp", proxyAddr)
	if err != nil {
		slog.Error("failed to start proxy listener", "error", err)
		os.Exit(1)
	}

	go func() {
		slog.Info("tcp proxy starting", "port", cfg.Server.ProxyPort, "target_port", cfg.Server.TargetPort)
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					slog.Error("proxy accept error", "error", err)
					continue
				}
			}
			go proxy.HandleConn(conn)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down...")

	listener.Close()
	httpServer.Shutdown(context.Background())

	slog.Info("server stopped")
}
