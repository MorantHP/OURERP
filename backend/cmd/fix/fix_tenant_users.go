package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := "host=localhost user=ourerp password=ourerp_dev_2024 dbname=ourerp port=5432 sslmode=disable client_encoding=UTF8"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}

	// 检查tenant_users表是否存在
	var tableExists int
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'tenant_users'").Scan(&tableExists)
	
	if tableExists == 0 {
		fmt.Println("创建tenant_users表...")
		db.Exec(`
			CREATE TABLE tenant_users (
				id BIGSERIAL PRIMARY KEY,
				tenant_id BIGINT NOT NULL,
				user_id BIGINT NOT NULL,
				role VARCHAR(20) NOT NULL DEFAULT 'member',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(tenant_id, user_id)
			)
		`)
	}

	// 检查用户4的关联
	var count int
	db.Raw("SELECT COUNT(*) FROM tenant_users WHERE user_id = 4").Scan(&count)
	fmt.Printf("用户4的租户关联数: %d\n", count)

	if count == 0 {
		fmt.Println("添加用户4的租户关联...")
		db.Exec("INSERT INTO tenant_users (tenant_id, user_id, role, created_at, updated_at) VALUES (1, 4, 'owner', NOW(), NOW())")
		db.Exec("INSERT INTO tenant_users (tenant_id, user_id, role, created_at, updated_at) VALUES (2, 4, 'owner', NOW(), NOW())")
		fmt.Println("已添加!")
	}

	// 验证
	var tus []struct {
		ID       int64
		TenantID int64
		UserID   int64
		Role     string
	}
	db.Raw("SELECT id, tenant_id, user_id, role FROM tenant_users").Scan(&tus)
	fmt.Println("\ntenant_users表内容:")
	for _, tu := range tus {
		fmt.Printf("  TenantID:%d, UserID:%d, Role:%s\n", tu.TenantID, tu.UserID, tu.Role)
	}
}
