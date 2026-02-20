package repository

import (
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"gorm.io/gorm"
)

// FinanceRepository 财务数据仓库
type FinanceRepository struct {
	db *gorm.DB
}

// NewFinanceRepository 创建财务仓库
func NewFinanceRepository(db *gorm.DB) *FinanceRepository {
	return &FinanceRepository{db: db}
}

// ==================== 收支记录 ====================

// CreateFinanceRecord 创建收支记录
func (r *FinanceRepository) CreateFinanceRecord(record *models.FinanceRecord) error {
	return r.db.Create(record).Error
}

// UpdateFinanceRecord 更新收支记录
func (r *FinanceRepository) UpdateFinanceRecord(record *models.FinanceRecord) error {
	return r.db.Save(record).Error
}

// DeleteFinanceRecord 删除收支记录
func (r *FinanceRepository) DeleteFinanceRecord(tenantID, id int64) error {
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&models.FinanceRecord{}).Error
}

// GetFinanceRecord 获取收支记录
func (r *FinanceRepository) GetFinanceRecord(tenantID, id int64) (*models.FinanceRecord, error) {
	var record models.FinanceRecord
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ListFinanceRecords 获取收支记录列表
func (r *FinanceRepository) ListFinanceRecords(tenantID int64, filter *FinanceRecordFilter) ([]models.FinanceRecord, int64, error) {
	var records []models.FinanceRecord
	var total int64

	query := r.db.Model(&models.FinanceRecord{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.Type != "" {
			query = query.Where("type = ?", filter.Type)
		}
		if filter.Category != "" {
			query = query.Where("category = ?", filter.Category)
		}
		if filter.ShopID != nil {
			query = query.Where("shop_id = ?", *filter.ShopID)
		}
		if filter.Status >= 0 {
			query = query.Where("status = ?", filter.Status)
		}
		if !filter.StartDate.IsZero() {
			query = query.Where("record_date >= ?", filter.StartDate)
		}
		if !filter.EndDate.IsZero() {
			query = query.Where("record_date <= ?", filter.EndDate)
		}
	}

	query.Count(&total)

	if filter != nil && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Order("record_date DESC, id DESC").Find(&records).Error
	return records, total, err
}

// ApproveFinanceRecord 审核收支记录
func (r *FinanceRepository) ApproveFinanceRecord(tenantID, id, approverID int64) error {
	now := time.Now()
	return r.db.Model(&models.FinanceRecord{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]interface{}{
			"status":      1,
			"approved_by": approverID,
			"approved_at": now,
		}).Error
}

// ==================== 平台账单 ====================

// CreatePlatformBill 创建平台账单
func (r *FinanceRepository) CreatePlatformBill(bill *models.PlatformBill) error {
	return r.db.Create(bill).Error
}

// UpdatePlatformBill 更新平台账单
func (r *FinanceRepository) UpdatePlatformBill(bill *models.PlatformBill) error {
	return r.db.Save(bill).Error
}

// GetPlatformBill 获取平台账单
func (r *FinanceRepository) GetPlatformBill(tenantID, id int64) (*models.PlatformBill, error) {
	var bill models.PlatformBill
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).
		Preload("Shop").First(&bill).Error
	if err != nil {
		return nil, err
	}
	return &bill, nil
}

// ListPlatformBills 获取平台账单列表
func (r *FinanceRepository) ListPlatformBills(tenantID int64, filter *BillFilter) ([]models.PlatformBill, int64, error) {
	var bills []models.PlatformBill
	var total int64

	query := r.db.Model(&models.PlatformBill{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.ShopID != nil {
			query = query.Where("shop_id = ?", *filter.ShopID)
		}
		if filter.Platform != "" {
			query = query.Where("platform = ?", filter.Platform)
		}
		if filter.BillPeriod != "" {
			query = query.Where("bill_period = ?", filter.BillPeriod)
		}
		if filter.Status >= 0 {
			query = query.Where("status = ?", filter.Status)
		}
	}

	query.Count(&total)

	if filter != nil && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Order("bill_date DESC, id DESC").Preload("Shop").Find(&bills).Error
	return bills, total, err
}

// CreatePlatformBillDetail 创建账单明细
func (r *FinanceRepository) CreatePlatformBillDetail(detail *models.PlatformBillDetail) error {
	return r.db.Create(detail).Error
}

// ListPlatformBillDetails 获取账单明细列表
func (r *FinanceRepository) ListPlatformBillDetails(tenantID, billID int64) ([]models.PlatformBillDetail, error) {
	var details []models.PlatformBillDetail
	err := r.db.Where("tenant_id = ? AND bill_id = ?", tenantID, billID).
		Order("transaction_time DESC").Find(&details).Error
	return details, err
}

// ReconcileBillDetail 对账单明细
func (r *FinanceRepository) ReconcileBillDetail(tenantID, detailID int64, orderID int64) error {
	now := time.Now()
	return r.db.Model(&models.PlatformBillDetail{}).
		Where("tenant_id = ? AND id = ?", tenantID, detailID).
		Updates(map[string]interface{}{
			"order_id":     orderID,
			"status":       1,
			"reconciled_at": now,
		}).Error
}

// ==================== 供应商 ====================

// CreateSupplier 创建供应商
func (r *FinanceRepository) CreateSupplier(supplier *models.Supplier) error {
	return r.db.Create(supplier).Error
}

// UpdateSupplier 更新供应商
func (r *FinanceRepository) UpdateSupplier(supplier *models.Supplier) error {
	return r.db.Save(supplier).Error
}

// DeleteSupplier 删除供应商
func (r *FinanceRepository) DeleteSupplier(tenantID, id int64) error {
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&models.Supplier{}).Error
}

// GetSupplier 获取供应商
func (r *FinanceRepository) GetSupplier(tenantID, id int64) (*models.Supplier, error) {
	var supplier models.Supplier
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&supplier).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

// ListSuppliers 获取供应商列表
func (r *FinanceRepository) ListSuppliers(tenantID int64, filter *SupplierFilter) ([]models.Supplier, int64, error) {
	var suppliers []models.Supplier
	var total int64

	query := r.db.Model(&models.Supplier{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.Name != "" {
			query = query.Where("name LIKE ?", "%"+filter.Name+"%")
		}
		if filter.Status >= 0 {
			query = query.Where("status = ?", filter.Status)
		}
	}

	query.Count(&total)

	if filter != nil && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Order("id DESC").Find(&suppliers).Error
	return suppliers, total, err
}

// UpdateSupplierBalance 更新供应商余额
func (r *FinanceRepository) UpdateSupplierBalance(tenantID, supplierID int64, delta float64) error {
	return r.db.Model(&models.Supplier{}).
		Where("tenant_id = ? AND id = ?", tenantID, supplierID).
		UpdateColumn("balance", gorm.Expr("balance + ?", delta)).Error
}

// ==================== 采购结算 ====================

// CreatePurchaseSettlement 创建采购结算单
func (r *FinanceRepository) CreatePurchaseSettlement(settlement *models.PurchaseSettlement) error {
	return r.db.Create(settlement).Error
}

// UpdatePurchaseSettlement 更新采购结算单
func (r *FinanceRepository) UpdatePurchaseSettlement(settlement *models.PurchaseSettlement) error {
	return r.db.Save(settlement).Error
}

// GetPurchaseSettlement 获取采购结算单
func (r *FinanceRepository) GetPurchaseSettlement(tenantID, id int64) (*models.PurchaseSettlement, error) {
	var settlement models.PurchaseSettlement
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).
		Preload("Supplier").First(&settlement).Error
	if err != nil {
		return nil, err
	}
	return &settlement, nil
}

// ListPurchaseSettlements 获取采购结算单列表
func (r *FinanceRepository) ListPurchaseSettlements(tenantID int64, filter *SettlementFilter) ([]models.PurchaseSettlement, int64, error) {
	var settlements []models.PurchaseSettlement
	var total int64

	query := r.db.Model(&models.PurchaseSettlement{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.SupplierID != nil {
			query = query.Where("supplier_id = ?", *filter.SupplierID)
		}
		if filter.Status >= 0 {
			query = query.Where("status = ?", filter.Status)
		}
	}

	query.Count(&total)

	if filter != nil && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Order("settlement_date DESC, id DESC").Preload("Supplier").Find(&settlements).Error
	return settlements, total, err
}

// CreatePurchasePayment 创建采购付款记录
func (r *FinanceRepository) CreatePurchasePayment(payment *models.PurchasePayment) error {
	return r.db.Create(payment).Error
}

// ListPurchasePayments 获取采购付款记录列表
func (r *FinanceRepository) ListPurchasePayments(tenantID, settlementID int64) ([]models.PurchasePayment, error) {
	var payments []models.PurchasePayment
	err := r.db.Where("tenant_id = ? AND settlement_id = ?", tenantID, settlementID).
		Order("payment_date DESC").Find(&payments).Error
	return payments, err
}

// ==================== 商品成本 ====================

// CreateProductCost 创建商品成本
func (r *FinanceRepository) CreateProductCost(cost *models.ProductCost) error {
	return r.db.Create(cost).Error
}

// UpdateProductCost 更新商品成本
func (r *FinanceRepository) UpdateProductCost(cost *models.ProductCost) error {
	return r.db.Save(cost).Error
}

// GetProductCost 获取商品成本
func (r *FinanceRepository) GetProductCost(tenantID, productID int64) (*models.ProductCost, error) {
	var cost models.ProductCost
	err := r.db.Where("tenant_id = ? AND product_id = ?", tenantID, productID).First(&cost).Error
	if err != nil {
		return nil, err
	}
	return &cost, nil
}

// GetProductCostByID 通过ID获取商品成本
func (r *FinanceRepository) GetProductCostByID(tenantID, id int64) (*models.ProductCost, error) {
	var cost models.ProductCost
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&cost).Error
	if err != nil {
		return nil, err
	}
	return &cost, nil
}

// ListProductCosts 获取商品成本列表
func (r *FinanceRepository) ListProductCosts(tenantID int64, filter *ProductCostFilter) ([]models.ProductCost, int64, error) {
	var costs []models.ProductCost
	var total int64

	query := r.db.Model(&models.ProductCost{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.ProductSku != "" {
			query = query.Where("product_sku LIKE ?", "%"+filter.ProductSku+"%")
		}
	}

	query.Count(&total)

	if filter != nil && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Order("id DESC").Find(&costs).Error
	return costs, total, err
}

// ==================== 订单成本 ====================

// CreateOrderCost 创建订单成本
func (r *FinanceRepository) CreateOrderCost(cost *models.OrderCost) error {
	return r.db.Create(cost).Error
}

// UpdateOrderCost 更新订单成本
func (r *FinanceRepository) UpdateOrderCost(cost *models.OrderCost) error {
	return r.db.Save(cost).Error
}

// GetOrderCost 获取订单成本
func (r *FinanceRepository) GetOrderCost(tenantID, orderID int64) (*models.OrderCost, error) {
	var cost models.OrderCost
	err := r.db.Where("tenant_id = ? AND order_id = ?", tenantID, orderID).First(&cost).Error
	if err != nil {
		return nil, err
	}
	return &cost, nil
}

// GetOrderCostByID 通过ID获取订单成本
func (r *FinanceRepository) GetOrderCostByID(tenantID, id int64) (*models.OrderCost, error) {
	var cost models.OrderCost
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&cost).Error
	if err != nil {
		return nil, err
	}
	return &cost, nil
}

// ListOrderCosts 获取订单成本列表
func (r *FinanceRepository) ListOrderCosts(tenantID int64, filter *OrderCostFilter) ([]models.OrderCost, int64, error) {
	var costs []models.OrderCost
	var total int64

	query := r.db.Model(&models.OrderCost{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.OrderNo != "" {
			query = query.Where("order_no LIKE ?", "%"+filter.OrderNo+"%")
		}
		if filter.Status >= 0 {
			query = query.Where("status = ?", filter.Status)
		}
	}

	query.Count(&total)

	if filter != nil && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Order("id DESC").Find(&costs).Error
	return costs, total, err
}

// ==================== 库存成本快照 ====================

// CreateInventoryCostSnapshot 创建库存成本快照
func (r *FinanceRepository) CreateInventoryCostSnapshot(snapshot *models.InventoryCostSnapshot) error {
	return r.db.Create(snapshot).Error
}

// ListInventoryCostSnapshots 获取库存成本快照列表
func (r *FinanceRepository) ListInventoryCostSnapshots(tenantID int64, filter *SnapshotFilter) ([]models.InventoryCostSnapshot, int64, error) {
	var snapshots []models.InventoryCostSnapshot
	var total int64

	query := r.db.Model(&models.InventoryCostSnapshot{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.WarehouseID != nil {
			query = query.Where("warehouse_id = ?", *filter.WarehouseID)
		}
		if filter.ProductID != nil {
			query = query.Where("product_id = ?", *filter.ProductID)
		}
		if !filter.SnapshotDate.IsZero() {
			query = query.Where("snapshot_date = ?", filter.SnapshotDate.Format("2006-01-02"))
		}
	}

	query.Count(&total)

	if filter != nil && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Order("snapshot_date DESC, id DESC").Find(&snapshots).Error
	return snapshots, total, err
}

// ==================== 财务结算 ====================

// CreateFinancialSettlement 创建财务结算
func (r *FinanceRepository) CreateFinancialSettlement(settlement *models.FinancialSettlement) error {
	return r.db.Create(settlement).Error
}

// UpdateFinancialSettlement 更新财务结算
func (r *FinanceRepository) UpdateFinancialSettlement(settlement *models.FinancialSettlement) error {
	return r.db.Save(settlement).Error
}

// GetFinancialSettlement 获取财务结算
func (r *FinanceRepository) GetFinancialSettlement(tenantID int64, settlementType, period string) (*models.FinancialSettlement, error) {
	var settlement models.FinancialSettlement
	err := r.db.Where("tenant_id = ? AND settlement_type = ? AND period = ?", tenantID, settlementType, period).
		First(&settlement).Error
	if err != nil {
		return nil, err
	}
	return &settlement, nil
}

// ListFinancialSettlements 获取财务结算列表
func (r *FinanceRepository) ListFinancialSettlements(tenantID int64, settlementType string, filter *FinancialSettlementFilter) ([]models.FinancialSettlement, int64, error) {
	var settlements []models.FinancialSettlement
	var total int64

	query := r.db.Model(&models.FinancialSettlement{}).
		Where("tenant_id = ? AND settlement_type = ?", tenantID, settlementType)

	if filter != nil {
		if filter.Status >= 0 {
			query = query.Where("status = ?", filter.Status)
		}
	}

	query.Count(&total)

	if filter != nil && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Order("period DESC").Find(&settlements).Error
	return settlements, total, err
}

// ==================== 结算账户 ====================

// CreateFinanceBankAccount 创建结算账户
func (r *FinanceRepository) CreateFinanceBankAccount(account *models.FinanceBankAccount) error {
	return r.db.Create(account).Error
}

// UpdateFinanceBankAccount 更新结算账户
func (r *FinanceRepository) UpdateFinanceBankAccount(account *models.FinanceBankAccount) error {
	return r.db.Save(account).Error
}

// DeleteFinanceBankAccount 删除结算账户
func (r *FinanceRepository) DeleteFinanceBankAccount(tenantID, id int64) error {
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&models.FinanceBankAccount{}).Error
}

// GetFinanceBankAccount 获取结算账户
func (r *FinanceRepository) GetFinanceBankAccount(tenantID, id int64) (*models.FinanceBankAccount, error) {
	var account models.FinanceBankAccount
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// ListFinanceBankAccounts 获取结算账户列表
func (r *FinanceRepository) ListFinanceBankAccounts(tenantID int64) ([]models.FinanceBankAccount, error) {
	var accounts []models.FinanceBankAccount
	err := r.db.Where("tenant_id = ?", tenantID).Order("is_default DESC, id DESC").Find(&accounts).Error
	return accounts, err
}

// ==================== 过滤器定义 ====================

// FinanceRecordFilter 收支记录过滤器
type FinanceRecordFilter struct {
	Page      int
	PageSize  int
	Type      string
	Category  string
	ShopID    *int64
	Status    int
	StartDate time.Time
	EndDate   time.Time
}

// BillFilter 账单过滤器
type BillFilter struct {
	Page       int
	PageSize   int
	ShopID     *int64
	Platform   string
	BillPeriod string
	Status     int
}

// SupplierFilter 供应商过滤器
type SupplierFilter struct {
	Page     int
	PageSize int
	Name     string
	Status   int
}

// SettlementFilter 结算单过滤器
type SettlementFilter struct {
	Page       int
	PageSize   int
	SupplierID *int64
	Status     int
}

// ProductCostFilter 商品成本过滤器
type ProductCostFilter struct {
	Page       int
	PageSize   int
	ProductSku string
}

// OrderCostFilter 订单成本过滤器
type OrderCostFilter struct {
	Page     int
	PageSize int
	OrderNo  string
	Status   int
}

// SnapshotFilter 快照过滤器
type SnapshotFilter struct {
	Page         int
	PageSize     int
	WarehouseID  *int64
	ProductID    *int64
	SnapshotDate time.Time
}

// FinancialSettlementFilter 财务结算过滤器
type FinancialSettlementFilter struct {
	Page     int
	PageSize int
	Status   int
}
