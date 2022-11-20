package model

import (
	"encoding/json"
)

type LogMessage struct {
	RouterName string `json:"RouterName"`
}

func (l LogMessage) Decode(message string) (*LogMessage, error) {
	return &l, json.Unmarshal([]byte(message), &l)
}
