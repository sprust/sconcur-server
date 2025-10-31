package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sconcur/internal/api/socket_server"
	"sconcur/pkg/foundation/logging"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()

	if err != nil {
		panic(err)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.Init()
	defer func(logger *logging.Logger) {
		err := logger.Close()

		if err != nil {
			panic(err)
		}
	}(logger)

	signals := make(chan os.Signal, 4)
	defer signal.Stop(signals)

	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	network := os.Getenv("SOCKET_NETWORK")
	address := os.Getenv("SOCKET_ADDRESS")

	if network == "unix" {
		if _, err := os.Stat(address); err == nil {
			if err := os.Remove(address); err != nil {
				panic(err)
			}

			slog.Warn("Removed old socket file: " + address)
		}
	}

	server := socket_server.NewServer(network, address)

	done := make(chan error, 1)

	go func(ctx context.Context) {
		done <- server.Run(ctx)
	}(ctx)

	select {
	case err := <-done:
		if err != nil {
			panic(err)
		}

		slog.Warn("Completed successfully")
	case sgn := <-signals:
		switch sgn {
		case syscall.SIGTERM, os.Interrupt:
			if sgn == syscall.SIGTERM {
				slog.Warn("Received stop (SIGTERM) signal")
			} else {
				slog.Warn("Received interrupt signal (Ctrl+C)")
			}

			cancel()

			select {
			case err := <-done:
				if err != nil {
					panic(err)
				}

				slog.Warn("Completed successfully by signal")
			case <-time.After(5 * time.Second):
				slog.Error("shutdown by timeout")
			}
		}
	}
}
