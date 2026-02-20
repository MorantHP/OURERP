package services

import (
	"fmt"
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"gorm.io/gorm"
)

// FinanceService 财务服务
type FinanceService struct {
	db   *gorm.DB
	repo *repository.FinanceRepository
}

// NewFinanceService 创建财务服务
func NewFinanceService(db *gorm.DB, repo *repository.FinanceRepository) *FinanceService {
	return &FinanceService{db: db, repo: repo}
}

// ==================== 收支记录 ====================

// CreateFinanceRecord 创建收支记录
func (s *FinanceService) CreateFinanceRecord(record *models.FinanceRecord) error {
	record.CreatedAt = time.Now()
	record.UpdatedAt = time.Now()
	return s.repo.CreateFinanceRecord(record)
}

// UpdateFinanceRecord 更新收支记录
func (s *FinanceService) UpdateFinanceRecord(tenantID, recordID int64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return s.db.Model(&models.FinanceRecord{}).
		Where("tenant_id = ? AND id = ?", tenantID, recordID).
		Updates(updates).Error
}

// DeleteFinanceRecord 删除收支记录
func (s *FinanceService) DeleteFinanceRecord(tenantID, recordID int64) error {
	return s.repo.DeleteFinanceRecord(tenantID, recordID)
}

// GetFinanceRecord 获取收支记录
func (s *FinanceService) GetFinanceRecord(tenantID, recordID int64) (*models.FinanceRecord, error) {
	return s.repo.GetFinanceRecord(tenantID, recordID)
}

// ListFinanceRecords 获取收支记录列表
func (s *FinanceService) ListFinanceRecords(tenantID int64, filter *repository.FinanceRecordFilter) ([]models.FinanceRecord, int64, error) {
	return s.repo.ListFinanceRecords(tenantID, filter)
}

// ApproveFinanceRecord 审核收支记录
func (s *FinanceService) ApproveFinanceRecord(tenantID, recordID, approverID int64) error {
	return s.repo.ApproveFinanceRecord(tenantID, recordID, approverID)
}

// ==================== 平台账单 ====================

// CreatePlatformBill 创建平台账单
func (s *FinanceService) CreatePlatformBill(bill *models.PlatformBill) error {
	bill.CreatedAt = time.Now()
	bill.UpdatedAt = time.Now()
	return s.repo.CreatePlatformBill(bill)
}

// GetPlatformBill 获取平台账单
func (s *FinanceService) GetPlatformBill(tenantID, billID int64) (*models.PlatformBill, error) {
	return s.repo.GetPlatformBill(tenantID, billID)
}

// ListPlatformBills 获取平台账单列表
func (s *FinanceService) ListPlatformBills(tenantID int64, filter *repository.BillFilter) ([]models.PlatformBill, int64, error) {
	return s.repo.ListPlatformBills(tenantID, filter)
}

// GetBillDetails 获取账单明细
func (s *FinanceService) GetBillDetails(tenantID, billID int64) ([]models.PlatformBillDetail, error) {
	return s.repo.ListPlatformBillDetails(tenantID, billID)
}

// ReconcileBillDetail 对账单明细
func (s *FinanceService) ReconcileBillDetail(tenantID, detailID int64, orderID int64) error {
	return s.repo.ReconcileBillDetail(tenantID, detailID, orderID)
}

// ==================== 供应商 ====================

// CreateSupplier 创建供应商
func (s *FinanceService) CreateSupplier(supplier *models.Supplier) error {
	supplier.CreatedAt = time.Now()
	supplier.UpdatedAt = time.Now()
	return s.repo.CreateSupplier(supplier)
}

// UpdateSupplier 更新供应商
func (s *FinanceService) UpdateSupplier(tenantID, supplierID int64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return s.db.Model(&models.Supplier{}).
		Where("tenant_id = ? AND id = ?", tenantID, supplierID).
		Updates(updates).Error
}

// DeleteSupplier 删除供应商
func (s *FinanceService) DeleteSupplier(tenantID, supplierID int64) error {
	return s.repo.DeleteSupplier(tenantID, supplierID)
}

// GetSupplier 获取供应商
func (s *FinanceService) GetSupplier(tenantID, supplierID int64) (*models.Supplier, error) {
	return s.repo.GetSupplier(tenantID, supplierID)
}

// ListSuppliers 获取供应商列表
func (s *FinanceService) ListSuppliers(tenantID int64, filter *repository.SupplierFilter) ([]models.Supplier, int64, error) {
	return s.repo.ListSuppliers(tenantID, filter)
}

// ==================== 采购结算 ====================

// CreatePurchaseSettlement 创建采购结算单
func (s *FinanceService) CreatePurchaseSettlement(settlement *models.PurchaseSettlement) error {
	settlement.CreatedAt = time.Now()
	settlement.UpdatedAt = time.Now()
	settlement.SettlementNo = s.generateSettlementNo()
	return s.repo.CreatePurchaseSettlement(settlement)
}

// GetPurchaseSettlement 获取采购结算单
func (s *FinanceService) GetPurchaseSettlement(tenantID, settlementID int64) (*models.PurchaseSettlement, error) {
	return s.repo.GetPurchaseSettlement(tenantID, settlementID)
}

// ListPurchaseSettlements 获取采购结算单列表
func (s *FinanceService) ListPurchaseSettlements(tenantID int64, filter *repository.SettlementFilter) ([]models.PurchaseSettlement, int64, error) {
	return s.repo.ListPurchaseSettlements(tenantID, filter)
}

// PaySettlement 结算单付款
func (s *FinanceService) PaySettlement(tenantID, settlementID, creatorID int64, payment *models.PurchasePayment) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取结算单
		var settlement models.PurchaseSettlement
		if err := tx.Where("tenant_id = ? AND id = ?", tenantID, settlementID).First(&settlement).Error; err != nil {
			return err
		}

		// 创建付款记录
		payment.TenantID = tenantID
		payment.SettlementID = settlementID
		payment.PaymentNo = s.generatePaymentNo()
		payment.CreatedBy = creatorID
		payment.CreatedAt = time.Now()
		payment.UpdatedAt = time.Now()
		if err := tx.Create(payment).Error; err != nil {
			return err
		}

		// 更新结算单已付金额
		newPaidAmount := settlement.PaidAmount + payment.Amount
		newStatus := settlement.Status
		if newPaidAmount >= settlement.RealAmount {
			newStatus = 2 // 已付款
		} else if newPaidAmount > 0 {
			newStatus = 1 // 部分付款
		}

		if err := tx.Model(&models.PurchaseSettlement{}).
			Where("tenant_id = ? AND id = ?", tenantID, settlementID).
			Updates(map[string]interface{}{
				"paid_amount": newPaidAmount,
				"status":      newStatus,
				"updated_at":  time.Now(),
			}).Error; err != nil {
			return err
		}

		// 更新供应商余额
		if err := tx.Model(&models.Supplier{}).
			Where("tenant_id = ? AND id = ?", tenantID, settlement.SupplierID).
			UpdateColumn("balance", gorm.Expr("balance - ?", payment.Amount)).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetSettlementPayments 获取结算单付款记录
func (s *FinanceService) GetSettlementPayments(tenantID, settlementID int64) ([]models.PurchasePayment, error) {
	return s.repo.ListPurchasePayments(tenantID, settlementID)
}

// ==================== 商品成本 ====================

// GetProductCost 获取商品成本
func (s *FinanceService) GetProductCost(tenantID, productID int64) (*models.ProductCost, error) {
	return s.repo.GetProductCost(tenantID, productID)
}

// GetProductCostByID 通过ID获取商品成本
func (s *FinanceService) GetProductCostByID(tenantID, costID int64) (*models.ProductCost, error) {
	return s.repo.GetProductCostByID(tenantID, costID)
}

// ListProductCosts 获取商品成本列表
func (s *FinanceService) ListProductCosts(tenantID int64, filter *repository.ProductCostFilter) ([]models.ProductCost, int64, error) {
	return s.repo.ListProductCosts(tenantID, filter)
}

// UpdateProductCost 更新商品成本
func (s *FinanceService) UpdateProductCost(tenantID, costID int64, updates map[string]interface{}) error {
	// 重新计算总成本
	if purchaseCost, ok := updates["purchase_cost"].(float64); ok {
		shippingCost, _ := updates["shipping_cost"].(float64)
		packageCost, _ := updates["package_cost"].(float64)
		otherCost, _ := updates["other_cost"].(float64)
		updates["total_cost"] = purchaseCost + shippingCost + packageCost + otherCost
	}

	updates["updated_at"] = time.Now()
	return s.db.Model(&models.ProductCost{}).
		Where("tenant_id = ? AND id = ?", tenantID, costID).
		Updates(updates).Error
}

// BatchUpdateProductCosts 批量更新商品成本
func (s *FinanceService) BatchUpdateProductCosts(tenantID int64, costs []models.ProductCost) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, cost := range costs {
			cost.TenantID = tenantID
			cost.TotalCost = cost.PurchaseCost + cost.ShippingCost + cost.PackageCost + cost.OtherCost
			cost.UpdatedAt = time.Now()

			// 查找是否存在
			var existing models.ProductCost
			err := tx.Where("tenant_id = ? AND product_id = ?", tenantID, cost.ProductID).First(&existing).Error
			if err == gorm.ErrRecordNotFound {
				cost.CreatedAt = time.Now()
				if err := tx.Create(&cost).Error; err != nil {
					return err
				}
			} else if err == nil {
				if err := tx.Model(&existing).Updates(map[string]interface{}{
					"purchase_cost": cost.PurchaseCost,
					"shipping_cost": cost.ShippingCost,
					"package_cost":  cost.PackageCost,
					"other_cost":    cost.OtherCost,
					"total_cost":    cost.TotalCost,
					"cost_method":   cost.CostMethod,
					"updated_at":    cost.UpdatedAt,
				}).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
		return nil
	})
}

// ==================== 订单成本 ====================

// GetOrderCost 获取订单成本
func (s *FinanceService) GetOrderCost(tenantID, orderID int64) (*models.OrderCost, error) {
	return s.repo.GetOrderCost(tenantID, orderID)
}

// ListOrderCosts 获取订单成本列表
func (s *FinanceService) ListOrderCosts(tenantID int64, filter *repository.OrderCostFilter) ([]models.OrderCost, int64, error) {
	return s.repo.ListOrderCosts(tenantID, filter)
}

// CalculateOrderCost 计算订单成本
func (s *FinanceService) CalculateOrderCost(tenantID, orderID int64) (*models.OrderCost, error) {
	// 获取订单
	var order models.Order
	if err := s.db.Where("tenant_id = ? AND id = ?", tenantID, orderID).
		Preload("Items").First(&order).Error; err != nil {
		return nil, err
	}

	// 计算商品成本
	var productCost float64
	for _, item := range order.Items {
		// 获取商品成本 (通过SkuID查找ProductID)
		var product models.Product
		err := s.db.Where("tenant_id = ? AND id = ?", tenantID, item.SkuID).First(&product).Error
		if err == nil {
			// 获取商品成本
			var cost models.ProductCost
			err := s.db.Where("tenant_id = ? AND product_id = ?", tenantID, product.ID).First(&cost).Error
			if err == nil {
				productCost += cost.TotalCost * float64(item.Quantity)
			}
		}
	}

	// 创建或更新订单成本
	cost := models.OrderCost{
		TenantID:       tenantID,
		OrderID:        orderID,
		OrderNo:        order.OrderNo,
		ProductCost:    productCost,
		ShippingCost:   0, // 可从订单中获取
		PackageCost:    0,
		Commission:     0, // 从平台账单获取
		ServiceFee:     0,
		PromotionFee:   0,
		OtherFee:       0,
		SaleAmount:     order.TotalAmount,
		RefundAmount:   0,
		RealSaleAmount: order.TotalAmount,
		Status:         1,
		CreatedAt:      time.Now(),
		UpdatedAt:       time.Now(),
	}

	// 计算总成本
	cost.TotalCost = cost.ProductCost + cost.ShippingCost + cost.PackageCost +
		cost.Commission + cost.ServiceFee + cost.PromotionFee + cost.OtherFee

	// 计算毛利和利润率
	cost.GrossProfit = cost.RealSaleAmount - cost.TotalCost
	if cost.RealSaleAmount > 0 {
		cost.ProfitRate = cost.GrossProfit / cost.RealSaleAmount * 100
	}

	now := time.Now()
	cost.CalculatedAt = &now

	// 保存
	var existing models.OrderCost
	err := s.db.Where("tenant_id = ? AND order_id = ?", tenantID, orderID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		if err := s.db.Create(&cost).Error; err != nil {
			return nil, err
		}
	} else if err == nil {
		cost.ID = existing.ID
		cost.CreatedAt = existing.CreatedAt
		if err := s.db.Save(&cost).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return &cost, nil
}

// GetProfitAnalysis 获取利润分析
func (s *FinanceService) GetProfitAnalysis(tenantID int64, startDate, endDate time.Time) (*ProfitAnalysis, error) {
	var analysis ProfitAnalysis

	// 汇总订单成本
	var costs []models.OrderCost
	err := s.db.Where("tenant_id = ? AND status = 1", tenantID).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Find(&costs).Error
	if err != nil {
		return nil, err
	}

	for _, c := range costs {
		analysis.TotalSales += c.SaleAmount
		analysis.TotalRefund += c.RefundAmount
		analysis.NetSales += c.RealSaleAmount
		analysis.ProductCost += c.ProductCost
		analysis.ShippingCost += c.ShippingCost
		analysis.Commission += c.Commission
		analysis.ServiceFee += c.ServiceFee
		analysis.PromotionFee += c.PromotionFee
		analysis.OtherFee += c.OtherFee
		analysis.TotalCost += c.TotalCost
		analysis.GrossProfit += c.GrossProfit
		analysis.OrderCount++
	}

	if analysis.NetSales > 0 {
		analysis.ProfitRate = analysis.GrossProfit / analysis.NetSales * 100
	}

	return &analysis, nil
}

// ProfitAnalysis 利润分析
type ProfitAnalysis struct {
	TotalSales    float64 `json:"total_sales"`
	TotalRefund   float64 `json:"total_refund"`
	NetSales      float64 `json:"net_sales"`
	ProductCost   float64 `json:"product_cost"`
	ShippingCost  float64 `json:"shipping_cost"`
	Commission    float64 `json:"commission"`
	ServiceFee    float64 `json:"service_fee"`
	PromotionFee  float64 `json:"promotion_fee"`
	OtherFee      float64 `json:"other_fee"`
	TotalCost     float64 `json:"total_cost"`
	GrossProfit   float64 `json:"gross_profit"`
	ProfitRate    float64 `json:"profit_rate"`
	OrderCount    int64   `json:"order_count"`
}

// ==================== 库存成本快照 ====================

// ListInventoryCostSnapshots 获取库存成本快照列表
func (s *FinanceService) ListInventoryCostSnapshots(tenantID int64, filter *repository.SnapshotFilter) ([]models.InventoryCostSnapshot, int64, error) {
	return s.repo.ListInventoryCostSnapshots(tenantID, filter)
}

// GenerateInventorySnapshot 生成库存成本快照
func (s *FinanceService) GenerateInventorySnapshot(tenantID int64, date time.Time) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取所有库存
		var inventories []models.Inventory
		if err := tx.Where("tenant_id = ?", tenantID).Find(&inventories).Error; err != nil {
			return err
		}

		snapshotDate := date.Format("2006-01-02")

		for _, inv := range inventories {
			// 获取商品成本
			var productCost models.ProductCost
			unitCost := 0.0
			err := tx.Where("tenant_id = ? AND product_id = ?", tenantID, inv.ProductID).First(&productCost).Error
			if err == nil {
				unitCost = productCost.TotalCost
			}

			// 检查当天是否已有快照
			var existing models.InventoryCostSnapshot
			err = tx.Where("tenant_id = ? AND warehouse_id = ? AND product_id = ? AND DATE(snapshot_date) = ?",
				tenantID, inv.WarehouseID, inv.ProductID, snapshotDate).First(&existing).Error

			snapshot := models.InventoryCostSnapshot{
				TenantID:     tenantID,
				WarehouseID:  inv.WarehouseID,
				ProductID:    inv.ProductID,
				SnapshotDate: date,
				EndQty:       inv.Quantity,
				EndAmount:    float64(inv.Quantity) * unitCost,
				UnitCost:     unitCost,
				CreatedAt:    time.Now(),
			}

			if err == gorm.ErrRecordNotFound {
				if err := tx.Create(&snapshot).Error; err != nil {
					return err
				}
			} else if err == nil {
				snapshot.ID = existing.ID
				snapshot.BeginQty = existing.BeginQty
				snapshot.BeginAmount = existing.BeginAmount
				snapshot.InQty = existing.InQty
				snapshot.InAmount = existing.InAmount
				snapshot.OutQty = existing.OutQty
				snapshot.OutAmount = existing.OutAmount
				if err := tx.Save(&snapshot).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		return nil
	})
}

// ==================== 财务结算 ====================

// ListMonthlySettlements 获取月度结算列表
func (s *FinanceService) ListMonthlySettlements(tenantID int64, filter *repository.FinancialSettlementFilter) ([]models.FinancialSettlement, int64, error) {
	return s.repo.ListFinancialSettlements(tenantID, "monthly", filter)
}

// GenerateMonthlySettlement 生成月度结算
func (s *FinanceService) GenerateMonthlySettlement(tenantID int64, period string, shopID *int64) (*models.FinancialSettlement, error) {
	// 解析期间
	startDate, err := time.Parse("2006-01", period)
	if err != nil {
		return nil, fmt.Errorf("invalid period format: %v", err)
	}
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	// 获取利润分析
	analysis, err := s.GetProfitAnalysis(tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	settlement := models.FinancialSettlement{
		TenantID:       tenantID,
		SettlementType: "monthly",
		Period:         period,
		ShopID:         shopID,
		TotalSales:     analysis.TotalSales,
		TotalRefund:    analysis.TotalRefund,
		NetSales:       analysis.NetSales,
		ProductCost:    analysis.ProductCost,
		ShippingCost:   analysis.ShippingCost,
		Commission:     analysis.Commission,
		ServiceFee:     analysis.ServiceFee,
		PromotionFee:   analysis.PromotionFee,
		OtherCost:      analysis.OtherFee,
		TotalCost:      analysis.TotalCost,
		GrossProfit:    analysis.GrossProfit,
		ProfitRate:     analysis.ProfitRate,
		Status:         0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 检查是否已存在
	var existing models.FinancialSettlement
	err = s.db.Where("tenant_id = ? AND settlement_type = ? AND period = ?", tenantID, "monthly", period).
		First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		if err := s.db.Create(&settlement).Error; err != nil {
			return nil, err
		}
	} else if err == nil {
		settlement.ID = existing.ID
		settlement.CreatedAt = existing.CreatedAt
		if err := s.db.Save(&settlement).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return &settlement, nil
}

// ConfirmMonthlySettlement 确认月度结算
func (s *FinanceService) ConfirmMonthlySettlement(tenantID, confirmerID int64, period string) error {
	now := time.Now()
	return s.db.Model(&models.FinancialSettlement{}).
		Where("tenant_id = ? AND settlement_type = ? AND period = ?", tenantID, "monthly", period).
		Updates(map[string]interface{}{
			"status":     1,
			"settled_at": now,
			"settled_by": confirmerID,
			"updated_at": now,
		}).Error
}

// ListYearlySettlements 获取年度结算列表
func (s *FinanceService) ListYearlySettlements(tenantID int64, filter *repository.FinancialSettlementFilter) ([]models.FinancialSettlement, int64, error) {
	return s.repo.ListFinancialSettlements(tenantID, "yearly", filter)
}

// GenerateYearlySettlement 生成年度结算
func (s *FinanceService) GenerateYearlySettlement(tenantID int64, year string) (*models.FinancialSettlement, error) {
	// 解析年份
	yearInt, err := fmt.Sscanf(year, "%d", new(int))
	if err != nil || yearInt != 1 {
		return nil, fmt.Errorf("invalid year format: %v", err)
	}

	startDate, _ := time.Parse("2006", year)
	_ = startDate // 用于年份验证

	// 汇总月度结算
	var monthlySettlements []models.FinancialSettlement
	err = s.db.Where("tenant_id = ? AND settlement_type = ? AND period LIKE ?", tenantID, "monthly", year+"%").
		Find(&monthlySettlements).Error
	if err != nil {
		return nil, err
	}

	settlement := models.FinancialSettlement{
		TenantID:       tenantID,
		SettlementType: "yearly",
		Period:         year,
		Status:         0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	for _, ms := range monthlySettlements {
		settlement.TotalSales += ms.TotalSales
		settlement.TotalRefund += ms.TotalRefund
		settlement.NetSales += ms.NetSales
		settlement.OtherIncome += ms.OtherIncome
		settlement.ProductCost += ms.ProductCost
		settlement.ShippingCost += ms.ShippingCost
		settlement.Commission += ms.Commission
		settlement.ServiceFee += ms.ServiceFee
		settlement.PromotionFee += ms.PromotionFee
		settlement.OtherCost += ms.OtherCost
		settlement.TotalCost += ms.TotalCost
		settlement.GrossProfit += ms.GrossProfit
		settlement.InventoryChange += ms.InventoryChange
	}

	if settlement.NetSales > 0 {
		settlement.ProfitRate = settlement.GrossProfit / settlement.NetSales * 100
	}

	// 检查是否已存在
	var existing models.FinancialSettlement
	err = s.db.Where("tenant_id = ? AND settlement_type = ? AND period = ?", tenantID, "yearly", year).
		First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		if err := s.db.Create(&settlement).Error; err != nil {
			return nil, err
		}
	} else if err == nil {
		settlement.ID = existing.ID
		settlement.CreatedAt = existing.CreatedAt
		if err := s.db.Save(&settlement).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return &settlement, nil
}

// ==================== 结算账户 ====================

// CreateBankAccount 创建结算账户
func (s *FinanceService) CreateBankAccount(account *models.FinanceBankAccount) error {
	account.CreatedAt = time.Now()
	account.UpdatedAt = time.Now()
	return s.repo.CreateFinanceBankAccount(account)
}

// UpdateBankAccount 更新结算账户
func (s *FinanceService) UpdateBankAccount(tenantID, accountID int64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return s.db.Model(&models.FinanceBankAccount{}).
		Where("tenant_id = ? AND id = ?", tenantID, accountID).
		Updates(updates).Error
}

// DeleteBankAccount 删除结算账户
func (s *FinanceService) DeleteBankAccount(tenantID, accountID int64) error {
	return s.repo.DeleteFinanceBankAccount(tenantID, accountID)
}

// GetBankAccount 获取结算账户
func (s *FinanceService) GetBankAccount(tenantID, accountID int64) (*models.FinanceBankAccount, error) {
	return s.repo.GetFinanceBankAccount(tenantID, accountID)
}

// ListBankAccounts 获取结算账户列表
func (s *FinanceService) ListBankAccounts(tenantID int64) ([]models.FinanceBankAccount, error) {
	return s.repo.ListFinanceBankAccounts(tenantID)
}

// ==================== 辅助方法 ====================

// generateSettlementNo 生成结算单号
func (s *FinanceService) generateSettlementNo() string {
	return fmt.Sprintf("PS%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
}

// generatePaymentNo 生成付款单号
func (s *FinanceService) generatePaymentNo() string {
	return fmt.Sprintf("PP%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
}
