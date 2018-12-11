package rmq

import (
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

func (r *rabbitMQ) InitConnection(url string) *rabbitMQ {
	if url == "" {
		panic("Cannot initialize connection to broker, bad url.")
	}

	var err error
	r.connection, err = amqp.Dial(fmt.Sprintf("%s/", url))
	if err != nil {
		panic("Failed to connect to AMQP compatible broker at: " + url)
	}

	return r
}

func (r *rabbitMQ) InitChannels(channels map[string]ChannelSettings) {
	r.channels = map[string]*amqp.Channel{}
	keys := GetKeys(channels)

	for _, channelName := range keys {
		channel, err := r.connection.Channel()
		FailOnError(err, "Failed to open connection with channel")

		r.channels[channelName] = channel
		settings := channels[channelName]

		r.declareExchange(channel, settings)
		r.declareQueue(channel, settings)

		if settings.BindingKey != "" {
			r.bindQueue(channel, settings)
		}

		if settings.ConsumeActivate == true {
			r.declareCunsumer(channel, settings)
		}
	}
}

func (r *rabbitMQ) declareExchange(ch *amqp.Channel, settings ChannelSettings) {
	err := ch.ExchangeDeclare(
		settings.ExchangeName, // name
		settings.ExchangeType, // type
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // no-wait
		nil,                   // arguments
	)
	FailOnError(err, "Failed to declare an exchange")
}

func (r *rabbitMQ) declareQueue(ch *amqp.Channel, settings ChannelSettings) {
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
	FailOnError(err, "Failed to declare a queue")
}

func (r *rabbitMQ) bindQueue(ch *amqp.Channel, settings ChannelSettings) {
	err := ch.QueueBind(
		settings.QueueName,    // queue name
		settings.BindingKey,   // routing key
		settings.ExchangeName, // exchange
		false,
		nil)
	FailOnError(err, "Failed to bind a queue")
}

func (r *rabbitMQ) declareCunsumer(channel *amqp.Channel, settings ChannelSettings) {
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
	FailOnError(err, "Failed to register a consumer")

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

func (r *rabbitMQ) sendToQueue(body []byte, queueName string) error {
	if r.connection == nil {
		panic("Tried to send message before connection was initialized.")
	}
	channel, err := r.connection.Channel() // Get a channel from the connection
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

func (r *rabbitMQ) sendToInternal(request Request) {
	_requestByte, marshalErr := json.Marshal(request)
	FailOnError(marshalErr, "Failed on marshal request message.")

	err := r.channels[NamespaceInternal].Publish(
		"",                // exchange
		NamespaceInternal, // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(_requestByte),
		})
	FailOnError(err, "Failed to publish a message to internal service.")
	log.Printf("%s Sent message to [* %s *]: Message %s", Header, NamespaceInternal, _requestByte)
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s %s: %s", Header, msg, err)
	}
}
