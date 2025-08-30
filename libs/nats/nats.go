package nats

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stardustagi/TopLib/libs/logs"
	"go.uber.org/zap"
)

type Subscription = nats.Subscription
type Msg = nats.Msg

type NatsConnection struct {
	conn       *nats.Conn
	config     *NatsConfig
	name       string
	js         nats.JetStreamContext
	streamInfo *nats.StreamInfo
	subject    []*nats.Subscription
	useStream  bool
	logger     *zap.Logger
	url        string
	stopChan   chan struct{}
	ctx        context.Context
	cannel     context.CancelFunc
	// 在结构体中添加消息通道
	messageChan chan *nats.Msg
	// handler 映射表
	handlers map[string]func(*nats.Msg)
}

// NatsConfig NATS配置结构体
type NatsConfig struct {
	Name       string   `json:"name"`
	Url        string   `json:"url"`
	UseStream  bool     `json:"use_stream"`
	StreamName string   `json:"stream_name"`
	Type       string   `json:"type"`
	Subject    []string `json:"subjects"`
	Username   string   `json:"username"` // 新增用户名
	Password   string   `json:"password"` // 新增密码
}

func NewNatsConnect(name string, natsConfig *NatsConfig) (*NatsConnection, error) {
	if natsConfig == nil {
		panic("NATS 配置未初始化")
	}
	var opts []nats.Option
	opts = append(opts,
		nats.MaxReconnects(10),
		nats.ReconnectWait(5*time.Second),
	)
	if natsConfig.Username != "" && natsConfig.Password != "" {
		opts = append(opts, nats.UserInfo(natsConfig.Username, natsConfig.Password))
	}
	conn, err := nats.Connect(natsConfig.Url, opts...)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	sub := &NatsConnection{
		name:      name,
		conn:      conn,
		config:    natsConfig,
		useStream: natsConfig.UseStream,
		logger:    logs.GetLogger("nats"),
		url:       natsConfig.Url,
		stopChan:  make(chan struct{}),
		ctx:       ctx,
		cannel:    cancel,
		// 初始化消息通道
		messageChan: make(chan *nats.Msg, 64),
		handlers:    make(map[string]func(*nats.Msg)),
	}
	if natsConfig.UseStream {
		js, err := conn.JetStream(nats.PublishAsyncMaxPending(100))
		if err != nil {
			panic(err)
			return nil, err
		}
		sub.js = js
	}
	return sub, nil
}

func (s *NatsConnection) IsConnected() bool {
	return s.conn.IsConnected()
}

func (s *NatsConnection) GetConfig() *NatsConfig {
	return s.config
}

func (s *NatsConnection) EnsureStream() error {
	if !s.useStream {
		return nil // 如果未启用 JetStream，则无需创建 Stream
	}

	// 检查 Stream 是否存在
	stream, err := s.js.StreamInfo(s.config.StreamName)
	if err != nil && stream == nil {
		s.logger.Info("Stream not found, creating a new one", zap.String("stream", s.config.StreamName))
		// 创建 Stream
		_, err = s.js.AddStream(&nats.StreamConfig{
			Name:      s.config.StreamName,
			Subjects:  s.config.Subject,
			Retention: nats.WorkQueuePolicy, // 确保每条消息只能被消费一次
		})
		if err != nil {
			s.logger.Error("Failed to create stream", zap.Error(err))
			return err
		}
		s.logger.Info("Stream created successfully", zap.String("stream", s.config.StreamName))
	}
	return nil
}
