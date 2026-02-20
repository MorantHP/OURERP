#!/bin/bash

# 数据中心模块API测试脚本
# 使用方法: ./test_datacenter.sh <base_url> <token>
# 示例: ./test_datacenter.sh http://localhost:8080 "your-jwt-token"

BASE_URL="${1:-http://localhost:8080}"
TOKEN="${2}"

if [ -z "$TOKEN" ]; then
    echo "错误: 请提供JWT token"
    echo "使用方法: $0 <base_url> <token>"
    exit 1
fi

# 通用请求头
HEADERS="-H 'Content-Type: application/json' -H 'Authorization: Bearer $TOKEN'"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试计数
PASS_COUNT=0
FAIL_COUNT=0

# 测试函数
test_api() {
    local name="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    local expected_status="$5"

    echo -e "${YELLOW}测试: $name${NC}"
    echo "请求: $method $endpoint"

    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $TOKEN" \
            -d "$data" 2>/dev/null)
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $TOKEN" 2>/dev/null)
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" == "$expected_status" ]; then
        echo -e "${GREEN}✓ 通过 (HTTP $http_code)${NC}"
        echo "响应: $body" | head -c 200
        echo ""
        ((PASS_COUNT++))
    else
        echo -e "${RED}✗ 失败 (期望: $expected_status, 实际: $http_code)${NC}"
        echo "响应: $body"
        ((FAIL_COUNT++))
    fi
    echo "----------------------------------------"
}

echo "========================================"
echo "    数据中心模块 API 测试"
echo "========================================"
echo "Base URL: $BASE_URL"
echo "========================================"
echo ""

# =============== 实时监控测试 ===============
echo ">>> 实时监控模块测试 <<<"
echo ""

test_api "获取实时概览" "GET" "/api/v1/datacenter/realtime/overview" "" "200"
test_api "获取实时库存状态" "GET" "/api/v1/datacenter/realtime/inventory" "" "200"
test_api "获取小时趋势" "GET" "/api/v1/datacenter/realtime/hourly-trend" "" "200"

# =============== 客户分析测试 ===============
echo ""
echo ">>> 客户分析模块测试 <<<"
echo ""

test_api "获取客户分析" "GET" "/api/v1/datacenter/customers/analysis?start_date=2025-01-01&end_date=2025-12-31" "" "200"
test_api "获取客户价值分布" "GET" "/api/v1/datacenter/customers/value-distribution" "" "200"
test_api "获取地域分布" "GET" "/api/v1/datacenter/customers/geography?start_date=2025-01-01&end_date=2025-12-31" "" "200"
test_api "获取复购分析" "GET" "/api/v1/datacenter/customers/repurchase?start_date=2025-01-01&end_date=2025-12-31" "" "200"

# =============== 商品分析测试 ===============
echo ""
echo ">>> 商品分析模块测试 <<<"
echo ""

test_api "获取商品动销率" "GET" "/api/v1/datacenter/products/turnover?start_date=2025-01-01&end_date=2025-12-31" "" "200"
test_api "获取库存水位" "GET" "/api/v1/datacenter/products/inventory-level" "" "200"
test_api "获取进货策略" "GET" "/api/v1/datacenter/products/purchase-strategy?days=30" "" "200"
test_api "获取低库存商品" "GET" "/api/v1/datacenter/products/low-stock" "" "200"
test_api "获取库存汇总" "GET" "/api/v1/datacenter/products/inventory-summary" "" "200"

# =============== 对比分析测试 ===============
echo ""
echo ">>> 对比分析模块测试 <<<"
echo ""

test_api "同比分析" "GET" "/api/v1/datacenter/compare/yoy?start_date=2025-01-01&end_date=2025-12-31" "" "200"
test_api "环比分析" "GET" "/api/v1/datacenter/compare/mom?start_date=2025-01-01&end_date=2025-12-31" "" "200"
test_api "期间对比" "GET" "/api/v1/datacenter/compare/period?current_start_date=2025-01-01&current_end_date=2025-06-30&compare_start_date=2024-01-01&compare_end_date=2024-06-30" "" "200"
test_api "平台对比" "GET" "/api/v1/datacenter/compare/platform?start_date=2025-01-01&end_date=2025-12-31" "" "200"

# =============== 预警管理测试 ===============
echo ""
echo ">>> 预警管理模块测试 <<<"
echo ""

test_api "获取预警类型" "GET" "/api/v1/datacenter/alerts/types" "" "200"
test_api "获取预警级别" "GET" "/api/v1/datacenter/alerts/levels" "" "200"
test_api "获取预警汇总" "GET" "/api/v1/datacenter/alerts/summary" "" "200"
test_api "获取预警规则列表" "GET" "/api/v1/datacenter/alerts/rules" "" "200"
test_api "获取预警记录列表" "GET" "/api/v1/datacenter/alerts/records" "" "200"

# 创建预警规则
test_api "创建预警规则" "POST" "/api/v1/datacenter/alerts/rules" \
    '{"name":"库存预警测试","type":"inventory","threshold":10,"level":"warning","notify_type":"system","description":"测试库存预警"}' \
    "200"

# 检查预警
test_api "检查预警" "POST" "/api/v1/datacenter/alerts/check" "" "200"

# =============== 测试结果汇总 ===============
echo ""
echo "========================================"
echo "    测试结果汇总"
echo "========================================"
echo -e "${GREEN}通过: $PASS_COUNT${NC}"
echo -e "${RED}失败: $FAIL_COUNT${NC}"
echo "总计: $((PASS_COUNT + FAIL_COUNT))"
echo "========================================"

if [ $FAIL_COUNT -eq 0 ]; then
    echo -e "${GREEN}所有测试通过!${NC}"
    exit 0
else
    echo -e "${RED}有测试失败!${NC}"
    exit 1
fi
