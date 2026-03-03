package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MorantHP/OURERP/internal/kafka"
)

func main() {
	// 命令行参数
	action := flag.String("action", "produce", "Action: produce, consume, simulate")
	platform := flag.String("platform", "taobao", "Platform: taobao, jd, douyin, kuaishou")
	count := flag.Int("count", 10, "Number of orders to generate")
	interval := flag.Int("interval", 5, "Interval in seconds for simulation")
	brokers := flag.String("brokers", "localhost:9093", "Kafka brokers")
	flag.Parse()

	// 加载配置
	config := kafka.LoadConfig()
	config.Brokers = []string{*brokers}

	switch *action {
	case "produce":
		produceOrders(config, *platform, *count)
	case "consume":
		consumeOrders(config)
	case "simulate":
		simulateOrders(config, *interval, *count)
	default:
		fmt.Println("Unknown action. Use: produce, consume, or simulate")
		flag.Usage()
		os.Exit(1)
	}
}

// produceOrders 生成并发送订单
func produceOrders(config *kafka.Config, platform string, count int) {
	producer, err := kafka.NewProducer(config)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	simulator := kafka.NewOrderSimulator(producer)

	fmt.Printf("🚀 开始生成 %d 个订单...\n", count)

	ctx := context.Background()
	for i := 0; i < count; i++ {
		// 使用批量生成方法
		if err := simulator.GenerateBatch(ctx, 1); err != nil {
			log.Printf("❌ 生成订单失败: %v", err)
			continue
		}
		fmt.Printf("✅ [%d/%d] 订单已发送\n", i+1, count)
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("🎉 完成! 共生成 %d 个订单\n", count)
}

// consumeOrders 消费订单
func consumeOrders(config *kafka.Config) {
	producer, err := kafka.NewProducer(config)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	// 创建简单的处理器
	handler := &consoleOrderHandler{}

	consumer, err := kafka.NewConsumer(config, handler, producer)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\n🛑 收到停止信号...")
		cancel()
	}()

	fmt.Println("👂 开始监听订单消息... (按 Ctrl+C 停止)")
	fmt.Println("Topics:", config.GetAllTopics())

	if err := consumer.Start(ctx, config.GetAllTopics()); err != nil {
		log.Printf("Consumer error: %v", err)
	}
}

// simulateOrders 持续模拟订单
func simulateOrders(config *kafka.Config, intervalSec, ordersPerInterval int) {
	producer, err := kafka.NewProducer(config)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	simulator := kafka.NewOrderSimulator(producer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\n🛑 收到停止信号...")
		cancel()
	}()

	fmt.Printf("🔄 开始模拟订单生成...\n")
	fmt.Printf("   间隔: %d 秒\n", intervalSec)
	fmt.Printf("   每次生成: %d 个订单\n", ordersPerInterval)
	fmt.Println("   按 Ctrl+C 停止")
	fmt.Println()

	if err := simulator.Start(ctx, time.Duration(intervalSec)*time.Second, ordersPerInterval); err != nil {
		if err != context.Canceled {
			log.Printf("Simulator error: %v", err)
		}
	}

	fmt.Println("✅ 模拟已停止")
}

// consoleOrderHandler 控制台订单处理器
type consoleOrderHandler struct{}

func (h *consoleOrderHandler) HandleOrderCreate(ctx context.Context, msg *kafka.OrderMessage) error {
	fmt.Printf("\n📦 [新订单] 平台: %s\n", msg.Platform)
	fmt.Printf("   订单号: %s\n", msg.Data.PlatformOrderID)
	fmt.Printf("   金额: %.2f\n", msg.Data.TotalAmount)
	fmt.Printf("   买家: %s\n", msg.Data.BuyerInfo.BuyerNick)
	fmt.Printf("   商品数: %d\n", len(msg.Data.Items))
	if msg.Data.ReceiverInfo != nil {
		fmt.Printf("   收货人: %s (%s)\n", msg.Data.ReceiverInfo.ReceiverName, msg.Data.ReceiverInfo.ReceiverPhone)
	}
	return nil
}

func (h *consoleOrderHandler) HandleOrderUpdate(ctx context.Context, msg *kafka.OrderMessage) error {
	fmt.Printf("\n📝 [订单更新] %s - %s -> %s\n", msg.Platform, msg.Data.PlatformOrderID, msg.Data.OrderStatus)
	return nil
}

func (h *consoleOrderHandler) HandleOrderCancel(ctx context.Context, msg *kafka.OrderMessage) error {
	fmt.Printf("\n❌ [订单取消] %s - %s\n", msg.Platform, msg.Data.PlatformOrderID)
	return nil
}

func (h *consoleOrderHandler) HandleOrderRefund(ctx context.Context, msg *kafka.OrderMessage) error {
	fmt.Printf("\n💰 [订单退款] %s - %s\n", msg.Platform, msg.Data.PlatformOrderID)
	return nil
}
