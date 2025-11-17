#!/bin/bash

# 后台任务处理完整演示脚本
# 演示AI分析任务的三种处理方式

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 配置
API_BASE="http://localhost:8080"
TOKEN="${1:-}"

if [ -z "$TOKEN" ]; then
    echo -e "${RED}错误: 需要提供JWT token${NC}"
    echo "用法: $0 <jwt_token>"
    echo ""
    echo "获取token的方法:"
    echo "  go run generate_token.go -openid 'om8hB12mqHOp1BiTf3KZ_ew8eWH4' -expire 72"
    exit 1
fi

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}AI分析后台任务处理演示${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# ============================================
# 方式1: 创建任务 + 手动触发处理
# ============================================
echo -e "${YELLOW}【方式1】创建任务 + 手动触发处理 (延迟处理)${NC}"
echo ""

echo -e "${GREEN}步骤1: 创建分析任务（立即返回，状态为pending）${NC}"
RESPONSE=$(curl -s -X POST "$API_BASE/v1/ai-analysis" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "baby_id": 1,
    "analysis_type": "feeding",
    "start_date": "2025-11-01",
    "end_date": "2025-11-08"
  }')

ANALYSIS_ID=$(echo "$RESPONSE" | grep -o '"analysis_id":"[^"]*"' | cut -d'"' -f4)
echo "响应:"
echo "$RESPONSE" | jq '.'
echo ""

if [ -z "$ANALYSIS_ID" ]; then
    echo -e "${RED}无法获取任务ID，请检查token和宝宝ID${NC}"
    echo ""
    # 继续演示其他方式
else
    echo -e "${GREEN}✓ 任务已创建，ID: $ANALYSIS_ID，状态: pending${NC}"
    echo ""

    echo -e "${GREEN}步骤2: 检查任务状态（应该是pending）${NC}"
    curl -s -X GET "$API_BASE/v1/ai-analysis/$ANALYSIS_ID" \
      -H "Authorization: Bearer $TOKEN" | jq '.data | {analysis_id, status}'
    echo ""

    echo -e "${GREEN}步骤3: 手动触发后台任务处理${NC}"
    echo "执行: curl -X POST $API_BASE/v1/jobs/process-pending-analyses -H \"Authorization: Bearer \$TOKEN\""
    PROCESS_RESPONSE=$(curl -s -X POST "$API_BASE/v1/jobs/process-pending-analyses" \
      -H "Authorization: Bearer $TOKEN")
    echo "响应:"
    echo "$PROCESS_RESPONSE" | jq '.'
    echo ""

    echo -e "${GREEN}步骤4: 等待后处理任务完成...${NC}"
    sleep 2

    echo -e "${GREEN}步骤5: 再次检查任务状态（应该是completed）${NC}"
    curl -s -X GET "$API_BASE/v1/ai-analysis/$ANALYSIS_ID" \
      -H "Authorization: Bearer $TOKEN" | jq '.data | {analysis_id, status}'
    echo ""
fi

# ============================================
# 方式2: 批量分析（立即处理）
# ============================================
echo -e "${YELLOW}【方式2】批量分析 (立即处理，推荐)${NC}"
echo ""

echo -e "${GREEN}步骤1: 调用批量分析端点${NC}"
echo "执行: curl -X POST \"$API_BASE/v1/ai-analysis/batch?baby_id=1&start_date=2025-11-01&end_date=2025-11-08\" -H \"Authorization: Bearer \$TOKEN\""
echo ""

BATCH_RESPONSE=$(curl -s -X POST "$API_BASE/v1/ai-analysis/batch?baby_id=1&start_date=2025-11-01&end_date=2025-11-08" \
  -H "Authorization: Bearer $TOKEN")

echo "响应 (包含所有分析结果):"
echo "$BATCH_RESPONSE" | jq '.'
echo ""

echo -e "${GREEN}✓ 所有分析任务已立即完成${NC}"
echo ""

# ============================================
# 方式3: 手动处理待处理任务
# ============================================
echo -e "${YELLOW}【方式3】直接调用处理待处理任务端点${NC}"
echo ""

echo -e "${GREEN}步骤1: 创建多个任务${NC}"
for i in 1 2 3; do
    curl -s -X POST "$API_BASE/v1/ai-analysis" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"baby_id\": 1,
        \"analysis_type\": \"sleep\",
        \"start_date\": \"2025-11-01\",
        \"end_date\": \"2025-11-08\"
      }" > /dev/null
    echo "✓ 创建任务 $i"
done
echo ""

echo -e "${GREEN}步骤2: 处理所有待处理任务${NC}"
curl -s -X POST "$API_BASE/v1/jobs/process-pending-analyses" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""

# ============================================
# 总结
# ============================================
echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}总结: 三种处理方式${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""
echo "1️⃣  创建+延后处理"
echo "   - POST /v1/ai-analysis          → 创建任务（返回pending）"
echo "   - POST /v1/jobs/process-pending-analyses → 手动触发处理"
echo "   - 适合: 批量创建任务，稍后统一处理"
echo ""
echo "2️⃣  批量分析（推荐）"
echo "   - POST /v1/ai-analysis/batch    → 创建+立即处理4种分析"
echo "   - 适合: 获取完整分析，需要立即返回结果"
echo ""
echo "3️⃣  处理待处理任务"
echo "   - POST /v1/jobs/process-pending-analyses → 处理所有pending任务"
echo "   - 适合: 批量处理，一次处理10个任务"
echo ""
echo -e "${GREEN}✓ 演示完成${NC}"
echo ""
