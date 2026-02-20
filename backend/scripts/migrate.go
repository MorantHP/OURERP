package scripts

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// CreateIndexes 创建数据库索引
func CreateIndexes(db *gorm.DB) error {
	log.Println("开始创建数据库索引...")

	// 用户表索引
	indexes := []struct {
		table  string
		name   string
		columns string
	}{
		// 用户表
		{"users", "idx_users_email", "email"},
		{"users", "idx_users_status", "status"},

		// 租户表
		{"tenants", "idx_tenants_code", "code"},
		{"tenants", "idx_tenants_owner_id", "owner_id"},
		{"tenants", "idx_tenants_status", "status"},

		// 租户用户关联表
		{"tenant_users", "idx_tenant_users_tenant_id", "tenant_id"},
		{"tenant_users", "idx_tenant_users_user_id", "user_id"},
		{"idx_tenant_users_unique", "tenant_users", "tenant_id, user_id"},

		// 店铺表
		{"shops", "idx_shops_tenant_id", "tenant_id"},
		{"shops", "idx_shops_platform", "platform"},
		{"shops", "idx_shops_status", "status"},

		// 商品表
		{"products", "idx_products_tenant_id", "tenant_id"},
		{"products", "idx_products_sku_code", "sku_code"},
		{"products", "idx_products_category", "category"},
		{"products", "idx_products_brand", "brand"},
		{"products", "idx_products_status", "status"},

		// 仓库表
		{"warehouses", "idx_warehouses_tenant_id", "tenant_id"},
		{"warehouses", "idx_warehouses_code", "code"},

		// 库存表
		{"inventories", "idx_inventories_tenant_id", "tenant_id"},
		{"inventories", "idx_inventories_product_id", "product_id"},
		{"inventories", "idx_inventories_warehouse_id", "warehouse_id"},
		{"idx_inventories_product_warehouse", "inventories", "product_id, warehouse_id"},

		// 库存日志表
		{"inventory_logs", "idx_inventory_logs_tenant_id", "tenant_id"},
		{"inventory_logs", "idx_inventory_logs_product_id", "product_id"},
		{"inventory_logs", "idx_inventory_logs_created_at", "created_at"},

		// 订单表
		{"orders", "idx_orders_tenant_id", "tenant_id"},
		{"orders", "idx_orders_order_no", "order_no"},
		{"orders", "idx_orders_platform", "platform"},
		{"orders", "idx_orders_shop_id", "shop_id"},
		{"orders", "idx_orders_status", "status"},
		{"orders", "idx_orders_created_at", "created_at"},
		{"orders", "idx_orders_paid_at", "paid_at"},

		// 订单商品表
		{"order_items", "idx_order_items_order_id", "order_id"},
		{"order_items", "idx_order_items_sku_id", "sku_id"},

		// 客户表
		{"customers", "idx_customers_tenant_id", "tenant_id"},
		{"customers", "idx_customers_phone", "phone"},
		{"customers", "idx_customers_level", "level"},
		{"customers", "idx_customers_province", "province"},
		{"customers", "idx_customers_city", "city"},

		// 预警规则表
		{"alert_rules", "idx_alert_rules_tenant_id", "tenant_id"},
		{"alert_rules", "idx_alert_rules_type", "type"},
		{"alert_rules", "idx_alert_rules_status", "status"},

		// 预警记录表
		{"alert_records", "idx_alert_records_tenant_id", "tenant_id"},
		{"alert_records", "idx_alert_records_rule_id", "rule_id"},
		{"alert_records", "idx_alert_records_status", "status"},
		{"alert_records", "idx_alert_records_created_at", "created_at"},

		// 财务记录表
		{"finance_records", "idx_finance_records_tenant_id", "tenant_id"},
		{"finance_records", "idx_finance_records_type", "type"},
		{"finance_records", "idx_finance_records_record_date", "record_date"},
		{"finance_records", "idx_finance_records_status", "status"},

		// 供应商表
		{"suppliers", "idx_suppliers_tenant_id", "tenant_id"},
		{"suppliers", "idx_suppliers_code", "code"},

		// 商品成本表
		{"product_costs", "idx_product_costs_tenant_id", "tenant_id"},
		{"product_costs", "idx_product_costs_product_id", "product_id"},

		// 订单成本表
		{"order_costs", "idx_order_costs_tenant_id", "tenant_id"},
		{"order_costs", "idx_order_costs_order_id", "order_id"},

		// 用户角色关联表
		{"user_roles", "idx_user_roles_user_id", "user_id"},
		{"user_roles", "idx_user_roles_role_id", "role_id"},
		{"user_roles", "idx_user_roles_tenant_id", "tenant_id"},

		// 用户资源权限表
		{"user_resource_permissions", "idx_user_resource_permissions_user_id", "user_id"},
		{"user_resource_permissions", "idx_user_resource_permissions_tenant_id", "tenant_id"},
	}

	for _, idx := range indexes {
		sql := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", idx.name, idx.table, idx.columns)
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("创建索引失败 %s: %v", idx.name, err)
			continue
		}
		log.Printf("创建索引成功: %s", idx.name)
	}

	log.Println("数据库索引创建完成")
	return nil
}

// AnalyzeTables 分析表统计信息
func AnalyzeTables(db *gorm.DB) error {
	tables := []string{
		"users", "tenants", "tenant_users", "shops",
		"products", "warehouses", "inventories", "inventory_logs",
		"orders", "order_items", "customers",
		"alert_rules", "alert_records",
		"finance_records", "suppliers", "product_costs", "order_costs",
		"user_roles", "user_resource_permissions",
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("ANALYZE %s", table)).Error; err != nil {
			log.Printf("分析表失败 %s: %v", table, err)
		}
	}

	return nil
}

// DropIndexes 删除索引（谨慎使用）
func DropIndexes(db *gorm.DB, indexNames []string) error {
	for _, name := range indexNames {
		sql := fmt.Sprintf("DROP INDEX IF EXISTS %s", name)
		if err := db.Exec(sql).Error; err != nil {
			return fmt.Errorf("删除索引失败 %s: %w", name, err)
		}
	}
	return nil
}
