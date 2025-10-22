/**
 * 排泄记录状态管理 - API 渐进式集成版本
 *
 * 已集成 API:
 * - POST /diaper-records (创建记录)
 * - GET /diaper-records (查询记录列表)
 *
 * 待集成 API (使用本地实现):
 * - PUT /diaper-records/{recordId} (更新记录) - API 待实现
 * - DELETE /diaper-records/{recordId} (删除记录) - API 待实现
 */
import { ref } from 'vue'
import type { DiaperRecord } from '@/types'
import { StorageKeys, getStorage, setStorage } from '@/utils/storage'
import { get, post } from '@/utils/request'
import { getTodayStart, getTodayEnd } from '@/utils/date'

// 排泄记录列表
const diaperRecords = ref<DiaperRecord[]>(
  getStorage<DiaperRecord[]>(StorageKeys.DIAPER_RECORDS) || []
)

/**
 * 从服务器获取换尿布记录列表
 *
 * API: GET /diaper-records?babyId={babyId}&startTime={startTime}&endTime={endTime}&page={page}&pageSize={pageSize}
 */
export async function fetchDiaperRecords(params: {
  babyId: string
  startTime?: number
  endTime?: number
  page?: number
  pageSize?: number
}): Promise<DiaperRecord[]> {
  try {
    const response = await get<{
      records: any[]
      total: number
      page: number
      pageSize: number
    }>('/diaper-records', params)

    if (response.code === 0 && response.data) {
      // 映射 API 响应到本地类型
      const records: DiaperRecord[] = response.data.records.map((item: any) => ({
        id: item.recordId,
        babyId: item.babyId,
        time: item.changeTime,
        type: mapDiaperTypeFromAPI(item.diaperType),
        note: item.note,
        createBy: item.createBy,
        createTime: item.createTime,
        // API 暂不支持以下字段，使用默认值
        poopColor: undefined,
        poopTexture: undefined,
      }))

      diaperRecords.value = records
      setStorage(StorageKeys.DIAPER_RECORDS, records)

      return records
    } else {
      throw new Error(response.message || '获取换尿布记录失败')
    }
  } catch (error: any) {
    console.error('fetch diaper records error:', error)
    throw error
  }
}

/**
 * 添加换尿布记录
 *
 * API: POST /diaper-records
 */
export async function addDiaperRecord(
  babyId: string,
  type: DiaperRecord['type'],
  createBy: string,
  options?: {
    poopColor?: DiaperRecord['poopColor']
    poopTexture?: DiaperRecord['poopTexture']
    note?: string
    time?: number
  }
): Promise<DiaperRecord> {
  try {
    // 构建备注，包含大便颜色和性状信息
    let fullNote = options?.note || ''
    if (options?.poopColor || options?.poopTexture) {
      const details: string[] = []
      if (options.poopColor) details.push(`颜色: ${options.poopColor}`)
      if (options.poopTexture) details.push(`性状: ${options.poopTexture}`)
      fullNote = fullNote
        ? `${fullNote} (${details.join(', ')})`
        : details.join(', ')
    }

    const response = await post<any>('/diaper-records', {
      babyId,
      diaperType: mapDiaperTypeToAPI(type),
      note: fullNote,
      changeTime: options?.time || Date.now(),
    })

    if (response.code === 0 && response.data) {
      const newRecord: DiaperRecord = {
        id: response.data.recordId,
        babyId: response.data.babyId,
        time: response.data.changeTime,
        type,
        poopColor: options?.poopColor,
        poopTexture: options?.poopTexture,
        note: options?.note,
        createBy: response.data.createBy,
        createTime: response.data.createTime,
      }

      diaperRecords.value.unshift(newRecord)
      setStorage(StorageKeys.DIAPER_RECORDS, diaperRecords.value)

      uni.showToast({
        title: '记录成功',
        icon: 'success',
      })

      return newRecord
    } else {
      throw new Error(response.message || '添加记录失败')
    }
  } catch (error: any) {
    console.error('add diaper record error:', error)
    uni.showToast({
      title: error.message || '记录失败',
      icon: 'none',
    })
    throw error
  }
}

/**
 * 删除换尿布记录 (本地实现,待 API 完善)
 * TODO: 集成 DELETE /diaper-records/{recordId} API
 */
export function deleteDiaperRecord(id: string): boolean {
  const index = diaperRecords.value.findIndex((record) => record.id === id)
  if (index === -1) {
    return false
  }

  diaperRecords.value.splice(index, 1)
  setStorage(StorageKeys.DIAPER_RECORDS, diaperRecords.value)
  return true
}

/**
 * 映射尿布类型到 API 格式
 */
function mapDiaperTypeToAPI(type: DiaperRecord['type']): string {
  const mapping: Record<string, string> = {
    wet: 'pee',
    dirty: 'poop',
    both: 'both',
  }
  return mapping[type] || type
}

/**
 * 映射 API 尿布类型到本地格式
 */
function mapDiaperTypeFromAPI(apiType: string): DiaperRecord['type'] {
  const mapping: Record<string, DiaperRecord['type']> = {
    pee: 'wet',
    poop: 'dirty',
    both: 'both',
  }
  return mapping[apiType] || ('wet' as DiaperRecord['type'])
}

/**
 * 本地查询方法
 */
export function getDiaperRecords(): DiaperRecord[] {
  return diaperRecords.value
}

export function getDiaperRecordsByBabyId(babyId: string): DiaperRecord[] {
  return diaperRecords.value.filter((record) => record.babyId === babyId)
}

export function getTodayDiaperRecords(babyId: string): DiaperRecord[] {
  const todayStart = getTodayStart()
  const todayEnd = getTodayEnd()

  return diaperRecords.value.filter(
    (record) =>
      record.babyId === babyId &&
      record.time >= todayStart &&
      record.time <= todayEnd
  )
}

export function getTodayDiaperCount(babyId: string): number {
  return getTodayDiaperRecords(babyId).length
}

export function getLastDiaperRecord(babyId: string): DiaperRecord | null {
  const records = diaperRecords.value
    .filter((record) => record.babyId === babyId)
    .sort((a, b) => b.time - a.time)

  return records.length > 0 ? records[0] : null
}

export { diaperRecords }
