package models

import (
	"time"
)

// 角色类型
type RoleType string

const (
	RoleOwner      RoleType = "owner"       // 主账号（所有者）
	RoleAdmin      RoleType = "admin"       // 管理员
	RoleManager    RoleType = "manager"     // 经理
	RoleOperator   RoleType = "operator"    // 运营
	RoleFinance    RoleType = "finance"     // 财务
	RoleWarehouse  RoleType = "warehouse"   // 仓库
	RoleCustomer   RoleType = "customer"    // 客服
	RoleViewer     RoleType = "viewer"      // 只读
)

// 系统预设角色（不可删除）
var SystemRoles = []RoleType{
	RoleOwner, RoleAdmin, RoleManager, RoleOperator,
	RoleFinance, RoleWarehouse, RoleCustomer, RoleViewer,
}

func IsSystemRole(code string) bool {
	for _, r := range SystemRoles {
		if string(r) == code {
			return true
		}
	}
	return false
}

// Role 角色
type Role struct {
	ID          int64        `json:"id" gorm:"primaryKey"`
	TenantID    int64        `json:"tenant_id" gorm:"index;default:0"` // 0=系统预设角色，>0=租户自定义角色
	Code        string       `json:"code" gorm:"index;size:50"`
	Name        string       `json:"name" gorm:"size:50"`
	Description string       `json:"description" gorm:"size:200"`
	IsSystem    bool         `json:"is_system" gorm:"default:false"` // 是否系统预设角色
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Permission 权限
type Permission struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	Code        string    `json:"code" gorm:"uniqueIndex;size:100"` // 如: order:read, order:write
	Name        string    `json:"name" gorm:"size:100"`
	Resource    string    `json:"resource" gorm:"size:50"` // 资源: order, product, user
	Action      string    `json:"action" gorm:"size:50"`   // 操作: read, write, delete, approve
	Description string    `json:"description" gorm:"size:200"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserRole 用户角色关联（租户级别）
type UserRole struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"uniqueIndex:idx_tenant_user_role;not null"`
	UserID    int64     `json:"user_id" gorm:"uniqueIndex:idx_tenant_user_role;not null"`
	RoleID    int64     `json:"role_id" gorm:"index;not null"`
	Role      *Role     `json:"role" gorm:"foreignKey:RoleID"`
	CreatedAt time.Time `json:"created_at"`
}

// UserResourcePermission 用户资源级权限（覆盖角色权限）
type UserResourcePermission struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	TenantID     int64     `json:"tenant_id" gorm:"index;not null"`
	UserID       int64     `json:"user_id" gorm:"index;not null"`
	ResourceType string    `json:"resource_type" gorm:"size:20;not null"` // shop, warehouse
	ResourceID   int64     `json:"resource_id"`                           // 店铺ID或仓库ID
	CanRead      bool      `json:"can_read"`
	CanWrite     bool      `json:"can_write"`
	CanDelete    bool      `json:"can_delete"`
	CreatedAt    time.Time `json:"created_at"`
}

// 资源类型常量
const (
	ResourceTypeShop     = "shop"
	ResourceTypeWarehouse = "warehouse"
)

// UserGroup 用户组
type UserGroup struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:50"`
	Description string    `json:"description" gorm:"size:200"`
	Users       []User    `json:"users" gorm:"many2many:group_users;"`
	Roles       []Role    `json:"roles" gorm:"many2many:group_roles;"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 权限常量定义
const (
	// 订单权限
	PermOrderRead   = "order:read"
	PermOrderWrite  = "order:write"
	PermOrderDelete = "order:delete"
	PermOrderAudit  = "order:audit"
	PermOrderShip   = "order:ship"

	// 商品权限
	PermProductRead   = "product:read"
	PermProductWrite  = "product:write"
	PermProductDelete = "product:delete"

	// 库存权限
	PermInventoryRead  = "inventory:read"
	PermInventoryWrite = "inventory:write"

	// 仓库权限
	PermWarehouseRead  = "warehouse:read"
	PermWarehouseWrite = "warehouse:write"

	// 用户权限
	PermUserRead   = "user:read"
	PermUserWrite  = "user:write"
	PermUserDelete = "user:delete"

	// 店铺权限
	PermShopRead   = "shop:read"
	PermShopWrite  = "shop:write"
	PermShopDelete = "shop:delete"

	// 财务权限
	PermFinanceRead   = "finance:read"
	PermFinanceWrite  = "finance:write"
	PermFinanceExport = "finance:export"

	// 报表权限
	PermReportRead   = "report:read"
	PermReportExport = "report:export"

	// 角色权限
	PermRoleRead  = "role:read"
	PermRoleWrite = "role:write"

	// 授权权限
	PermPermissionAssign = "permission:assign"

	// 系统权限
	PermSystemConfig = "system:config"
	PermSystemLog    = "system:log"
)

// 预定义角色权限
var RolePermissions = map[RoleType][]string{
	RoleOwner: {
		PermOrderRead, PermOrderWrite, PermOrderDelete, PermOrderAudit, PermOrderShip,
		PermProductRead, PermProductWrite, PermProductDelete,
		PermInventoryRead, PermInventoryWrite,
		PermUserRead, PermUserWrite, PermUserDelete,
		PermShopRead, PermShopWrite, PermShopDelete,
		PermWarehouseRead, PermWarehouseWrite,
		PermFinanceRead, PermFinanceWrite, PermFinanceExport,
		PermReportRead, PermReportExport,
		PermSystemConfig, PermSystemLog,
		PermRoleRead, PermRoleWrite,
		PermPermissionAssign,
	},
	RoleAdmin: {
		PermOrderRead, PermOrderWrite, PermOrderAudit, PermOrderShip,
		PermProductRead, PermProductWrite,
		PermInventoryRead, PermInventoryWrite,
		PermUserRead, PermUserWrite,
		PermShopRead, PermShopWrite,
		PermWarehouseRead, PermWarehouseWrite,
		PermFinanceRead, PermFinanceExport,
		PermReportRead, PermReportExport,
	},
	RoleManager: {
		PermOrderRead, PermOrderWrite, PermOrderAudit, PermOrderShip,
		PermProductRead, PermProductWrite,
		PermInventoryRead, PermInventoryWrite,
		PermUserRead,
		PermShopRead,
		PermWarehouseRead,
		PermFinanceRead, PermFinanceExport,
		PermReportRead, PermReportExport,
	},
	RoleOperator: {
		PermOrderRead, PermOrderWrite, PermOrderAudit, PermOrderShip,
		PermProductRead,
		PermInventoryRead,
		PermReportRead,
	},
	RoleFinance: {
		PermOrderRead,
		PermFinanceRead, PermFinanceWrite, PermFinanceExport,
		PermReportRead, PermReportExport,
	},
	RoleWarehouse: {
		PermOrderRead, PermOrderShip,
		PermInventoryRead, PermInventoryWrite,
		PermProductRead,
		PermWarehouseRead, PermWarehouseWrite,
	},
	RoleCustomer: {
		PermOrderRead, PermOrderWrite,
	},
	RoleViewer: {
		PermOrderRead,
		PermProductRead,
		PermInventoryRead,
		PermReportRead,
	},
}
