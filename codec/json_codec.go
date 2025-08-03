package codec

import (
	"github.com/stardustagi/TopLib/utils"
)

type ICodec interface {
	// Decode 解码
	Decode(data []byte) (IMessage, error)
	Encode(message IMessage) (string, error)
}

type JsonCodec struct {
}

func NewJsonCodec() ICodec {
	return &JsonCodec{}
}

func (c *JsonCodec) Decode(data []byte) (IMessage, error) {
	msg, err := utils.Bytes2Struct[Message](data)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (c *JsonCodec) Encode(message IMessage) (string, error) {
	byteInfo, err := utils.Struct2Bytes(message)
	if err != nil {
		return "", err
	}
	return byteInfo, nil
}
