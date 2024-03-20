package rabbitmq

import (
	"context"
	"mq_perf_test/internal"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/mus-format/mus-go/varint"
	amqp "github.com/rabbitmq/amqp091-go"
)

var err error

type AmqpTestClient struct {
	conn *amqp.Connection
	ch   *amqp.Channel

	PubConfigs []*internal.PublisherConfig
	SubConfigs []*internal.SubscriberConfig

	exchangeQueueMapping map[string][]string
}

func CreateAmqpClient(uri string) *AmqpTestClient {
	log.Debug().Msgf("amqp uri: %s", uri)
	conn, err := amqp.Dial(uri)
	if err != nil {
		panic(err)
	}

	c := &AmqpTestClient{conn: conn}

	c.Init()

	return c
}

// Init connect the amqp server and create channel
func (c *AmqpTestClient) Init() {
	c.ch, err = c.conn.Channel()
	if err != nil {
		panic(err)
	}

	c.exchangeQueueMapping = make(map[string][]string)
}

// BuildPublishers in rmq branch, only declare the exchanges
func (c *AmqpTestClient) BuildPublishers(pubConfigs []*internal.PublisherConfig) {
	c.PubConfigs = pubConfigs
	for _, pubConf := range c.PubConfigs {
		// create exchange
		err = c.ch.ExchangeDeclare(pubConf.Topic, "topic", true, pubConf.AutoDelete, false, false, nil)
		if err != nil {
			panic(err)
		}
	}
}
func (c *AmqpTestClient) StartPublishers(pubConfigs []*internal.PublisherConfig) {
	for _, pubConfig := range pubConfigs {
		log.Debug().Msgf("publish config: %+v", pubConfig)
		for i := range pubConfig.ConcurrencyCount {
			go func(i int) {
				log.Debug().Msgf("start publisher %d", i)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				for j := 0; j < pubConfig.MessageCount; j++ {
					currentTs := time.Now().UnixMilli()
					tsSize := varint.SizeInt64(currentTs)
					tsBuf := make([]byte, tsSize)
					varint.MarshalInt64(currentTs, tsBuf)

					err = c.ch.PublishWithContext(ctx, pubConfig.Topic, pubConfig.RoutingKey, false, false, amqp.Publishing{
						ContentType: "text/plain",
						Body:        tsBuf,
					})
					if err != nil {
						panic(err)
					}
				}
			}(i)
		}
	}
}

// BuildSubscribers in rmq branch, declare exchanges, create queue and bind queue with exchange and routing key
func (c *AmqpTestClient) BuildSubscribers(subConfigs []*internal.SubscriberConfig) {
	c.SubConfigs = subConfigs
	for _, subConf := range subConfigs {
		// create exchange
		err = c.ch.ExchangeDeclare(subConf.Topic, "topic", true, subConf.AutoDeleteExchange, false, false, nil)
		if err != nil {
			panic(err)
		}

		qName := uuid.NewString()

		// Declare queue
		_, err := c.ch.QueueDeclare(qName, false, subConf.AutoDeleteQueue, true, false, nil)
		if err != nil {
			panic(err)

		}

		// Bind queue
		err = c.ch.QueueBind(qName, subConf.RoutingKey, subConf.Topic, false, nil)
		if err != nil {
			panic(err)
		}

		if _, ok := c.exchangeQueueMapping[subConf.Topic]; !ok {
			c.exchangeQueueMapping[subConf.Topic] = make([]string, 0)
		}
		c.exchangeQueueMapping[subConf.Topic] = append(c.exchangeQueueMapping[subConf.Topic], qName)
	}
}

func (c *AmqpTestClient) StartSubscribers(subConfigs []*internal.SubscriberConfig) {
	for _, subConfig := range subConfigs {
		if queues, ok := c.exchangeQueueMapping[subConfig.Topic]; ok {
			for _, qName := range queues {
				deliveries, err := c.ch.Consume(qName, subConfig.RoutingKey, subConfig.AutoAck, false, false, false, nil)
				if err != nil {
					panic(err)
				}
				go processMsg(deliveries)
			}
		} else {
			log.Debug().Msgf("no queue bind: %s", subConfig.Topic)
		}

	}
}

func processMsg(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		//log.Printf("got %dB delivery: [%v] %q", len(d.Body), d.DeliveryTag, d.Body)
		sendTs, _, err := varint.UnmarshalInt64(d.Body)
		if err != nil {
			panic(err)
		}
		log.Printf("ts diff: %d", time.Now().UnixMilli()-sendTs)
	}
}

// LatencyTest publish messages with publishers, and consume messages with subscribers.
func LatencyTest(pubClient, subClient *AmqpTestClient) {
	if len(pubClient.PubConfigs) == 0 {
		panic("no publisher")
	}

	if len(subClient.SubConfigs) == 0 {
		panic("no subscriber")
	}

}
