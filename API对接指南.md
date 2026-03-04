# OURERP 外部数据对接指南

## 📋 目录
1. [快速开始](#快速开始)
2. [认证方式](#认证方式)
3. [API 接口说明](#api-接口说明)
4. [数据格式](#数据格式)
5. [代码示例](#代码示例)
6. [常见问题](#常见问题)

---

## 🚀 快速开始

### 1. 获取访问令牌

```bash
# 登录获取 Token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "root@ourerp.com",
    "password": "root123456"
  }'

# 响应示例
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 3,
    "email": "root@ourerp.com",
    "name": "Root"
  }
}
```

### 2. 切换账套

```bash
# 切换到目标账套（获取账套ID）
curl -X POST http://localhost:8080/api/v1/tenants/switch \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": 1
  }'
```

### 3. 同步订单

```bash
# 推送订单数据
curl -X POST http://localhost:8080/api/v1/sync/orders \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-Tenant-ID: 1" \
  -H "Content-Type: application/json" \
  -d @orders.json
```

---

## 🔐 认证方式

### 请求头设置

```http
Authorization: Bearer YOUR_TOKEN
X-Tenant-ID: 1
Content-Type: application/json
```

| 参数 | 说明 | 示例 |
|------|------|------|
| Authorization | JWT 访问令牌 | `Bearer eyJhbGci...` |
| X-Tenant-ID | 账套ID | `1` |
| Content-Type | 内容类型 | `application/json` |

---

## 📡 API 接口说明

### 1. 批量同步订单

**接口地址：** `POST /api/v1/sync/orders`

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| source | string | 否 | 数据来源标识 |
| orders | array | 是 | 订单列表 |

**响应示例：**

```json
{
  "code": "SUCCESS",
  "message": "success",
  "data": {
    "success": true,
    "total": 10,
    "created": 8,
    "updated": 2,
    "failed": 0,
    "errors": [],
    "process_time": 156
  }
}
```

### 2. 查询同步统计

**接口地址：** `GET /api/v1/sync/statistics`

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| start_time | string | 否 | 开始时间 (RFC3339) |
| end_time | string | 否 | 结束时间 (RFC3339) |

**示例：**
```
GET /api/v1/sync/statistics?start_time=2026-03-01T00:00:00Z&end_time=2026-03-04T23:59:59Z
```

---

## 📦 数据格式

### 订单对象 (ExternalOrder)

```json
{
  "platform": "taobao",
  "platform_order_id": "TB123456789",
  "shop_id": 1,
  "status": "paid",
  "order_time": "2026-03-04T10:00:00Z",
  "pay_time": "2026-03-04T10:05:00Z",
  "total_amount": 599.00,
  "pay_amount": 549.00,
  "buyer_nick": "买家昵称",
  "buyer_message": "买家留言",
  "receiver_name": "收货人姓名",
  "receiver_phone": "13800138000",
  "receiver_state": "浙江省",
  "receiver_city": "杭州市",
  "receiver_district": "余杭区",
  "receiver_street": "文一西路",
  "receiver_address": "浙江省杭州市余杭区文一西路",
  "logistics_company": "顺丰速运",
  "logistics_no": "SF1234567890",
  "items": [
    {
      "sku_id": "12345",
      "sku_name": "商品名称",
      "quantity": 2,
      "price": 299.00,
      "total_amount": 598.00
    }
  ]
}
```

### 字段说明

#### 订单基本信息

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| platform | string | 是 | 平台类型 (taobao/tmall/jd/douyin/kuaishou/pdd/custom) |
| platform_order_id | string | 是 | 平台订单号（唯一标识） |
| shop_id | number | 否 | 店铺ID（不填则自动匹配或创建） |
| status | string | 否 | 订单状态 |
| order_time | string | 否 | 下单时间 (RFC3339格式) |
| pay_time | string | 否 | 支付时间 (RFC3339格式) |

#### 金额信息

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| total_amount | number | 是 | 订单总金额 |
| pay_amount | number | 是 | 实付金额 |

#### 买家信息

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| buyer_nick | string | 是 | 买家昵称 |
| buyer_message | string | 否 | 买家留言 |

#### 收货信息

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| receiver_name | string | 是 | 收货人姓名 |
| receiver_phone | string | 是 | 收货人电话 |
| receiver_address | string | 是* | 完整收货地址 |
| receiver_state | string | 否 | 省份 |
| receiver_city | string | 否 | 城市 |
| receiver_district | string | 否 | 区县 |
| receiver_street | string | 否 | 街道 |

* 如果提供了省市区街道字段，会自动组装完整地址

#### 物流信息

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| logistics_company | string | 否 | 物流公司 |
| logistics_no | string | 否 | 物流单号 |

#### 订单明细 (items)

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sku_id | string | 是 | 商品SKU ID |
| sku_name | string | 是 | 商品名称 |
| quantity | number | 是 | 数量 |
| price | number | 是 | 单价 |

---

## 💻 代码示例

### Python 示例

```python
import requests
import json
from datetime import datetime

class OURERPClient:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url
        self.token = None
        self.tenant_id = None

    def login(self, email, password):
        """登录获取 Token"""
        response = requests.post(
            f"{self.base_url}/api/v1/auth/login",
            json={"email": email, "password": password}
        )
        data = response.json()
        self.token = data["token"]
        return self

    def switch_tenant(self, tenant_id):
        """切换账套"""
        headers = self._get_headers()
        requests.post(
            f"{self.base_url}/api/v1/tenants/switch",
            headers=headers,
            json={"tenant_id": tenant_id}
        )
        self.tenant_id = tenant_id
        return self

    def sync_orders(self, orders, source="api"):
        """同步订单"""
        headers = self._get_headers()
        response = requests.post(
            f"{self.base_url}/api/v1/sync/orders",
            headers=headers,
            json={
                "source": source,
                "orders": orders
            }
        )
        return response.json()

    def _get_headers(self):
        """构建请求头"""
        headers = {
            "Authorization": f"Bearer {self.token}",
            "Content-Type": "application/json"
        }
        if self.tenant_id:
            headers["X-Tenant-ID"] = str(self.tenant_id)
        return headers

# 使用示例
client = OURERPClient()

# 1. 登录
client.login("root@ourerp.com", "root123456")

# 2. 切换账套
client.switch_tenant(1)

# 3. 同步订单
orders = [{
    "platform": "taobao",
    "platform_order_id": "TB" + str(int(datetime.now().timestamp())),
    "status": "paid",
    "pay_time": datetime.now().isoformat() + "Z",
    "total_amount": 299.00,
    "pay_amount": 299.00,
    "buyer_nick": "测试买家",
    "receiver_name": "张三",
    "receiver_phone": "13800138000",
    "receiver_address": "浙江省杭州市余杭区",
    "items": [{
        "sku_id": "1",
        "sku_name": "测试商品",
        "quantity": 1,
        "price": 299.00
    }]
}]

result = client.sync_orders(orders, source="python_api")
print(f"同步结果: {result}")
```

### Java 示例

```java
import okhttp3.*;
import com.google.gson.Gson;
import java.util.*;

public class OURERPClient {
    private final String baseUrl;
    private String token;
    private Long tenantId;

    public OURERPClient(String baseUrl) {
        this.baseUrl = baseUrl;
    }

    public void login(String email, String password) throws Exception {
        OkHttpClient client = new OkHttpClient();

        Map<String, String> credentials = new HashMap<>();
        credentials.put("email", email);
        credentials.put("password", password);

        RequestBody body = RequestBody.create(
            MediaType.parse("application/json"),
            new Gson().toJson(credentials)
        );

        Request request = new Request.Builder()
            .url(baseUrl + "/api/v1/auth/login")
            .post(body)
            .build();

        try (Response response = client.newCall(request).execute()) {
            Map<String, Object> data = new Gson().fromJson(
                response.body().string(),
                Map.class
            );
            this.token = (String) data.get("token");
        }
    }

    public void switchTenant(long tenantId) throws Exception {
        OkHttpClient client = new OkHttpClient();

        Map<String, Object> params = new HashMap<>();
        params.put("tenant_id", tenantId);

        RequestBody body = RequestBody.create(
            MediaType.parse("application/json"),
            new Gson().toJson(params)
        );

        Request request = new Request.Builder()
            .url(baseUrl + "/api/v1/tenants/switch")
            .post(body)
            .addHeader("Authorization", "Bearer " + token)
            .build();

        client.newCall(request).execute();
        this.tenantId = tenantId;
    }

    public Map<String, Object> syncOrders(
        List<Map<String, Object>> orders,
        String source
    ) throws Exception {
        OkHttpClient client = new OkHttpClient();

        Map<String, Object> requestBody = new HashMap<>();
        requestBody.put("source", source);
        requestBody.put("orders", orders);

        RequestBody body = RequestBody.create(
            MediaType.parse("application/json"),
            new Gson().toJson(requestBody)
        );

        Request request = new Request.Builder()
            .url(baseUrl + "/api/v1/sync/orders")
            .post(body)
            .addHeader("Authorization", "Bearer " + token)
            .addHeader("X-Tenant-ID", String.valueOf(tenantId))
            .build();

        try (Response response = client.newCall(request).execute()) {
            return new Gson().fromJson(
                response.body().string(),
                Map.class
            );
        }
    }
}
```

### JavaScript/Node.js 示例

```javascript
class OURERPClient {
    constructor(baseUrl = 'http://localhost:8080') {
        this.baseUrl = baseUrl;
        this.token = null;
        this.tenantId = null;
    }

    async login(email, password) {
        const response = await fetch(`${this.baseUrl}/api/v1/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        const data = await response.json();
        this.token = data.token;
        return this;
    }

    async switchTenant(tenantId) {
        await fetch(`${this.baseUrl}/api/v1/tenants/switch`, {
            method: 'POST',
            headers: this._getHeaders(),
            body: JSON.stringify({ tenant_id: tenantId })
        });
        this.tenantId = tenantId;
        return this;
    }

    async syncOrders(orders, source = 'api') {
        const response = await fetch(`${this.baseUrl}/api/v1/sync/orders`, {
            method: 'POST',
            headers: this._getHeaders(),
            body: JSON.stringify({ source, orders })
        });
        return await response.json();
    }

    _getHeaders() {
        const headers = {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.token}`
        };
        if (this.tenantId) {
            headers['X-Tenant-ID'] = this.tenantId;
        }
        return headers;
    }
}

// 使用示例
(async () => {
    const client = new OURERPClient();

    await client.login('root@ourerp.com', 'root123456');
    await client.switchTenant(1);

    const orders = [{
        platform: 'taobao',
        platform_order_id: 'TB' + Date.now(),
        status: 'paid',
        pay_time: new Date().toISOString(),
        total_amount: 299.00,
        pay_amount: 299.00,
        buyer_nick: '测试买家',
        receiver_name: '张三',
        receiver_phone: '13800138000',
        receiver_address: '浙江省杭州市余杭区',
        items: [{
            sku_id: '1',
            sku_name: '测试商品',
            quantity: 1,
            price: 299.00
        }]
    }];

    const result = await client.syncOrders(orders, 'nodejs_api');
    console.log('同步结果:', result);
})();
```

### PHP 示例

```php
<?php
class OURERPClient {
    private $baseUrl;
    private $token;
    private $tenantId;

    public function __construct($baseUrl = 'http://localhost:8080') {
        $this->baseUrl = $baseUrl;
    }

    public function login($email, $password) {
        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $this->baseUrl . '/api/v1/auth/login');
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, ['Content-Type: application/json']);
        curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode([
            'email' => $email,
            'password' => $password
        ]));
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);

        $response = curl_exec($ch);
        curl_close($ch);

        $data = json_decode($response, true);
        $this->token = $data['token'];
        return $this;
    }

    public function switchTenant($tenantId) {
        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $this->baseUrl . '/api/v1/tenants/switch');
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, $this->_getHeaders());
        curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode([
            'tenant_id' => $tenantId
        ]));
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);

        curl_exec($ch);
        curl_close($ch);

        $this->tenantId = $tenantId;
        return $this;
    }

    public function syncOrders($orders, $source = 'api') {
        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $this->baseUrl . '/api/v1/sync/orders');
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, $this->getHeaders());
        curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode([
            'source' => $source,
            'orders' => $orders
        ]));
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);

        $response = curl_exec($ch);
        curl_close($ch);

        return json_decode($response, true);
    }

    private function _getHeaders() {
        $headers = [
            'Content-Type: application/json',
            'Authorization: Bearer ' . $this->token
        ];

        if ($this->tenantId) {
            $headers[] = 'X-Tenant-ID: ' . $this->tenantId;
        }

        return $headers;
    }
}

// 使用示例
$client = new OURERPClient();
$client->login('root@ourerp.com', 'root123456');
$client->switchTenant(1);

$orders = [[
    'platform' => 'taobao',
    'platform_order_id' => 'TB' . time(),
    'status' => 'paid',
    'pay_time' => date('c'),
    'total_amount' => 299.00,
    'pay_amount' => 299.00,
    'buyer_nick' => '测试买家',
    'receiver_name' => '张三',
    'receiver_phone' => '13800138000',
    'receiver_address' => '浙江省杭州市余杭区',
    'items' => [[
        'sku_id' => '1',
        'sku_name' => '测试商品',
        'quantity' => 1,
        'price' => 299.00
    ]]
]];

$result = $client->syncOrders($orders, 'php_api');
print_r($result);
?>
```

---

## 🔧 订单状态映射

### 平台状态 → 系统状态

| 平台状态 | 系统状态 | 说明 |
|---------|---------|------|
| pending_payment | 1 | 待付款 |
| paid / pending_ship | 2 | 待发货 |
| shipped | 3 | 已发货 |
| completed | 4 | 已完成 |
| cancelled | 5 | 已取消 |

### 支持的平台

| 平台代码 | 平台名称 |
|---------|---------|
| taobao | 淘宝 |
| tmall | 天猫 |
| jd | 京东 |
| douyin | 抖音 |
| kuaishou | 快手 |
| pdd | 拼多多 |
| custom | 自定义 |

---

## ❓ 常见问题

### 1. 如何处理 Token 过期？

**错误响应：**
```json
{
  "code": "UNAUTHORIZED",
  "message": "token已过期"
}
```

**解决方案：** 重新调用登录接口获取新 Token

### 2. 如何获取账套 ID？

```bash
# 获取我的账套列表
curl http://localhost:8080/api/v1/tenants/my \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 3. 订单重复推送会怎样？

系统会根据 `platform_order_id` 去重：
- 如果订单已存在，会更新订单信息
- 不会创建重复订单

### 4. 如何处理大批量订单？

建议分批推送，每批 100-500 个订单：

```python
def sync_orders_batch(client, orders, batch_size=200):
    for i in range(0, len(orders), batch_size):
        batch = orders[i:i+batch_size]
        result = client.sync_orders(batch)
        print(f"批次 {i//batch_size + 1}: {result}")
```

### 5. 如何验证对接成功？

```bash
# 查询订单数量
curl http://localhost:8080/api/v1/orders?page=1&size=10 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-Tenant-ID: 1"
```

---

## 📞 技术支持

如有问题，请查看：
- 后端日志: `/tmp/backend-sync-ready.log`
- API 文档: `/docs/swagger.md`
- GitHub Issues: `https://github.com/MorantHP/OURERP/issues`
