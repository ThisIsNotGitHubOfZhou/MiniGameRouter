package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	name     string
	exchange string
}

// New 创建一个新的 RabbitMQ 实例并返回
// s: RabbitMQ 服务地址
func New(s string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(s)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // exclusive
		false, // auto-delete
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
		name:    q.Name,
	}, nil
}

// Bind 绑定消息队列到交换机
func (r *RabbitMQ) Bind(exchange string) error {
	err := r.channel.QueueBind(
		r.name,   // queue name
		"",       // routing key
		exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind a queue: %w", err)
	}
	r.exchange = exchange
	return nil
}

// Send 向某个队列发送信息
func (r *RabbitMQ) Send(queue string, body interface{}) error {
	str, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	err = r.channel.Publish(
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			ReplyTo:     r.name,
			Body:        str,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}
	return nil
}

// Publish 向交换机发布信息
func (r *RabbitMQ) Publish(exchange string, body []byte) error {

	err := r.channel.Publish(
		exchange, // exchange
		"",       // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			ReplyTo:     r.name,
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}
	return nil
}

// Consume 消费消息
func (r *RabbitMQ) Consume() (<-chan amqp.Delivery, error) {
	c, err := r.channel.Consume(
		r.name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer: %w", err)
	}
	return c, nil
}

// Close 关闭 RabbitMQ 连接和通道
func (r *RabbitMQ) Close() {
	if err := r.channel.Close(); err != nil {
		log.Printf("failed to close channel: %v", err)
	}
	if err := r.conn.Close(); err != nil {
		log.Printf("failed to close connection: %v", err)
	}
}

// GetName 返回队列名称
func (r *RabbitMQ) GetName() string {
	return r.name
}

// GetExchange 返回交换机名称
func (r *RabbitMQ) GetExchange() string {
	return r.exchange
}
