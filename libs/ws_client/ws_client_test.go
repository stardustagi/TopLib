package wsclient

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"testing"

	"github.com/stardustagi/TopLib/codec"
	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/protocol"
)

func TestClient(t *testing.T) {
	// 配置服务器地址和头部信息
	serverURL := "ws://localhost:8080/api/test/ws"
	headers := map[string]string{
		"userId": "testUser123", // 服务器要求的 userId
	}
	logger := logs.GetLogger("wsclient")
	newCodec := codec.NewJsonCodec()
	newHandler := protocol.NewDefaultMessageHandler()
	// 创建客户端
	client, err := NewWSClient(serverURL, headers, newCodec, newHandler, logger)
	if err != nil {
		logger.Info("create client error:", logs.ErrorInfo(err))
	}
	defer client.Close()

	// 启动客户端
	client.Start()
	logger.Info("WebSocket client connected to", logs.String("serverUrl", serverURL))

	// 处理中断信号
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// 启动输入处理
	go func() {
		// scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Enter messages (type 'quit' to exit):")

		// 发送消息
		err := client.SendMessage("bill", "echo", "{\"message\": \""+"Hello, WebSocket!"+"\"}")
		if err != nil {
			log.Printf("send message error: %v", err)
		}
	}()

	// 等待中断信号
	select {
	case <-interrupt:
		log.Println("interrupt received, shutting down...")
	case <-client.done:
		log.Println("connection closed")
	}

	// 优雅关闭
	log.Println("client shutting down...")
}
