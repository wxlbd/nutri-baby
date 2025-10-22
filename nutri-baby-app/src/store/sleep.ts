/**
 * 睡眠记录状态管理 - API 渐进式集成版本
 *
 * 已集成 API:
 * - POST /sleep-records (创建记录)
 * - GET /sleep-records (查询记录列表)
 *
 * 待集成 API (使用本地实现):
 * - PUT /sleep-records/{recordId} (更新记录) - API 待实现
 * - DELETE /sleep-records/{recordId} (删除记录) - API 待实现
 */
import { ref } from 'vue'
import type { SleepRecord } from '@/types'
import { StorageKeys, getStorage, setStorage } from '@/utils/storage'
import { get, post } from '@/utils/request'
import { getTodayStart, getTodayEnd } from '@/utils/date'
import { generateId } from '@/utils/common'

// 睡眠记录列表
const sleepRecords = ref<SleepRecord[]>(
  getStorage<SleepRecord[]>(StorageKeys.SLEEP_RECORDS) || []
)

/**
 * 从服务器获取睡眠记录列表
 *
 * API: GET /sleep-records?babyId={babyId}&startTime={startTime}&endTime={endTime}&page={page}&pageSize={pageSize}
 */
export async function fetchSleepRecords(params: {
  babyId: string
  startTime?: number
  endTime?: number
  page?: number
  pageSize?: number
}): Promise<SleepRecord[]> {
  try {
    const response = await get<{
      records: any[]
      total: number
      page: number
      pageSize: number
    }>('/sleep-records', params)

    if (response.code === 0 && response.data) {
      // 映射 API 响应到本地类型
      const records: SleepRecord[] = response.data.records.map((item: any) => ({
        id: item.recordId,
        babyId: item.babyId,
        startTime: item.startTime,
        endTime: item.endTime,
        duration: item.duration,
        quality: item.quality,
        type: item.quality === 'good' ? 'night' : 'nap', // 根据质量推断类型
        note: item.note,
        createBy: item.createBy,
        createTime: item.createTime,
      }))

      sleepRecords.value = records
      setStorage(StorageKeys.SLEEP_RECORDS, records)

      return records
    } else {
      throw new Error(response.message || '获取睡眠记录失败')
    }
  } catch (error: any) {
    console.error('fetch sleep records error:', error)
    throw error
  }
}

/**
 * 添加睡眠记录 (完整记录)
 *
 * API: POST /sleep-records
 */
export async function addSleepRecord(
  record: Omit<SleepRecord, 'id' | 'createTime'>
): Promise<SleepRecord> {
  try {
    const response = await post<any>('/sleep-records', {
      babyId: record.babyId,
      startTime: record.startTime,
      endTime: record.endTime,
      duration: record.duration,
      quality: record.quality,
      note: record.note,
    })

    if (response.code === 0 && response.data) {
      const newRecord: SleepRecord = {
        id: response.data.recordId,
        babyId: response.data.babyId,
        startTime: response.data.startTime,
        endTime: response.data.endTime,
        duration: response.data.duration,
        quality: response.data.quality,
        type: record.type,
        note: response.data.note,
        createBy: response.data.createBy,
        createTime: response.data.createTime,
      }

      sleepRecords.value.unshift(newRecord)
      setStorage(StorageKeys.SLEEP_RECORDS, sleepRecords.value)

      uni.showToast({
        title: '记录成功',
        icon: 'success',
      })

      return newRecord
    } else {
      throw new Error(response.message || '添加记录失败')
    }
  } catch (error: any) {
    console.error('add sleep record error:', error)
    uni.showToast({
      title: error.message || '记录失败',
      icon: 'none',
    })
    throw error
  }
}

/**
 * 开始睡眠记录 (本地实现,用于计时器功能)
 */
export function startSleepRecord(
  babyId: string,
  type: SleepRecord['type'],
  createBy: string,
  startTime?: number
): SleepRecord {
  // 检查是否有进行中的睡眠记录
  const ongoing = getOngoingSleepRecord(babyId)
  if (ongoing) {
    throw new Error('已有进行中的睡眠记录,请先结束')
  }

  const record: SleepRecord = {
    id: generateId(),
    babyId,
    startTime: startTime || Date.now(),
    type,
    createBy,
    createTime: Date.now(),
  }

  sleepRecords.value.unshift(record)
  setStorage(StorageKeys.SLEEP_RECORDS, sleepRecords.value)

  return record
}

/**
 * 结束睡眠记录 (本地实现,结束后可调用 addSleepRecord 同步到服务器)
 */
export function endSleepRecord(id: string, endTime?: number): boolean {
  const record = sleepRecords.value.find((r) => r.id === id)
  if (!record) {
    return false
  }

  const end = endTime || Date.now()
  const duration = Math.floor((end - record.startTime) / 1000 / 60) // 转为分钟

  record.endTime = end
  record.duration = duration

  setStorage(StorageKeys.SLEEP_RECORDS, sleepRecords.value)
  return true
}

/**
 * 结束睡眠记录并同步到服务器
 */
export async function endSleepRecordAndSync(
  id: string,
  quality?: 'good' | 'fair' | 'poor',
  note?: string
): Promise<SleepRecord | null> {
  // 先在本地结束记录
  const success = endSleepRecord(id)
  if (!success) return null

  // 获取更新后的记录
  const record = sleepRecords.value.find((r) => r.id === id)
  if (!record || !record.endTime) return null

  try {
    // 同步到服务器 (会创建新的记录ID)
    const serverRecord = await addSleepRecord({
      babyId: record.babyId,
      startTime: record.startTime,
      endTime: record.endTime,
      duration: record.duration,
      quality: quality || 'good',
      type: record.type,
      note: note || record.note,
      createBy: record.createBy,
    })

    // 删除本地的临时记录
    const index = sleepRecords.value.findIndex((r) => r.id === id)
    if (index !== -1) {
      sleepRecords.value.splice(index, 1)
    }

    return serverRecord
  } catch (error) {
    console.error('sync sleep record error:', error)
    // 同步失败，本地记录仍然保留
    return record
  }
}

/**
 * 删除睡眠记录 (本地实现,待 API 完善)
 * TODO: 集成 DELETE /sleep-records/{recordId} API
 */
export function deleteSleepRecord(id: string): boolean {
  const index = sleepRecords.value.findIndex((record) => record.id === id)
  if (index === -1) {
    return false
  }

  sleepRecords.value.splice(index, 1)
  setStorage(StorageKeys.SLEEP_RECORDS, sleepRecords.value)
  return true
}

/**
 * 本地查询方法
 */
export function getSleepRecords(): SleepRecord[] {
  return sleepRecords.value
}

export function getSleepRecordsByBabyId(babyId: string): SleepRecord[] {
  return sleepRecords.value.filter((record) => record.babyId === babyId)
}

export function getTodaySleepRecords(babyId: string): SleepRecord[] {
  const todayStart = getTodayStart()
  const todayEnd = getTodayEnd()

  return sleepRecords.value.filter(
    (record) =>
      record.babyId === babyId &&
      record.startTime >= todayStart &&
      record.startTime <= todayEnd
  )
}

export function getTodayTotalSleepDuration(babyId: string): number {
  const todayRecords = getTodaySleepRecords(babyId)
  let total = 0

  todayRecords.forEach((record) => {
    if (record.duration) {
      total += record.duration
    }
  })

  return total
}

export function getOngoingSleepRecord(babyId: string): SleepRecord | null {
  const records = sleepRecords.value.filter(
    (record) => record.babyId === babyId && !record.endTime
  )

  return records.length > 0 ? records[0] : null
}

export function getLastSleepRecord(babyId: string): SleepRecord | null {
  const records = sleepRecords.value
    .filter((record) => record.babyId === babyId)
    .sort((a, b) => b.startTime - a.startTime)

  return records.length > 0 ? records[0] : null
}

export { sleepRecords }
