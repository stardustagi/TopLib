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
	url := nats.DefaultURL

	// Initialize NatsConnection
	var mq *NatsConnection
	var err error
	if mq, err = NewNatsConnect(url, true); err != nil {
		t.Error(err)
		defer mq.Close()
	}

	// Publish a message
	message := []byte("\"{\\\"publisherId\\\":\\\"91\\\",\\\"eventId\\\":\\\"nil\\\",\\\"userId\\\":\\\"1184241832\\\",\\\"firstName\\\":\\\"Snake ??\\\",\\\"lastName\\\":\\\"LN1184241832\\\",\\\"userName\\\":\\\"sd_hong\\\",\\\"timeStamp\\\":\\\"1731590619.434\\\",\\\"signature\\\":\\\"\\\",\\\"language\\\":\\\"ko\\\",\\\"channel\\\":\\\"TG\\\",\\\"version\\\":\\\"3.0.0\\\",\\\"fromType\\\":\\\"script\\\",\\\"traceId\\\":\\\"1731506981989-913fa511-f776-4ea7-ac00-b45f49f30297\\\",\\\"requestType\\\":\\\"getAd\\\",\\\"ip_address\\\":\\\"210.181.107.48\\\",\\\"location\\\":\\\"aHR0cHM6Ly9yZXN0YXVyYW50LXYyLnBpZ2d5cGlnZ3kuaW8vYnIvaW5kZXguaHRtbA==\\\",\\\"platform\\\":\\\"MAC\\\",\\\"zoneId\\\":\\\"137\\\"}\"")
	mq.Publish("ad_info.clickinfo", message)

	// Keep the program running to allow time for message processing
	time.Sleep(2 * time.Second)

}

func handler(msg *nats.Msg) error {
	msgChan <- msg
	return nil
}
