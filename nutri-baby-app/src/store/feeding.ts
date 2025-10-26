/**
 * 喂养记录状态管理
 * 职责: 状态管理 + 本地计算,API 调用委托给 api 层
 *
 * ⚠️ 向后兼容: 所有导出函数的签名保持不变,页面组件无需修改
 */
import { ref } from 'vue'
import type { FeedingRecord } from '@/types'
import { StorageKeys, getStorage, setStorage } from '@/utils/storage'
import { getTodayStart, getTodayEnd } from '@/utils/date'
import * as feedingApi from '@/api/feeding'

// ============ 状态定义 ============

// 喂养记录列表 - 延迟初始化,避免模块加载时同步读取存储
const feedingRecords = ref<FeedingRecord[]>([])

// 初始化标记
let isInitialized = false

// 延迟初始化函数
function initializeIfNeeded() {
  if (!isInitialized) {
    feedingRecords.value = getStorage<FeedingRecord[]>(StorageKeys.FEEDING_RECORDS) || []
    isInitialized = true
  }
}

// ============ API 调用函数(委托给 api 层) ============

/**
 * 从服务器获取喂养记录列表
 *
 * API: GET /feeding-records?babyId={babyId}&startTime={startTime}&endTime={endTime}&page={page}&pageSize={pageSize}
 *
 * ⚠️ 向后兼容: 函数签名保持不变
 */
export async function fetchFeedingRecords(params: {
  babyId: string
  startTime?: number
  endTime?: number
  page?: number
  pageSize?: number
}): Promise<FeedingRecord[]> {
  try {
    const response = await feedingApi.apiFetchFeedingRecords(params)

    // 映射 API 响应到本地类型
    const records: FeedingRecord[] = response.records.map((apiRecord) => ({
      id: apiRecord.recordId,
      babyId: apiRecord.babyId,
      time: apiRecord.feedingTime,
      detail: mapApiDetailToLocal(apiRecord),
      createTime: apiRecord.createTime,
    }))

    feedingRecords.value = records
    setStorage(StorageKeys.FEEDING_RECORDS, records)

    return records
  } catch (error: any) {
    console.error('fetch feeding records error:', error)
    throw error
  }
}

/**
 * 添加喂养记录
 *
 * API: POST /feeding-records
 *
 * ⚠️ 向后兼容: 函数签名保持不变
 */
export async function addFeedingRecord(
  record: Omit<FeedingRecord, 'id' | 'createTime'>
): Promise<FeedingRecord> {
  try {
    // 映射本地类型到 API 请求格式
    const requestData = mapLocalToApiRequest(record)

    // 调用 API 层
    const response = await feedingApi.apiCreateFeedingRecord(requestData)

    // 映射 API 响应到本地类型
    const newRecord: FeedingRecord = {
      id: response.recordId,
      babyId: response.babyId,
      time: response.feedingTime,
      detail: mapApiDetailToLocal(response),
      createTime: response.createTime,
    }

    feedingRecords.value.unshift(newRecord)
    setStorage(StorageKeys.FEEDING_RECORDS, feedingRecords.value)

    uni.showToast({
      title: '记录成功',
      icon: 'success',
    })

    return newRecord
  } catch (error: any) {
    console.error('add feeding record error:', error)
    uni.showToast({
      title: error.message || '记录失败',
      icon: 'none',
    })
    throw error
  }
}

/**
 * 删除喂养记录 (本地实现,待 API 完善)
 * TODO: 集成 DELETE /feeding-records/{recordId} API
 *
 * ⚠️ 向后兼容: 函数签名保持不变
 */
export function deleteFeedingRecord(id: string): boolean {
  const index = feedingRecords.value.findIndex((record) => record.id === id)
  if (index === -1) {
    return false
  }

  feedingRecords.value.splice(index, 1)
  setStorage(StorageKeys.FEEDING_RECORDS, feedingRecords.value)
  return true
}

// ============ 本地查询和计算函数 ============

/**
 * 获取喂养记录列表(本地)
 *
 * ⚠️ 向后兼容: 函数签名保持不变
 */
export function getFeedingRecords(): FeedingRecord[] {
  initializeIfNeeded()
  return feedingRecords.value
}

/**
 * 根据宝宝ID获取喂养记录(本地)
 *
 * ⚠️ 向后兼容: 函数签名保持不变
 */
export function getFeedingRecordsByBabyId(babyId: string): FeedingRecord[] {
  return feedingRecords.value.filter((record) => record.babyId === babyId)
}

/**
 * 获取今日喂养记录(本地)
 *
 * ⚠️ 向后兼容: 函数签名保持不变
 */
export function getTodayFeedingRecords(babyId: string): FeedingRecord[] {
  const todayStart = getTodayStart()
  const todayEnd = getTodayEnd()

  return feedingRecords.value.filter(
    (record) =>
      record.babyId === babyId &&
      record.time >= todayStart &&
      record.time <= todayEnd
  )
}

/**
 * 获取今日总奶量（仅统计奶瓶喂养的实际奶量，不包含母乳喂养）
 * 注：母乳喂养只记录时间，无法准确测量毫升数，因此不计入总奶量
 *
 * ⚠️ 向后兼容: 函数签名保持不变
 */
export function getTodayTotalMilk(babyId: string): number {
  initializeIfNeeded()
  const todayRecords = getTodayFeedingRecords(babyId)
  let total = 0

  todayRecords.forEach((record) => {
    // 只统计奶瓶喂养的奶量
    if (record.detail.type === 'bottle') {
      const amount =
        record.detail.unit === 'oz'
          ? record.detail.amount * 29.5735
          : record.detail.amount
      total += amount
    }
    // 母乳喂养不计入奶量统计
  })

  return Math.round(total)
}

/**
 * 获取今日母乳喂养统计
 * 返回：{ count: 次数, totalDuration: 总时长(秒) }
 *
 * ⚠️ 向后兼容: 函数签名保持不变
 */
export function getTodayBreastfeedingStats(babyId: string): { count: number; totalDuration: number } {
  const todayRecords = getTodayFeedingRecords(babyId)
  let count = 0
  let totalDuration = 0

  todayRecords.forEach((record) => {
    if (record.detail.type === 'breast') {
      count++
      totalDuration += record.detail.duration
    }
  })

  return { count, totalDuration }
}

/**
 * 获取最后一条喂养记录(本地)
 *
 * ⚠️ 向后兼容: 函数签名保持不变
 */
export function getLastFeedingRecord(babyId: string): FeedingRecord | null {
  const records = feedingRecords.value
    .filter((record) => record.babyId === babyId)
    .sort((a, b) => b.time - a.time)

  return records.length > 0 ? records[0] : null
}

// ============ 内部辅助函数 ============

/**
 * 映射本地 FeedingRecord 到 API 请求格式
 * 与后端 DTO 保持一致: CreateFeedingRecordRequest + FeedingDetail
 */
function mapLocalToApiRequest(
  record: Omit<FeedingRecord, 'id' | 'createTime'>
): feedingApi.CreateFeedingRecordRequest {
  const detail = record.detail
  const base: feedingApi.CreateFeedingRecordRequest = {
    babyId: record.babyId,
    feedingType: detail.type, // 'breast' | 'bottle' | 'food'
    feedingTime: record.time,
    detail: {},
  }

  if (detail.type === 'breast') {
    base.duration = detail.duration
    base.detail = {
      breastSide: detail.side,
      leftTime: detail.leftDuration,
      rightTime: detail.rightDuration,
      duration: detail.duration,
    }
  } else if (detail.type === 'bottle') {
    base.amount = detail.amount
    base.detail = {
      bottleType: detail.bottleType,
      unit: detail.unit,
      remaining: detail.remaining,
    }
  } else {
    // food
    base.detail = {
      foodName: detail.foodName,
      note: detail.note,
    }
  }

  return base
}

/**
 * 映射 API 响应到本地 FeedingRecord.detail 格式
 */
function mapApiDetailToLocal(apiRecord: feedingApi.FeedingRecordResponse): FeedingRecord['detail'] {
  const feedingType = apiRecord.feedingType

  if (feedingType === 'breast') {
    return {
      type: 'breast',
      side: apiRecord.detail.breastSide || 'left',
      duration: apiRecord.duration || 0,
      leftDuration: apiRecord.detail.leftTime || 0,
      rightDuration: apiRecord.detail.rightTime || 0,
    }
  } else if (feedingType === 'bottle') {
    return {
      type: 'bottle',
      bottleType: apiRecord.detail.bottleType || 'formula',
      amount: apiRecord.amount || 0,
      unit: apiRecord.detail.unit || 'ml',
      remaining: apiRecord.detail.remaining,
    }
  } else {
    // food
    return {
      type: 'food',
      foodName: apiRecord.detail.foodName || '',
      note: apiRecord.detail.note || '',
    }
  }
}

// ============ 导出 ============

export { feedingRecords }
