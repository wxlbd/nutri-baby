#!/bin/bash

# AI分析API测试脚本
# 支持多种日期格式测试

API_BASE_URL="http://localhost:8080/v1"
BABY_ID=1
TOKEN="your-jwt-token-here"

echo "=== AI分析API测试 ==="
echo ""

# 测试1: 创建分析任务 - 使用 YYYY-MM-DD 格式
echo "测试1: 创建分析任务（YYYY-MM-DD 格式）"
curl -X POST "${API_BASE_URL}/ai-analysis" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "baby_id": 1,
    "analysis_type": "feeding",
    "start_date": "2025-11-01",
    "end_date": "2025-11-08"
  }' | jq '.'
echo ""
echo "---"
echo ""

# 测试2: 创建分析任务 - 使用 RFC3339 格式
echo "测试2: 创建分析任务（RFC3339 格式）"
curl -X POST "${API_BASE_URL}/ai-analysis" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "baby_id": 1,
    "analysis_type": "sleep",
    "start_date": "2025-11-01T00:00:00Z",
    "end_date": "2025-11-08T23:59:59Z"
  }' | jq '.'
echo ""
echo "---"
echo ""

# 测试3: 批量分析
echo "测试3: 批量分析"
curl -X POST "${API_BASE_URL}/ai-analysis/batch?baby_id=1&start_date=2025-11-01&end_date=2025-11-08" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" | jq '.'
echo ""
echo "---"
echo ""

# 测试4: 获取每日建议
echo "测试4: 获取每日建议"
curl -X GET "${API_BASE_URL}/ai-analysis/daily-tips?baby_id=1&date=2025-11-08" \
  -H "Authorization: Bearer ${TOKEN}" | jq '.'
echo ""
echo "---"
echo ""

# 测试5: 获取分析统计
echo "测试5: 获取分析统计"
curl -X GET "${API_BASE_URL}/ai-analysis/stats?baby_id=1" \
  -H "Authorization: Bearer ${TOKEN}" | jq '.'
echo ""
