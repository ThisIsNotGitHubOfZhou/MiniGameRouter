package main

import (
	"github.com/streadway/amqp"
	"log"
)

func client() {
	conn, err := amqp.Dial("amqp://thisiszhou:thisiszhou@9.135.95.71:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// 声明交换机
	err = ch.ExchangeDeclare(
		"route_logs", // name
		"fanout",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare an exchange: %v", err)
	}

	// 声明队列
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	//绑定队列到交换机
	err = ch.QueueBind(
		q.Name,       // queue name
		"",           // routing key
		"route_logs", // exchange
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind a queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			// 判断是否需要这条路由信息
			if isMyRoute(d.Body) {
				log.Printf("Route information processed: %s", d.Body)
			} else {
				log.Printf("Route information ignored: %s", d.Body)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func isMyRoute(body []byte) bool {
	return (body[1] == 'A' && body[2] == 'B' && body[3] == 'C')
}
