# AI分析功能 - 使用指南

## 快速开始

### 1️⃣ 生成JWT Token

```bash
cd nutri-baby-server
go run generate_token.go \
  -openid "om8hB12mqHOp1BiTf3KZ_ew8eWH4" \
  -secret "your-secret-key-change-in-production" \
  -expire 72
```

### 2️⃣ 创建分析任务

```bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl -X POST http://localhost:8080/v1/ai-analysis \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "baby_id": 1,
    "analysis_type": "feeding",
    "start_date": "2025-11-01",
    "end_date": "2025-11-08"
  }'
```

### 3️⃣ 查看结果

```bash
curl -X GET http://localhost:8080/v1/ai-analysis/1 \
  -H "Authorization: Bearer $TOKEN"
```

## 📚 文档导航

| 文档 | 用途 | 读者 |
|------|------|------|
| [AI_ANALYSIS_QUICK_START.md](./AI_ANALYSIS_QUICK_START.md) | **快速入门** | 所有人 |
| [APIFOX_GUIDE.md](./APIFOX_GUIDE.md) | API测试指南 | 测试者 |
| [DATE_FORMAT_GUIDE.md](./DATE_FORMAT_GUIDE.md) | 日期格式详解 | 前端开发者 |
| [API_ANALYSIS_API.md](./AI_ANALYSIS_API.md) | 完整API参考 | 后端开发者 |
| [JWT_AUTH_FIX_REPORT.md](./JWT_AUTH_FIX_REPORT.md) | 技术实现报告 | 架构师 |
| [TEST_REPORT_20251112.md](./TEST_REPORT_20251112.md) | 测试报告 | QA |
| [COMPLETION_SUMMARY.md](./COMPLETION_SUMMARY.md) | 项目总结 | 项目经理 |

## 🎯 支持的功能

✅ **5种分析类型**
- 喂养分析 (feeding)
- 睡眠分析 (sleep)
- 成长分析 (growth)
- 健康分析 (health)
- 行为分析 (behavior)

✅ **5种日期格式**
- 简单日期: `2025-11-01`
- 日期时间: `2025-11-01 10:30:00`
- ISO 8601: `2025-11-01T10:30:00`
- RFC3339: `2025-11-01T10:30:00Z`
- RFC3339+TZ: `2025-11-01T10:30:00+08:00`

✅ **7个API端点**
- POST `/v1/ai-analysis` - 创建分析
- GET `/v1/ai-analysis/{id}` - 获取结果
- GET `/v1/ai-analysis/baby/{babyId}/latest` - 获取最新
- GET `/v1/ai-analysis/baby/{babyId}/history` - 获取统计
- POST `/v1/ai-analysis/batch` - 批量分析
- GET `/v1/ai-analysis/daily-tips/{babyId}` - 获取建议
- POST `/v1/ai-analysis/daily-tips/{babyId}/generate` - 生成建议

## ⚠️ 常见问题

### 401 Unauthorized
**原因**: Token过期或无效
**解决**: 重新生成token

### 1003 获取宝宝信息失败
**原因**: baby_id不存在
**解决**: 使用正确的baby_id，或先创建宝宝

### 日期格式错误
**原因**: 日期格式不支持
**解决**: 使用支持的日期格式之一

## 🔧 工具

| 工具 | 用途 |
|------|------|
| `generate_token.go` | 生成JWT token |
| `test_ai_analysis.sh` | 测试API所有端点 |
| APIfox | 图形化API测试 |
| curl | 命令行API测试 |

## 📊 测试状态

- ✅ JWT认证: 通过
- ✅ 日期格式: 5/5通过
- ✅ 分析类型: 5/5通过
- ✅ 错误处理: 4/4通过
- ✅ API响应: 正常
- ✅ 性能指标: 达标

## 🚀 部署准备

- ✅ 代码编译通过
- ✅ 所有测试通过
- ✅ 文档完整
- ✅ 可以上线

## 💡 最佳实践

1. **日期处理**: 使用简单格式 `YYYY-MM-DD`
2. **Token管理**: 定期刷新token
3. **错误处理**: 检查错误码而不是消息
4. **缓存**: 利用分析结果缓存
5. **权限**: 实现完整的权限检查

## 📞 支持

遇到问题？
1. 查看相关文档
2. 查看代码注释
3. 运行测试脚本
4. 查看日志文件

## 📝 文件清单

```
nutri-baby-server/
├── AI_ANALYSIS_API.md              # API完整文档
├── AI_ANALYSIS_QUICK_START.md      # 快速开始指南
├── APIFOX_GUIDE.md                 # APIfox测试指南
├── COMPLETION_SUMMARY.md           # 项目完成总结
├── DATE_FORMAT_GUIDE.md            # 日期格式详解
├── JWT_AUTH_FIX_REPORT.md          # 认证问题报告
├── TEST_REPORT_20251112.md         # 测试报告
├── README_AI_ANALYSIS.md           # 本文件
├── generate_token.go               # Token生成工具
└── test_ai_analysis.sh             # 测试脚本
```

---

**最后更新**: 2025-11-12
**状态**: ✅ 准备上线
