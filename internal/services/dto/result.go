package dto

import "sconcur/internal/services/types"

type Result struct {
	FlowUuid string       `json:"fu"`
	Method   types.Method `json:"md"`
	TaskKey  string       `json:"tk"`
	Waitable bool         `json:"wt"`
	IsError  bool         `json:"er"`
	Payload  string       `json:"pl"`
	HasNext  bool         `json:"hn"`
}
