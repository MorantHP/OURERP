package seed

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"gorm.io/gorm"
)

type Seeder struct {
	db *gorm.DB
	// repositories
	userRepo     *repository.UserRepository
	tenantRepo   *repository.TenantRepository
	shopRepo     *repository.ShopRepository
	productRepo  *repository.ProductRepository
	warehouseRepo *repository.WarehouseRepository
	inventoryRepo *repository.InventoryRepository
	orderRepo    *repository.OrderRepository
	financeRepo  *repository.FinanceRepository
}

// clearData 清空所有表数据（保留root用户）
func (s *Seeder) clearData() {
	fmt.Println("清空现有数据...")

	// 按依赖顺序删除（先删除外键依赖的表）
	s.db.Exec("TRUNCATE TABLE alert_records RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE alert_rules RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE report_templates RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE customer_analyses RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE product_analyses RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE region_analyses RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE compare_analyses RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE realtime_snapshots RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE dashboard_widgets RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE customers RESTART IDENTITY CASCADE")

	// 财务相关
	s.db.Exec("TRUNCATE TABLE finance_records RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE platform_bill_details RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE platform_bills RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE purchase_payments RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE purchase_settlement_details RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE purchase_settlements RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE product_costs RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE order_costs RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE inventory_cost_snapshots RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE financial_settlements RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE finance_bank_accounts RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE suppliers RESTART IDENTITY CASCADE")

	// 订单相关
	s.db.Exec("TRUNCATE TABLE order_items RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE orders RESTART IDENTITY CASCADE")

	// 库存相关
	s.db.Exec("TRUNCATE TABLE inventory_logs RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE inventories RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE inbound_items RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE inbound_orders RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE outbound_items RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE outbound_orders RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE stocktake_items RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE stocktakes RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE transfer_items RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE transfer_orders RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE products RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE warehouses RESTART IDENTITY CASCADE")

	// 权限相关
	s.db.Exec("TRUNCATE TABLE user_resource_permissions RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE user_roles RESTART IDENTITY CASCADE")

	// 店铺和租户
	s.db.Exec("TRUNCATE TABLE shops RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE tenant_users RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE tenants RESTART IDENTITY CASCADE")

	// 删除非root用户
	s.db.Exec("DELETE FROM users WHERE is_root = false")

	fmt.Println("数据清空完成")
}

func NewSeeder(db *gorm.DB) *Seeder {
	return &Seeder{
		db:            db,
		userRepo:      repository.NewUserRepository(db),
		tenantRepo:    repository.NewTenantRepository(db),
		shopRepo:      repository.NewShopRepository(db),
		productRepo:   repository.NewProductRepository(db),
		warehouseRepo: repository.NewWarehouseRepository(db),
		inventoryRepo: repository.NewInventoryRepository(db),
		orderRepo:     repository.NewOrderRepository(db),
		financeRepo:   repository.NewFinanceRepository(db),
	}
}

// SeedAll 生成所有演示数据
func (s *Seeder) SeedAll() error {
	// 检查是否已有演示租户
	var tenantCount int64
	s.db.Model(&models.Tenant{}).Where("code = ?", "DEMO001").Count(&tenantCount)
	if tenantCount > 0 {
		fmt.Println("演示数据已存在，跳过生成")
		return nil
	}

	fmt.Println("开始生成演示数据...")

	// 清空现有数据（按依赖顺序删除）
	s.clearData()

	// 1. 创建演示用户

	// 1. 创建演示用户
	users := s.seedUsers()
	fmt.Printf("✓ 创建 %d 个演示用户\n", len(users))

	// 2. 创建演示租户
	tenant := s.seedTenant(users[0])
	fmt.Printf("✓ 创建演示租户\n")

	// 3. 创建演示店铺
	shops := s.seedShops(tenant.ID)
	fmt.Printf("✓ 创建 %d 个演示店铺\n", len(shops))

	// 4. 创建演示仓库
	warehouses := s.seedWarehouses(tenant.ID)
	fmt.Printf("✓ 创建 %d 个演示仓库\n", len(warehouses))

	// 5. 创建演示商品
	products := s.seedProducts(tenant.ID)
	fmt.Printf("✓ 创建 %d 个演示商品\n", len(products))

	// 6. 创建演示库存
	s.seedInventory(tenant.ID, products, warehouses)
	fmt.Printf("✓ 创建演示库存数据\n")

	// 7. 创建演示客户
	customers := s.seedCustomers(tenant.ID)
	fmt.Printf("✓ 创建 %d 个演示客户\n", len(customers))

	// 8. 创建演示订单
	orders := s.seedOrders(tenant.ID, shops, products, customers)
	fmt.Printf("✓ 创建 %d 个演示订单\n", len(orders))

	// 9. 创建演示供应商
	suppliers := s.seedSuppliers(tenant.ID)
	fmt.Printf("✓ 创建 %d 个演示供应商\n", len(suppliers))

	// 10. 创建演示财务记录
	s.seedFinanceRecords(tenant.ID, shops, suppliers)
	fmt.Printf("✓ 创建演示财务记录\n")

	// 11. 创建预警规则
	s.seedAlertRules(tenant.ID)
	fmt.Printf("✓ 创建演示预警规则\n")

	// 12. 创建租户用户关联
	s.seedTenantUsers(tenant, users)
	fmt.Printf("✓ 创建租户用户关联\n")

	fmt.Println("\n========================================")
	fmt.Println("演示数据生成完成!")
	fmt.Println("========================================")
	fmt.Println("\n登录账号:")
	fmt.Println("  Root: root@ourerp.com / root123456")
	fmt.Println("  管理员: admin@demo.com / demo123456")
	fmt.Println("  操作员: operator@demo.com / demo123456")
	fmt.Println("\n租户编码: DEMO001")

	return nil
}

// seedUsers 创建演示用户
func (s *Seeder) seedUsers() []*models.User {
	users := []struct {
		email    string
		name     string
		phone    string
		isRoot   bool
		approved bool
	}{
		{"admin@demo.com", "演示管理员", "13800000001", false, true},
		{"operator@demo.com", "演示操作员", "13800000002", false, true},
		{"finance@demo.com", "财务小王", "13800000003", false, true},
		{"warehouse@demo.com", "仓库小李", "13800000004", false, true},
		{"viewer@demo.com", "访客用户", "13800000005", false, true},
		{"pending@demo.com", "待审核用户", "13800000006", false, false},
	}

	var result []*models.User
	for _, u := range users {
		user := &models.User{
			Email:      u.email,
			Name:       u.name,
			Phone:      u.phone,
			IsRoot:     u.isRoot,
			IsApproved: u.approved,
			Status:     1,
		}
		user.SetPassword("demo123456")
		s.userRepo.Create(user)
		result = append(result, user)
	}
	return result
}

// seedTenant 创建演示租户
func (s *Seeder) seedTenant(owner *models.User) *models.Tenant {
	tenant := &models.Tenant{
		Code:        "DEMO001",
		Name:        "演示公司",
		Platform:    "multi",
		Description: "这是一个演示租户，包含完整的演示数据",
		Status:      1,
		OwnerID:     owner.ID,
	}
	s.tenantRepo.Create(tenant)
	return tenant
}

// seedShops 创建演示店铺
func (s *Seeder) seedShops(tenantID int64) []*models.Shop {
	shops := []struct {
		name     string
		platform string
		shopID   string
	}{
		{"天猫旗舰店", "taobao", "TM001"},
		{"淘宝C店", "taobao", "TB001"},
		{"抖音小店", "douyin", "DY001"},
		{"京东自营店", "jd", "JD001"},
		{"拼多多店铺", "pdd", "PDD001"},
	}

	var result []*models.Shop
	for _, shopData := range shops {
		shop := &models.Shop{
			TenantID:       tenantID,
			Name:           shopData.name,
			Platform:       shopData.platform,
			PlatformShopID: shopData.shopID,
			Status:         1,
			SyncInterval:   30,
		}
		s.db.Create(shop)
		result = append(result, shop)
	}
	return result
}

// seedWarehouses 创建演示仓库
func (s *Seeder) seedWarehouses(tenantID int64) []*models.Warehouse {
	warehouses := []struct {
		code    string
		name    string
		address string
		isDefault bool
	}{
		{"WH001", "主仓库", "浙江省杭州市余杭区仓前街道", true},
		{"WH002", "北京分仓", "北京市朝阳区望京街道", false},
		{"WH003", "广州分仓", "广东省广州市白云区太和镇", false},
		{"WH004", "保税仓", "上海自贸区", false},
	}

	var result []*models.Warehouse
	for _, wh := range warehouses {
		w := &models.Warehouse{
			TenantID:  tenantID,
			Code:      wh.code,
			Name:      wh.name,
			Address:   wh.address,
			Contact:   "仓库管理员",
			Phone:     "0571-88888888",
			Type:      "normal",
			Status:    1,
			IsDefault: wh.isDefault,
		}
		s.db.Create(w)
		result = append(result, w)
	}
	return result
}

// seedProducts 创建演示商品
func (s *Seeder) seedProducts(tenantID int64) []*models.Product {
	products := []struct {
		skuCode   string
		name      string
		category  string
		brand     string
		costPrice float64
		salePrice float64
		barcode   string
	}{
		// 手机类目
		{"SKU001", "iPhone 15 Pro Max 256G 原色钛金属", "手机", "Apple", 8999, 9999, "6901234567001"},
		{"SKU002", "iPhone 15 Pro 128G 蓝色钛金属", "手机", "Apple", 6999, 7999, "6901234567002"},
		{"SKU003", "iPhone 15 128G 黑色", "手机", "Apple", 4999, 5999, "6901234567003"},
		{"SKU004", "华为 Mate60 Pro 12+512G 雅丹黑", "手机", "华为", 5999, 6999, "6901234567004"},
		{"SKU005", "小米14 Pro 16+512G 黑色", "手机", "小米", 3999, 4999, "6901234567005"},
		{"SKU006", "OPPO Find X6 Pro 12+256G 云白", "手机", "OPPO", 4499, 5499, "6901234567006"},

		// 电脑类目
		{"SKU007", "MacBook Pro 14寸 M3 Pro 18+512G", "电脑", "Apple", 12999, 14999, "6901234567007"},
		{"SKU008", "MacBook Air 15寸 M3 8+256G", "电脑", "Apple", 8999, 10499, "6901234567008"},
		{"SKU009", "ThinkPad X1 Carbon Gen11 i7", "电脑", "Lenovo", 9999, 12999, "6901234567009"},
		{"SKU010", "华为 MateBook X Pro 2024款", "电脑", "华为", 7999, 9999, "6901234567010"},

		// 平板类目
		{"SKU011", "iPad Pro 12.9寸 M2 256G", "平板", "Apple", 7999, 9299, "6901234567011"},
		{"SKU012", "iPad Air 5 64G WiFi版", "平板", "Apple", 3999, 4799, "6901234567012"},
		{"SKU013", "华为 MatePad Pro 12.6寸", "平板", "华为", 3499, 4299, "6901234567013"},

		// 配件类目
		{"SKU014", "AirPods Pro 2 USB-C版", "配件", "Apple", 1499, 1899, "6901234567014"},
		{"SKU015", "AirPods Max 银色", "配件", "Apple", 3999, 4399, "6901234567015"},
		{"SKU016", "Apple Watch Ultra 2", "配件", "Apple", 5999, 6499, "6901234567016"},
		{"SKU017", "Apple Watch S9 45mm GPS", "配件", "Apple", 2999, 3499, "6901234567017"},
		{"SKU018", "MagSafe充电器", "配件", "Apple", 299, 399, "6901234567018"},

		// 家电类目
		{"SKU019", "戴森 V15 Detect吸尘器", "家电", "戴森", 4999, 5999, "6901234567019"},
		{"SKU020", "戴森 Supersonic吹风机 HD08", "家电", "戴森", 2999, 3599, "6901234567020"},

		// 音频类目
		{"SKU021", "索尼 WH-1000XM5 黑色", "耳机", "索尼", 2299, 2999, "6901234567021"},
		{"SKU022", "Bose QuietComfort 45", "耳机", "Bose", 1999, 2499, "6901234567022"},
		{"SKU023", "JBL Flip 6 便携音箱", "音箱", "JBL", 799, 999, "6901234567023"},

		// 游戏类目
		{"SKU024", "任天堂 Switch OLED 白色", "游戏", "任天堂", 2099, 2599, "6901234567024"},
		{"SKU025", "PlayStation 5 光驱版", "游戏", "Sony", 3599, 4299, "6901234567025"},
		{"SKU026", "Xbox Series X 1TB", "游戏", "Microsoft", 3499, 4199, "6901234567026"},

		// 摄影类目
		{"SKU027", "索尼 A7M4 全画幅微单", "相机", "索尼", 15999, 18999, "6901234567027"},
		{"SKU028", "佳能 EOS R6 Mark II", "相机", "佳能", 14999, 17999, "6901234567028"},
		{"SKU029", "大疆 DJI Mini 4 Pro", "无人机", "大疆", 5999, 7199, "6901234567029"},
		{"SKU030", "Insta360 X4 全景相机", "相机", "Insta360", 3999, 4999, "6901234567030"},
	}

	var result []*models.Product
	for _, p := range products {
		product := &models.Product{
			TenantID:  tenantID,
			SkuCode:   p.skuCode,
			Name:      p.name,
			Category:  p.category,
			Brand:     p.brand,
			CostPrice: p.costPrice,
			SalePrice: p.salePrice,
			Barcode:   p.barcode,
			Unit:      "件",
			Status:    1,
		}
		s.db.Create(product)
		result = append(result, product)
	}
	return result
}

// seedInventory 创建演示库存
func (s *Seeder) seedInventory(tenantID int64, products []*models.Product, warehouses []*models.Warehouse) {
	rand.Seed(time.Now().UnixNano())

	for _, product := range products {
		for _, wh := range warehouses {
			// 每个商品在每个仓库都有库存，数量随机
			qty := rand.Intn(500) + 10
			lockedQty := rand.Intn(20)
			alertQty := rand.Intn(50) + 10

			inventory := &models.Inventory{
				TenantID:    tenantID,
				ProductID:   product.ID,
				WarehouseID: wh.ID,
				Quantity:    qty,
				LockedQty:   lockedQty,
				TotalQty:    qty + lockedQty,
				AlertQty:    alertQty,
				Location:    fmt.Sprintf("A-%d-%d", rand.Intn(10)+1, rand.Intn(20)+1),
			}
			s.db.Create(inventory)
		}
	}
}

// seedCustomers 创建演示客户
func (s *Seeder) seedCustomers(tenantID int64) []*models.Customer {
	customers := []struct {
		name     string
		phone    string
		level    string
		province string
		city     string
	}{
		{"张三", "13800138001", "vip", "浙江省", "杭州市"},
		{"李四", "13800138002", "vip", "广东省", "深圳市"},
		{"王五", "13800138003", "normal", "北京市", "北京市"},
		{"赵六", "13800138004", "normal", "上海市", "上海市"},
		{"钱七", "13800138005", "normal", "江苏省", "南京市"},
		{"孙八", "13800138006", "new", "四川省", "成都市"},
		{"周九", "13800138007", "new", "湖北省", "武汉市"},
		{"吴十", "13800138008", "vip", "陕西省", "西安市"},
		{"郑十一", "13800138009", "normal", "山东省", "济南市"},
		{"王十二", "13800138010", "new", "福建省", "厦门市"},
		{"刘十三", "13800138011", "vip", "河南省", "郑州市"},
		{"陈十四", "13800138012", "normal", "湖南省", "长沙市"},
		{"杨十五", "13800138013", "new", "安徽省", "合肥市"},
		{"黄十六", "13800138014", "vip", "重庆市", "重庆市"},
		{"林十七", "13800138015", "normal", "云南省", "昆明市"},
		{"何十八", "13800138016", "new", "辽宁省", "沈阳市"},
		{"马十九", "13800138017", "vip", "天津市", "天津市"},
		{"高二十", "13800138018", "normal", "河北省", "石家庄市"},
		{"罗二一", "13800138019", "new", "江西省", "南昌市"},
		{"梁二二", "13800138020", "vip", "广西", "南宁市"},
	}

	var result []*models.Customer
	rand.Seed(time.Now().UnixNano())

	for i, c := range customers {
		totalOrders := rand.Intn(50) + 1
		totalAmount := float64(rand.Intn(50000)+1000)

		customer := &models.Customer{
			TenantID:    tenantID,
			Code:        fmt.Sprintf("C%06d", i+1),
			Name:        c.name,
			Phone:       c.phone,
			Type:        "b2c",
			Level:       c.level,
			Source:      "online",
			Province:    c.province,
			City:        c.city,
			Address:     "某某街道某某小区",
			TotalOrders: totalOrders,
			TotalAmount: totalAmount,
			Status:      1,
		}

		// 设置首单和末单时间
		firstOrder := time.Now().AddDate(0, -rand.Intn(12), -rand.Intn(30))
		lastOrder := time.Now().AddDate(0, 0, -rand.Intn(30))
		customer.FirstOrderAt = &firstOrder
		customer.LastOrderAt = &lastOrder

		s.db.Create(customer)
		result = append(result, customer)
	}
	return result
}

// seedOrders 创建演示订单
func (s *Seeder) seedOrders(tenantID int64, shops []*models.Shop, products []*models.Product, customers []*models.Customer) []*models.Order {
	var orders []*models.Order
	rand.Seed(time.Now().UnixNano())

	statuses := []int{100, 200, 300, 400, 500, 600, 999}
	statusWeights := []int{5, 10, 20, 25, 20, 15, 5} // 权重

	logistics := []struct {
		company string
		code    string
	}{
		{"顺丰速运", "SF1234567890"},
		{"中通快递", "ZT9876543210"},
		{"圆通速递", "YT1122334455"},
		{"韵达速递", "YD5566778899"},
		{"极兔速递", "JT9988776655"},
	}

	// 生成最近90天的订单
	for day := 0; day < 90; day++ {
		orderDate := time.Now().AddDate(0, 0, -day)

		// 每天随机生成5-20个订单
		dailyOrders := rand.Intn(16) + 5

		for i := 0; i < dailyOrders; i++ {
			// 随机选择店铺
			shop := shops[rand.Intn(len(shops))]

			// 随机选择状态（按权重）
			status := weightedRandom(statuses, statusWeights)

			// 随机选择客户
			customer := customers[rand.Intn(len(customers))]

			// 生成订单号
			orderNo := fmt.Sprintf("%s%d%04d", shop.Platform, orderDate.Format("20060102"), rand.Intn(10000))

			// 创建订单
			order := &models.Order{
				TenantID:        tenantID,
				OrderNo:         orderNo,
				Platform:        shop.Platform,
				PlatformOrderID: fmt.Sprintf("PO%d", rand.Int63n(1000000000)),
				ShopID:          shop.ID,
				Status:          status,
				BuyerNick:       fmt.Sprintf("买家%d", rand.Intn(10000)),
				ReceiverName:    customer.Name,
				ReceiverPhone:   customer.Phone,
				ReceiverAddress: fmt.Sprintf("%s %s %s", customer.Province, customer.City, "某某街道"),
			}

			// 根据状态设置时间和金额
			order.CreatedAt = orderDate

			if status >= 200 {
				// 已付款
				order.PaidAt = ptrTime(orderDate.Add(time.Duration(rand.Intn(60)) * time.Minute))
				order.PayAmount = float64(rand.Intn(5000)+100) + float64(rand.Intn(100))/100
				order.TotalAmount = order.PayAmount + float64(rand.Intn(50))
			}

			if status >= 400 {
				// 已发货
				shipInfo := logistics[rand.Intn(len(logistics))]
				order.LogisticsCompany = shipInfo.company
				order.LogisticsNo = shipInfo.code + fmt.Sprintf("%d", rand.Intn(1000))
				order.ShippedAt = ptrTime(orderDate.Add(time.Duration(rand.Intn(48)+12) * time.Hour))
			}

			s.db.Create(order)

			// 创建订单商品
			itemCount := rand.Intn(3) + 1
			for j := 0; j < itemCount; j++ {
				product := products[rand.Intn(len(products))]
				qty := rand.Intn(3) + 1

				item := &models.OrderItem{
					TenantID: tenantID,
					OrderID:  order.ID,
					SkuID:    product.ID,
					SkuName:  product.Name,
					Quantity: qty,
					Price:    product.SalePrice,
				}
				s.db.Create(item)
			}

			orders = append(orders, order)
		}
	}

	return orders
}

// seedSuppliers 创建演示供应商
func (s *Seeder) seedSuppliers(tenantID int64) []*models.Supplier {
	suppliers := []struct {
		code    string
		name    string
		contact string
		phone   string
		address string
	}{
		{"SUP001", "深圳华强电子有限公司", "张经理", "0755-88888881", "深圳市福田区华强北路"},
		{"SUP002", "广州数码科技有限公司", "李经理", "020-88888882", "广州市天河区天河路"},
		{"SUP003", "北京创新科技有限公司", "王经理", "010-88888883", "北京市海淀区中关村"},
		{"SUP004", "上海智能设备有限公司", "赵经理", "021-88888884", "上海市浦东新区张江"},
		{"SUP005", "杭州电子商务有限公司", "钱经理", "0571-88888885", "杭州市滨江区网商路"},
		{"SUP006", "成都西部数码有限公司", "孙经理", "028-88888886", "成都市武侯区天府大道"},
		{"SUP007", "武汉光谷电子有限公司", "周经理", "027-88888887", "武汉市洪山区光谷大道"},
		{"SUP008", "南京科技贸易有限公司", "吴经理", "025-88888888", "南京市玄武区长江路"},
	}

	var result []*models.Supplier
	for _, sup := range suppliers {
		supplier := &models.Supplier{
			TenantID:    tenantID,
			Code:        sup.code,
			Name:        sup.name,
			Contact:     sup.contact,
			Phone:       sup.phone,
			Address:     sup.address,
			CreditLimit: 1000000,
			Balance:     float64(rand.Intn(100000)),
			Status:      1,
		}
		s.db.Create(supplier)
		result = append(result, supplier)
	}
	return result
}

// seedFinanceRecords 创建演示财务记录
func (s *Seeder) seedFinanceRecords(tenantID int64, shops []*models.Shop, suppliers []*models.Supplier) {
	rand.Seed(time.Now().UnixNano())

	// 收入类型
	incomeCategories := []string{"销售收入", "退款收回", "平台补贴", "其他收入"}
	// 支出类型
	expenseCategories := []string{"采购货款", "物流费用", "平台佣金", "推广费用", "退款支出", "其他支出"}

	// 生成最近90天的财务记录
	for day := 0; day < 90; day++ {
		recordDate := time.Now().AddDate(0, 0, -day)

		// 每天生成5-15条记录
		recordCount := rand.Intn(11) + 5

		for i := 0; i < recordCount; i++ {
			var record models.FinanceRecord
			record.TenantID = tenantID
			record.RecordDate = recordDate
			record.Source = "manual"
			record.Status = 1
			record.VoucherNo = fmt.Sprintf("V%s%04d", recordDate.Format("20060102"), rand.Intn(10000))

			// 随机选择收入或支出
			if rand.Intn(100) < 60 {
				// 60% 收入
				record.Type = "income"
				record.Category = incomeCategories[rand.Intn(len(incomeCategories))]
				record.Amount = float64(rand.Intn(10000)+100) + float64(rand.Intn(100))/100
				record.Description = fmt.Sprintf("%s - %s", record.Category, recordDate.Format("2006-01-02"))

				// 随机关联店铺
				if rand.Intn(100) < 70 {
					shop := shops[rand.Intn(len(shops))]
					record.ShopID = &shop.ID
				}
			} else {
				// 40% 支出
				record.Type = "expense"
				record.Category = expenseCategories[rand.Intn(len(expenseCategories))]
				record.Amount = float64(rand.Intn(5000)+50) + float64(rand.Intn(100))/100
				record.Description = fmt.Sprintf("%s - %s", record.Category, recordDate.Format("2006-01-02"))

				// 采购货款关联供应商
				if record.Category == "采购货款" {
					supplier := suppliers[rand.Intn(len(suppliers))]
					record.Description = fmt.Sprintf("采购货款 - %s", supplier.Name)
				}
			}

			s.db.Create(&record)
		}
	}

	// 创建银行账户
	bankAccounts := []struct {
		name        string
		accountNo   string
		bankName    string
		accountType string
	}{
		{"公司基本账户", "6222021234567890001", "工商银行杭州分行", "bank"},
		{"支付宝账户", "finance@demo.com", "支付宝", "alipay"},
		{"微信账户", "wx_demo_company", "微信支付", "wechat"},
	}

	for i, acc := range bankAccounts {
		account := models.FinanceBankAccount{
			TenantID:    tenantID,
			AccountName: acc.name,
			AccountNo:   acc.accountNo,
			BankName:    acc.bankName,
			AccountType: acc.accountType,
			Currency:    "CNY",
			Balance:     float64(rand.Intn(1000000) + 10000),
			Status:      1,
			IsDefault:   i == 0,
		}
		s.db.Create(&account)
	}

	// 创建商品成本
	var products []models.Product
	s.db.Where("tenant_id = ?", tenantID).Find(&products)

	for _, product := range products {
		shippingCost := float64(rand.Intn(20) + 5)
		packageCost := float64(rand.Intn(5) + 1)
		otherCost := float64(rand.Intn(10))

		cost := models.ProductCost{
			TenantID:      tenantID,
			ProductID:     product.ID,
			ProductSku:    product.SkuCode,
			PurchaseCost:  product.CostPrice,
			ShippingCost:  shippingCost,
			PackageCost:   packageCost,
			OtherCost:     otherCost,
			TotalCost:     product.CostPrice + shippingCost + packageCost + otherCost,
			CostMethod:    "weighted",
			EffectiveDate: time.Now(),
		}
		s.db.Create(&cost)
	}
}

// seedAlertRules 创建演示预警规则
func (s *Seeder) seedAlertRules(tenantID int64) {
	rules := []struct {
		name      string
		alertType string
		condition string
		threshold float64
		level     string
	}{
		{"库存不足预警", "inventory", `{"field": "quantity", "operator": "<"}`, 20, "warning"},
		{"库存严重不足", "inventory", `{"field": "quantity", "operator": "<"}`, 5, "critical"},
		{"订单超时未发货", "order", `{"field": "pending_hours", "operator": ">"}`, 48, "warning"},
		{"大额订单提醒", "order", `{"field": "amount", "operator": ">"}`, 10000, "info"},
		{"日销售额下降", "sales", `{"field": "drop_rate", "operator": ">"}`, 30, "warning"},
		{"退货率异常", "sales", `{"field": "refund_rate", "operator": ">"}`, 15, "critical"},
		{"新客户流失预警", "customer", `{"field": "churn_days", "operator": ">"}`, 30, "warning"},
	}

	for _, r := range rules {
		rule := models.AlertRule{
			TenantID:    tenantID,
			Name:        r.name,
			Type:        r.alertType,
			Condition:   r.condition,
			Threshold:   r.threshold,
			NotifyType:  "system",
			Level:       r.level,
			Status:      1,
			Description: r.name,
		}
		s.db.Create(&rule)
	}

	// 创建一些预警记录
	var rule models.AlertRule
	s.db.Where("tenant_id = ? AND name = ?", tenantID, "库存不足预警").First(&rule)

	for i := 0; i < 5; i++ {
		record := models.AlertRecord{
			TenantID:   tenantID,
			RuleID:     rule.ID,
			Title:      fmt.Sprintf("库存预警 #%d", i+1),
			Content:    fmt.Sprintf("商品 SKU00%d 库存不足，当前库存: %d", rand.Intn(30)+1, rand.Intn(20)),
			Level:      "warning",
			SourceType: "inventory",
			Status:     rand.Intn(2), // 0-未处理 1-已处理
		}
		s.db.Create(&record)
	}
}

// seedTenantUsers 创建租户用户关联
func (s *Seeder) seedTenantUsers(tenant *models.Tenant, users []*models.User) {
	roles := []string{"admin", "member", "member", "member", "member", "member"}

	for i, user := range users {
		if i >= len(roles) {
			break
		}

		tenantUser := &models.TenantUser{
			TenantID: tenant.ID,
			UserID:   user.ID,
			Role:     roles[i],
		}
		s.db.Create(tenantUser)
	}
}

// 辅助函数：加权随机
func weightedRandom(items []int, weights []int) int {
	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}

	r := rand.Intn(totalWeight)
	for i, w := range weights {
		r -= w
		if r <= 0 {
			return items[i]
		}
	}
	return items[len(items)-1]
}

// 辅助函数：创建时间指针
func ptrTime(t time.Time) *time.Time {
	return &t
}
