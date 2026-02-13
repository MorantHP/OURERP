package mock

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/MorantHP/OURERP/internal/models"
)

type OrderGenerator struct {
	rand *rand.Rand
}

func NewOrderGenerator() *OrderGenerator {
	return &OrderGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateOrder 生成单个模拟订单
func (g *OrderGenerator) GenerateOrder(shopID int64, platform string) *models.Order {
	buyer := MockBuyers[g.rand.Intn(len(MockBuyers))]
	product := MockProducts[g.rand.Intn(len(MockProducts))]
	
	// 随机数量 1-3
	qty := g.rand.Intn(3) + 1
	totalAmount := product.Price * float64(qty)
	
	// 根据权重选择状态
	status := g.randomStatus()
	
	order := &models.Order{
		OrderNo:         g.generateOrderNo(platform),
		Platform:        platform,
		PlatformOrderID: g.generatePlatformOrderID(platform),
		ShopID:          shopID,
		Status:          status,
		TotalAmount:     totalAmount,
		PayAmount:       totalAmount,
		BuyerNick:       buyer.Nick,
		ReceiverName:    buyer.Name,
		ReceiverPhone:   buyer.Phone,
		ReceiverAddress: fmt.Sprintf("%s%s%s%s", buyer.Province, buyer.City, buyer.District, buyer.Address),
		Items: []models.OrderItem{
			{
				SkuID:    product.SkuID,
				SkuName:  product.SkuName,
				Quantity: qty,
				Price:    product.Price,
			},
		},
	}

	// 设置时间
	now := time.Now()
	order.CreatedAt = now.Add(-time.Duration(g.rand.Intn(72)) * time.Hour) // 随机72小时内
	
	if status >= 200 { // 已支付
		order.PaidAt = &order.CreatedAt
	}
	if status >= 400 { // 已发货
		shippedAt := order.CreatedAt.Add(time.Duration(g.rand.Intn(24)+1) * time.Hour)
		order.ShippedAt = &shippedAt
		order.LogisticsCompany = LogisticsCompanies[g.rand.Intn(len(LogisticsCompanies))].Name
		order.LogisticsNo = g.generateLogisticsNo()
	}

	return order
}

// GenerateOrders 批量生成订单
func (g *OrderGenerator) GenerateOrders(count int, shopID int64, platform string) []*models.Order {
	orders := make([]*models.Order, count)
	for i := 0; i < count; i++ {
		orders[i] = g.GenerateOrder(shopID, platform)
	}
	return orders
}

// randomStatus 根据权重随机选择状态
func (g *OrderGenerator) randomStatus() int {
	totalWeight := 0
	for _, sw := range StatusWeights {
		totalWeight += sw.Weight
	}
	
	r := g.rand.Intn(totalWeight)
	cumulative := 0
	for _, sw := range StatusWeights {
		cumulative += sw.Weight
		if r < cumulative {
			return sw.Status
		}
	}
	return 100
}

// generateOrderNo 生成系统订单号
func (g *OrderGenerator) generateOrderNo(platform string) string {
	prefix := "ORD"
	switch platform {
	case "taobao":
		prefix = "TB"
	case "jd":
		prefix = "JD"
	case "douyin":
		prefix = "DY"
	case "pdd":
		prefix = "PD"
	}
	return fmt.Sprintf("%s%s%04d", 
		prefix, 
		time.Now().Format("20060102"), 
		g.rand.Intn(9000)+1000)
}

// generatePlatformOrderID 生成平台订单ID
func (g *OrderGenerator) generatePlatformOrderID(platform string) string {
	switch platform {
	case "taobao":
		return fmt.Sprintf("%d%d", time.Now().Unix(), g.rand.Intn(1000000))
	case "jd":
		return fmt.Sprintf("JD%d", g.rand.Intn(900000000000)+100000000000)
	case "douyin":
		return fmt.Sprintf("DY%s%d", time.Now().Format("20060102150405"), g.rand.Intn(1000))
	case "pdd":
		return fmt.Sprintf("PD%d-%d", time.Now().Unix(), g.rand.Intn(10000))
	default:
		return fmt.Sprintf("%d", time.Now().Unix())
	}
}

// generateLogisticsNo 生成物流单号
func (g *OrderGenerator) generateLogisticsNo() string {
	// 顺丰格式: SF 开头 + 数字
	return fmt.Sprintf("SF%d%d", g.rand.Intn(9)+1, g.rand.Intn(9000000000)+1000000000)
}

// GenerateRealtimeOrders 实时生成新订单（模拟持续接单）
func (g *OrderGenerator) GenerateRealtimeOrders(shopID int64) <-chan *models.Order {
	ch := make(chan *models.Order)
	
	go func() {
		for {
			// 随机间隔 1-10 秒生成一个订单
			interval := time.Duration(g.rand.Intn(10)+1) * time.Second
			time.Sleep(interval)
			
			platform := Platforms[g.rand.Intn(len(Platforms))]
			order := g.GenerateOrder(shopID, platform)
			
			// 新订单默认待付款或待审核
			if g.rand.Intn(2) == 0 {
				order.Status = 100 // 待付款
			} else {
				order.Status = 200 // 待审核
				order.PaidAt = &order.CreatedAt
			}
			
			ch <- order
		}
	}()
	
	return ch
}