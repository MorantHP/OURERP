package main

import (
	"fmt"
	"log"

	"github.com/MorantHP/OURERP/internal/config"
	"github.com/MorantHP/OURERP/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg := config.Load()
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s client_encoding=UTF8",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.Port, cfg.Database.SSLMode)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// 创建所有缺失的表
	db.AutoMigrate(
		&models.Product{},
		&models.Inventory{},
		&models.InventoryLog{},
		&models.Warehouse{},
		&models.Customer{},
		&models.AlertRule{},
		&models.AlertRecord{},
		&models.ReportTemplate{},
		&models.RealtimeSnapshot{},
	)
	
	// 添加 tenant_id 到 inventory 表（如果不存在）
	db.Exec("ALTER TABLE inventory ADD COLUMN IF NOT EXISTS tenant_id BIGINT DEFAULT 0")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_inventory_tenant_id ON inventory(tenant_id)")
	
	fmt.Println("Tables created successfully")
}
