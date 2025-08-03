package server

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stardustagi/TopLib/codec"
	"github.com/stardustagi/TopLib/libs/logs"
	"go.uber.org/zap"
)

type IClient interface {
	GetSessionID() string
	GetUserID() string
	GetConn() *websocket.Conn
	SendMessage() error
	ReceivedMessage() <-chan []byte
	Send(message []byte)
	Close() error
	Listen()
}

type Client struct {
	sessionId string          // 客户端ID
	conn      *websocket.Conn // WebSocket连接
	codec     codec.ICodec    // 编解码器
	logger    *zap.Logger
	ctx       context.Context
	handler   codec.IMessageProcessor
	userId    string // 用户ID,方便根据用户ID获取客户端
	send      chan []byte
	closed    chan struct{}  // 用于关闭连接的通道
	cm        IClientManager // 客户端管理器接口
}

func NewClient(userId, sessionId string, conn *websocket.Conn, codec codec.ICodec, logger *zap.Logger, ctx context.Context, handlerInterface codec.IMessageProcessor, cm IClientManager) IClient {
	return &Client{
		sessionId: sessionId,
		logger:    logger,
		codec:     codec,
		ctx:       ctx,
		userId:    userId,
		conn:      conn,
		send:      make(chan []byte, 1024), // 缓冲区大小为1024
		closed:    make(chan struct{}),
		handler:   handlerInterface,
		cm:        cm,
	}
}

func (c *Client) GetSessionID() string {
	return c.sessionId
}

func (c *Client) GetUserID() string {
	return c.userId
}

func (c *Client) GetConn() *websocket.Conn {
	return c.conn
}

func (c *Client) Send(message []byte) {
	c.send <- message
}

// 发送数据
func (c *Client) SendMessage() error {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				err := c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					c.logger.Error("Failed to write close message", logs.String("SessionId", c.sessionId), logs.ErrorInfo(err))
					<-c.closed
					return err
				}
				c.logger.Info("Send channel closed, closing connection", logs.String("SessionId", c.sessionId))
				return nil
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.logger.Error("Failed to get next writer", logs.String("SessionId", c.sessionId), logs.ErrorInfo(err))
				return err
			}
			w.Write(message)
			w.Close()
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return nil
			}
		case <-c.closed:
			c.logger.Info("Client closed", logs.String("SessionId", c.sessionId))
			return nil
		}
	}
}

// 接收数据
func (c *Client) ReceivedMessage() <-chan []byte {
	c.logger.Info("Client receiving messages", logs.String("sessionId", c.sessionId))
	messageChan := make(chan []byte, 100)
	go func() {
		defer func() {
			// manager.UnregisterClient(c)
			c.cm.UnregisterClient(c) // Unregister the client from the manager
			c.conn.Close()
			close(messageChan)
			c.logger.Info("Connection closed", logs.String("sessionId", c.sessionId))
		}()

		c.conn.SetReadLimit(512)
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.conn.SetPongHandler(func(string) error {
			c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		for {
			_, msg, err := c.conn.ReadMessage()
			if err != nil {
				c.logger.Info("read error:", logs.String("sessionId", c.sessionId), logs.ErrorInfo(err))
				break
			}
			if string(msg) == "" {
				continue
			}
			message, err := c.codec.Decode(msg)
			if err != nil {
				c.logger.Info("decode error:", logs.ErrorInfo(err))
				continue
			}

			resultMsg, err := c.handler.HandlerMessage(message)
			if err != nil {
				c.logger.Error("message handler error:", logs.ErrorInfo(err))
				continue
			}
			c.send <- []byte(resultMsg)
		}
	}()
	return messageChan
}

func (c *Client) Close() error {
	close(c.send)         // 关闭发送通道
	close(c.closed)       // 关闭closed通道
	return c.conn.Close() // 关闭WebSocket连接
}

func (c *Client) Listen() {
	c.logger.Info("Client listening", zap.String("sessionId", c.sessionId))
	go c.ReceivedMessage() // 启动收消息协程
	go c.SendMessage()     // 启动发送消息的协程
}
