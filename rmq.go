package rmq

import (
	core "bb_core"
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// // Defines our interface for connecting, producing and consuming messages.
// type IMessagingClient interface {
// 	InitConnection(url string)
// 	Publish(msg []byte, exchangeName string, exchangeType string) error
// 	sendToQueue(msg []byte, queueName string) error
// 	Subscribe(exchangeName string, exchangeType string, consumerName string, handlerFunc func(amqp.Delivery)) error
// 	SubscribeToQueue(queueName string, consumerName string, handlerFunc func(amqp.Delivery)) error
// 	Close()
// }

var Rabbit rabbitMQ

type rabbitMQ struct {
	connection *amqp.Connection
	channels   map[string]*amqp.Channel
}

func (*rabbitMQ) InitConnection(url string) {
	var err error

	fmt.Print(url)
	if url == "" {
		panic("Cannot initialize connection to broker, bad url.")
	}

	Rabbit.connection, err = amqp.Dial(fmt.Sprintf("%s/", url))
	if err != nil {
		panic("Failed to connect to AMQP compatible broker at: " + url)
	}
}

func (*rabbitMQ) InitChannels(channels map[string]core.ChannelSettings) {
	Rabbit.channels = map[string]*amqp.Channel{}
	keys := getKeys(channels)

	for _, channelName := range keys {
		channel, err := Rabbit.connection.Channel()
		core.FailOnError(Header, "Failed to open connection with channel", err)

		Rabbit.channels[channelName] = channel
		settings := channels[channelName]

		Rabbit.declareExchange(channel, settings)
		Rabbit.declareQueue(channel, settings)

		if settings.BindingKey != "" {
			Rabbit.bindQueue(channel, settings)
		}

		if settings.ConsumeActivate == true {
			Rabbit.declareCunsumer(channel, settings)
		}
	}
}

func (*rabbitMQ) declareExchange(ch *amqp.Channel, settings core.ChannelSettings) {
	err := ch.ExchangeDeclare(
		settings.ExchangeName, // name
		settings.ExchangeType, // type
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // no-wait
		nil,                   // arguments
	)
	core.FailOnError(Header, "Failed to declare an exchange", err)
}

func (*rabbitMQ) declareQueue(ch *amqp.Channel, settings core.ChannelSettings) {
	args := make(amqp.Table)
	args["x-message-ttl"] = settings.QueueOptions.MessageTTL

	_, err := ch.QueueDeclare(
		settings.QueueName,               // name
		settings.QueueOptions.Durable,    // durable
		settings.QueueOptions.AutoDelete, // delete when unused
		false,                            // exclusive
		false,                            // no-wait
		args,                             // arguments
	)
	core.FailOnError(Header, "Failed to declare a queue", err)
}

func (*rabbitMQ) bindQueue(ch *amqp.Channel, settings core.ChannelSettings) {
	err := ch.QueueBind(
		settings.QueueName,    // queue name
		settings.BindingKey,   // routing key
		settings.ExchangeName, // exchange
		false,
		nil)
	core.FailOnError(Header, "Failed to bind a queue", err)
}

func (*rabbitMQ) declareCunsumer(channel *amqp.Channel, settings core.ChannelSettings) {
	queueName := settings.QueueName
	msgs, err := channel.Consume(
		queueName,                   // queue
		"",                          // consumer
		settings.QueueOptions.NoAck, // auto-ack
		false,                       // exclusive
		false,                       // no-local
		false,                       // no-wait
		nil,                         // args
	)
	core.FailOnError(Header, "Failed to register a consumer", err)

	go func() {
		for message := range msgs {
			log.Printf("%s Received a message from [* %s *]: Message %s", Header, queueName, message.Body)
			// rmqProcessing(message.Body)
		}
	}()
	if queueName == "" {
		if settings.BindingKey != "" {
			queueName = settings.BindingKey
		} else {
			queueName = "current"
		}
	}
	log.Printf("%s Waiting for messages from %s channel. To exit press CTRL+C", Header, queueName)
}

func (*rabbitMQ) sendToQueue(body []byte, queueName string) error {
	if Rabbit.connection == nil {
		panic("Tried to send message before connection was initialized.")
	}
	channel, err := Rabbit.connection.Channel() // Get a channel from the connection
	defer channel.Close()

	// Declare a queue that will be created if not exists with some args
	queue, err := channel.QueueDeclare(
		queueName, // our queue name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	// Publishes a message onto the queue.
	err = channel.Publish(
		"",         // use the default exchange
		queue.Name, // routing key, e.g. our queue name
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body, // Our JSON body as []byte
		})
	fmt.Printf("A message was sent to queue %v: %v", queueName, body)
	return err
}

func (*rabbitMQ) sendToInternal(request core.Request) {
	_requestByte, marshalErr := json.Marshal(request)
	core.FailOnError(Header, "Failed on marshal request message.", marshalErr)

	err := Rabbit.channels[core.NAMESPACE_INTERNAL].Publish(
		"",                      // exchange
		core.NAMESPACE_INTERNAL, // routing key
		false,                   // mandatory
		false,                   // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(_requestByte),
		})
	core.FailOnError(Header, "Failed to publish a message to internal service.", err)
	log.Printf("%s Sent message to [* %s *]: Message %s", Header, core.NAMESPACE_INTERNAL, _requestByte)
}
