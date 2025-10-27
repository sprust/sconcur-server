package contracts

import (
	"context"
	"sconcur/internal/services/dto"
)

type MessageHandler interface {
	Handle(ctx context.Context, message *dto.Message) *dto.Result
}
