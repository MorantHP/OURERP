@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM 数据中心模块API测试脚本 (Windows版本)
REM 使用方法: test_datacenter.bat <base_url> <token>
REM 示例: test_datacenter.bat http://localhost:8080 "your-jwt-token"

set "BASE_URL=%~1"
set "TOKEN=%~2"

if "%BASE_URL%"=="" set "BASE_URL=http://localhost:8080"
if "%TOKEN%"=="" (
    echo 错误: 请提供JWT token
    echo 使用方法: %0 ^<base_url^> ^<token^>
    exit /b 1
)

set PASS_COUNT=0
set FAIL_COUNT=0

echo ========================================
echo     数据中心模块 API 测试
echo ========================================
echo Base URL: %BASE_URL%
echo ========================================
echo.

REM 测试函数
goto :main

:test_api
set "name=%~1"
set "method=%~2"
set "endpoint=%~3"
set "data=%~4"
set "expected=%~5"

echo 测试: %name%
echo 请求: %method% %endpoint%

if "%data%"=="" (
    curl -s -o response.txt -w "%%{http_code}" -X %method% "%BASE_URL%%endpoint%" ^
        -H "Content-Type: application/json" ^
        -H "Authorization: Bearer %TOKEN%" > status.txt 2>nul
) else (
    curl -s -o response.txt -w "%%{http_code}" -X %method% "%BASE_URL%%endpoint%" ^
        -H "Content-Type: application/json" ^
        -H "Authorization: Bearer %TOKEN%" ^
        -d "%data%" > status.txt 2>nul
)

set /p HTTP_CODE=<status.txt

if "%HTTP_CODE%"=="%expected%" (
    echo [92m✓ 通过 (HTTP %HTTP_CODE%)[0m
    set /a PASS_COUNT+=1
) else (
    echo [91m✗ 失败 (期望: %expected%, 实际: %HTTP_CODE%)[0m
    type response.txt
    set /a FAIL_COUNT+=1
)
echo ----------------------------------------
goto :eof

:main

REM 实时监控测试
echo ^>^>^> 实时监控模块测试 ^<^<^<^<^<^
echo.

call :test_api "获取实时概览" "GET" "/api/v1/datacenter/realtime/overview" "" "200"
call :test_api "获取实时库存状态" "GET" "/api/v1/datacenter/realtime/inventory" "" "200"
call :test_api "获取小时趋势" "GET" "/api/v1/datacenter/realtime/hourly-trend" "" "200"

REM 客户分析测试
echo.
echo ^>^>^> 客户分析模块测试 ^<^<^<^<^<^
echo.

call :test_api "获取客户分析" "GET" "/api/v1/datacenter/customers/analysis?start_date=2025-01-01^&end_date=2025-12-31" "" "200"
call :test_api "获取客户价值分布" "GET" "/api/v1/datacenter/customers/value-distribution" "" "200"
call :test_api "获取地域分布" "GET" "/api/v1/datacenter/customers/geography?start_date=2025-01-01^&end_date=2025-12-31" "" "200"
call :test_api "获取复购分析" "GET" "/api/v1/datacenter/customers/repurchase?start_date=2025-01-01^&end_date=2025-12-31" "" "200"

REM 商品分析测试
echo.
echo ^>^>^> 商品分析模块测试 ^<^<^<^<^<^
echo.

call :test_api "获取商品动销率" "GET" "/api/v1/datacenter/products/turnover?start_date=2025-01-01^&end_date=2025-12-31" "" "200"
call :test_api "获取库存水位" "GET" "/api/v1/datacenter/products/inventory-level" "" "200"
call :test_api "获取进货策略" "GET" "/api/v1/datacenter/products/purchase-strategy?days=30" "" "200"
call :test_api "获取低库存商品" "GET" "/api/v1/datacenter/products/low-stock" "" "200"
call :test_api "获取库存汇总" "GET" "/api/v1/datacenter/products/inventory-summary" "" "200"

REM 对比分析测试
echo.
echo ^>^>^> 对比分析模块测试 ^<^<^<^<^<^
echo.

call :test_api "同比分析" "GET" "/api/v1/datacenter/compare/yoy?start_date=2025-01-01^&end_date=2025-12-31" "" "200"
call :test_api "环比分析" "GET" "/api/v1/datacenter/compare/mom?start_date=2025-01-01^&end_date=2025-12-31" "" "200"
call :test_api "平台对比" "GET" "/api/v1/datacenter/compare/platform?start_date=2025-01-01^&end_date=2025-12-31" "" "200"

REM 预警管理测试
echo.
echo ^>^>^> 预警管理模块测试 ^<^<^<^<^<^
echo.

call :test_api "获取预警类型" "GET" "/api/v1/datacenter/alerts/types" "" "200"
call :test_api "获取预警级别" "GET" "/api/v1/datacenter/alerts/levels" "" "200"
call :test_api "获取预警汇总" "GET" "/api/v1/datacenter/alerts/summary" "" "200"
call :test_api "获取预警规则列表" "GET" "/api/v1/datacenter/alerts/rules" "" "200"
call :test_api "获取预警记录列表" "GET" "/api/v1/datacenter/alerts/records" "" "200"

REM 测试结果汇总
echo.
echo ========================================
echo     测试结果汇总
echo ========================================
echo 通过: %PASS_COUNT%
echo 失败: %FAIL_COUNT%
set /a TOTAL=%PASS_COUNT%+%FAIL_COUNT%
echo 总计: %TOTAL%
echo ========================================

if %FAIL_COUNT%==0 (
    echo 所有测试通过!
    exit /b 0
) else (
    echo 有测试失败!
    exit /b 1
)

REM 清理临时文件
del response.txt status.txt 2>nul
