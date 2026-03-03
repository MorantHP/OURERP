package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// OrderHandler 订单消息处理器接口
type OrderHandler interface {
	HandleOrderCreate(ctx context.Context, msg *OrderMessage) error
	HandleOrderUpdate(ctx context.Context, msg *OrderMessage) error
	HandleOrderCancel(ctx context.Context, msg *OrderMessage) error
	HandleOrderRefund(ctx context.Context, msg *OrderMessage) error
}

// Consumer Kafka消费者
type Consumer struct {
	consumer sarama.ConsumerGroup
	config   *Config
	handler  OrderHandler
	producer *Producer
	ready    chan bool
	mu       sync.Mutex
	running  bool
}

// NewConsumer 创建Kafka消费者
func NewConsumer(config *Config, handler OrderHandler, producer *Producer) (*Consumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = true
	saramaConfig.Consumer.Offsets.AutoCommit.Interval = time.Second * 5
	saramaConfig.Consumer.Group.Session.Timeout = config.Consumer.SessionTimeout
	saramaConfig.Consumer.Group.Heartbeat.Interval = config.Consumer.HeartbeatInterval

	consumer, err := sarama.NewConsumerGroup(config.Brokers, config.Consumer.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	return &Consumer{
		consumer: consumer,
		config:   config,
		handler:  handler,
		producer: producer,
		ready:    make(chan bool),
	}, nil
}

// Start 启动消费者
func (c *Consumer) Start(ctx context.Context, topics []string) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = true
	c.mu.Unlock()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := c.consumer.Consume(ctx, topics, c)
				if err != nil {
					log.Printf("[Kafka] 消费错误: %v", err)
					time.Sleep(time.Second * 5)
				}
			}
		}
	}()

	<-c.ready
	log.Printf("[Kafka] 消费者已启动, 订阅Topics: %v", topics)
	return nil
}

// Stop 停止消费者
func (c *Consumer) Stop() error {
	c.mu.Lock()
	c.running = false
	c.mu.Unlock()
	return c.consumer.Close()
}

// Setup 实现ConsumerGroupHandler接口
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

// Cleanup 实现ConsumerGroupHandler接口
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 实现ConsumerGroupHandler接口
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		session.MarkMessage(message, "")

		// 解析消息
		var orderMsg OrderMessage
		if err := json.Unmarshal(message.Value, &orderMsg); err != nil {
			log.Printf("[Kafka] 解析消息失败: %v, Topic: %s, Offset: %d", err, message.Topic, message.Offset)
			continue
		}

		// 处理消息
		ctx := context.Background()
		if err := c.handleMessage(ctx, message.Topic, &orderMsg); err != nil {
			log.Printf("[Kafka] 处理消息失败: %v, OrderID: %s", err, orderMsg.Data.PlatformOrderID)

			// 发送到死信队列
			if c.producer != nil {
				_ = c.producer.SendToDLQ(ctx, message.Topic, &orderMsg, err)
			}
		}
	}
	return nil
}

// handleMessage 处理消息
func (c *Consumer) handleMessage(ctx context.Context, topic string, msg *OrderMessage) error {
	var err error

	switch msg.MessageType {
	case MessageTypeOrderCreate:
		err = c.handler.HandleOrderCreate(ctx, msg)
	case MessageTypeOrderUpdate:
		err = c.handler.HandleOrderUpdate(ctx, msg)
	case MessageTypeOrderCancel:
		err = c.handler.HandleOrderCancel(ctx, msg)
	case MessageTypeOrderRefund:
		err = c.handler.HandleOrderRefund(ctx, msg)
	default:
		log.Printf("[Kafka] 未知消息类型: %s", msg.MessageType)
	}

	return err
}

// MultiTopicConsumer 多Topic消费者
type MultiTopicConsumer struct {
	consumers map[string]*Consumer
	config    *Config
	producer  *Producer
	mu        sync.RWMutex
}

// NewMultiTopicConsumer 创建多Topic消费者
func NewMultiTopicConsumer(config *Config, producer *Producer) *MultiTopicConsumer {
	return &MultiTopicConsumer{
		consumers: make(map[string]*Consumer),
		config:    config,
		producer:  producer,
	}
}

// Subscribe 订阅Topic
func (m *MultiTopicConsumer) Subscribe(ctx context.Context, topics []string, handler OrderHandler) error {
	for _, topic := range topics {
		groupID := fmt.Sprintf("%s-%s", m.config.Consumer.GroupID, topic)
		config := *m.config
		config.Consumer.GroupID = groupID

		consumer, err := NewConsumer(&config, handler, m.producer)
		if err != nil {
			return fmt.Errorf("failed to create consumer for topic %s: %w", topic, err)
		}

		m.mu.Lock()
		m.consumers[topic] = consumer
		m.mu.Unlock()

		if err := consumer.Start(ctx, []string{topic}); err != nil {
			return fmt.Errorf("failed to start consumer for topic %s: %w", topic, err)
		}
	}

	return nil
}

// StopAll 停止所有消费者
func (m *MultiTopicConsumer) StopAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var lastErr error
	for topic, consumer := range m.consumers {
		if err := consumer.Stop(); err != nil {
			lastErr = fmt.Errorf("failed to stop consumer for topic %s: %w", topic, err)
		}
	}
	return lastErr
}
