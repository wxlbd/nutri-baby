import type { ApiResponse } from '@/types'
import { request } from '@/utils/request'

// ======================== 时间线查询参数 ========================

export interface TimelineQuery {
  babyId: string
  startTime?: number
  endTime?: number
  page?: number
  pageSize?: number
}

// ======================== 时间线记录项 ========================

export interface TimelineItem {
  recordType: 'feeding' | 'sleep' | 'diaper' | 'growth'
  recordId: string
  babyId: string
  eventTime: number
  detail: any // 具体记录详情 (FeedingRecordDTO | SleepRecordDTO | DiaperRecordDTO | GrowthRecordDTO)
  createBy: string
  createTime: number
}

// ======================== 时间线响应 ========================

export interface TimelineResponse {
  items: TimelineItem[]
  total: number
  page: number
  pageSize: number
}

// ======================== API 接口 ========================

/**
 * 获取时间线记录
 * @param query 查询参数
 * @returns 时间线响应
 */
export const apiFetchTimeline = (
  query: TimelineQuery
): Promise<ApiResponse<TimelineResponse>> => {
  return request<TimelineResponse>({
    url: '/record/timeline',
    method: 'GET',
    data: query,
  })
}
