package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const baseURL = "http://localhost:8080/api/v1"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "generate":
		generateOrders()
	case "realtime":
		startRealtime()
	case "list":
		listOrders()
	case "help":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("使用方法:")
	fmt.Println("  go run cmd/mock/main.go generate    # 生成100个测试订单")
	fmt.Println("  go run cmd/mock/main.go realtime    # 启动实时订单生成")
	fmt.Println("  go run cmd/mock/main.go list        # 查看模拟订单列表")
}

func generateOrders() {
	data := map[string]interface{}{
		"count":     100,
		"platform":  "taobao",
		"shop_id":   1,
	}
	
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(baseURL+"/mock/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	fmt.Printf("生成结果: %+v\n", result)
}

func startRealtime() {
	resp, err := http.Post(baseURL+"/mock/realtime/start", "application/json", nil)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	fmt.Println("实时订单生成已启动")
	fmt.Println("按 Ctrl+C 停止")
	
	// 保持运行
	for {
		time.Sleep(1 * time.Second)
	}
}

func listOrders() {
	resp, err := http.Get(baseURL + "/mock/orders?limit=10")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	fmt.Printf("总订单数: %.0f\n", result["total"])
	fmt.Println("\n前10个订单:")
	
	orders, _ := result["orders"].([]interface{})
	for i, o := range orders {
		order := o.(map[string]interface{})
		fmt.Printf("%d. %s - %s - ¥%.2f - %s\n", 
			i+1,
			order["order_no"],
			order["buyer_nick"],
			order["total_amount"],
			getStatusText(order["status"].(float64)),
		)
	}
}

func getStatusText(status float64) string {
	switch int(status) {
	case 100:
		return "待付款"
	case 200:
		return "待审核"
	case 300:
		return "待发货"
	case 400:
		return "已发货"
	case 500:
		return "已签收"
	case 600:
		return "已完成"
	default:
		return "未知"
	}
}