package server

import (
	"context"
	"fmt"
	"os"

	"github.com/stardustagi/TopLib/libs/jwt"
	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/libs/redis"
	"github.com/stardustagi/TopLib/utils"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Access 用于判断是否具有访问权限
func Access() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			//在这里处理拦截请求的逻辑
			jwtstr := c.Request().Header.Get("jwt")

			if jwtstr != "" {
				appName := os.Getenv("APP_NAME")
				appVersion := os.Getenv("APP_VERSION")
				secret := fmt.Sprintf("%s-%s", appName, appVersion)
				jwtobj, ok := jwt.JWTDecrypt(jwtstr, secret)
				if !ok || jwtobj == nil || jwtobj["token"] == nil || jwtobj["id"] == nil {
					return c.JSON(401, map[string]interface{}{
						"errcode": 2,
						"errmsg":  "jwt解析错误",
					})
				}

				id, ok := jwtobj["id"].(string)
				if !ok {
					return c.JSON(401, map[string]interface{}{
						"errcode": 2,
						"errmsg":  "用户信息获取失败",
					})
				}
				if !(os.Getenv("DEBUG") == "true") {
					redisCmd := redis.GetRedisDb()
					key := jwt.UserTokenKey(id)
					logs.Info("jwt token :", zap.String("key", key))
					oldToken, err1 := redisCmd.Get(context.Background(), key).Result()
					if err1 != nil {
						return c.JSON(401, map[string]interface{}{
							"errcode": 2,
							"errmsg":  "获取token失败",
						})
					}
					if jwtobj["token"] != oldToken {
						return c.JSON(401, map[string]interface{}{
							"errcode": 2,
							"errmsg":  "账户已经在其他终端登录",
						})
					}
				}
				c.QueryParams().Add("id", id)
			} else {
				ip := echo.ExtractIPFromXFFHeader()(c.Request())
				// 判断是否内网IP
				ok := utils.IsInnerIp(ip)
				if !ok {
					return c.JSON(401, map[string]interface{}{
						"errcode": 2,
						"errmsg":  "服务器繁忙,请稍后再试!",
					})
				}
			}
			return next(c)
		}
	}
}

func SetHeaders(headers map[string]string, next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		for key, value := range headers {
			c.Response().Header().Set(key, value)
		}
		return next(c)
	}
}
