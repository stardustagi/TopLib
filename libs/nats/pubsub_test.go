// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	topic = "topic"
	//chansPrefix = "channels"
	channel  = "9b7b1b3f-b1b0-46a8-a717-b8213f9eda3b"
	subtopic = "engine"
)

var (
	msgChan = make(chan *nats.Msg)
	data    = []byte("payload")
)

func TestPubsub(t *testing.T) {
	configMap := map[string]interface{}{
		"publisher_name": "publisher_worker",
		"consumer_name":  "consumer_worker",
		"url":            "nats://localhost:4222",
		"use_stream":     true,
		"type":           "publisher",
		"subject":        []string{"billing.*"},
	}
	// Initialize NatsConnection
	Init(configMap)
	natsInstance := GetNatsInstance()
	err := natsInstance.AddStream("billing", []string{"billing.*"})
	if err != nil {
		t.Errorf("Failed to add stream: %v", err)
		panic(err)
	}
	// Publish a message
	message := []byte("\"{\\\"publisherId\\\":\\\"91\\\",\\\"eventId\\\":\\\"nil\\\",\\\"userId\\\":\\\"1184241832\\\",\\\"firstName\\\":\\\"Snake ??\\\",\\\"lastName\\\":\\\"LN1184241832\\\",\\\"userName\\\":\\\"sd_hong\\\",\\\"timeStamp\\\":\\\"1731590619.434\\\",\\\"signature\\\":\\\"\\\",\\\"language\\\":\\\"ko\\\",\\\"channel\\\":\\\"TG\\\",\\\"version\\\":\\\"3.0.0\\\",\\\"fromType\\\":\\\"script\\\",\\\"traceId\\\":\\\"1731506981989-913fa511-f776-4ea7-ac00-b45f49f30297\\\",\\\"requestType\\\":\\\"getAd\\\",\\\"ip_address\\\":\\\"210.181.107.48\\\",\\\"location\\\":\\\"aHR0cHM6Ly9yZXN0YXVyYW50LXYyLnBpZ2d5cGlnZ3kuaW8vYnIvaW5kZXguaHRtbA==\\\",\\\"platform\\\":\\\"MAC\\\",\\\"zoneId\\\":\\\"137\\\"}\"")
	err = natsInstance.Publish("billing.clickinfo", message)
	if err != nil {
		t.Errorf("Failed to publish message: %v", err)
	}

	time.Sleep(2 * time.Second)
}

func TestSubscribe(t *testing.T) {
	configMap := map[string]interface{}{
		"publisher_name": "publisher_worker",
		"consumer_name":  "consumer_worker",
		"url":            "nats://localhost:4222",
		"use_stream":     true,
		"type":           "consumer",
		"subject":        []string{"billing.*"},
	}
	// 初始化 NatsConnection
	Init(configMap)
	natsInstance := GetNatsInstance()
	// 订阅主题
	_, err := natsInstance.conn.Subscribe("billing.clickinfo", func(msg *nats.Msg) {
		t.Logf("收到消息: %s", string(msg.Data))
	})
	if err != nil {
		t.Errorf("订阅失败: %v", err)
		return
	}
	// 等待消息到达
	t.Log("等待消息...")
	time.Sleep(30 * time.Second)
}
