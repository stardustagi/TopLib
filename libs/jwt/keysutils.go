package jwt

import (
	"fmt"
	"os"

	"github.com/stardustagi/TopLib/libs/logs"

	"go.uber.org/zap"
)

var (
	appName    = os.Getenv("APP_NAME")
	appVersion = os.Getenv("APP_VERSION")
)

func AdminTokenKey(id string) string {
	logs.Infof("%s:admin:token:%s", appName, id)
	return fmt.Sprintf("%s:admin:token:%s", appName, id)
}

func UserTokenKey(id string) string {
	logs.Info("user token", zap.String("id", id))
	return fmt.Sprintf("%s:user:token:%s", appName, id)
}

func NodeUserTokenKey(id string) string {
	logs.Info("node token", zap.String("id", id))
	return fmt.Sprintf("%s:node:token:%s", appName, id)
}

func NodeAccesslKey(nodeId string) string {
	return fmt.Sprintf("%s:node:access:%s", appName, nodeId)
}

func NodeModeKey(nodeId string) string {
	return fmt.Sprintf("%s:node:mode:%s", appName, nodeId)
}
