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

// Init 初始化基础服务
func (bs *BaseService) Init(config ...interface{}) {
	bs.isRun = false
}

// Start 启动服务
func (bs *BaseService) Start() {
	if bs.isRun {
		return
	}
	bs.ctx, bs.cancel = context.WithCancel(context.Background())
	bs.isRun = true
	if bs.logger != nil {
		bs.logger.Info("BaseService started")
	}
}

// Stop 优雅关闭服务
func (bs *BaseService) Stop() {
	if !bs.isRun {
		return
	}
	if bs.cancel != nil {
		bs.cancel()
	}
	bs.isRun = false
	if bs.logger != nil {
		bs.logger.Info("BaseService stopped")
	}
}

func (bs *BaseService) IsRunning() bool { return bs.isRun }

// HttpService 实现

type HttpService struct {
	BaseService
	config string // 示例配置
}

func (h *HttpService) Init(config ...interface{}) {
	h.BaseService.Init(config...)
	if len(config) > 0 {
		if v, ok := config[0].(string); ok {
			h.config = v
		}
	}
}

// QueueService 实现

type QueueService struct {
	BaseService
	queueName string
}

func (q *QueueService) Init(config ...interface{}) {
	q.BaseService.Init(config...)
	if len(config) > 0 {
		if v, ok := config[0].(string); ok {
			q.queueName = v
		}
	}
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
