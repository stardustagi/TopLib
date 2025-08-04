package wsclient

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stardustagi/TopLib/codec"
	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/protocol"
	"github.com/stardustagi/TopLib/utils"
	"go.uber.org/zap"
)

// Message 对应服务器端的消息结构

type WSClient struct {
	conn     *websocket.Conn
	logger   *zap.Logger
	sendChan chan []byte
	done     chan struct{}
	ctx      context.Context
	cancel   context.CancelFunc
	handler  protocol.IMessageProcessor
}

func NewWSClient(serverURL string, headers map[string]string, pcodec codec.ICodec, msgHandler protocol.IMessageProcessor, logger *zap.Logger) (*WSClient, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	// 设置请求头
	header := make(map[string][]string)
	for k, v := range headers {
		header[k] = []string{v}
	}

	// 连接到 WebSocket 服务器
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &WSClient{
		conn:     conn,
		sendChan: make(chan []byte, 256),
		done:     make(chan struct{}),
		ctx:      ctx,
		cancel:   cancel,
		logger:   logger,
		handler:  msgHandler,
	}

	return client, nil
}

// Start 启动客户端
func (c *WSClient) Start() {
	go c.readPump()
	go c.writePump()
	go c.pingPump()
}

// readPump 处理接收消息
func (c *WSClient) readPump() {
	defer func() {
		c.conn.Close()
		close(c.done)
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				return
			}

			// 处理接收到的消息
			c.handleMessage(message)
		}
	}
}

// writePump 处理发送消息
func (c *WSClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.sendChan:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			w.Close()

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.ctx.Done():
			return
		}
	}
}

// pingPump 发送心跳
func (c *WSClient) pingPump() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.logger.Info("ping error", logs.ErrorInfo(err))
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}

// handleMessage 处理接收到的消息
func (c *WSClient) handleMessage(data []byte) {
	msg, err := utils.Bytes2Struct[protocol.Message](data)
	if err != nil {
		c.logger.Error("decode message error", zap.Error(err))
		return
	}

	c.logger.Info("Received message - Type: %s, Payload: %v\n", logs.String("main", msg.Main),
		logs.String("sub", msg.Sub), logs.String("payload", msg.Payload),
	)
	_, err = c.handler.HandlerMessage(&msg)
	if err != nil {
		c.logger.Error("decode message error", zap.Error(err))
		return
	}
}

// SendMessage 发送消息
func (c *WSClient) SendMessage(main, sub, payload string) error {
	msg := protocol.Message{
		Main:    main,
		Sub:     sub,
		Payload: payload,
	}

	data, err := utils.Struct2Bytes(msg)
	if err != nil {
		return fmt.Errorf("marshal message error: %v", err)
	}

	select {
	case c.sendChan <- []byte(data):
		return nil
	case <-c.ctx.Done():
		return fmt.Errorf("client is closed")
	}
}

// Close 关闭客户端
func (c *WSClient) Close() error {
	c.cancel()
	close(c.sendChan)
	return c.conn.Close()
}
