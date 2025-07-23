package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestRedisTls(t *testing.T) {
	//tlsConfig := &tls.Config{
	//	InsecureSkipVerify: false, // Set to true if you want to skip certificate verification (not recommended for production)
	//}

	redisConfig := &Config{
		Addrs: []string{"localhost:6379"},
		//Password:  "yourpassword",
		DbIndex: 0,
		//TLSConfig: tlsConfig,
	}

	// Marshal redisConfig to JSON if Init expects []byte (adjust as needed)
	configBytes, err := json.Marshal(redisConfig)
	if err != nil {
		panic(err)
	}
	redisCmd, err := Init(configBytes)
	if err != nil {
		panic(err)
	}
	_, err = redisCmd.Set(context.TODO(), "key", "hello", time.Minute*60).Result()
	if err != nil {
		panic(err)
	}
}
