package services

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type HttpService struct {
	BaseService
	config []byte
}

var (
	instance     interface{}
	instanceOnce sync.Once
)

func InitConfig(config []byte) {
	// 这里可以添加配置初始化逻辑
}

// GetHttpServiceInstance 获取单例
func GetHttpServiceInstance() *HttpService {
	instanceOnce.Do(func() {
		instance = &HttpService{}
	})
	return instance.(*HttpService)
}

// Init 初始化
func (s *HttpService) Init(config []byte) {
	s.config = config
}

func (s *HttpService) Start() {
	s.isRun = true
}

func (s *HttpService) Stop() {
	s.isRun = false
}

func TestBaseService_IsRunning(t *testing.T) {
	ins := GetHttpServiceInstance()
	ins.Start()
	assert.True(t, ins.IsRunning())
	ins.Stop()
	assert.False(t, ins.IsRunning())
}

func TestServiceFactory_HttpService(t *testing.T) {
	httpSrv := ServiceFactory(ServiceTypeHTTP)
	assert.NotNil(t, httpSrv)
	httpSrv.Init("http-config")
	httpSrv.Start()
	assert.True(t, httpSrv.IsRunning())
	httpSrv.Stop()
	assert.False(t, httpSrv.IsRunning())
}

func TestServiceFactory_QueueService(t *testing.T) {
	queueSrv := ServiceFactory(ServiceTypeQueue)
	assert.NotNil(t, queueSrv)
	queueSrv.Init("queue-name")
	queueSrv.Start()
	assert.True(t, queueSrv.IsRunning())
	queueSrv.Stop()
	assert.False(t, queueSrv.IsRunning())
}

func TestServiceFactory_UnknownType(t *testing.T) {
	unknownSrv := ServiceFactory("unknown")
	assert.Nil(t, unknownSrv)
}
