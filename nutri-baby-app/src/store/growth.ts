/**
 * 成长记录状态管理 - API 渐进式集成版本
 *
 * 已集成 API:
 * - POST /growth-records (创建记录)
 * - GET /growth-records (查询记录列表)
 *
 * 待集成 API (使用本地实现):
 * - PUT /growth-records/{recordId} (更新记录) - API 待实现
 * - DELETE /growth-records/{recordId} (删除记录) - API 待实现
 */
import { ref } from 'vue'
import type { GrowthRecord } from '@/types'
import { StorageKeys, getStorage, setStorage } from '@/utils/storage'
import { get, post } from '@/utils/request'
import { generateId } from '@/utils/common'

/**
 * 成长记录接口（包含完整记录信息）
 */
export interface GrowthRecordItem {
  id: string
  babyId: string
  time: number
  height?: number // 身高(cm)
  weight?: number // 体重(kg)
  headCircumference?: number // 头围(cm)
  note?: string // 备注
  createBy: string
  createTime: number
}

// 成长记录列表
const growthRecords = ref<GrowthRecordItem[]>(
  getStorage<GrowthRecordItem[]>(StorageKeys.GROWTH_RECORDS) || []
)

/**
 * 从服务器获取成长记录列表
 *
 * API: GET /growth-records?babyId={babyId}&startTime={startTime}&endTime={endTime}&page={page}&pageSize={pageSize}
 */
export async function fetchGrowthRecords(params: {
  babyId: string
  startTime?: number
  endTime?: number
  page?: number
  pageSize?: number
}): Promise<GrowthRecordItem[]> {
  try {
    const response = await get<{
      records: any[]
      total: number
      page: number
      pageSize: number
    }>('/growth-records', params)

    if (response.code === 0 && response.data) {
      // 映射 API 响应到本地类型
      const records: GrowthRecordItem[] = response.data.records.map((item: any) => ({
        id: item.recordId,
        babyId: item.babyId,
        time: item.recordTime,
        height: item.height, // cm
        weight: item.weight ? item.weight / 1000 : undefined, // API 是克(g), 转为公斤(kg)
        headCircumference: item.headCircum, // cm
        note: item.note,
        createBy: item.createBy,
        createTime: item.createTime,
      }))

      growthRecords.value = records
      setStorage(StorageKeys.GROWTH_RECORDS, records)

      return records
    } else {
      throw new Error(response.message || '获取成长记录失败')
    }
  } catch (error: any) {
    console.error('fetch growth records error:', error)
    throw error
  }
}

/**
 * 添加成长记录
 *
 * API: POST /growth-records
 */
export async function addGrowthRecord(
  record: Omit<GrowthRecordItem, 'id' | 'createBy' | 'createTime'>
): Promise<GrowthRecordItem> {
  try {
    const response = await post<any>('/growth-records', {
      babyId: record.babyId,
      height: record.height, // cm
      weight: record.weight ? Math.round(record.weight * 1000) : undefined, // 公斤(kg) 转为 克(g)
      headCircum: record.headCircumference, // cm
      note: record.note,
      recordTime: record.time,
    })

    if (response.code === 0 && response.data) {
      const newRecord: GrowthRecordItem = {
        id: response.data.recordId,
        babyId: response.data.babyId,
        time: response.data.recordTime,
        height: response.data.height,
        weight: response.data.weight ? response.data.weight / 1000 : undefined,
        headCircumference: response.data.headCircum,
        note: response.data.note,
        createBy: response.data.createBy,
        createTime: response.data.createTime,
      }

      growthRecords.value.unshift(newRecord)
      setStorage(StorageKeys.GROWTH_RECORDS, growthRecords.value)

      uni.showToast({
        title: '记录成功',
        icon: 'success',
      })

      return newRecord
    } else {
      throw new Error(response.message || '添加记录失败')
    }
  } catch (error: any) {
    console.error('add growth record error:', error)
    uni.showToast({
      title: error.message || '记录失败',
      icon: 'none',
    })
    throw error
  }
}

/**
 * 更新成长记录 (本地实现,待 API 完善)
 * TODO: 集成 PUT /growth-records/{recordId} API
 */
export function updateGrowthRecord(
  id: string,
  data: Partial<Omit<GrowthRecordItem, 'id' | 'createBy' | 'createTime'>>
): boolean {
  const index = growthRecords.value.findIndex((record) => record.id === id)
  if (index === -1) {
    return false
  }

  growthRecords.value[index] = {
    ...growthRecords.value[index],
    ...data,
  }

  setStorage(StorageKeys.GROWTH_RECORDS, growthRecords.value)
  return true
}

/**
 * 删除成长记录 (本地实现,待 API 完善)
 * TODO: 集成 DELETE /growth-records/{recordId} API
 */
export function deleteGrowthRecord(id: string): boolean {
  const index = growthRecords.value.findIndex((record) => record.id === id)
  if (index === -1) {
    return false
  }

  growthRecords.value.splice(index, 1)
  setStorage(StorageKeys.GROWTH_RECORDS, growthRecords.value)
  return true
}

/**
 * 本地查询方法
 */
export function getGrowthRecordsByBabyId(babyId: string): GrowthRecordItem[] {
  return growthRecords.value
    .filter((record) => record.babyId === babyId)
    .sort((a, b) => b.time - a.time)
}

export function getLatestGrowthRecord(babyId: string): GrowthRecordItem | null {
  const records = getGrowthRecordsByBabyId(babyId)
  return records.length > 0 ? records[0] : null
}

/**
 * 获取成长曲线数据
 */
export function getGrowthCurveData(babyId: string): {
  dates: string[]
  heights: number[]
  weights: number[]
  headCircumferences: number[]
} {
  const records = getGrowthRecordsByBabyId(babyId).reverse() // 按时间正序

  const dates: string[] = []
  const heights: number[] = []
  const weights: number[] = []
  const headCircumferences: number[] = []

  records.forEach((record) => {
    const date = new Date(record.time)
    dates.push(`${date.getMonth() + 1}/${date.getDate()}`)

    if (record.height) heights.push(record.height)
    if (record.weight) weights.push(record.weight)
    if (record.headCircumference) headCircumferences.push(record.headCircumference)
  })

  return { dates, heights, weights, headCircumferences }
}

export { growthRecords }
