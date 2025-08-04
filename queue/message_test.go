package queue

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stardustagi/TopLib/codec"
	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/protocol"
)

func TestQueue(t *testing.T) {
	// 创建 Echo 实例
	e := echo.New()

	// 创建 WebSocket 升级器
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// 添加 WebSocket 路由
	e.GET("/ws", func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		defer conn.Close()

		// 简单的回声服务器
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			// 回复收到的消息
			err = conn.WriteMessage(messageType, message)
			if err != nil {
				break
			}
		}
		return nil
	})

	// 启动服务器
	go func() {
		e.Logger.Fatal(e.Start(":8888"))
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 创建客户端连接
	u := url.URL{Scheme: "ws", Host: "localhost:8888", Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	// 初始化组件
	logger := logs.GetLogger("message_test")
	codecInstance := codec.NewJsonCodec()
	processor := protocol.NewDefaultMessageHandler()

	// 创建发送器和接收器
	sender := NewMessageSender(conn, codecInstance, logger)
	receiver := NewMessageReceiver(conn, codecInstance, processor, logger)

	// 设置发送器到接收器（用于回复消息）
	receiver.SetSender(sender)

	// 启动发送器和接收器
	sender.Start()
	receiver.Start()

	// 用于接收消息的通道
	receivedMessages := make(chan protocol.IMessage, 10)

	// 监听接收到的消息
	go func() {
		for msg := range receiver.GetMessageChan() {
			logger.Info("Received message",
				logs.String("main", msg.GetMain()),
				logs.String("sub", msg.GetSub()),
				logs.String("payload", msg.GetPayload()))
			receivedMessages <- msg
		}
	}()

	// 发送测试消息
	testMsg := protocol.NewJsonMessage("test", "hello", map[string]string{
		"message": "Hello World",
		"id":      "test123",
	})

	err = sender.SendMessage(testMsg)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 等待接收消息
	select {
	case receivedMsg := <-receivedMessages:
		if receivedMsg.GetMain() != "test" {
			t.Errorf("Expected main='test', got main='%s'", receivedMsg.GetMain())
		}
		if receivedMsg.GetSub() != "hello" {
			t.Errorf("Expected sub='hello', got sub='%s'", receivedMsg.GetSub())
		}
		t.Logf("Successfully received message: main=%s, sub=%s",
			receivedMsg.GetMain(), receivedMsg.GetSub())
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message")
	}

	// 停止发送器和接收器
	sender.Stop()
	receiver.Stop()

	// 优雅关闭 Echo 服务器
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		t.Logf("Server shutdown error: %v", err)
	}

	t.Log("Test completed successfully")
}

// 更好的测试实现 - 使用随机端口避免冲突
func TestQueueWithRandomPort(t *testing.T) {
	// 创建 Echo 实例
	e := echo.New()
	e.HideBanner = true // 隐藏启动横幅

	// 创建 WebSocket 升级器
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// 添加 WebSocket 路由
	e.GET("/ws", func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		defer conn.Close()

		// 简单的回声服务器
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			// 回复收到的消息
			err = conn.WriteMessage(messageType, message)
			if err != nil {
				break
			}
		}
		return nil
	})

	// 使用 0 端口让系统分配可用端口
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// 启动服务器
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- e.Start(fmt.Sprintf(":%d", port))
	}()

	// 等待服务器启动
	time.Sleep(200 * time.Millisecond)

	// 创建客户端连接
	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("localhost:%d", port), Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	// 其余测试代码保持不变...
	logger := logs.GetLogger("message_test")
	codecInstance := codec.NewJsonCodec()
	processor := protocol.NewDefaultMessageHandler()

	sender := NewMessageSender(conn, codecInstance, logger)
	receiver := NewMessageReceiver(conn, codecInstance, processor, logger)

	receiver.SetSender(sender)

	sender.Start()
	receiver.Start()

	receivedMessages := make(chan protocol.IMessage, 10)

	go func() {
		for msg := range receiver.GetMessageChan() {
			logger.Info("Received message",
				logs.String("main", msg.GetMain()),
				logs.String("sub", msg.GetSub()),
				logs.String("payload", msg.GetPayload()))
			receivedMessages <- msg
		}
	}()

	testMsg := protocol.NewJsonMessage("test", "hello", map[string]string{
		"message": "Hello World",
		"id":      "test123",
	})

	err = sender.SendMessage(testMsg)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	select {
	case receivedMsg := <-receivedMessages:
		if receivedMsg.GetMain() != "test" {
			t.Errorf("Expected main='test', got main='%s'", receivedMsg.GetMain())
		}
		if receivedMsg.GetSub() != "hello" {
			t.Errorf("Expected sub='hello', got sub='%s'", receivedMsg.GetSub())
		}
		t.Logf("Successfully received message: main=%s, sub=%s",
			receivedMsg.GetMain(), receivedMsg.GetSub())
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message")
	}

	sender.Stop()
	receiver.Stop()

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		t.Logf("Server shutdown error: %v", err)
	}

	t.Log("Test completed successfully")
}
