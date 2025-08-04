package protocol

import (
	"fmt"

	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/utils"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Message struct {
	Main    string `json:"main"`
	Sub     string `json:"sub"`
	Payload string `json:"payload"`
}

type IMessage interface {
	GetPayload() string
	GetMain() string
	GetSub() string
}

func (m *Message) GetPayload() string {
	return m.Payload
}

func (m *Message) GetMain() string {
	return m.Main
}

func (m *Message) GetSub() string {
	return m.Sub
}

// NewMessage 创建一个新的消息
func NewJsonMessage[T any](main, sub string, data T) IMessage {
	payload, err := utils.Struct2Bytes(data)
	if err != nil {
		return nil // 处理错误
	}
	return &Message{
		Main:    main,
		Sub:     sub,
		Payload: payload,
	}
}

// NewProtobufMessage 创建一个新的Protobuf消息
func NewProtobufMessage(main, sub string, msg proto.Message) IMessage {
	payload, err := proto.Marshal(msg)
	if err != nil {
		return nil // 处理错误
	}
	return &Message{
		Main:    main,
		Sub:     sub,
		Payload: string(payload),
	}
}

// 消息处理接口
type IMessageProcessor interface {
	HandlerMessage(msg IMessage) (string, error)
}

// DefaultMessageProcessor 默认消息处理器
type DefaultMessageProcessor struct {
	logger  *zap.Logger
	mainCmd string
}

func NewDefaultMessageHandler() IMessageProcessor {
	return &DefaultMessageProcessor{}
}

func (m *DefaultMessageProcessor) HandlerMessage(msg IMessage) (string, error) {
	result := fmt.Sprintf("you send message %s:", string(msg.GetPayload()))
	resultMsg := NewJsonMessage("system", "echo", result)
	bResult, err := (utils.Struct2Bytes(resultMsg)) // Convert the message to bytes for logging
	if err != nil {
		logs.Error("Failed to convert message to bytes", logs.ErrorInfo(err))
		return "", err
	}
	result = string(bResult) // Convert bytes back to string for logging
	logs.Info("message handler", logs.String("result", result))
	return result, nil
}
