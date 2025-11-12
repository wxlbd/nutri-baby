import { reactive, ref, computed } from 'vue'
import type {
  AIAnalysis,
  AnalysisResponse,
  DailyTipsResponse,
  AnalysisStatsResponse,
  AIAnalysisType,
  AIAnalysisStatus,
  DailyTip
} from '@/types/ai'
import {
  createAIAnalysis,
  getAIAnalysisResult,
  getLatestAIAnalysis,
  getAIAnalysisStats,
  getDailyTips,
  generateDailyTips,
  pollAnalysisStatus
} from '@/api/ai'

/**
 * AI分析状态管理
 */
export const useAIStore = () => {
  // 状态定义
  const analyses = reactive<Record<number, AIAnalysis>>({}) // 分析记录映射
  const dailyTips = reactive<Record<string, DailyTip[]>>({}) // 每日建议映射
  const stats = reactive<AnalysisStatsResponse | null>(null) // 分析统计
  const isAnalyzing = ref(false) // 是否正在分析
  const analyzingIds = reactive<Set<number>>(new Set()) // 正在分析的ID集合
  const currentAnalysis = ref<AIAnalysis | null>(null) // 当前分析

  // 计算属性
  const completedAnalyses = computed(() => {
    return Object.values(analyses).filter(analysis => analysis.status === 'completed')
  })

  const pendingAnalyses = computed(() => {
    return Object.values(analyses).filter(analysis =>
      analysis.status === 'pending' || analysis.status === 'analyzing'
    )
  })

  const failedAnalyses = computed(() => {
    return Object.values(analyses).filter(analysis => analysis.status === 'failed')
  })

  const todayTips = computed(() => {
    const today = new Date().toISOString().split('T')[0]
    return dailyTips[today] || []
  })

  const hasUnexpiredTips = computed(() => {
    const now = new Date()
    return Object.entries(dailyTips).some(([date, tips]) => {
      const tipDate = new Date(date)
      const expiryDate = new Date(tipDate.getTime() + 24 * 60 * 60 * 1000) // 24小时后过期
      return now < expiryDate && tips.length > 0
    })
  })

  // 方法

  /**
   * 创建AI分析任务
   */
  const createAnalysis = async (babyId: number, analysisType: AIAnalysisType, startDate: string, endDate: string) => {
    try {
      isAnalyzing.value = true

      const response = await createAIAnalysis({
        baby_id: babyId,
        analysis_type: analysisType,
        start_date: startDate,
        end_date: endDate
      })

      // 添加到分析记录
      const analysis: AIAnalysis = {
        id: response.analysis_id,
        baby_id: babyId,
        analysis_type: analysisType,
        status: response.status,
        start_date: startDate,
        end_date: endDate,
        created_at: response.created_at,
        updated_at: response.created_at
      }

      analyses[analysis.id] = analysis
      analyzingIds.add(analysis.id)

      // 开始轮询状态
      pollAnalysisStatus(analysis.id)

      return analysis
    } catch (error) {
      console.error('创建AI分析任务失败:', error)
      throw error
    } finally {
      isAnalyzing.value = false
    }
  }

  /**
   * 轮询分析状态
   */
  const pollAnalysisStatus = async (analysisId: number) => {
    try {
      const result = await pollAnalysisStatus(analysisId, (status) => {
        // 更新分析状态
        if (analyses[analysisId]) {
          analyses[analysisId].status = status as AIAnalysisStatus
        }
      })

      // 更新完整分析结果
      if (result.result) {
        analyses[analysisId] = {
          ...analyses[analysisId],
          status: result.status as AIAnalysisStatus,
          result: result.result,
          score: result.result.score,
          insights: result.result.insights?.map(insight => JSON.stringify(insight)),
          alerts: result.result.alerts?.map(alert => JSON.stringify(alert)),
          updated_at: new Date().toISOString()
        }
      }

      // 从分析中移除
      analyzingIds.delete(analysisId)

      return result
    } catch (error) {
      console.error('轮询分析状态失败:', error)

      // 更新为失败状态
      if (analyses[analysisId]) {
        analyses[analysisId].status = 'failed'
        analyses[analysisId].updated_at = new Date().toISOString()
      }

      analyzingIds.delete(analysisId)
      throw error
    }
  }

  /**
   * 获取分析结果
   */
  const getAnalysisResult = async (analysisId: number): Promise<AIAnalysis> => {
    try {
      const response = await getAIAnalysisResult(analysisId)

      // 更新本地缓存
      if (analyses[analysisId]) {
        analyses[analysisId].status = response.status as AIAnalysisStatus
        if (response.result) {
          analyses[analysisId].result = response.result
          analyses[analysisId].score = response.result.score
          analyses[analysisId].insights = response.result.insights?.map(insight => JSON.stringify(insight))
          analyses[analysisId].alerts = response.result.alerts?.map(alert => JSON.stringify(alert))
        }
        analyses[analysisId].updated_at = new Date().toISOString()
      }

      return analyses[analysisId]
    } catch (error) {
      console.error('获取分析结果失败:', error)
      throw error
    }
  }

  /**
   * 获取最新分析结果
   */
  const getLatestAnalysis = async (babyId: number, analysisType: AIAnalysisType): Promise<AIAnalysis | null> => {
    try {
      const response = await getLatestAIAnalysis(babyId, analysisType)

      if (response.result) {
        const analysis: AIAnalysis = {
          id: response.analysis_id,
          baby_id: babyId,
          analysis_type: analysisType,
          status: response.status as AIAnalysisStatus,
          start_date: '', // 将从result中获取
          end_date: '',   // 将从result中获取
          result: response.result,
          score: response.result.score,
          insights: response.result.insights?.map(insight => JSON.stringify(insight)),
          alerts: response.result.alerts?.map(alert => JSON.stringify(alert)),
          created_at: response.created_at,
          updated_at: response.created_at
        }

        analyses[analysis.id] = analysis
        return analysis
      }

      return null
    } catch (error) {
      if (error.response?.status === 404) {
        return null // 未找到分析结果
      }
      console.error('获取最新分析结果失败:', error)
      throw error
    }
  }

  /**
   * 获取分析统计
   */
  const getAnalysisStats = async (babyId: number): Promise<AnalysisStatsResponse> => {
    try {
      const response = await getAIAnalysisStats(babyId)
      stats = response
      return response
    } catch (error) {
      console.error('获取分析统计失败:', error)
      throw error
    }
  }

  /**
   * 生成每日建议
   */
  const generateDailyTips = async (babyId: number, date?: string): Promise<DailyTip[]> => {
    try {
      const response = await generateDailyTips(babyId, date)

      const targetDate = date || new Date().toISOString().split('T')[0]
      dailyTips[targetDate] = response.tips

      return response.tips
    } catch (error) {
      console.error('生成每日建议失败:', error)
      throw error
    }
  }

  /**
   * 获取每日建议
   */
  const getDailyTips = async (babyId: number, date?: string): Promise<DailyTip[]> => {
    try {
      const targetDate = date || new Date().toISOString().split('T')[0]

      // 如果已有缓存且未过期，直接返回
      if (dailyTips[targetDate] && dailyTips[targetDate].length > 0) {
        return dailyTips[targetDate]
      }

      // 否则从服务器获取
      const response = await getDailyTips(babyId, date)
      dailyTips[targetDate] = response.tips

      return response.tips
    } catch (error) {
      console.error('获取每日建议失败:', error)
      throw error
    }
  }

  /**
   * 清除分析缓存
   */
  const clearAnalysisCache = (analysisId?: number) => {
    if (analysisId) {
      delete analyses[analysisId]
    } else {
      // 清除所有分析缓存
      Object.keys(analyses).forEach(key => {
        delete analyses[parseInt(key)]
      })
    }
  }

  /**
   * 清除每日建议缓存
   */
  const clearDailyTipsCache = (date?: string) => {
    if (date) {
      delete dailyTips[date]
    } else {
      // 清除所有每日建议缓存
      Object.keys(dailyTips).forEach(key => {
        delete dailyTips[key]
      })
    }
  }

  /**
   * 检查是否有活跃的分析任务
   */
  const hasActiveAnalysis = computed(() => {
    return analyzingIds.size > 0
  })

  /**
   * 获取指定类型的最新分析
   */
  const getLatestAnalysisByType = (analysisType: AIAnalysisType): AIAnalysis | null => {
    const typeAnalyses = Object.values(analyses).filter(
      analysis => analysis.analysis_type === analysisType && analysis.status === 'completed'
    )

    if (typeAnalyses.length === 0) return null

    // 按创建时间排序，返回最新的
    return typeAnalyses.sort((a, b) =>
      new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    )[0]
  }

  /**
   * 获取分析概览
   */
  const getAnalysisOverview = (babyId: number) => {
    const babyAnalyses = Object.values(analyses).filter(analysis => analysis.baby_id === babyId)

    return {
      total: babyAnalyses.length,
      completed: babyAnalyses.filter(a => a.status === 'completed').length,
      pending: babyAnalyses.filter(a => a.status === 'pending' || a.status === 'analyzing').length,
      failed: babyAnalyses.filter(a => a.status === 'failed').length,
      averageScore: babyAnalyses
        .filter(a => a.score !== undefined)
        .reduce((sum, a) => sum + (a.score || 0), 0) / babyAnalyses.filter(a => a.score !== undefined).length || 0
    }
  }

  /**
   * 获取需要关注的事项
   */
  const getAttentionItems = (babyId: number) => {
    const babyAnalyses = Object.values(analyses).filter(analysis =>
      analysis.baby_id === babyId && analysis.status === 'completed'
    )

    const attentionItems = []

    babyAnalyses.forEach(analysis => {
      // 检查警告
      if (analysis.alerts && analysis.alerts.length > 0) {
        analysis.alerts.forEach(alertStr => {
          try {
            const alert = JSON.parse(alertStr)
            if (alert.level === 'critical' || alert.level === 'warning') {
              attentionItems.push({
                type: 'alert',
                title: alert.title,
                description: alert.description,
                level: alert.level,
                analysisType: analysis.analysis_type
              })
            }
          } catch (e) {
            console.error('解析警告失败:', e)
          }
        })
      }

      // 检查低分分析
      if (analysis.score !== undefined && analysis.score < 60) {
        attentionItems.push({
          type: 'low_score',
          title: `${analysis.analysis_type}分析评分较低`,
          description: `评分为${analysis.score}分，建议关注`,
          level: 'warning',
          analysisType: analysis.analysis_type,
          score: analysis.score
        })
      }
    })

    return attentionItems.sort((a, b) => {
      const levelPriority = { critical: 3, warning: 2, info: 1 }
      return (levelPriority[b.level] || 0) - (levelPriority[a.level] || 0)
    })
  }

  return {
    // 状态
    analyses,
    dailyTips,
    stats,
    isAnalyzing,
    analyzingIds,
    currentAnalysis,

    // 计算属性
    completedAnalyses,
    pendingAnalyses,
    failedAnalyses,
    todayTips,
    hasUnexpiredTips,
    hasActiveAnalysis,

    // 方法
    createAnalysis,
    pollAnalysisStatus,
    getAnalysisResult,
    getLatestAnalysis,
    getAnalysisStats,
    generateDailyTips,
    getDailyTips,
    clearAnalysisCache,
    clearDailyTipsCache,
    getLatestAnalysisByType,
    getAnalysisOverview,
    getAttentionItems
  }
}

// 导出单例实例
export const aiStore = useAIStore()