/**
 * ⚠️ 此文件已废弃 - 临时兼容层
 *
 * Store 架构已重构,喂养记录数据应直接调用 API 层:
 * import * as feedingApi from '@/api/feeding'
 *
 * 请尽快迁移页面代码!
 */

console.error('[Deprecated] store/feeding.ts 已废弃,请使用 @/api/feeding')

// 导出空函数避免编译错误
export const getTodayTotalMilk = () => {
  console.error('getTodayTotalMilk() 已废弃,请直接调用 API')
  return 0
}

export const getLastFeedingRecord = () => {
  console.error('getLastFeedingRecord() 已废弃,请直接调用 API')
  return null
}

export const getTodayBreastfeedingStats = () => {
  console.error('getTodayBreastfeedingStats() 已废弃,请直接调用 API')
  return { count: 0, totalDuration: 0 }
}

export const getTodayFeedingRecords = () => {
  console.error('getTodayFeedingRecords() 已废弃,请直接调用 API')
  return []
}

export const getFeedingRecords = () => {
  console.error('getFeedingRecords() 已废弃,请直接调用 API')
  return []
}

export const addFeedingRecord = async () => {
  console.error('addFeedingRecord() 已废弃,请使用 feedingApi.apiCreateFeedingRecord()')
  throw new Error('此方法已废弃')
}

export const fetchFeedingRecords = async () => {
  console.error('fetchFeedingRecords() 已废弃,请使用 feedingApi.apiFetchFeedingRecords()')
  return []
}

export const deleteFeedingRecord = () => {
  console.error('deleteFeedingRecord() 已废弃,请使用 feedingApi.apiDeleteFeedingRecord()')
  return false
}
