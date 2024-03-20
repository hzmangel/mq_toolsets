package internal

import (
	"time"

	"github.com/samber/lo"

	"github.com/rs/zerolog/log"
)

type AppConfig struct {
	Uri string `json:"uri" yaml:"uri"`

	//// goroutine number and number of each goroutine
	//Concurrency      int `json:"concurrency" yaml:"concurrency"`
	//MsgCountPerRound int `json:"msg_count" yaml:"msg_count"`
	//
	//// How many queues will be launched in total, here are different means in different brokers
	//// For rmq: the queue count means topic exchange count
	//// For redis pubsub: the queue count means pubsub channel count
	//// For redis stream: TODO
	//QueueCount int `json:"queue_count" yaml:"queue_count"`

	// Publishers
	Publishers []*PublisherConfig `json:"publishers" yaml:"publishers"`

	// Subscribers
	Subscribers []*SubscriberConfig `json:"subscribers" yaml:"subscribers"`
}

type PublishClient interface {
	StartPublishers([]*PublisherConfig)
}

type SubscribeClient interface {
	StartSubscribers([]*SubscriberConfig, chan int64)
}

// PublisherConfig defines configuration for each single publisher.
// Each publisher should Each one topic with configurable concurrency and message count
type PublisherConfig struct {
	Topic            string `json:"topic" yaml:"topic"`
	RoutingKey       string `json:"routing_key" yaml:"routing_key"`
	ConcurrencyCount int    `json:"concurrency_count" yaml:"concurrency_count"`
	MessageCount     int    `json:"message_count" yaml:"message_count"`

	// Extra config
	AutoDelete bool `json:"auto_delete" yaml:"auto_delete"`
}

// SubscriberConfig defines a single subscriber which processing messages match routing key
type SubscriberConfig struct {
	Topic      string `json:"topic" yaml:"topic"`
	RoutingKey string `json:"routing_key" yaml:"routing_key"`

	// Extra configs
	AutoDeleteExchange bool `yaml:"auto_delete_exchange" json:"auto_delete_exchange"`
	AutoDeleteQueue    bool `json:"auto_delete_queue" yaml:"auto_delete_queue"`
	AutoAck            bool `json:"auto_ack" yaml:"auto_ack"`
}

// LatencyTest publish messages with publishers, and consume messages with subscribers.
func LatencyTest(pubClient PublishClient, subClient SubscribeClient, cfg *AppConfig) {
	if len(cfg.Publishers) == 0 || len(cfg.Subscribers) == 0 {
		panic("no publisher or subscriber")
	}
	tsDiffCh := make(chan int64)
	totalMsgCnt := lo.Sum(lo.Map(cfg.Publishers, func(c *PublisherConfig, _ int) int { return c.ConcurrencyCount * c.MessageCount }))
	go subClient.StartSubscribers(cfg.Subscribers, tsDiffCh)
	time.Sleep(time.Second)
	go pubClient.StartPublishers(cfg.Publishers)

	var tsDiffData []int64
	timer := time.NewTimer(time.Second)

	for done := false; !done; {
		select {
		case tsDiff := <-tsDiffCh:
			tsDiffData = append(tsDiffData, tsDiff)
		case <-timer.C:
			receivedMsgCount := len(tsDiffData)
			log.Debug().Msgf("received %d message, avg ts diff: %f", receivedMsgCount, float64(lo.Sum(tsDiffData))/float64(receivedMsgCount))
			if receivedMsgCount == totalMsgCnt {
				done = true
			}
			timer.Reset(time.Second)
		}
	}
}
