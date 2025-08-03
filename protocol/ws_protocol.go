package protocol

import (
	"github.com/stardustagi/TopLib/codec"
)

func NewWsKeepalive() *codec.IMessage {
	payload := map[string]string{"data": "ping"}
	keepLive := codec.NewJsonMessage("system", "keepalive", payload)
	return &keepLive
}
