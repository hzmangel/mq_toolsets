package internal

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
