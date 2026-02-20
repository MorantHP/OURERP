// +build ignore

package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := "host=localhost user=ourerp password=ourerp_dev_2024 dbname=ourerp port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	fmt.Println("开始修复数据库Schema...")

	// 1. 为shops表添加tenant_id列
	fmt.Println("\n1. 检查shops表...")
	var hasTenantID int
	db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'shops' AND column_name = 'tenant_id'").Scan(&hasTenantID)
	if hasTenantID == 0 {
		fmt.Println("   添加shops.tenant_id列...")
		if err := db.Exec("ALTER TABLE shops ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 0").Error; err != nil {
			fmt.Println("   警告:", err)
		} else {
			fmt.Println("   ✓ 成功添加tenant_id列")
		}
	} else {
		fmt.Println("   ✓ tenant_id列已存在")
	}

	// 2. 创建warehouses表
	fmt.Println("\n2. 检查warehouses表...")
	var warehousesExist int
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'warehouses'").Scan(&warehousesExist)
	if warehousesExist == 0 {
		fmt.Println("   创建warehouses表...")
		createWarehousesTable := `
		CREATE TABLE warehouses (
			id BIGSERIAL PRIMARY KEY,
			tenant_id BIGINT NOT NULL DEFAULT 0,
			code VARCHAR(50) NOT NULL,
			name VARCHAR(100) NOT NULL,
			address VARCHAR(200),
			contact VARCHAR(50),
			phone VARCHAR(20),
			type VARCHAR(20) DEFAULT 'normal',
			status INTEGER DEFAULT 1,
			is_default BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			UNIQUE(tenant_id, code)
		);
		CREATE INDEX IF NOT EXISTS idx_warehouses_tenant_id ON warehouses(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_warehouses_deleted_at ON warehouses(deleted_at);
		`
		if err := db.Exec(createWarehousesTable).Error; err != nil {
			fmt.Println("   警告:", err)
		} else {
			fmt.Println("   ✓ 成功创建warehouses表")
		}
	} else {
		fmt.Println("   ✓ warehouses表已存在")
	}

	// 3. 创建alert_rules表
	fmt.Println("\n3. 检查alert_rules表...")
	var alertRulesExist int
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'alert_rules'").Scan(&alertRulesExist)
	if alertRulesExist == 0 {
		fmt.Println("   创建alert_rules表...")
		createAlertRulesTable := `
		CREATE TABLE alert_rules (
			id BIGSERIAL PRIMARY KEY,
			tenant_id BIGINT NOT NULL DEFAULT 0,
			name VARCHAR(100) NOT NULL,
			type VARCHAR(20) NOT NULL,
			condition TEXT,
			threshold DOUBLE PRECISION DEFAULT 0,
			notify_type VARCHAR(50),
			notify_target VARCHAR(200),
			level VARCHAR(20) DEFAULT 'warning',
			status INTEGER DEFAULT 1,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_alert_rules_tenant_id ON alert_rules(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_alert_rules_type ON alert_rules(type);
		`
		if err := db.Exec(createAlertRulesTable).Error; err != nil {
			fmt.Println("   警告:", err)
		} else {
			fmt.Println("   ✓ 成功创建alert_rules表")
		}
	} else {
		fmt.Println("   ✓ alert_rules表已存在")
	}

	// 4. 创建alert_records表
	fmt.Println("\n4. 检查alert_records表...")
	var alertRecordsExist int
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'alert_records'").Scan(&alertRecordsExist)
	if alertRecordsExist == 0 {
		fmt.Println("   创建alert_records表...")
		createAlertRecordsTable := `
		CREATE TABLE alert_records (
			id BIGSERIAL PRIMARY KEY,
			tenant_id BIGINT NOT NULL DEFAULT 0,
			rule_id BIGINT,
			title VARCHAR(200),
			content TEXT,
			level VARCHAR(20) DEFAULT 'warning',
			source_type VARCHAR(50),
			source_id BIGINT,
			status INTEGER DEFAULT 0,
			handled_by BIGINT,
			handled_at TIMESTAMP,
			note TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_alert_records_tenant_id ON alert_records(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_alert_records_rule_id ON alert_records(rule_id);
		CREATE INDEX IF NOT EXISTS idx_alert_records_status ON alert_records(status);
		`
		if err := db.Exec(createAlertRecordsTable).Error; err != nil {
			fmt.Println("   警告:", err)
		} else {
			fmt.Println("   ✓ 成功创建alert_records表")
		}
	} else {
		fmt.Println("   ✓ alert_records表已存在")
	}

	// 5. 创建report_templates表
	fmt.Println("\n5. 检查report_templates表...")
	var reportTemplatesExist int
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'report_templates'").Scan(&reportTemplatesExist)
	if reportTemplatesExist == 0 {
		fmt.Println("   创建report_templates表...")
		createReportTemplatesTable := `
		CREATE TABLE report_templates (
			id BIGSERIAL PRIMARY KEY,
			tenant_id BIGINT NOT NULL DEFAULT 0,
			name VARCHAR(100) NOT NULL,
			type VARCHAR(20),
			data_source VARCHAR(50),
			columns TEXT,
			filters TEXT,
			chart_type VARCHAR(20),
			created_by BIGINT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_report_templates_tenant_id ON report_templates(tenant_id);
		`
		if err := db.Exec(createReportTemplatesTable).Error; err != nil {
			fmt.Println("   警告:", err)
		} else {
			fmt.Println("   ✓ 成功创建report_templates表")
		}
	} else {
		fmt.Println("   ✓ report_templates表已存在")
	}

	// 6. 创建customers表
	fmt.Println("\n6. 检查customers表...")
	var customersExist int
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'customers'").Scan(&customersExist)
	if customersExist == 0 {
		fmt.Println("   创建customers表...")
		createCustomersTable := `
		CREATE TABLE customers (
			id BIGSERIAL PRIMARY KEY,
			tenant_id BIGINT NOT NULL DEFAULT 0,
			code VARCHAR(50),
			name VARCHAR(100),
			phone VARCHAR(20),
			email VARCHAR(100),
			type VARCHAR(20),
			level VARCHAR(20),
			province VARCHAR(50),
			city VARCHAR(50),
			address VARCHAR(200),
			total_orders INTEGER DEFAULT 0,
			total_amount DOUBLE PRECISION DEFAULT 0,
			first_order_at TIMESTAMP,
			last_order_at TIMESTAMP,
			tags VARCHAR(200),
			remark VARCHAR(500),
			status INTEGER DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_customers_tenant_id ON customers(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone);
		`
		if err := db.Exec(createCustomersTable).Error; err != nil {
			fmt.Println("   警告:", err)
		} else {
			fmt.Println("   ✓ 成功创建customers表")
		}
	} else {
		fmt.Println("   ✓ customers表已存在")
	}

	// 7. 创建realtime_snapshots表
	fmt.Println("\n7. 检查realtime_snapshots表...")
	var snapshotsExist int
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'realtime_snapshots'").Scan(&snapshotsExist)
	if snapshotsExist == 0 {
		fmt.Println("   创建realtime_snapshots表...")
		createSnapshotsTable := `
		CREATE TABLE realtime_snapshots (
			id BIGSERIAL PRIMARY KEY,
			tenant_id BIGINT NOT NULL DEFAULT 0,
			snapshot_time TIMESTAMP NOT NULL,
			order_count INTEGER DEFAULT 0,
			order_amount DOUBLE PRECISION DEFAULT 0,
			pending_orders INTEGER DEFAULT 0,
			shipped_orders INTEGER DEFAULT 0,
			completed_orders INTEGER DEFAULT 0,
			low_stock_items INTEGER DEFAULT 0,
			new_customers INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_realtime_snapshots_tenant_id ON realtime_snapshots(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_realtime_snapshots_time ON realtime_snapshots(snapshot_time);
		`
		if err := db.Exec(createSnapshotsTable).Error; err != nil {
			fmt.Println("   警告:", err)
		} else {
			fmt.Println("   ✓ 成功创建realtime_snapshots表")
		}
	} else {
		fmt.Println("   ✓ realtime_snapshots表已存在")
	}

	// 8. 为orders表添加tenant_id列（如果不存在）
	fmt.Println("\n8. 检查orders表...")
	var orderHasTenantID int
	db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'orders' AND column_name = 'tenant_id'").Scan(&orderHasTenantID)
	if orderHasTenantID == 0 {
		fmt.Println("   添加orders.tenant_id列...")
		if err := db.Exec("ALTER TABLE orders ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 0").Error; err != nil {
			fmt.Println("   警告:", err)
		} else {
			fmt.Println("   ✓ 成功添加tenant_id列")
		}
	} else {
		fmt.Println("   ✓ tenant_id列已存在")
	}

	// 9. 为order_items表添加tenant_id列（如果不存在）
	fmt.Println("\n9. 检查order_items表...")
	var orderItemHasTenantID int
	db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'order_items' AND column_name = 'tenant_id'").Scan(&orderItemHasTenantID)
	if orderItemHasTenantID == 0 {
		fmt.Println("   添加order_items.tenant_id列...")
		if err := db.Exec("ALTER TABLE order_items ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 0").Error; err != nil {
			fmt.Println("   警告:", err)
		} else {
			fmt.Println("   ✓ 成功添加tenant_id列")
		}
	} else {
		fmt.Println("   ✓ tenant_id列已存在")
	}

	fmt.Println("\n========================================")
	fmt.Println("数据库Schema修复完成!")
	fmt.Println("========================================")
}
