package redis

import (
	"context"
	"mq_perf_test/internal"
	"time"

	"github.com/mus-format/mus-go/varint"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

var err error

type TestClient struct {
	conn   *redis.Client
	pubsub *redis.PubSub

	PubConfigs []*internal.PublisherConfig
	SubConfigs []*internal.SubscriberConfig

	exchangeQueueMapping map[string][]string
}

func CreateClient(uri string) *TestClient {
	log.Debug().Msgf("redis uri: %s", uri)
	connOpts, err := redis.ParseURL(uri)
	if err != nil {
		panic(err)
	}

	conn := redis.NewClient(connOpts)

	c := &TestClient{conn: conn}

	return c
}
func (c *TestClient) BuildPublishers(pubConfigs []*internal.PublisherConfig) {
}

func (c *TestClient) StartPublishers(pubConfigs []*internal.PublisherConfig) {
	for _, pubConfig := range pubConfigs {
		//log.Debug().Msgf("publish config: %+v", pubConfig)
		for i := range pubConfig.ConcurrencyCount {
			go func(i int) {
				//log.Debug().Msgf("start publisher %d", i)
				ctx := context.Background()

				for j := 0; j < pubConfig.MessageCount; j++ {
					currentTs := time.Now().UnixMilli()
					tsSize := varint.SizeInt64(currentTs)
					tsBuf := make([]byte, tsSize)
					varint.MarshalInt64(currentTs, tsBuf)

					err = c.conn.Publish(ctx, pubConfig.Topic+pubConfig.RoutingKey, tsBuf).Err()
					if err != nil {
						panic(err)
					}
				}
			}(i)
		}
	}
}

func (c *TestClient) BuildSubscribers(subConfigs []*internal.SubscriberConfig) {

}

func (c *TestClient) StartSubscribers(subConfigs []*internal.SubscriberConfig, tsDiffCh chan int64) {
	ctx := context.Background()
	for i, subConf := range subConfigs {
		go func(i int) {
			//log.Debug().Msgf("start subscriber %d", i)
			pubsub := c.conn.Subscribe(ctx, subConf.Topic+subConf.RoutingKey)
			ch := pubsub.Channel()
			for {
				select {
				case msg := <-ch:
					sendTs, _, err := varint.UnmarshalInt64([]byte(msg.Payload))
					if err != nil {
						panic(err)
					}
					tsDiffCh <- time.Now().UnixMilli() - sendTs
				}
			}
		}(i)
	}
}
