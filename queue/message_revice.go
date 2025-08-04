package queue

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stardustagi/TopLib/codec"
	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/protocol"
	"go.uber.org/zap"
)

// MessageReceiver 消息接收器
type MessageReceiver struct {
	conn        *websocket.Conn
	codec       codec.ICodec
	processor   protocol.IMessageProcessor
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	logger      *zap.Logger
	messageChan chan protocol.IMessage
	sender      IMessageSender
	closed      chan struct{}
}

// IMessageReceiver 消息接收器接口
type IMessageReceiver interface {
	Start()
	Stop()
	GetMessageChan() <-chan protocol.IMessage
	SetProcessor(processor protocol.IMessageProcessor)
	SetSender(sender IMessageSender)
	IsRunning() bool
}

// NewMessageReceiver 创建新的消息接收器
func NewMessageReceiver(conn *websocket.Conn, codec codec.ICodec, processor protocol.IMessageProcessor, logger *zap.Logger) IMessageReceiver {
	ctx, cancel := context.WithCancel(context.Background())

	return &MessageReceiver{
		conn:        conn,
		codec:       codec,
		processor:   processor,
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger,
		messageChan: make(chan protocol.IMessage, 100),
		closed:      make(chan struct{}),
	}
}

// Start 启动消息接收器
func (r *MessageReceiver) Start() {
	r.wg.Add(1)
	go r.receivePump()
	r.logger.Info("Message receiver started")
}

// Stop 停止消息接收器
func (r *MessageReceiver) Stop() {
	r.logger.Info("Stopping message receiver...")

	r.cancel()

	// 等待接收协程结束
	r.wg.Wait()

	close(r.messageChan)
	close(r.closed)
	r.logger.Info("Message receiver stopped")
}

// GetMessageChan 获取消息通道
func (r *MessageReceiver) GetMessageChan() <-chan protocol.IMessage {
	return r.messageChan
}

// SetProcessor 设置消息处理器
func (r *MessageReceiver) SetProcessor(processor protocol.IMessageProcessor) {
	r.processor = processor
}

// SetSender 设置消息发送器（用于回复消息）
func (r *MessageReceiver) SetSender(sender IMessageSender) {
	r.sender = sender
}

// IsRunning 检查接收器是否运行中
func (r *MessageReceiver) IsRunning() bool {
	select {
	case <-r.ctx.Done():
		return false
	default:
		return true
	}
}

// receivePump 消息接收泵
func (r *MessageReceiver) receivePump() {
	defer r.wg.Done()

	r.conn.SetReadLimit(512)
	r.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	r.conn.SetPongHandler(func(string) error {
		r.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-r.ctx.Done():
			r.logger.Info("Receive pump cancelled")
			return
		default:
			_, msgData, err := r.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					r.logger.Error("Unexpected close error", logs.ErrorInfo(err))
				} else {
					r.logger.Info("Connection closed", logs.ErrorInfo(err))
				}
				return
			}

			// 处理接收到的消息
			r.handleRawMessage(msgData)
		}
	}
}

// handleRawMessage 处理原始消息数据
func (r *MessageReceiver) handleRawMessage(data []byte) {
	if len(data) == 0 {
		return
	}

	// 解码消息
	message, err := r.codec.Decode(data)
	if err != nil {
		r.logger.Error("Failed to decode message", logs.ErrorInfo(err))
		return
	}

	// 将消息转换为 IMessage 接口
	msg, ok := message.(protocol.IMessage)
	if !ok {
		r.logger.Error("Message does not implement IMessage interface")
		return
	}

	r.logger.Debug("Message received",
		logs.String("main", msg.GetMain()),
		logs.String("sub", msg.GetSub()))

	// 发送到消息通道
	select {
	case r.messageChan <- msg:
	case <-r.ctx.Done():
		return
	default:
		r.logger.Warn("Message channel is full, dropping message")
	}

	// 如果有处理器，则处理消息
	if r.processor != nil {
		go r.processMessage(msg)
	}
}

// processMessage 处理消息并发送回复
func (r *MessageReceiver) processMessage(msg protocol.IMessage) {
	// 检查上下文是否已取消
	select {
	case <-r.ctx.Done():
		r.logger.Debug("Processing cancelled, skipping message",
			logs.String("main", msg.GetMain()),
			logs.String("sub", msg.GetSub()))
		return
	default:
	}

	resultMsg, err := r.processor.HandlerMessage(msg)
	if err != nil {
		r.logger.Error("Message processing failed",
			logs.String("main", msg.GetMain()),
			logs.String("sub", msg.GetSub()),
			logs.ErrorInfo(err))
		return
	}

	// 再次检查上下文状态，避免在处理过程中被取消
	select {
	case <-r.ctx.Done():
		r.logger.Debug("Processing cancelled after handler, skipping reply")
		return
	default:
	}

	// 如果有发送器且有结果，发送回复
	if r.sender != nil && resultMsg != "" {
		// 检查发送器是否还在运行
		if !r.sender.IsRunning() {
			r.logger.Warn("Sender is not running, skipping reply")
			return
		}

		err = r.sender.SendRawBytes([]byte(resultMsg))
		if err != nil {
			r.logger.Error("Failed to send reply", logs.ErrorInfo(err))
		}
	}
}
