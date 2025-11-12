import { request } from '@/utils/request'
import type {
  AIAnalysisParams,
  AnalysisResponse,
  BatchAnalysisResponse,
  CreateAnalysisRequest,
  DailyTipsResponse,
  AnalysisStatsResponse,
  AIAnalysisType
} from '@/types/ai'

/**
 * AI分析相关API
 */

/**
 * 创建AI分析任务
 */
export const createAIAnalysis = (data: CreateAnalysisRequest): Promise<AnalysisResponse> => {
  return request<AnalysisResponse>({
    url: '/ai/analysis',
    method: 'POST',
    data
  })
}

/**
 * 获取AI分析结果
 */
export const getAIAnalysisResult = (analysisId: number): Promise<AnalysisResponse> => {
  return request<AnalysisResponse>({
    url: `/ai/analysis/${analysisId}`,
    method: 'GET'
  })
}

/**
 * 获取最新AI分析结果
 */
export const getLatestAIAnalysis = (babyId: number, analysisType: AIAnalysisType): Promise<AnalysisResponse> => {
  return request<AnalysisResponse>({
    url: '/ai/analysis/latest',
    method: 'GET',
    data: {
      baby_id: babyId,
      analysis_type: analysisType
    }
  })
}

/**
 * 批量AI分析
 */
export const batchAIAnalysis = (babyId: number, startDate: string, endDate: string): Promise<BatchAnalysisResponse> => {
  return request<BatchAnalysisResponse>({
    url: '/ai/analysis/batch',
    method: 'POST',
    data: {
      baby_id: babyId,
      start_date: startDate,
      end_date: endDate
    }
  })
}

/**
 * 获取AI分析统计
 */
export const getAIAnalysisStats = (babyId: number): Promise<AnalysisStatsResponse> => {
  return request<AnalysisStatsResponse>({
    url: '/ai/analysis/stats',
    method: 'GET',
    data: {
      baby_id: babyId
    }
  })
}

/**
 * 生成每日建议
 */
export const generateDailyTips = (babyId: number, date?: string): Promise<DailyTipsResponse> => {
  const params: any = { baby_id: babyId }
  if (date) {
    params.date = date
  }

  return request<DailyTipsResponse>({
    url: '/ai/daily-tips',
    method: 'POST',
    data: params
  })
}

/**
 * 获取每日建议
 */
export const getDailyTips = (babyId: number, date?: string): Promise<DailyTipsResponse> => {
  const params: any = { baby_id: babyId }
  if (date) {
    params.date = date
  }

  return request<DailyTipsResponse>({
    url: '/ai/daily-tips',
    method: 'GET',
    data: params
  })
}

/**
 * 轮询分析状态
 */
export const pollAnalysisStatus = async (
  analysisId: number,
  onStatusUpdate: (status: string) => void,
  maxAttempts = 30,
  interval = 2000
): Promise<AnalysisResponse> => {
  for (let attempt = 0; attempt < maxAttempts; attempt++) {
    try {
      const result = await getAIAnalysisResult(analysisId)

      // 更新状态
      onStatusUpdate(result.status)

      // 如果分析完成或失败，返回结果
      if (result.status === 'completed' || result.status === 'failed') {
        return result
      }

      // 等待下次轮询
      await new Promise(resolve => setTimeout(resolve, interval))
    } catch (error) {
      console.error(`轮询分析状态失败 (attempt ${attempt + 1}):`, error)

      // 最后一次尝试失败则抛出错误
      if (attempt === maxAttempts - 1) {
        throw error
      }

      // 继续下一次尝试
      await new Promise(resolve => setTimeout(resolve, interval))
    }
  }

  throw new Error('分析超时')
}

/**
 * 获取分析图表数据
 */
export const getAnalysisChartData = (analysisType: AIAnalysisType, data: any) => {
  switch (analysisType) {
    case 'feeding':
      return getFeedingAnalysisChartData(data)
    case 'sleep':
      return getSleepAnalysisChartData(data)
    case 'growth':
      return getGrowthAnalysisChartData(data)
    case 'health':
      return getHealthAnalysisChartData(data)
    default:
      return null
  }
}

/**
 * 获取喂养分析图表数据
 */
const getFeedingAnalysisChartData = (data: any) => {
  if (!data.patterns || !data.patterns.length) return null

  const pattern = data.patterns[0]
  const categories = ['规律性', '适量性', '及时性', '多样性']
  const scores = [
    pattern.regularity || 0,
    pattern.adequacy || 0,
    pattern.timeliness || 0,
    pattern.diversity || 0
  ]

  return {
    categories,
    series: [{
      name: '喂养模式评分',
      data: scores.map(score => Math.round(score * 100)),
      color: '#7dd3a2'
    }],
    title: '喂养模式分析',
    subtitle: '基于AI智能分析的综合评分'
  }
}

/**
 * 获取睡眠分析图表数据
 */
const getSleepAnalysisChartData = (data: any) => {
  if (!data.patterns || !data.patterns.length) return null

  const pattern = data.patterns[0]
  const categories = ['连续性', '时长', '规律性', '深度']
  const scores = [
    pattern.continuity || 0,
    pattern.duration || 0,
    pattern.regularity || 0,
    pattern.depth || 0
  ]

  return {
    categories,
    series: [{
      name: '睡眠质量评分',
      data: scores.map(score => Math.round(score * 100)),
      color: '#52c41a'
    }],
    title: '睡眠质量分析',
    subtitle: '基于AI智能分析的综合评分'
  }
}

/**
 * 获取成长分析图表数据
 */
const getGrowthAnalysisChartData = (data: any) => {
  if (!data.predictions || !data.predictions.length) return null

  const predictions = data.predictions.filter((p: any) => p.prediction_type === 'growth')
  const categories = predictions.map((p: any) => p.time_frame)
  const values = predictions.map((p: any) => parseFloat(p.value) || 0)

  return {
    categories,
    series: [{
      name: '预测值',
      data: values,
      color: '#ff6b6b'
    }],
    title: '成长趋势预测',
    subtitle: '基于AI智能分析的预测结果'
  }
}

/**
 * 获取健康分析图表数据
 */
const getHealthAnalysisChartData = (data: any) => {
  if (!data.alerts || !data.alerts.length) return null

  const alerts = data.alerts
  const levelCounts = {
    critical: alerts.filter((a: any) => a.level === 'critical').length,
    warning: alerts.filter((a: any) => a.level === 'warning').length,
    info: alerts.filter((a: any) => a.level === 'info').length
  }

  return {
    categories: ['严重', '警告', '提示'],
    series: [{
      name: '健康预警',
      data: [levelCounts.critical, levelCounts.warning, levelCounts.info],
      color: '#ff4757',
      type: 'column'
    }],
    title: '健康预警分布',
    subtitle: '基于AI智能分析的风险评估'
  }
}

/**
 * 分析状态文本
 */
export const getAnalysisStatusText = (status: string): string => {
  const statusMap: Record<string, string> = {
    pending: '等待分析',
    analyzing: '分析中...',
    completed: '分析完成',
    failed: '分析失败'
  }
  return statusMap[status] || status
}

/**
 * 分析状态颜色
 */
export const getAnalysisStatusColor = (status: string): string => {
  const colorMap: Record<string, string> = {
    pending: '#ffa940',
    analyzing: '#1890ff',
    completed: '#52c41a',
    failed: '#ff4d4f'
  }
  return colorMap[status] || '#8c8c8c'
}

/**
 * 优先级颜色
 */
export const getPriorityColor = (priority: string): string => {
  const colorMap: Record<string, string> = {
    high: '#ff4d4f',
    medium: '#faad14',
    low: '#52c41a'
  }
  return colorMap[priority] || '#8c8c8c'
}

/**
 * 警告级别颜色
 */
export const getAlertLevelColor = (level: string): string => {
  const colorMap: Record<string, string> = {
    critical: '#ff4d4f',
    warning: '#faad14',
    info: '#1890ff'
  }
  return colorMap[level] || '#8c8c8c'
}

/**
 * 分析类型图标
 */
export const getAnalysisTypeIcon = (type: AIAnalysisType): string => {
  const iconMap: Record<AIAnalysisType, string> = {
    feeding: '🍼',
    sleep: '😴',
    growth: '📈',
    health: '❤️',
    behavior: '🧠'
  }
  return iconMap[type] || '🤖'
}

/**
 * 分析类型名称
 */
export const getAnalysisTypeName = (type: AIAnalysisType): string => {
  const nameMap: Record<AIAnalysisType, string> = {
    feeding: '喂养分析',
    sleep: '睡眠分析',
    growth: '成长分析',
    health: '健康分析',
    behavior: '行为分析'
  }
  return nameMap[type] || '未知分析'
}