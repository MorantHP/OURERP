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

// Producer Kafka生产者
type Producer struct {
	producer sarama.SyncProducer
	config   *Config
	mu       sync.Mutex
}

// NewProducer 创建Kafka生产者
func NewProducer(config *Config) (*Producer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.RequiredAcks(config.Producer.Acks)
	saramaConfig.Producer.Retry.Max = config.Producer.Retries
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Partitioner = sarama.NewHashPartitioner
	saramaConfig.Producer.Timeout = config.Producer.WriteTimeout

	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &Producer{
		producer: producer,
		config:   config,
	}, nil
}

// SendMessage 发送消息到指定Topic
func (p *Producer) SendMessage(ctx context.Context, topic string, message *OrderMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(message.Data.PlatformOrderID),
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("[Kafka] 消息已发送 - Topic: %s, Partition: %d, Offset: %d, OrderID: %s",
		topic, partition, offset, message.Data.PlatformOrderID)

	return nil
}

// SendOrderMessage 发送订单消息（自动选择Topic）
func (p *Producer) SendOrderMessage(ctx context.Context, message *OrderMessage) error {
	topic := p.config.GetTopicByPlatform(message.Platform)
	return p.SendMessage(ctx, topic, message)
}

// SendOrderToAll 发送订单到合并Topic
func (p *Producer) SendOrderToAll(ctx context.Context, message *OrderMessage) error {
	return p.SendMessage(ctx, p.config.Topics.All, message)
}

// SendToDLQ 发送到死信队列
func (p *Producer) SendToDLQ(ctx context.Context, originalTopic string, message *OrderMessage, err error) error {
	// 添加错误信息到扩展数据
	if message.Data.ExtendData == nil {
		message.Data.ExtendData = make(map[string]interface{})
	}
	message.Data.ExtendData["dlq_original_topic"] = originalTopic
	message.Data.ExtendData["dlq_error"] = err.Error()
	message.Data.ExtendData["dlq_timestamp"] = time.Now().Format(time.RFC3339)

	return p.SendMessage(ctx, p.config.Topics.DLQ, message)
}

// Close 关闭生产者
func (p *Producer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}

// BatchProducer 批量生产者
type BatchProducer struct {
	producer   *Producer
	messageCh  chan *batchMessage
	batchSize  int
	batchWait  time.Duration
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

type batchMessage struct {
	topic   string
	message *OrderMessage
	errCh   chan error
}

// NewBatchProducer 创建批量生产者
func NewBatchProducer(producer *Producer, batchSize int, batchWait time.Duration) *BatchProducer {
	bp := &BatchProducer{
		producer:  producer,
		messageCh: make(chan *batchMessage, 1000),
		batchSize: batchSize,
		batchWait: batchWait,
		stopCh:    make(chan struct{}),
	}

	bp.wg.Add(1)
	go bp.runBatcher()

	return bp
}

func (bp *BatchProducer) runBatcher() {
	defer bp.wg.Done()

	batch := make([]*batchMessage, 0, bp.batchSize)
	timer := time.NewTimer(bp.batchWait)
	defer timer.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}

		// 批量发送
		for _, bm := range batch {
			err := bp.producer.SendOrderMessage(context.Background(), bm.message)
			bm.errCh <- err
		}
		batch = batch[:0]
	}

	for {
		select {
		case bm := <-bp.messageCh:
			batch = append(batch, bm)
			if len(batch) >= bp.batchSize {
				flush()
				timer.Reset(bp.batchWait)
			}

		case <-timer.C:
			flush()
			timer.Reset(bp.batchWait)

		case <-bp.stopCh:
			flush()
			return
		}
	}
}

// SendAsync 异步发送消息
func (bp *BatchProducer) SendAsync(topic string, message *OrderMessage) <-chan error {
	errCh := make(chan error, 1)
	bp.messageCh <- &batchMessage{
		topic:   topic,
		message: message,
		errCh:   errCh,
	}
	return errCh
}

// Close 关闭批量生产者
func (bp *BatchProducer) Close() error {
	close(bp.stopCh)
	bp.wg.Wait()
	return bp.producer.Close()
}
