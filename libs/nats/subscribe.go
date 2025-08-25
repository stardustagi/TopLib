package nats

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stardustagi/TopLib/libs/logs"
	"go.uber.org/zap"
)

type NatsConnection struct {
	conn       *nats.Conn
	name       string
	js         nats.JetStreamContext
	streamInfo *nats.StreamInfo
	useStream  bool
	logger     *zap.Logger
	url        string
}

func NewNatsConnect(name, url string, useStream bool) (*NatsConnection, error) {
	conn, err := nats.Connect(url,
		nats.MaxReconnects(10), // 最大重连次数
		//nats.DontRandomize(),  // 不随机连接
		nats.ReconnectWait(5*time.Second), // 重连等待时间
	)
	if err != nil {
		panic(err)
	}
	sub := &NatsConnection{
		name:      name,
		conn:      conn,
		useStream: useStream,
		logger:    logs.GetLogger("nats"),
		url:       url,
	}
	if useStream {
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

func (s *NatsConnection) Subscribe(subject, durableName string, handler func(msg *nats.Msg)) (_subject *nats.Subscription, err error) {
	if s.useStream {
		_sub, err := s.js.QueueSubscribe(subject, durableName, handler, nats.Durable(durableName))
		if err != nil {
			s.logger.Error("Error subscribing to subject", logs.ErrorInfo(err))
			return nil, err
		}
		s.logger.Info("Subscribed to subject", logs.String("subject", subject), logs.String("durable", durableName))
		return _sub, nil
	} else {
		if _subject, err = s.conn.QueueSubscribe(subject, durableName, handler); err != nil {
			s.logger.Error("Error subscribing to subject", logs.ErrorInfo(err))
			return nil, err
		}
	}
	s.logger.Info("Subscribed to subject", logs.String("subject", subject))
	return
}

func (s *NatsConnection) Close() {
	if s.conn != nil {
		s.conn.Close()
		s.logger.Info("NATS connection closed", logs.String("name", s.name), logs.String("url", s.url))
	}
}

func (s *NatsConnection) AddStream(streamName string, subjects []string) error {
	if !s.useStream {
		return nats.ErrNoStreamResponse
	}
	stream, err := s.js.StreamInfo(streamName)
	if err != nil && stream != nil {
		return err
	}
	if stream == nil {
		s.streamInfo, err = s.js.AddStream(&nats.StreamConfig{
			Name:      streamName,
			Subjects:  subjects,
			Retention: nats.WorkQueuePolicy, // 使用工作队列策略，确保每条消息只能被消费一次
		})
		if err != nil {
			s.logger.Error("Failed to add stream",
				zap.String("stream", streamName),
				zap.Strings("subjects", subjects),
				zap.Error(err))
			return err
		}
		s.logger.Info("Stream added",
			zap.String("stream", streamName),
			zap.Strings("subjects", subjects))
	}
	return nil
}

func (s *NatsConnection) GetJetStream() nats.JetStreamContext {
	return s.js
}

func (s *NatsConnection) GetNativeConn() *nats.Conn {
	return s.conn
}
