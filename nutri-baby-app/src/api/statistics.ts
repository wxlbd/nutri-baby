import { request } from '@/utils/request'
import type { ApiResponse } from '@/types'
// ============ 统计 DTO ============

// 今日喂养统计
export interface TodayFeedingStats {
  breastCount: number  // 母乳喂养次数
  bottleMl: number     // 奶瓶总毫升数
  totalCount: number   // 总喂养次数
  lastFeedingTime?: number // 最后一次喂养时间戳(毫秒)
}

// 今日睡眠统计
export interface TodaySleepStats {
  totalMinutes: number   // 总睡眠分钟数
  lastSleepMinutes: number // 上次睡眠分钟数
  sessionCount: number   // 睡眠次数
}

// 今日换尿布统计
export interface TodayDiaperStats {
  totalCount: number // 总换尿布次数
  wetCount: number   // 尿湿次数
  dirtyCount: number // 排便次数
}

// 今日成长统计
export interface TodayGrowthStats {
  latestWeight?: number            // 最新体重 (kg)
  latestHeight?: number            // 最新身高 (cm)
  latestHeadCircumference?: number // 最新头围 (cm)
}

// 今日统计
export interface TodayStatistics {
  feeding: TodayFeedingStats
  sleep: TodaySleepStats
  diaper: TodayDiaperStats
  growth: TodayGrowthStats
}

// 本周喂养统计
export interface WeeklyFeedingStats {
  totalCount: number  // 本周总喂养次数
  trend: number       // 趋势对比（与上周的差异）
  avgPerDay: number   // 日均喂养次数
}

// 本周睡眠统计
export interface WeeklySleepStats {
  totalHours: number  // 本周总睡眠小时数
  trend: number       // 趋势对比（与上周的小时数差异）
  avgPerDay: number   // 日均睡眠小时数
}

// 本周成长统计
export interface WeeklyGrowthStats {
  weightGain: number       // 周内体重增长 (kg)
  heightGain: number       // 周内身高增长 (cm)
  weekStartWeight?: number // 周初体重 (kg)
}

// 本周统计
export interface WeeklyStatistics {
  feeding: WeeklyFeedingStats
  sleep: WeeklySleepStats
  growth: WeeklyGrowthStats
}

// 完整统计响应
export interface BabyStatisticsResponse {
  today: TodayStatistics
  weekly: WeeklyStatistics
}

// ============ API 接口 ============

/**
 * 获取宝宝统计数据
 * @param babyId 宝宝ID
 * @returns 统计数据响应
 */
export const apiFetchBabyStatistics = (babyId: string): Promise<ApiResponse<BabyStatisticsResponse>> => {
  return request({
    method: 'GET',
    url: `/babies/${babyId}/statistics`,
  })
}
