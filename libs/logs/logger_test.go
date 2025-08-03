package logs

import (
	"encoding/json"
	"testing"
)

func TestLogger(t *testing.T) {
	configMap := map[string]interface{}{
		"global": map[string]interface{}{
			"appName":    "testApp",
			"appVersion": "1.0.0",
		}}
	conf, err := json.Marshal(configMap)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	Init(conf)

	Log.Info("This is an info message")
	Log.Warn("This is a warning message")
	Log.Error("This is an error message")
}
