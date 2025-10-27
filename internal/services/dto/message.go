package dto

import "sconcur/internal/services/types"

type Message struct {
	FlowUuid string       `json:"fu"`
	Method   types.Method `json:"md"`
	TaskKey  string       `json:"tk"`
	Payload  string       `json:"pl"`
}
