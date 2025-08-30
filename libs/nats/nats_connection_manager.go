package nats

import (
	"sync"

	"github.com/goccy/go-json"
)

type NatsConnManager struct {
	mu      sync.RWMutex
	clients map[string]*NatsConnection // key 可以是 URL 或自定义名字
}

// 内部单例实例
var (
	manager       *NatsConnManager
	managerOnce   sync.Once
	managerConfig []*NatsConfig
)

func Init(config []byte) {
	err := json.Unmarshal(config, &managerConfig)
	if err != nil {
		panic(err)
	}
}

// GetNatsManager 返回全局唯一的连接管理器实例
func GetNatsManager() *NatsConnManager {
	managerOnce.Do(func() {
		manager = &NatsConnManager{
			clients: make(map[string]*NatsConnection),
		}
		for _, cfg := range managerConfig {
			err := manager.addClient(cfg.Name, cfg)
			if err != nil {
				panic(err)
			}
		}
	})
	return manager
}

// AddClient 在管理器里添加一条新连接
// key 由调用方决定，常见的做法是使用 NATS URL 或业务层的别名
func (m *NatsConnManager) addClient(key string, config *NatsConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clients[key]; exists {
		return nil // 已存在，直接返回
	}
	nc, err := NewNatsConnect(key, config)
	if err != nil {
		return err
	}
	m.clients[key] = nc
	return nil
}

// GetClient 根据 key 获取连接
func (m *NatsConnManager) GetClient(key string) (*NatsConnection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.clients[key]
	return c, ok
}

// CloseAll 关闭所有活跃连接，并清空 map
func (m *NatsConnManager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, c := range m.clients {
		c.Stop()
		delete(m.clients, k)
	}
	return nil
}

func (m *NatsConnManager) StartAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, c := range m.clients {
		c.Start()
	}
}
