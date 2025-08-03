package server

import (
	"github.com/stardustagi/TopLib/libs/logs"
	"go.uber.org/zap"
)

// 客户端管理器
type ClientManager struct {
	clients    map[string]IClient
	userIdMap  map[string]string // 用户ID到客户端的映射
	register   chan IClient
	unregister chan IClient
	broadcast  chan []byte
	stopChan   chan struct{}
	startChan  chan struct{}
	logger     *zap.Logger
}

var Manager IClientManager

type IClientManager interface {
	Init()
	Start()
	ClientCount() int
	ClientKeepLive()
	KickClientByUserId(userId string)
	KickClientBySessionId(sessionId string)
	RegisterClient(client IClient)
	UnregisterClient(client IClient)
	BroadcastMessage(message []byte)
	Stop()
}

func NewClientManager(logger *zap.Logger) IClientManager {
	return &ClientManager{
		clients:    make(map[string]IClient),
		register:   make(chan IClient),
		unregister: make(chan IClient),
		broadcast:  make(chan []byte),
		stopChan:   make(chan struct{}),
		startChan:  make(chan struct{}),
		userIdMap:  make(map[string]string),
		logger:     logger,
	}
}

func (m *ClientManager) Init() {
	Manager = NewClientManager(logs.GetLogger("ClientManager"))
}

func (m *ClientManager) Start() {
	m.logger.Info("Starting client manager")
	close(m.startChan)
	for {
		select {
		//注册客户端
		case client := <-m.register:
			m.logger.Info("Registering new client", zap.String("sessionId", client.GetSessionID()), zap.String("userId", client.GetUserID()))
			m.clients[client.GetSessionID()] = client
			m.userIdMap[client.GetUserID()] = client.GetSessionID()
		// 删除客户端，并关健连接
		case client := <-m.unregister:
			m.logger.Info("Unregistering client", zap.String("sessionId", client.GetSessionID()), zap.String("userId", client.GetUserID()))
			if _, ok := m.clients[client.GetSessionID()]; ok {
				client.Close()
				delete(m.clients, client.GetSessionID())
			}
		// 所有消息广播
		case message := <-m.broadcast:
			for _, client := range m.clients {
				m.logger.Debug("Broadcasting message to client", zap.String("sessionId", client.GetSessionID()))
				client.Send(message)
			}
		case <-m.stopChan:
			m.logger.Debug("Stopping client manager")
			// 关闭所有连接
			m.Stop()
			m.logger.Info("Client manager stopped")
			return
		}
	}
}

func (m *ClientManager) RegisterClient(client IClient) {
	m.register <- client
}

func (m *ClientManager) UnregisterClient(client IClient) {
	m.unregister <- client
}

func (m *ClientManager) BroadcastMessage(message []byte) {
	m.broadcast <- message
}

func (m *ClientManager) ClientCount() int {
	return len(m.clients)
}

// ClientKeepLive 定时发送心跳包
func (m *ClientManager) ClientKeepLive() {
	for _, client := range m.clients {
		client.Send([]byte("keepalive"))
	}
}

func (m *ClientManager) KickClientByUserId(userId string) {
	if sessionId, ok := m.userIdMap[userId]; ok {
		if client, ok := m.clients[sessionId]; ok {
			client.Close()
			delete(m.clients, sessionId)
			delete(m.userIdMap, userId)
		}
	}
}

func (m *ClientManager) KickClientBySessionId(sessionId string) {
	if client, ok := m.clients[sessionId]; ok {
		client.Close()
		delete(m.clients, sessionId)
		delete(m.userIdMap, client.GetUserID())
	}
}

func (m *ClientManager) Stop() {
	for _, client := range m.clients {
		m.logger.Debug("Shutting down client", zap.String("sessionId", client.GetSessionID()), zap.String("userId", client.GetUserID()))
		client.Close()
		delete(m.clients, client.GetSessionID())
		delete(m.userIdMap, client.GetUserID())
	}
	m.logger.Info("All clients have been kicked")
	m.stopChan <- struct{}{}
}
