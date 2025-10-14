package nats

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stretchr/testify/assert"
)

func TestSubscription(t *testing.T) {
	// 首先测试非 JetStream 模式
	t.Run("Non-JetStream", func(t *testing.T) {
		natsConfig := &NatsConfig{
			Name:      "test",
			Url:       nats.DefaultURL,
			UseStream: false, // 使用非 JetStream 模式
		}
		s, err := NewNatsConnect("test", natsConfig)
		assert.NoError(t, err)

		subject := "test.subject"
		durableName := "test-durable"
		messageReceived := make(chan string, 1)

		// 订阅
		err = s.StartSubscription(subject, durableName, func(msg *nats.Msg) {
			messageReceived <- string(msg.Data)
		})
		assert.NoError(t, err, "Failed to subscribe to subject")

		// 等待订阅就绪
		time.Sleep(100 * time.Millisecond)

		// 发布消息
		message := "Hello, NATS!"
		err = s.Publish(subject, []byte(message))
		assert.NoError(t, err, "Failed to publish message")

		// 等待接收消息
		select {
		case received := <-messageReceived:
			assert.Equal(t, message, received)
			t.Logf("Successfully received message: %s", received)
		case <-time.After(2 * time.Second):
			t.Fatal("Did not receive message in time")
		}

		// 测试停止订阅
		err = s.StopSubscription(subject)
		assert.NoError(t, err, "Failed to stop subscription")
	})
}

func TestJetStreamOperations(t *testing.T) {
	// 测试 JetStream 的完整功能
	natsConfig := &NatsConfig{
		Name:       "jetstream-test",
		Url:        nats.DefaultURL,
		Subject:    []string{"BILLING.*", "ORDER.*"},
		UseStream:  true,
		StreamName: "TEST_BILLING_STREAM",
	}

	s, err := NewNatsConnect("jetstream-test", natsConfig)
	if err != nil {
		t.Skipf("Cannot connect to NATS: %v", err)
		return
	}
	// 启动统一消息拉取与消费
	go s.Start()
	defer s.Stop()

	// 首先验证 JetStream 是否可用
	_, err = s.GetJetStream().AccountInfo()
	if err != nil {
		t.Skipf("JetStream not available on NATS server: %v", err)
		return
	}

	t.Run("AddStream", func(t *testing.T) {
		// 测试添加 Stream
		streamName := "TEST_BILLING_STREAM"
		// subjects 与初始化配置保持一致
		subjects := []string{"BILLING.*", "ORDER.*"}

		err := s.AddStream(streamName, subjects)
		if err != nil {
			t.Skipf("JetStream not available: %v", err)
			return
		}

		// 验证 Stream 是否创建成功
		streamInfo, err := s.GetJetStream().StreamInfo(streamName)
		assert.NoError(t, err, "Failed to get stream info")
		assert.Equal(t, streamName, streamInfo.Config.Name)
		assert.ElementsMatch(t, subjects, streamInfo.Config.Subjects)
		t.Logf("Stream created successfully: %s with subjects: %v", streamName, subjects)
	})

	t.Run("AddConsumer", func(t *testing.T) {
		// 确保 Stream 存在
		streamName := "TEST_BILLING_STREAM"
		subjects := []string{"BILLING.created", "BILLING.updated", "BILLING.deleted"}
		err := s.AddStream(streamName, subjects)
		if err != nil {
			t.Skipf("JetStream not available: %v", err)
			return
		}

		// 测试添加 Consumer
		durableName := "billing-processor"
		subject := "BILLING.created"

		err = s.AddConsumer(streamName, durableName, subject)
		if err != nil {
			t.Skipf("Consumer creation not supported on this NATS server: %v", err)
			return
		}

		// 等待Consumer创建完成
		time.Sleep(100 * time.Millisecond)

		// 验证 Consumer 是否创建成功
		consumerInfo, err := s.GetJetStream().ConsumerInfo(streamName, durableName)
		if err != nil {
			t.Skipf("Cannot verify consumer creation: %v", err)
			return
		}

		assert.Equal(t, durableName, consumerInfo.Config.Durable)
		assert.Equal(t, subject, consumerInfo.Config.FilterSubject)
		t.Logf("Consumer created successfully: %s for subject: %s", durableName, subject)
	})

	t.Run("SimpleJetStreamTest", func(t *testing.T) {
		// 简化的 JetStream 测试，不依赖预先创建的 Consumer
		streamName := "SIMPLE_TEST_STREAM"
		subjects := []string{"SIMPLE.test"}

		// 创建 Stream
		err := s.AddStream(streamName, subjects)
		if err != nil {
			t.Skipf("Cannot create stream: %v", err)
			return
		}
		subject := "SIMPLE.test"
		durableName := "simple-consumer"
		err = s.AddConsumer(streamName, durableName, "SIMPLE.test")
		if err != nil {
			t.Skipf("Cannot create consumer: %v", err)
			return
		}

		messageReceived := make(chan string, 1)
		err = s.StartSubscription(subject, durableName, func(msg *nats.Msg) {
			messageReceived <- string(msg.Data)
		})
		assert.NoError(t, err, "Failed to start JetStream subscription")

		// 发布消息
		message := "Simple JetStream test message"
		err = s.Publish(subject, []byte(message))
		assert.NoError(t, err, "Failed to publish message")

		// 拉取消息
		time.Sleep(200 * time.Millisecond)
		select {
		case received := <-messageReceived:
			assert.Equal(t, message, received)
			t.Logf("Successfully received JetStream message: %s", received)
		case <-time.After(2 * time.Second):
			t.Fatal("Did not receive JetStream message in time")
		}

		// 测试停止订阅
		err = s.StopSubscription(subject)
		assert.NoError(t, err, "Failed to stop subscription")
	})

	t.Run("AsyncPublish", func(t *testing.T) {
		// 测试异步发布功能
		streamName := "ASYNC_TEST_STREAM"
		subjects := []string{"ASYNC.test"}
		err := s.AddStream(streamName, subjects)
		if err != nil {
			t.Skipf("Cannot create stream for async test: %v", err)
			return
		}

		subject := "ASYNC.test"
		message := "Async test message"

		// 测试异步发布
		err = s.PublishAsync(subject, []byte(message))
		assert.NoError(t, err, "Failed to publish async message")
		t.Logf("Successfully published async message: %s", message)

		// 等待异步发布完成
		time.Sleep(100 * time.Millisecond)
	})
}

func TestNatsConnectionManager(t *testing.T) {
	// 测试连接管理器功能
	config1 := &NatsConfig{
		Name:      "test-conn-1",
		Url:       nats.DefaultURL,
		UseStream: false,
	}

	config2 := &NatsConfig{
		Name:       "test-conn-2",
		Url:        nats.DefaultURL,
		UseStream:  true,
		StreamName: "MANAGER_TEST_STREAM",
		Subject:    []string{"MANAGER.*"},
	}

	manager := &NatsConnManager{
		clients: make(map[string]*NatsConnection),
	}

	// 测试添加连接
	err := manager.addClient("conn1", config1)
	assert.NoError(t, err, "Failed to add first connection")

	err = manager.addClient("conn2", config2)
	assert.NoError(t, err, "Failed to add second connection")

	// 测试获取连接
	conn1, exists := manager.GetClient("conn1")
	assert.True(t, exists, "Connection 1 should exist")
	assert.NotNil(t, conn1, "Connection 1 should not be nil")

	conn2, exists := manager.GetClient("conn2")
	assert.True(t, exists, "Connection 2 should exist")
	assert.NotNil(t, conn2, "Connection 2 should not be nil")

	// 测试连接状态
	assert.True(t, conn1.IsConnected(), "Connection 1 should be connected")
	assert.True(t, conn2.IsConnected(), "Connection 2 should be connected")

	// 测试关闭所有连接
	err = manager.CloseAll()
	assert.NoError(t, err, "Failed to close all connections")

	t.Log("Connection manager test completed successfully")
}

func TestStartStopSubscription(t *testing.T) {
	// 专门测试 StartSubscription 和 StopSubscription 方法
	natsConfig := &NatsConfig{
		Name:      "start-stop-test",
		Url:       nats.DefaultURL,
		UseStream: false,
	}

	s, err := NewNatsConnect("start-stop-test", natsConfig)
	assert.NoError(t, err)
	go s.Start()
	defer s.Stop()

	subject := "start.stop.test"
	durableName := "start-stop-durable"
	messageReceived := make(chan string, 2)

	// 测试 StartSubscription
	err = s.StartSubscription(subject, durableName, func(msg *nats.Msg) {
		messageReceived <- string(msg.Data)
	})
	assert.NoError(t, err, "Failed to start subscription")

	// 等待订阅就绪
	time.Sleep(100 * time.Millisecond)

	// 发布第一条消息
	message1 := "First message"
	err = s.Publish(subject, []byte(message1))
	assert.NoError(t, err, "Failed to publish first message")

	// 验证第一条消息被接收
	select {
	case received := <-messageReceived:
		assert.Equal(t, message1, received)
		t.Logf("Received first message: %s", received)
	case <-time.After(2 * time.Second):
		t.Fatal("Did not receive first message in time")
	}

	// 测试 StopSubscription
	err = s.StopSubscription(subject)
	assert.NoError(t, err, "Failed to stop subscription")

	// 发布第二条消息（应该不会被接收）
	message2 := "Second message after stop"
	err = s.Publish(subject, []byte(message2))
	assert.NoError(t, err, "Failed to publish second message")

	// 验证第二条消息不会被接收
	select {
	case received := <-messageReceived:
		t.Fatalf("Should not receive message after stop, but got: %s", received)
	case <-time.After(1 * time.Second):
		t.Log("Correctly did not receive message after stop")
	}

	t.Log("StartSubscription and StopSubscription test completed successfully")
}

func TestJetStreamHandlerDispatch(t *testing.T) {
	natsConfig := &NatsConfig{
		Name:       "handler-dispatch-test",
		Url:        nats.DefaultURL,
		Subject:    []string{"DISPATCH.A", "DISPATCH.B"},
		UseStream:  true,
		StreamName: "DISPATCH_STREAM",
	}

	s, err := NewNatsConnect("handler-dispatch-test", natsConfig)
	assert.NoError(t, err)
	go s.Start()
	defer s.Stop()

	err = s.StartSubscription("DISPATCH.A", "dispatch-a", func(msg *nats.Msg) {
		fmt.Printf("hello one:%s\n", string(msg.Data))
	})
	defer s.StopSubscription("DISPATCH.A")
	assert.NoError(t, err)
	err = s.StartSubscription("DISPATCH.B", "dispatch-b", func(msg *nats.Msg) {
		fmt.Printf("hello two:%s\n", string(msg.Data))
	})
	defer s.StopSubscription("DISPATCH.B")
	assert.NoError(t, err)

	// 发布消息
	err = s.Publish("DISPATCH.A", []byte("MessageA"))
	assert.NoError(t, err)
	err = s.Publish("DISPATCH.B", []byte("MessageB"))
	assert.NoError(t, err)
	err = s.Publish("DISPATCH.A", []byte("MessageA1"))
	assert.NoError(t, err)
	err = s.Publish("DISPATCH.B", []byte("MessageB1"))
	assert.NoError(t, err)
	// 创建信号通道，监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	t.Log("Service started successfully, waiting for shutdown signal...")

	// 阻塞等待信号
	sig := <-sigChan
	t.Log("Received shutdown signal", logs.String("signal", sig.String()))

	t.Log("Service is shutting down...")

}
