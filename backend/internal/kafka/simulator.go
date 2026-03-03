package kafka

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// OrderSimulator 订单模拟器
type OrderSimulator struct {
	producer   *Producer
	platforms  []PlatformGenerator
	shopIDMap  map[string]int64 // platform -> shop_id
	stopCh     chan struct{}
	running    bool
}

// PlatformGenerator 平台订单生成器接口
type PlatformGenerator interface {
	Platform() string
	GenerateOrder() *OrderData
	GenerateUpdate(orderID string) *OrderData
}

// NewOrderSimulator 创建订单模拟器
func NewOrderSimulator(producer *Producer) *OrderSimulator {
	return &OrderSimulator{
		producer:  producer,
		platforms: []PlatformGenerator{
			NewTaobaoGenerator(),
			NewJDGenerator(),
			NewDouyinGenerator(),
			NewKuaishouGenerator(),
		},
		shopIDMap: map[string]int64{
			"taobao":   1,
			"tmall":    1,
			"jd":       2,
			"douyin":   3,
			"kuaishou": 4,
		},
		stopCh: make(chan struct{}),
	}
}

// SetShopID 设置店铺ID映射
func (s *OrderSimulator) SetShopID(platform string, shopID int64) {
	s.shopIDMap[platform] = shopID
}

// Start 启动模拟器
func (s *OrderSimulator) Start(ctx context.Context, interval time.Duration, ordersPerInterval int) error {
	s.running = true
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logPrefix := "[OrderSimulator]"
	fmt.Printf("%s 已启动, 每 %v 生成 %d 个订单\n", logPrefix, interval, ordersPerInterval)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.stopCh:
			return nil
		case <-ticker.C:
			for i := 0; i < ordersPerInterval; i++ {
				s.generateAndSend(ctx)
			}
		}
	}
}

// Stop 停止模拟器
func (s *OrderSimulator) Stop() {
	if s.running {
		close(s.stopCh)
		s.running = false
	}
}

// generateAndSend 生成并发送订单
func (s *OrderSimulator) generateAndSend(ctx context.Context) {
	// 随机选择平台
	platform := s.platforms[rand.Intn(len(s.platforms))]

	// 生成订单数据
	orderData := platform.GenerateOrder()

	// 创建消息
	msg := NewOrderCreateMessage(platform.Platform(), orderData)
	msg.ShopID = s.shopIDMap[platform.Platform()]

	// 发送到Kafka
	if err := s.producer.SendOrderMessage(ctx, msg); err != nil {
		fmt.Printf("[OrderSimulator] 发送失败: %v\n", err)
	} else {
		fmt.Printf("[OrderSimulator] 已发送 - 平台: %s, 订单号: %s, 金额: %.2f\n",
			platform.Platform(), orderData.PlatformOrderID, orderData.TotalAmount)
	}
}

// GenerateSingleOrder 生成单个订单
func (s *OrderSimulator) GenerateSingleOrder(ctx context.Context, platform string) (*OrderMessage, error) {
	var gen PlatformGenerator
	for _, p := range s.platforms {
		if p.Platform() == platform {
			gen = p
			break
		}
	}
	if gen == nil {
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}

	orderData := gen.GenerateOrder()
	msg := NewOrderCreateMessage(platform, orderData)
	msg.ShopID = s.shopIDMap[platform]

	if err := s.producer.SendOrderMessage(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

// GenerateBatch 批量生成订单
func (s *OrderSimulator) GenerateBatch(ctx context.Context, count int) error {
	for i := 0; i < count; i++ {
		s.generateAndSend(ctx)
	}
	return nil
}

// ==================== 淘宝/天猫生成器 ====================

// TaobaoGenerator 淘宝订单生成器
type TaobaoGenerator struct {
	statuses  []string
	products  []productInfo
	provinces []string
	cities    map[string][]string
}

type productInfo struct {
	name  string
	price float64
	spec  string
}

func NewTaobaoGenerator() *TaobaoGenerator {
	return &TaobaoGenerator{
		statuses: []string{"WAIT_BUYER_PAY", "WAIT_SELLER_SEND", "SELLER_CONSIGNED", "TRADE_FINISHED"},
		products: []productInfo{
			{"男士休闲外套春季新款", 199.00, "颜色:黑色;尺码:XL"},
			{"女士连衣裙夏季新款", 159.00, "颜色:白色;尺码:M"},
			{"运动鞋跑步鞋透气", 299.00, "颜色:蓝色;尺码:42"},
			{"手机壳全包防摔", 29.90, "颜色:透明;型号:iPhone15"},
			{"蓝牙耳机无线降噪", 399.00, "颜色:白色"},
			{"保温杯大容量不锈钢", 89.00, "颜色:银色;容量:500ml"},
			{"双肩包电脑包商务", 159.00, "颜色:黑色"},
			{"护肤套装补水保湿", 299.00, "规格:标准装"},
			{"零食大礼包坚果", 99.00, "规格:1kg装"},
			{"充电宝20000毫安", 129.00, "颜色:白色"},
		},
		provinces: []string{"浙江省", "江苏省", "广东省", "北京市", "上海市", "四川省", "湖北省", "山东省"},
		cities: map[string][]string{
			"浙江省": {"杭州市", "宁波市", "温州市", "嘉兴市"},
			"江苏省": {"南京市", "苏州市", "无锡市", "常州市"},
			"广东省": {"广州市", "深圳市", "东莞市", "佛山市"},
			"北京市": {"北京市"},
			"上海市": {"上海市"},
			"四川省": {"成都市", "绵阳市", "德阳市"},
			"湖北省": {"武汉市", "宜昌市", "襄阳市"},
			"山东省": {"济南市", "青岛市", "烟台市"},
		},
	}
}

func (g *TaobaoGenerator) Platform() string {
	if rand.Intn(10) < 7 {
		return "taobao"
	}
	return "tmall"
}

func (g *TaobaoGenerator) GenerateOrder() *OrderData {
	now := time.Now()
	province := g.provinces[rand.Intn(len(g.provinces))]
	cities := g.cities[province]
	city := cities[rand.Intn(len(cities))]

	// 随机选择1-3个商品
	itemCount := rand.Intn(3) + 1
	items := make([]*OrderItem, 0, itemCount)
	var totalAmount float64

	selectedProducts := make(map[int]bool)
	for len(items) < itemCount {
		idx := rand.Intn(len(g.products))
		if selectedProducts[idx] {
			continue
		}
		selectedProducts[idx] = true

		product := g.products[idx]
		quantity := rand.Intn(3) + 1
		itemTotal := product.price * float64(quantity)
		totalAmount += itemTotal

		items = append(items, &OrderItem{
			SKUID:       fmt.Sprintf("TB_SKU_%d", rand.Intn(100000)),
			SKUName:     product.name,
			SKUSpec:     product.spec,
			Quantity:    quantity,
			Price:       product.price,
			TotalAmount: itemTotal,
			ProductID:   fmt.Sprintf("TB_PID_%d", rand.Intn(10000)),
			ProductName: product.name,
		})
	}

	// 随机折扣
	discount := float64(rand.Intn(30)) / 100.0 * totalAmount
	payAmount := totalAmount - discount

	return &OrderData{
		PlatformOrderID: fmt.Sprintf("%d%010d", now.Year(), rand.Intn(1000000000)),
		OrderStatus:     g.statuses[rand.Intn(len(g.statuses))],
		TotalAmount:     totalAmount,
		PayAmount:       payAmount,
		DiscountAmount:  discount,
		PostFee:         0,
		BuyerInfo: &BuyerInfo{
			BuyerID:    fmt.Sprintf("tb_buyer_%d", rand.Intn(100000)),
			BuyerNick:  fmt.Sprintf("淘宝用户%d", rand.Intn(10000)),
			VipLevel:   rand.Intn(7),
			IsNewBuyer: rand.Intn(10) < 2,
		},
		ReceiverInfo: &ReceiverInfo{
			ReceiverName:     fmt.Sprintf("收货人%d", rand.Intn(100)),
			ReceiverPhone:    fmt.Sprintf("1%d%08d", 3+rand.Intn(5), rand.Intn(100000000)),
			ReceiverProvince: province,
			ReceiverCity:     city,
			ReceiverDistrict: "某区",
			ReceiverAddress:  fmt.Sprintf("某街道%d号", rand.Intn(1000)),
		},
		Items:       items,
		CreatedAt:   now.Add(-time.Duration(rand.Intn(24)) * time.Hour),
		UpdatedAt:   now,
	}
}

func (g *TaobaoGenerator) GenerateUpdate(orderID string) *OrderData {
	order := g.GenerateOrder()
	order.PlatformOrderID = orderID
	return order
}

// ==================== 京东生成器 ====================

type JDGenerator struct {
	TaobaoGenerator
}

func NewJDGenerator() *JDGenerator {
	g := &JDGenerator{}
	g.statuses = []string{"WAIT_PAY", "WAIT_DELIVER", "DELIVERED", "FINISHED", "CANCEL"}
	g.products = []productInfo{
		{"Apple iPhone 15 Pro", 7999.00, "颜色:深空黑;存储:256GB"},
		{"小米14 Ultra", 5999.00, "颜色:白色;存储:512GB"},
		{"戴森吸尘器V15", 4990.00, "型号:标准版"},
		{"索尼WH-1000XM5耳机", 2699.00, "颜色:黑色"},
		{"戴尔显示器27寸4K", 2499.00, "型号:U2723QE"},
		{"罗技MX Master 3S鼠标", 799.00, "颜色:黑色"},
		{"iPad Air 5", 4799.00, "颜色:深空灰;存储:256GB"},
		{"海尔冰箱双开门", 4999.00, "型号:BCT-500"},
		{"联想笔记本ThinkPad", 6999.00, "配置:i7/16G/512G"},
		{"美的空调1.5匹", 2999.00, "型号:KFR-35GW"},
	}
	g.provinces = []string{"北京市", "上海市", "广东省", "江苏省", "浙江省"}
	g.cities = map[string][]string{
		"北京市": {"北京市"},
		"上海市": {"上海市"},
		"广东省": {"广州市", "深圳市", "东莞市"},
		"江苏省": {"南京市", "苏州市", "无锡市"},
		"浙江省": {"杭州市", "宁波市", "温州市"},
	}
	return g
}

func (g *JDGenerator) Platform() string {
	return "jd"
}

func (g *JDGenerator) GenerateOrder() *OrderData {
	order := g.TaobaoGenerator.GenerateOrder()
	order.PlatformOrderID = fmt.Sprintf("JD%d%010d", time.Now().Year(), rand.Intn(1000000000))

	// 京东特有字段
	for _, item := range order.Items {
		item.SKUID = fmt.Sprintf("JD_SKU_%d", rand.Intn(1000000))
		item.ProductID = fmt.Sprintf("JD_PID_%d", rand.Intn(100000))
	}
	order.BuyerInfo.BuyerID = fmt.Sprintf("jd_user_%d", rand.Intn(100000))

	return order
}

// ==================== 抖音生成器 ====================

type DouyinGenerator struct {
	TaobaoGenerator
}

func NewDouyinGenerator() *DouyinGenerator {
	g := &DouyinGenerator{}
	g.statuses = []string{"10", "20", "30", "40", "50"} // 抖音订单状态是数字
	g.products = []productInfo{
		{"网红零食大礼包", 59.90, "规格:1kg装"},
		{"抖音同款小风扇", 29.90, "颜色:粉色"},
		{"手机支架直播神器", 39.90, "颜色:黑色"},
		{"美妆蛋套装", 19.90, "规格:6个装"},
		{"洁面乳氨基酸", 49.90, "规格:100g"},
		{"防晒喷雾防晒霜", 69.90, "规格:150ml"},
		{"面膜补水保湿", 39.90, "规格:10片装"},
		{"指甲油套装", 29.90, "规格:12色"},
		{"洗脸巾一次性", 19.90, "规格:200抽"},
		{"口红套装礼盒", 99.90, "规格:6支装"},
	}
	g.provinces = []string{"广东省", "浙江省", "江苏省", "河南省", "山东省"}
	g.cities = map[string][]string{
		"广东省": {"广州市", "深圳市", "佛山市"},
		"浙江省": {"杭州市", "宁波市", "温州市"},
		"江苏省": {"南京市", "苏州市", "无锡市"},
		"河南省": {"郑州市", "洛阳市", "开封市"},
		"山东省": {"济南市", "青岛市", "烟台市"},
	}
	return g
}

func (g *DouyinGenerator) Platform() string {
	return "douyin"
}

func (g *DouyinGenerator) GenerateOrder() *OrderData {
	order := g.TaobaoGenerator.GenerateOrder()
	order.PlatformOrderID = fmt.Sprintf("DY%d%012d", time.Now().Year(), rand.Intn(100000000000))

	for _, item := range order.Items {
		item.SKUID = fmt.Sprintf("DY_SKU_%d", rand.Intn(1000000))
		item.ProductID = fmt.Sprintf("DY_PID_%d", rand.Intn(100000))
	}
	order.BuyerInfo.BuyerID = fmt.Sprintf("douyin_user_%d", rand.Intn(100000))
	order.BuyerInfo.BuyerNick = fmt.Sprintf("抖音用户%d", rand.Intn(10000))

	return order
}

// ==================== 快手生成器 ====================

type KuaishouGenerator struct {
	TaobaoGenerator
}

func NewKuaishouGenerator() *KuaishouGenerator {
	g := &KuaishouGenerator{}
	g.statuses = []string{"10", "20", "30", "40", "50"}
	g.products = []productInfo{
		{"快手同款手机壳", 19.90, "型号:通用"},
		{"网红同款帽子", 29.90, "颜色:黑色"},
		{"直播补光灯", 89.90, "规格:10寸"},
		{"声卡套装直播", 199.90, "型号:标准版"},
		{"数据线快充", 15.90, "长度:1米"},
		{"充电器快充头", 39.90, "功率:65W"},
		{"手机稳定器", 299.90, "型号:Pro版"},
		{"蓝牙音箱迷你", 49.90, "颜色:黑色"},
		{"自拍杆三脚架", 29.90, "长度:1.5米"},
		{"直播间背景布", 59.90, "尺寸:2x3米"},
	}
	g.provinces = []string{"河北省", "山东省", "河南省", "安徽省", "四川省"}
	g.cities = map[string][]string{
		"河北省": {"石家庄市", "保定市", "邯郸市"},
		"山东省": {"济南市", "青岛市", "临沂市"},
		"河南省": {"郑州市", "洛阳市", "南阳市"},
		"安徽省": {"合肥市", "芜湖市", "蚌埠市"},
		"四川省": {"成都市", "绵阳市", "德阳市"},
	}
	return g
}

func (g *KuaishouGenerator) Platform() string {
	return "kuaishou"
}

func (g *KuaishouGenerator) GenerateOrder() *OrderData {
	order := g.TaobaoGenerator.GenerateOrder()
	order.PlatformOrderID = fmt.Sprintf("KS%d%012d", time.Now().Year(), rand.Intn(100000000000))

	for _, item := range order.Items {
		item.SKUID = fmt.Sprintf("KS_SKU_%d", rand.Intn(1000000))
		item.ProductID = fmt.Sprintf("KS_PID_%d", rand.Intn(100000))
	}
	order.BuyerInfo.BuyerID = fmt.Sprintf("kuaishou_user_%d", rand.Intn(100000))
	order.BuyerInfo.BuyerNick = fmt.Sprintf("快手用户%d", rand.Intn(10000))

	return order
}

// 辅助函数：去除前缀空格
func init() {
	rand.Seed(time.Now().UnixNano())
}

// 随机字符串
func randomString(prefix string, length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return prefix + string(b)
}

// 随机选择
func randomChoice[T any](slice []T) T {
	return slice[rand.Intn(len(slice))]
}

// 随机中文姓名
func randomChineseName() string {
	surnames := []string{"张", "王", "李", "赵", "刘", "陈", "杨", "黄", "周", "吴"}
	names := []string{"伟", "芳", "娜", "秀英", "敏", "静", "丽", "强", "磊", "洋"}
	return surnames[rand.Intn(len(surnames))] + names[rand.Intn(len(names))]
}

// 随机手机号
func randomPhone() string {
	prefixes := []string{"138", "139", "136", "137", "135", "158", "159", "188", "189", "186"}
	return prefixes[rand.Intn(len(prefixes))] + fmt.Sprintf("%08d", rand.Intn(100000000))
}

// 随机地址
func randomAddress() string {
	return fmt.Sprintf("某路%d号某小区%d栋%d单元%d室",
		rand.Intn(999)+1,
		rand.Intn(20)+1,
		rand.Intn(4)+1,
		rand.Intn(30)+1,
	)
}

// 使用strings包的TrimSpace
var _ = strings.TrimSpace("")
