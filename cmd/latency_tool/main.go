package main

import (
	"encoding/json"
	"flag"
	"mq_perf_test/internal"
	"mq_perf_test/internal/rabbitmq"
	"mq_perf_test/internal/redis"
	"net/url"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var err error

func loadConfigFile(configFile string) *internal.AppConfig {
	configBytes, err := os.ReadFile(configFile)
	if err != nil {
		panic(err)
	}

	var cfg internal.AppConfig
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
		pubC := rabbitmq.CreateClient(cfg.Uri)
		pubC.BuildPublishers(cfg.Publishers)

		subC := rabbitmq.CreateClient(cfg.Uri)
		subC.BuildSubscribers(cfg.Subscribers)

		internal.LatencyTest(pubC, subC, cfg)
	case "redis":
		pubC := redis.CreateClient(cfg.Uri)
		pubC.BuildPublishers(cfg.Publishers)

		subC := redis.CreateClient(cfg.Uri)
		subC.BuildSubscribers(cfg.Subscribers)

		internal.LatencyTest(pubC, subC, cfg)
	}
}
