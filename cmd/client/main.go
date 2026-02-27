package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bytedance/ddns/internal/client"
	"github.com/bytedance/ddns/internal/config"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if cfg.Client.ServerURL == "" {
		slog.Error("client.server_url is required in config")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	c := client.New(cfg.Client.ServerURL, cfg.Client.Token, cfg.Client.Interval)

	slog.Info("client starting", "server", cfg.Client.ServerURL, "interval", cfg.Client.Interval)
	c.Run(ctx)
}
