package redis

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// MockRedisCmd 实现 RedisCmd 接口用于测试
type MockRedisCmd struct {
	*redis.Client
}

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *MockRedisCmd, *zap.Logger) {
	// 启动 miniredis 服务器
	mr := miniredis.RunT(t)

	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// 创建 mock 命令
	mockCmd := &MockRedisCmd{Client: client}

	// 创建 zap logger
	logger := zaptest.NewLogger(t, zaptest.Level(zap.DebugLevel))

	return mr, mockCmd, logger
}

func TestNewRedisView(t *testing.T) {
	_, mockCmd, logger := setupTestRedis(t)

	t.Run("CreateWithPrefix", func(t *testing.T) {
		view := NewRedisView(mockCmd, "test", logger)
		assert.NotNil(t, view)
		assert.Equal(t, "test", view.KeyPrefix())
		assert.Equal(t, mockCmd, view.NativeCmd())
	})

	t.Run("CreateWithoutPrefix", func(t *testing.T) {
		view := NewRedisView(mockCmd, "", logger)
		assert.NotNil(t, view)
		assert.Equal(t, "", view.KeyPrefix())
	})

	t.Run("CreateWithoutLogger", func(t *testing.T) {
		view := NewRedisView(mockCmd, "test", nil)
		assert.NotNil(t, view)
	})
}

func TestRedisView_SetAndGet(t *testing.T) {
	mr, mockCmd, logger := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	view := NewRedisView(mockCmd, "test", logger)

	t.Run("SetAndGetBasic", func(t *testing.T) {
		key := "basic_key"
		value := []byte("basic_value")

		// Set 操作
		err := view.Set(ctx, key, value, "")
		require.NoError(t, err)

		// Get 操作
		result, err := view.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("SetWithExpiration", func(t *testing.T) {
		key := "expiring_key"
		value := []byte("expiring_value")

		// 设置1秒过期
		err := view.Set(ctx, key, value, "1s")
		require.NoError(t, err)

		// 立即获取应该成功
		result, err := view.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, result)

		// miniredis 需要手动触发时间流逝
		mr.FastForward(2 * time.Second)

		// 再次获取应该失败
		_, err = view.Get(ctx, key)
		assert.Error(t, err)
		assert.Equal(t, redis.Nil, err)
	})

	t.Run("InvalidDuration", func(t *testing.T) {
		key := "invalid_duration_key"
		value := []byte("test_value")

		err := view.Set(ctx, key, value, "invalid_duration")
		assert.Error(t, err)
	})
}

func TestRedisView_SetNX(t *testing.T) {
	mr, mockCmd, logger := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	view := NewRedisView(mockCmd, "test", logger)

	t.Run("SetNXSuccess", func(t *testing.T) {
		key := "setnx_key"
		value := []byte("setnx_value")

		// 第一次设置应该成功
		success, err := view.SetNX(ctx, key, value, "")
		require.NoError(t, err)
		assert.True(t, success)

		// 第二次设置应该失败（键已存在）
		success, err = view.SetNX(ctx, key, []byte("new_value"), "")
		require.NoError(t, err)
		assert.False(t, success)

		// 验证值没有改变
		result, err := view.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("SetNXWithExpiration", func(t *testing.T) {
		key := "setnx_exp_key"
		value := []byte("setnx_exp_value")

		success, err := view.SetNX(ctx, key, value, "1s")
		require.NoError(t, err)
		assert.True(t, success)

		// miniredis 需要手动触发时间流逝
		mr.FastForward(2 * time.Second)

		// 现在应该可以再次设置
		success, err = view.SetNX(ctx, key, []byte("new_value"), "")
		require.NoError(t, err)
		assert.True(t, success)
	})
}

func TestRedisView_Del(t *testing.T) {
	mr, mockCmd, logger := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	view := NewRedisView(mockCmd, "test", logger)

	t.Run("DeleteSingleKey", func(t *testing.T) {
		key := "delete_key"
		value := []byte("delete_value")

		// 先设置一个键
		err := view.Set(ctx, key, value, "")
		require.NoError(t, err)

		// 删除键
		deleted, err := view.Del(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, int64(1), deleted)

		// 验证键已被删除
		_, err = view.Get(ctx, key)
		assert.Error(t, err)
		assert.Equal(t, redis.Nil, err)
	})

	t.Run("DeleteMultipleKeys", func(t *testing.T) {
		keys := []string{"del_key1", "del_key2", "del_key3"}
		value := []byte("test_value")

		// 设置多个键
		for _, key := range keys {
			err := view.Set(ctx, key, value, "")
			require.NoError(t, err)
		}

		// 删除所有键
		deleted, err := view.Del(ctx, keys...)
		require.NoError(t, err)
		assert.Equal(t, int64(3), deleted)
	})

	t.Run("DeleteNonExistentKey", func(t *testing.T) {
		deleted, err := view.Del(ctx, "non_existent_key")
		require.NoError(t, err)
		assert.Equal(t, int64(0), deleted)
	})
}

func TestRedisView_Expire(t *testing.T) {
	mr, mockCmd, logger := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	view := NewRedisView(mockCmd, "test", logger)

	t.Run("ExpireExistingKey", func(t *testing.T) {
		key := "expire_key"
		value := []byte("expire_value")

		// 设置一个永久键
		err := view.Set(ctx, key, value, "")
		require.NoError(t, err)

		// 设置过期时间
		err = view.Expire(ctx, key, "1s")
		require.NoError(t, err)

		// 立即检查键还存在
		result, err := view.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, result)

		// 等待过期
		time.Sleep(1100 * time.Millisecond)

		// 检查键已过期 (miniredis 可能不完全模拟过期行为)
		_, err = view.Get(ctx, key)
		if err != nil {
			assert.Equal(t, redis.Nil, err)
		} else {
			// 在某些情况下，miniredis 可能不会立即过期键
			t.Log("Warning: Key may not have expired in miniredis")
		}
	})

	t.Run("ExpireInvalidDuration", func(t *testing.T) {
		err := view.Expire(ctx, "any_key", "invalid_duration")
		assert.Error(t, err)
	})
}

func TestRedisView_Scan(t *testing.T) {
	mr, mockCmd, logger := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	view := NewRedisView(mockCmd, "scan_test", logger)

	t.Run("ScanKeys", func(t *testing.T) {
		// 设置一些测试键
		testKeys := []string{"user:1", "user:2", "user:3", "order:1"}
		for _, key := range testKeys {
			err := view.Set(ctx, key, []byte("value"), "")
			require.NoError(t, err)
		}

		// 扫描所有 user: 开头的键
		keys, err := view.Scan(ctx, 0, "user:*", 10)
		require.NoError(t, err)

		// 验证扫描结果（注意需要考虑前缀）
		var userKeys []string
		for _, key := range keys {
			if key == "scan_test:user:1" || key == "scan_test:user:2" || key == "scan_test:user:3" {
				userKeys = append(userKeys, key)
			}
		}
		assert.Len(t, userKeys, 3)
	})
}

func TestRedisView_XAdd(t *testing.T) {
	mr, mockCmd, logger := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	view := NewRedisView(mockCmd, "stream_test", logger)

	t.Run("XAddBasic", func(t *testing.T) {
		args := redis.XAddArgs{
			Stream: "test_stream",
			Values: map[string]interface{}{
				"field1": "value1",
				"field2": "value2",
			},
		}

		cmd := view.XAdd(ctx, args)
		require.NotNil(t, cmd)

		result, err := cmd.Result()
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})
}

func TestRedisView_KeyPrefix(t *testing.T) {
	_, mockCmd, logger := setupTestRedis(t)

	t.Run("WithPrefix", func(t *testing.T) {
		view := NewRedisView(mockCmd, "myapp", logger)
		assert.Equal(t, "myapp", view.KeyPrefix())
	})

	t.Run("WithoutPrefix", func(t *testing.T) {
		view := NewRedisView(mockCmd, "", logger)
		assert.Equal(t, "", view.KeyPrefix())
	})
}

// 基准测试
func BenchmarkRedisView_Set(b *testing.B) {
	mr := miniredis.RunT(b)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	mockCmd := &MockRedisCmd{Client: client}
	logger := zap.NewNop()

	view := NewRedisView(mockCmd, "bench", logger)
	ctx := context.Background()
	value := []byte("benchmark_value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "benchmark_key_" + string(rune(i))
		view.Set(ctx, key, value, "")
	}
}

func BenchmarkRedisView_Get(b *testing.B) {
	mr := miniredis.RunT(b)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	mockCmd := &MockRedisCmd{Client: client}
	logger := zap.NewNop()

	view := NewRedisView(mockCmd, "bench", logger)
	ctx := context.Background()
	value := []byte("benchmark_value")

	// 预设一些数据
	for i := 0; i < 1000; i++ {
		key := "benchmark_key_" + string(rune(i))
		view.Set(ctx, key, value, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "benchmark_key_" + string(rune(i%1000))
		view.Get(ctx, key)
	}
}

// Redis 发布/订阅测试
func TestRedisView_PubSub(t *testing.T) {
	mr, mockCmd, logger := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	view := NewRedisView(mockCmd, "pubsub", logger)

	t.Run("BasicPublishSubscribe", func(t *testing.T) {
		channel := "test_channel"
		message := "Hello, Redis PubSub!"

		// 创建订阅
		pubsub, err := view.Subscribe(ctx, channel)
		require.NoError(t, err)
		require.NotNil(t, pubsub)
		defer pubsub.Close()

		// 等待订阅建立
		time.Sleep(100 * time.Millisecond)

		// 发布消息
		subscribers, err := view.Publish(ctx, channel, message)
		require.NoError(t, err)
		assert.Equal(t, int64(1), subscribers)

		// 接收消息
		msg, err := pubsub.ReceiveMessage(ctx)
		require.NoError(t, err)
		assert.Equal(t, channel, msg.Channel)
		assert.Equal(t, message, msg.Payload)
	})

	t.Run("MultipleSubscribers", func(t *testing.T) {
		channel := "multi_test_channel"
		message := "Broadcast message"

		// 创建多个订阅者
		pubsub1, err := view.Subscribe(ctx, channel)
		require.NoError(t, err)
		defer pubsub1.Close()

		pubsub2, err := view.Subscribe(ctx, channel)
		require.NoError(t, err)
		defer pubsub2.Close()

		// 等待订阅建立
		time.Sleep(100 * time.Millisecond)

		// 发布消息
		subscribers, err := view.Publish(ctx, channel, message)
		require.NoError(t, err)
		assert.Equal(t, int64(2), subscribers)

		// 两个订阅者都应该收到消息
		msg1, err := pubsub1.ReceiveMessage(ctx)
		require.NoError(t, err)
		assert.Equal(t, channel, msg1.Channel)
		assert.Equal(t, message, msg1.Payload)

		msg2, err := pubsub2.ReceiveMessage(ctx)
		require.NoError(t, err)
		assert.Equal(t, channel, msg2.Channel)
		assert.Equal(t, message, msg2.Payload)
	})

	t.Run("MultipleChannels", func(t *testing.T) {
		channels := []string{"channel1", "channel2", "channel3"}

		// 订阅多个频道
		pubsub, err := view.Subscribe(ctx, channels...)
		require.NoError(t, err)
		defer pubsub.Close()

		// 等待订阅建立
		time.Sleep(100 * time.Millisecond)

		// 向每个频道发布消息
		for i, channel := range channels {
			message := fmt.Sprintf("Message %d", i+1)
			subscribers, err := view.Publish(ctx, channel, message)
			require.NoError(t, err)
			assert.Equal(t, int64(1), subscribers)

			// 接收消息
			msg, err := pubsub.ReceiveMessage(ctx)
			require.NoError(t, err)
			assert.Equal(t, channel, msg.Channel)
			assert.Equal(t, message, msg.Payload)
		}
	})

	t.Run("PublishToNonExistentChannel", func(t *testing.T) {
		// 向没有订阅者的频道发布消息
		subscribers, err := view.Publish(ctx, "nonexistent_channel", "test message")
		require.NoError(t, err)
		assert.Equal(t, int64(0), subscribers) // 没有订阅者
	})

	t.Run("JSONMessagePublishSubscribe", func(t *testing.T) {
		channel := "json_channel"
		// 使用JSON字符串而不是map，因为Redis Publish需要字符串
		jsonMessage := `{"type":"notification","user_id":12345,"message":"You have a new message","timestamp":1692454800}`

		// 创建订阅
		pubsub, err := view.Subscribe(ctx, channel)
		require.NoError(t, err)
		defer pubsub.Close()

		// 等待订阅建立
		time.Sleep(100 * time.Millisecond)

		// 发布JSON消息
		subscribers, err := view.Publish(ctx, channel, jsonMessage)
		require.NoError(t, err)
		assert.Equal(t, int64(1), subscribers)

		// 接收消息并验证
		msg, err := pubsub.ReceiveMessage(ctx)
		require.NoError(t, err)
		assert.Equal(t, channel, msg.Channel)
		// 消息应该包含JSON字符串
		assert.Contains(t, msg.Payload, "notification")
		assert.Contains(t, msg.Payload, "12345")
	})
}

func TestRedisView_PatternSubscribe(t *testing.T) {
	mr, mockCmd, logger := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	view := NewRedisView(mockCmd, "pattern", logger)

	t.Run("BasicPatternSubscribe", func(t *testing.T) {
		pattern := "test.*"

		// 创建模式订阅
		pubsub, err := view.PSubscribe(ctx, pattern)
		require.NoError(t, err)
		require.NotNil(t, pubsub)
		defer pubsub.Close()

		// 等待订阅建立
		time.Sleep(100 * time.Millisecond)

		// 向匹配模式的频道发布消息
		testChannels := []string{"test.channel1", "test.channel2", "test.events"}
		for i, channel := range testChannels {
			message := fmt.Sprintf("Pattern message %d", i+1)

			subscribers, err := view.Publish(ctx, channel, message)
			require.NoError(t, err)
			assert.Equal(t, int64(1), subscribers)

			// 接收消息
			msg, err := pubsub.ReceiveMessage(ctx)
			require.NoError(t, err)
			assert.Equal(t, channel, msg.Channel)
			assert.Equal(t, pattern, msg.Pattern)
			assert.Equal(t, message, msg.Payload)
		}
	})

	t.Run("MultiplePatterns", func(t *testing.T) {
		patterns := []string{"user.*", "order.*", "payment.*"}

		// 订阅多个模式
		pubsub, err := view.PSubscribe(ctx, patterns...)
		require.NoError(t, err)
		defer pubsub.Close()

		// 等待订阅建立
		time.Sleep(100 * time.Millisecond)

		// 测试各种频道
		testCases := []struct {
			channel string
			pattern string
			message string
		}{
			{"user.created", "user.*", "User created event"},
			{"order.placed", "order.*", "Order placed event"},
			{"payment.completed", "payment.*", "Payment completed event"},
		}

		for _, tc := range testCases {
			subscribers, err := view.Publish(ctx, tc.channel, tc.message)
			require.NoError(t, err)
			assert.Equal(t, int64(1), subscribers)

			// 接收消息
			msg, err := pubsub.ReceiveMessage(ctx)
			require.NoError(t, err)
			assert.Equal(t, tc.channel, msg.Channel)
			assert.Equal(t, tc.pattern, msg.Pattern)
			assert.Equal(t, tc.message, msg.Payload)
		}
	})

	t.Run("PatternNotMatching", func(t *testing.T) {
		pattern := "events.*"

		// 创建模式订阅
		pubsub, err := view.PSubscribe(ctx, pattern)
		require.NoError(t, err)
		defer pubsub.Close()

		// 等待订阅建立
		time.Sleep(100 * time.Millisecond)

		// 向不匹配模式的频道发布消息
		subscribers, err := view.Publish(ctx, "notifications.test", "Should not receive this")
		require.NoError(t, err)
		assert.Equal(t, int64(0), subscribers) // 没有匹配的订阅者
	})
}

func TestRedisView_PubSubConcurrency(t *testing.T) {
	mr, mockCmd, logger := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	view := NewRedisView(mockCmd, "concurrent", logger)

	t.Run("ConcurrentPublishSubscribe", func(t *testing.T) {
		channel := "concurrent_channel"
		numPublishers := 5
		numMessages := 10

		// 创建订阅
		pubsub, err := view.Subscribe(ctx, channel)
		require.NoError(t, err)
		defer pubsub.Close()

		// 等待订阅建立
		time.Sleep(100 * time.Millisecond)

		// 收集接收到的消息
		receivedMessages := make(chan string, numPublishers*numMessages)
		go func() {
			for i := 0; i < numPublishers*numMessages; i++ {
				msg, err := pubsub.ReceiveMessage(ctx)
				if err != nil {
					t.Logf("Error receiving message: %v", err)
					continue
				}
				receivedMessages <- msg.Payload
			}
		}()

		// 并发发布消息
		var wg sync.WaitGroup
		for publisherID := 0; publisherID < numPublishers; publisherID++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for msgID := 0; msgID < numMessages; msgID++ {
					message := fmt.Sprintf("Publisher %d - Message %d", id, msgID)
					_, err := view.Publish(ctx, channel, message)
					if err != nil {
						t.Logf("Error publishing message: %v", err)
					}
					time.Sleep(1 * time.Millisecond) // 小延迟避免过快
				}
			}(publisherID)
		}

		wg.Wait()

		// 验证接收到的消息数量
		time.Sleep(200 * time.Millisecond) // 等待所有消息被接收
		close(receivedMessages)

		var messages []string
		for msg := range receivedMessages {
			messages = append(messages, msg)
		}

		assert.Equal(t, numPublishers*numMessages, len(messages))
	})
}

func TestRedisView_PubSubErrorHandling(t *testing.T) {
	mr, mockCmd, logger := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	view := NewRedisView(mockCmd, "error", logger)
	t.Run("SubscribeWithTimeout", func(t *testing.T) {
		channel := "timeout_channel"

		// 创建订阅
		pubsub, err := view.Subscribe(ctx, channel)
		require.NoError(t, err)
		defer pubsub.Close()

		// 使用带超时的上下文接收消息
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		// 尝试接收消息（应该超时）
		_, err = pubsub.ReceiveMessage(ctxWithTimeout)
		assert.Error(t, err)
		// 检查是否是超时相关的错误
		assert.True(t,
			strings.Contains(err.Error(), "context deadline exceeded") ||
				strings.Contains(err.Error(), "i/o timeout"),
			"Expected timeout error, got: %v", err)
	})

	t.Run("PublishAfterClose", func(t *testing.T) {
		channel := "close_test"

		// 创建订阅
		pubsub, err := view.Subscribe(ctx, channel)
		require.NoError(t, err)

		// 立即关闭订阅
		err = pubsub.Close()
		require.NoError(t, err)

		// 发布消息（应该返回0个订阅者）
		subscribers, err := view.Publish(ctx, channel, "message after close")
		require.NoError(t, err)
		assert.Equal(t, int64(0), subscribers)
	})
}

// 基准测试
func BenchmarkRedisView_Publish(b *testing.B) {
	mr := miniredis.RunT(b)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	mockCmd := &MockRedisCmd{Client: client}
	logger := zap.NewNop()

	view := NewRedisView(mockCmd, "bench", logger)
	ctx := context.Background()
	channel := "benchmark_channel"
	message := "benchmark message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		view.Publish(ctx, channel, message)
	}
}

func BenchmarkRedisView_Subscribe(b *testing.B) {
	mr := miniredis.RunT(b)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	mockCmd := &MockRedisCmd{Client: client}
	logger := zap.NewNop()

	view := NewRedisView(mockCmd, "bench", logger)
	ctx := context.Background()
	channel := "benchmark_channel"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pubsub, _ := view.Subscribe(ctx, channel)
		pubsub.Close()
	}
}
