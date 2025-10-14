package services

import (
	"context"

	"github.com/stardustagi/TopLib/libs/databases"
	"github.com/stardustagi/TopLib/libs/redis"
	"go.uber.org/zap"
)

// ServiceType 定义
const (
	ServiceTypeHTTP  = "http"
	ServiceTypeQueue = "queue"
)

// Service 接口
// 统一所有服务的行为
// Init 可选参数，具体服务可自定义
// Start/Stop/IsRunning 通用
// 也可扩展更多方法

type Service interface {
	Init(config ...interface{})
	Start()
	Stop()
	IsRunning() bool
}

type BaseService struct {
	logger *zap.Logger
	rds    redis.RedisCli
	dao    databases.BaseDao
	ctx    context.Context
	cancel context.CancelFunc
	isRun  bool
}

func (bs *BaseService) Stop() {}

func (bs *BaseService) Start() {}

func (bs *BaseService) IsRunning() bool { return bs.isRun }

// HttpService 实现

type HttpService struct {
	BaseService
	config string // 示例配置
}

func (h *HttpService) Init(config ...interface{}) {
	if len(config) > 0 {
		if v, ok := config[0].(string); ok {
			h.config = v
		}
	}
}

func (h *HttpService) Start() {
	h.isRun = true
}

func (h *HttpService) Stop() {
	h.isRun = false
}

// QueueService 实现

type QueueService struct {
	BaseService
	queueName string
}

func (q *QueueService) Init(config ...interface{}) {
	if len(config) > 0 {
		if v, ok := config[0].(string); ok {
			q.queueName = v
		}
	}
}

func (q *QueueService) Start() {
	q.isRun = true
}

func (q *QueueService) Stop() {
	q.isRun = false
}

// ServiceFactory 工厂方法
func ServiceFactory(serviceType string) Service {
	switch serviceType {
	case ServiceTypeHTTP:
		return &HttpService{}
	case ServiceTypeQueue:
		return &QueueService{}
	default:
		return nil
	}
}
