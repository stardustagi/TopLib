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
		panic("app name is missing in the global configuration")
	}
	appVersion, ok := globalInfo["app_version"].(string)
	if !ok {
		panic("app version is missing in the global configuration")
	}
	redisKeyPrefix, ok := globalInfo["redis_key_prefix"].(string)
	if !ok {
		panic("redis key prefix is missing in the global configuration")
	}

	os.Setenv("APP_NAME", appName)
	os.Setenv("APP_VERSION", appVersion)
	os.Setenv("REDIS_KEY_PREFIX", redisKeyPrefix)
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
