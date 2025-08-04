package server

type HttpServerConfig struct {
	Port       int    `json:"port" yaml:"port"`               // HTTP服务器端口
	Address    string `json:"address" yaml:"address"`         // HTTP服务器主机名
	Path       string `json:"path" yaml:"path"`               // HTTP服务器路径
	Cors       bool   `json:"cors" yaml:"cors"`               // 是否启用CORS
	RequestLog bool   `json:"request_log" yaml:"request_log"` // 是否启用请求日志
	Access     bool   `json:"access" yaml:"access"`           // 是否启用访问日志
}

type HttpWebSocketConfig struct {
	ReadBufferSize    int `json:"read_buffer_size" yaml:"read_buffer_size"`   // WebSocket读取缓冲区大小
	WriteBufferSize   int `json:"write_buffer_size" yaml:"write_buffer_size"` // WebSocket写入缓冲区大小
	TickerTime        int `json:"ticker_time" yaml:"ticker_time"`             // WebSocket心跳间隔时间
	WriteDeadlineTime int `json:"deadline_time" yaml:"deadline_time"`         // WebSocket连接超时时间
	ReadLimit         int `json:"read_limit" yaml:"read_limit"`               // WebSocket读取限制
	HandshakeTimeout  int `json:"handshake_timeout" yaml:"handshake_timeout"` // Web
}
