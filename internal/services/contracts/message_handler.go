package contracts

import (
	"sconcur/internal/services/dto"
	"sconcur/internal/services/flows"
)

type MessageHandler interface {
	Handle(flow *flows.Flow, message *dto.Message) *dto.Result
}
