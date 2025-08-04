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

// MessageSender 消息发送器
type MessageSender struct {
	conn       *websocket.Conn
	codec      codec.ICodec
	sendChan   chan []byte
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	logger     *zap.Logger
	closed     chan struct{}
	pingTicker *time.Ticker
}

// IMessageSender 消息发送器接口
type IMessageSender interface {
	Start()
	Stop()
	SendMessage(msg protocol.IMessage) error
	SendRawBytes(data []byte) error
	IsRunning() bool
}

// NewMessageSender 创建新的消息发送器
func NewMessageSender(conn *websocket.Conn, codec codec.ICodec, logger *zap.Logger) IMessageSender {
	ctx, cancel := context.WithCancel(context.Background())

	return &MessageSender{
		conn:       conn,
		codec:      codec,
		sendChan:   make(chan []byte, 256),
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
		closed:     make(chan struct{}),
		pingTicker: time.NewTicker(54 * time.Second),
	}
}

// Start 启动消息发送器
func (s *MessageSender) Start() {
	s.wg.Add(1)
	go s.sendPump()
	s.logger.Info("Message sender started")
}

// Stop 停止消息发送器
func (s *MessageSender) Stop() {
	s.logger.Info("Stopping message sender...")

	s.cancel()
	s.pingTicker.Stop()
	close(s.sendChan)

	// 等待发送协程结束
	s.wg.Wait()

	close(s.closed)
	s.logger.Info("Message sender stopped")
}

// SendMessage 发送消息对象
func (s *MessageSender) SendMessage(msg protocol.IMessage) error {
	// 将消息编码为字节
	data, err := s.codec.Encode(msg)
	if err != nil {
		s.logger.Error("Failed to encode message",
			logs.String("main", msg.GetMain()),
			logs.String("sub", msg.GetSub()),
			logs.ErrorInfo(err))
		return err
	}

	return s.SendRawBytes([]byte(data))
}

// SendRawBytes 发送原始字节数据
func (s *MessageSender) SendRawBytes(data []byte) error {
	select {
	case s.sendChan <- data:
		s.logger.Debug("Message queued for sending", logs.Int("size", len(data)))
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	case <-time.After(5 * time.Second):
		s.logger.Warn("Send queue is full, message timeout")
		return context.DeadlineExceeded
	}
}

// IsRunning 检查发送器是否运行中
func (s *MessageSender) IsRunning() bool {
	select {
	case <-s.ctx.Done():
		return false
	default:
		return true
	}
}

// sendPump 消息发送泵
func (s *MessageSender) sendPump() {
	defer s.wg.Done()

	for {
		select {
		case message, ok := <-s.sendChan:
			if !ok {
				// 发送通道关闭
				s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			if err := s.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				s.logger.Error("Failed to write message", logs.ErrorInfo(err))
				return
			}

			s.logger.Debug("Message sent successfully", logs.Int("size", len(message)))

		case <-s.pingTicker.C:
			s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				s.logger.Error("Failed to send ping", logs.ErrorInfo(err))
				return
			}
			s.logger.Debug("Ping message sent")

		case <-s.ctx.Done():
			s.logger.Info("Send pump cancelled")
			return
		}
	}
}
