package queue

import (
    "context"
    "time"
    amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
    conn *amqp.Connection
    ch   *amqp.Channel
    queue string
}

func NewPublisher(amqpURL, queueName string) (*Publisher, error) {
    conn, err := amqp.Dial(amqpURL)
    if err != nil { return nil, err }
    ch, err := conn.Channel()
    if err != nil { conn.Close(); return nil, err }
    _, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
    if err != nil { ch.Close(); conn.Close(); return nil, err }
    return &Publisher{conn: conn, ch: ch, queue: queueName}, nil
}

func (p *Publisher) Publish(ctx context.Context, body []byte) error {
    return p.ch.PublishWithContext(ctx, "", p.queue, false, false, amqp.Publishing{
        ContentType: "application/json",
        DeliveryMode: amqp.Persistent,
        Timestamp: time.Now(),
        Body: body,
    })
}

func (p *Publisher) Close() {
    if p.ch != nil { _ = p.ch.Close() }
    if p.conn != nil { _ = p.conn.Close() }
}


