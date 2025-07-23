package conf

import (
	"encoding/json"
	"os"

	"github.com/BurntSushi/toml"
)

var (
	config  map[string]interface{}
	ISDEBUG = true
)

func Init() {
	var configPath string
	configPath = os.Getenv("runConfig")
	config = make(map[string]interface{})
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		panic(err)
	}
	globalInfo, ok := config["global"].(map[string]interface{})
	if !ok {
		panic("global configuration is missing in the config file")
	}
	// 设置配置的全局变量
	appName, ok := globalInfo["app_name"].(string)
	if !ok {
		appName = "logistics"
	}
	appVersion, ok := globalInfo["app_version"].(string)
	if !ok {
		appVersion = "1.0.0"
	}
	redisKeyPrefix, ok := globalInfo["redis_key_prefix"].(string)
	if !ok {
		redisKeyPrefix = "logistics"
	}

	os.Setenv("APP_NAME", appName)
	os.Setenv("APP_VERSION", appVersion)
	os.Setenv("REDIS_KEY_PREFIX", redisKeyPrefix)

	// 设置webservice 运行参数
	webSrv := config["websrv"].(map[string]interface{})
	port, ok := webSrv["port"].(string)
	if !ok {
		port = "10001"
	}
	host, ok := webSrv["host"].(string)
	if !ok {
		host = "127.0.0.1"
	}
	os.Setenv("WEB_PORT", port)
	os.Setenv("WEB_HOST", host)
}

func Get(key string) []byte {
	if config == nil {
		Init()
	}
	if value, exists := config[key]; exists {
		bytes, err := json.Marshal(value)
		if err != nil {
			return nil
		}
		return bytes
	}
	return nil
}
