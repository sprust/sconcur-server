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

	server := socket_server.NewServer("tcp", ":"+os.Getenv("SOCKET_PORT"))

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
			case <-time.After(3 * time.Second):
				slog.Error("shutdown by timeout")
			}
		}
	}
}
