//go:build debug
// +build debug

package client

type ClientState struct {
	ID      int   `json:"id"`
	Running int32 `json:"running"`
}
