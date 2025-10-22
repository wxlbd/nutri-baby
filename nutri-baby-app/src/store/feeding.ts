/**
 * 喂养记录状态管理 - API 渐进式集成版本
 *
 * 已集成 API:
 * - POST /feeding-records (创建记录)
 * - GET /feeding-records (查询记录列表)
 *
 * 待集成 API (使用本地实现):
 * - PUT /feeding-records/{recordId} (更新记录) - API 待实现
 * - DELETE /feeding-records/{recordId} (删除记录) - API 待实现
 */
import { ref } from 'vue'
import type { FeedingRecord } from '@/types'
import { StorageKeys, getStorage, setStorage } from '@/utils/storage'
import { get, post } from '@/utils/request'
import { getTodayStart, getTodayEnd } from '@/utils/date'

// 喂养记录列表
const feedingRecords = ref<FeedingRecord[]>(
  getStorage<FeedingRecord[]>(StorageKeys.FEEDING_RECORDS) || []
)

/**
 * 从服务器获取喂养记录列表
 *
 * API: GET /feeding-records?babyId={babyId}&startTime={startTime}&endTime={endTime}&page={page}&pageSize={pageSize}
 */
export async function fetchFeedingRecords(params: {
  babyId: string
  startTime?: number
  endTime?: number
  page?: number
  pageSize?: number
}): Promise<FeedingRecord[]> {
  try {
    const response = await get<{
      records: any[]
      total: number
      page: number
      pageSize: number
    }>('/feeding-records', params)

    if (response.code === 0 && response.data) {
      // TODO: 映射 API 响应到本地类型 (需根据实际 API 响应调整)
      const records = response.data.records as FeedingRecord[]

      feedingRecords.value = records
      setStorage(StorageKeys.FEEDING_RECORDS, records)

      return records
    } else {
      throw new Error(response.message || '获取喂养记录失败')
    }
  } catch (error: any) {
    console.error('fetch feeding records error:', error)
    throw error
  }
}

/**
 * 添加喂养记录
 *
 * API: POST /feeding-records
 */
export async function addFeedingRecord(
  record: Omit<FeedingRecord, 'id' | 'createTime'>
): Promise<FeedingRecord> {
  try {
    // TODO: 映射本地类型到 API 请求格式 (需根据实际 API 要求调整)
    const response = await post<any>('/feeding-records', {
      babyId: record.babyId,
      feedingTime: record.time,
      // 根据 detail 类型映射字段
      ...mapFeedingDetailToAPI(record.detail),
    })

    if (response.code === 0 && response.data) {
      const newRecord: FeedingRecord = {
        ...record,
        id: response.data.recordId,
        createTime: response.data.createTime,
      }

      feedingRecords.value.unshift(newRecord)
      setStorage(StorageKeys.FEEDING_RECORDS, feedingRecords.value)

      uni.showToast({
        title: '记录成功',
        icon: 'success',
      })

      return newRecord
    } else {
      throw new Error(response.message || '添加记录失败')
    }
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

/**
 * 映射喂养详情到 API 格式
 */
function mapFeedingDetailToAPI(detail: FeedingRecord['detail']): any {
  if (detail.type === 'breast') {
    return {
      feedingType: 'breast',
      duration: detail.duration,
      detail: {
        breastSide: detail.side,
        leftTime: detail.leftDuration,
        rightTime: detail.rightDuration,
      },
    }
  } else if (detail.type === 'bottle') {
    return {
      feedingType: detail.bottleType === 'formula' ? 'formula' : 'breast',
      amount: detail.amount,
      detail: {
        formulaType: detail.bottleType,
      },
    }
  } else {
    return {
      feedingType: 'mixed',
      note: detail.note,
    }
  }
}

/**
 * 本地查询方法
 */
export function getFeedingRecords(): FeedingRecord[] {
  return feedingRecords.value
}

export function getFeedingRecordsByBabyId(babyId: string): FeedingRecord[] {
  return feedingRecords.value.filter((record) => record.babyId === babyId)
}

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

export function getTodayTotalMilk(babyId: string): number {
  const todayRecords = getTodayFeedingRecords(babyId)
  let total = 0

  todayRecords.forEach((record) => {
    if (record.detail.type === 'bottle') {
      const amount =
        record.detail.unit === 'oz'
          ? record.detail.amount * 29.5735
          : record.detail.amount
      total += amount
    } else if (record.detail.type === 'breast') {
      total += record.detail.duration * 5
    }
  })

  return Math.round(total)
}

export function getLastFeedingRecord(babyId: string): FeedingRecord | null {
  const records = feedingRecords.value
    .filter((record) => record.babyId === babyId)
    .sort((a, b) => b.time - a.time)

  return records.length > 0 ? records[0] : null
}

export { feedingRecords }
