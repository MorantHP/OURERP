package kafka

import (
	"os"
	"strconv"
	"time"
)

// Kafka配置
type Config struct {
	// Broker地址，多个用逗号分隔
	Brokers []string
	// 消费者组ID
	ConsumerGroupID string
	// Topic配置
	Topics TopicsConfig
	// 生产者配置
	Producer ProducerConfig
	// 消费者配置
	Consumer ConsumerConfig
}

// TopicsConfig Topic配置
type TopicsConfig struct {
	Taobao    string
	JD        string
	Douyin    string
	Kuaishou  string
	All       string
	DLQ       string
}

// ProducerConfig 生产者配置
type ProducerConfig struct {
	// 发送超时
	WriteTimeout time.Duration
	// 是否等待所有副本确认
	Acks int
	// 重试次数
	Retries int
	// 批量大小
	BatchSize int
}

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	// 消费者组ID
	GroupID string
	// 初始偏移量
	InitialOffset string
	// 会话超时
	SessionTimeout time.Duration
	// 心跳间隔
	HeartbeatInterval time.Duration
	// 最大处理时间
	MaxProcessingTime time.Duration
	// 批量大小
	BatchSize int
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	return &Config{
		Brokers: parseBrokers(getEnv("KAFKA_BROKERS", "localhost:9093")),
		ConsumerGroupID: getEnv("KAFKA_CONSUMER_GROUP", "erp-order-consumer"),
		Topics: TopicsConfig{
			Taobao:    getEnv("KAFKA_TOPIC_TAOBAO", "orders.taobao"),
			JD:        getEnv("KAFKA_TOPIC_JD", "orders.jd"),
			Douyin:    getEnv("KAFKA_TOPIC_DOUYIN", "orders.douyin"),
			Kuaishou:  getEnv("KAFKA_TOPIC_KUAISHOU", "orders.kuaishou"),
			All:       getEnv("KAFKA_TOPIC_ALL", "orders.all"),
			DLQ:       getEnv("KAFKA_TOPIC_DLQ", "orders.dlq"),
		},
		Producer: ProducerConfig{
			WriteTimeout: time.Duration(getIntEnv("KAFKA_PRODUCER_TIMEOUT", 10)) * time.Second,
			Acks:         getIntEnv("KAFKA_PRODUCER_ACKS", 1),
			Retries:      getIntEnv("KAFKA_PRODUCER_RETRIES", 3),
			BatchSize:    getIntEnv("KAFKA_PRODUCER_BATCH_SIZE", 100),
		},
		Consumer: ConsumerConfig{
			GroupID:           getEnv("KAFKA_CONSUMER_GROUP", "erp-order-consumer"),
			InitialOffset:     getEnv("KAFKA_CONSUMER_OFFSET", "latest"),
			SessionTimeout:    time.Duration(getIntEnv("KAFKA_SESSION_TIMEOUT", 10)) * time.Second,
			HeartbeatInterval: time.Duration(getIntEnv("KAFKA_HEARTBEAT_INTERVAL", 3)) * time.Second,
			MaxProcessingTime: time.Duration(getIntEnv("KAFKA_MAX_PROCESSING_TIME", 60)) * time.Second,
			BatchSize:         getIntEnv("KAFKA_CONSUMER_BATCH_SIZE", 100),
		},
	}
}

// GetTopicByPlatform 根据平台获取Topic
func (c *Config) GetTopicByPlatform(platform string) string {
	switch platform {
	case "taobao", "tmall":
		return c.Topics.Taobao
	case "jd":
		return c.Topics.JD
	case "douyin":
		return c.Topics.Douyin
	case "kuaishou":
		return c.Topics.Kuaishou
	default:
		return c.Topics.All
	}
}

// GetAllTopics 获取所有订单Topic
func (c *Config) GetAllTopics() []string {
	return []string{
		c.Topics.Taobao,
		c.Topics.JD,
		c.Topics.Douyin,
		c.Topics.Kuaishou,
	}
}

func parseBrokers(brokers string) []string {
	if brokers == "" {
		return []string{"localhost:9092"}
	}
	// 简单分割，实际可能需要更复杂的解析
	result := []string{}
	for _, b := range splitString(brokers, ",") {
		if b != "" {
			result = append(result, b)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
