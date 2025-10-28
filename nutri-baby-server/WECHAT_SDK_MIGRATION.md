# 微信 SDK 迁移文档

## 概述

本次重构将项目中自行实现的微信 API 调用代码替换为成熟的第三方 SDK `github.com/silenceper/wechat/v2`，提高代码质量和可维护性。

## 迁移时间

**完成时间**: 2025-10-25

## 变更内容

### 1. 添加 SDK 依赖

**依赖包**: `github.com/silenceper/wechat/v2 v2.1.9`

**新增依赖**:
- `github.com/silenceper/wechat/v2` - 微信 SDK 核心包
- `github.com/bradfitz/gomemcache` - Memcache 缓存支持
- `github.com/fatih/structs` - 结构体工具
- `github.com/go-redis/redis/v8` - Redis 客户端
- `github.com/tidwall/gjson` - JSON 解析工具

### 2. 新建基础设施层微信客户端

**文件**: `internal/infrastructure/wechat/wechat.go`

**功能**:
- 封装微信 SDK 实例
- 配置小程序 AppID 和 AppSecret
- 使用 Redis 作为缓存后端存储 access_token
- 在初始化时就配置好小程序实例,避免运行时空指针异常

**主要方法**:
```go
// NewClient 创建微信客户端
func NewClient(cfg *config.Config, redisClient *redis.Client) *Client

// GetMiniProgram 获取小程序实例(已配置好,直接使用)
func (c *Client) GetMiniProgram() *miniprogram.MiniProgram
```

**关键设计**:
```go
type Client struct {
	wechat      *wechat.Wechat
	miniProgram *miniprogram.MiniProgram  // 保存已配置的实例
}

// 在 NewClient 时就配置好
mini := wc.GetMiniProgram(miniCfg)
return &Client{
	wechat:      wc,
	miniProgram: mini,  // 保存配置好的实例
}
```

### 3. 重构 AuthService (认证服务)

**文件**: `internal/application/service/auth_service.go`

**主要变更**:

#### 之前 (自行实现)
```go
// 手动构造 HTTP 请求
func (s *AuthService) getWechatSession(code string) (*WechatSession, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.cfg.Wechat.AppID,
		s.cfg.Wechat.AppSecret,
		code,
	)
	resp, err := http.Get(url)
	// ... 手动处理响应
}
```

#### 之后 (使用 SDK)
```go
// 使用 SDK 调用
func (s *AuthService) WechatLogin(ctx context.Context, req *dto.WechatLoginRequest) (*dto.LoginResponse, error) {
	// 直接获取已配置的小程序实例
	miniProgram := s.wechatClient.GetMiniProgram()
	auth := miniProgram.GetAuth()

	session, err := auth.Code2SessionContext(ctx, req.Code)
	// SDK 自动处理 HTTP 请求、错误处理、重试等
}
```

**优势**:
- ✅ 不再需要手动构造 URL 和处理 HTTP 请求
- ✅ SDK 内置错误处理和重试机制
- ✅ 支持 Context 传递,便于超时控制
- ✅ 代码更简洁,可读性更强

### 4. 重构 WechatService (微信服务)

**文件**: `internal/application/service/wechat_service.go`

**主要变更**:

#### 之前 (自行实现)
- 手动管理 access_token 缓存 (使用内存 + 互斥锁)
- 手动构造订阅消息请求体
- 手动调用微信 API
- 需要处理 token 过期、刷新等逻辑
- 代码约 **299 行**

#### 之后 (使用 SDK)
```go
func (s *WechatService) SendSubscribeMessage(
	openid string,
	templateID string,
	data map[string]interface{},
	page string,
	miniprogramState string,
) error {
	// 获取订阅消息实例
	miniProgram := s.wechatClient.GetMiniProgram().GetMiniProgram(nil)
	subscribeService := miniProgram.GetSubscribe()

	// 格式化数据
	formattedData := make(map[string]*subscribe.DataItem)
	for k, v := range data {
		formattedData[k] = &subscribe.DataItem{Value: v}
	}

	// 构造消息
	msg := &subscribe.Message{
		ToUser:           openid,
		TemplateID:       templateID,
		Page:             page,
		Data:             formattedData,
		MiniprogramState: miniprogramState,
		Lang:             "zh_CN",
	}

	// 发送订阅消息
	return subscribeService.Send(msg)
}
```
- 代码约 **83 行** (减少 72%)

**优势**:
- ✅ SDK 自动管理 access_token,使用 Redis 缓存
- ✅ 自动处理 token 过期和刷新
- ✅ 不再需要手动实现双重检查锁
- ✅ 代码量大幅减少,更易维护
- ✅ SDK 内置日志和错误处理

**返回值变更**:
- 之前: `(*WechatSubscribeMessageResult, error)`
- 之后: `error`

这个变更需要同步修改 `subscribe_service.go` 中的调用代码。

### 5. 更新 SubscribeService

**文件**: `internal/application/service/subscribe_service.go`

**变更**:
```go
// 之前
result, err := s.wechatService.SendSubscribeMessage(...)
if err != nil {
	log.SendStatus = "failed"
	log.ErrMsg = err.Error()
} else {
	log.SendStatus = "success"
	log.ErrCode = result.ErrCode  // ❌ 不再返回 result
	log.ErrMsg = result.ErrMsg
}

// 之后
err := s.wechatService.SendSubscribeMessage(...)
if err != nil {
	log.SendStatus = "failed"
	log.ErrMsg = err.Error()
} else {
	log.SendStatus = "success"
	log.SendTime = &now
}
```

### 6. 更新 Wire 依赖注入配置

**文件**: `wire/wire.go`

**变更**:
```go
wire.Build(
	// 基础设施层
	logger.NewLogger,
	persistence.NewDatabase,
	persistence.NewRedis,    // ✅ 启用 Redis
	wechat.NewClient,        // ✅ 新增微信客户端

	// ... 其他配置
)
```

**重新生成 Wire 代码**:
```bash
cd wire && wire
```

## 技术优势

### 1. 代码质量提升
- ❌ **之前**: 自行实现,容易出错
- ✅ **之后**: 使用成熟 SDK,经过大量项目验证

### 2. access_token 管理
- ❌ **之前**: 内存缓存 + 互斥锁,服务重启丢失
- ✅ **之后**: Redis 缓存,支持分布式部署

### 3. 错误处理
- ❌ **之前**: 需要手动处理各种微信 API 错误码
- ✅ **之后**: SDK 统一处理,提供友好错误信息

### 4. 可维护性
- ❌ **之前**: 代码量大,逻辑复杂
- ✅ **之后**: 代码简洁,易于理解和维护

### 5. 扩展性
- ❌ **之前**: 新增微信 API 需要手动实现
- ✅ **之后**: SDK 提供完整 API 支持,开箱即用

## SDK 功能支持

`github.com/silenceper/wechat/v2` 支持以下微信能力:

### 小程序 (已使用)
- ✅ **登录认证** - `miniprogram.GetAuth()`
  - Code2Session (登录凭证校验)
  - GetPhoneNumber (获取手机号)

- ✅ **订阅消息** - `miniprogram.GetSubscribe()`
  - Send (发送订阅消息)
  - GetTemplateList (获取模板列表)
  - AddTemplate (添加模板)

### 小程序 (可扩展使用)
- 📋 **客服消息** - `miniprogram.GetCustomerMessage()`
- 📋 **小程序码** - `miniprogram.GetQRCode()`
- 📋 **内容安全** - `miniprogram.GetContentSecurity()`
- 📋 **数据分析** - `miniprogram.GetAnalysis()`
- 📋 **URL Scheme** - `miniprogram.GetSURLScheme()`
- 📋 **URL Link** - `miniprogram.GetURLLink()`

### 其他平台 (SDK 支持)
- 📋 公众号 (Official Account)
- 📋 企业微信 (Work WeChat)
- 📋 微信支付 (WeChat Pay)
- 📋 开放平台 (Open Platform)

## 测试建议

### 1. 登录功能测试
```bash
# 测试小程序登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "code": "xxx",
    "nickName": "测试用户",
    "avatarUrl": "http://..."
  }'
```

### 2. 订阅消息测试
```bash
# 测试发送订阅消息
curl -X POST http://localhost:8080/api/v1/subscribe/send \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "openid": "xxx",
    "templateType": "feeding_reminder",
    "data": {
      "thing1": {"value": "喂养提醒"},
      "time2": {"value": "2025-10-25 12:00"}
    },
    "page": "pages/index/index"
  }'
```

### 3. Redis 缓存验证
```bash
# 检查 access_token 是否存储在 Redis
redis-cli
> KEYS *access_token*
```

## 注意事项

### 1. 配置文件
确保 `config/config.yaml` 中配置正确:
```yaml
wechat:
  app_id: wxf47340979046b474
  app_secret: c5b5f88848865bc3b0ac9ba5aa1f477e

redis:
  host: 101.200.47.93
  port: 6379
  password: ""
  db: 0
```

### 2. Redis 连接
- SDK 会自动将 access_token 存储到 Redis
- 默认过期时间为微信返回的 `expires_in` (通常 7200 秒)
- 无需手动管理 token 刷新

### 3. 兼容性
- 所有旧接口保持兼容
- 仅内部实现改为 SDK
- API 响应格式不变

## 回滚方案

如遇问题需要回滚,执行以下步骤:

1. **恢复旧代码**:
```bash
git checkout HEAD~1 -- internal/application/service/auth_service.go
git checkout HEAD~1 -- internal/application/service/wechat_service.go
git checkout HEAD~1 -- internal/application/service/subscribe_service.go
```

2. **删除新文件**:
```bash
rm internal/infrastructure/wechat/wechat.go
```

3. **恢复 Wire 配置**:
```bash
git checkout HEAD~1 -- wire/wire.go
cd wire && wire
```

4. **移除依赖**:
```bash
go mod tidy
```

## 相关链接

- **SDK 官方仓库**: https://github.com/silenceper/wechat
- **SDK 文档**: https://godoc.org/github.com/silenceper/wechat/v2
- **微信小程序官方文档**: https://developers.weixin.qq.com/miniprogram/dev/

## 总结

本次迁移成功将自行实现的微信 API 调用替换为成熟的 SDK,显著提升了:
- ✅ 代码质量和可维护性
- ✅ access_token 管理的可靠性
- ✅ 分布式部署的支持能力
- ✅ 未来功能扩展的便利性

**代码量对比**:
- `auth_service.go`: 删除 44 行手动实现代码
- `wechat_service.go`: 从 299 行减少到 83 行 (减少 72%)
- 新增 `wechat/wechat.go`: 49 行基础设施封装

**总净减少**: 约 211 行代码 ✨
