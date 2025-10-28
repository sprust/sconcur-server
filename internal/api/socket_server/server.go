package socket_server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sconcur/internal/services/connection"
	"sconcur/internal/services/features"
	"sconcur/pkg/foundation/errs"
	"sync/atomic"
	"time"
)

type Server struct {
	network  string
	address  string
	handler  *features.Handler
	listener net.Listener
	closing  atomic.Bool
}

func NewServer(network string, address string) *Server {
	return &Server{
		network: network,
		address: address,
		handler: features.NewHandler(),
	}
}

func (s *Server) Run(ctx context.Context) error {
	listener, err := net.Listen(s.network, s.address)

	if err != nil {
		return errs.Err(err)
	}

	slog.Info("Socket server started on " + s.network + "://" + s.address)

	s.listener = listener

	go func(ctx context.Context, s *Server) {
		select {
		case <-ctx.Done():
			slog.Warn("Shutting down [socket server] by context")

			s.closing.Store(true)

			if s.listener != nil {
				_ = s.listener.Close()
			}
		}
	}(ctx, s)

	for {
		conn, err := listener.Accept()

		if err != nil {
			if s.closing.Load() {
				break
			}

			var ne net.Error

			if errors.As(err, &ne) && ne.Temporary() {
				continue
			}

			return errs.Err(err)
		}

		go func(ctx context.Context, s *Server, conn net.Conn) {
			err := s.handleConnection(ctx, conn)

			if err != nil {
				slog.Error(err.Error())
			}
		}(ctx, s, conn)
	}

	for {
		flowsCount := s.handler.GetFlowsCount()

		if flowsCount == 0 {
			break
		}

		slog.Info(
			fmt.Sprintf(
				"Waiting for flows finishing [%d]",
				flowsCount,
			),
		)

		time.Sleep(1 * time.Second)
	}

	return nil
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) error {
	transport := connection.NewTransport(conn)

	defer func(transport *connection.Transport) {
		_ = transport.Close()
	}(transport)

	message, err := transport.ReadMessage()

	if err != nil {
		return errs.Err(err)
	}

	slog.Debug(fmt.Sprintf("received message: %+v", message))

	err = s.handler.Handle(ctx, transport, message)

	if err != nil {
		return errs.Err(err)
	}

	return nil
}
