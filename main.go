package main

import (
	"encoding/json"
	"flag"
	"mq_perf_test/internal"
	"mq_perf_test/internal/rabbitmq"
	"net/url"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var err error

type appConfig struct {
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
	Publishers []*internal.PublisherConfig `json:"publishers" yaml:"publishers"`

	// Subscribers
	Subscribers []*internal.SubscriberConfig `json:"subscribers" yaml:"subscribers"`
}

func loadConfigFile(configFile string) *appConfig {
	configBytes, err := os.ReadFile(configFile)
	if err != nil {
		panic(err)
	}

	var cfg appConfig
	if strings.HasSuffix(configFile, "yaml") {
		err = yaml.Unmarshal(configBytes, &cfg)
		if err != nil {
			panic(err)
		}
	} else if strings.HasSuffix(configFile, "json") {
		err = json.Unmarshal(configBytes, &cfg)
		if err != nil {
			panic(err)
		}
	} else {
		panic("unknown config file type")
	}

	return &cfg
}

func main() {
	configFile := flag.String("config", "config.yaml", "config file path")
	flag.Parse()

	cfg := loadConfigFile(*configFile)
	if err != nil {
		panic(err)
	}
	log.Info().Msgf("config: %+v", cfg)

	uri, err := url.Parse(cfg.Uri)
	if err != nil {
		panic(err)
	}

	switch uri.Scheme {

	case "amqp":
		// Build publishers
		pubC := rabbitmq.CreateAmqpClient(cfg.Uri)
		pubC.BuildPublishers(cfg.Publishers)

		subC := rabbitmq.CreateAmqpClient(cfg.Uri)
		subC.BuildSubscribers(cfg.Subscribers)

		subC.StartSubscribers(cfg.Subscribers)

		pubC.StartPublishers(cfg.Publishers)
		select {}
	}
}
