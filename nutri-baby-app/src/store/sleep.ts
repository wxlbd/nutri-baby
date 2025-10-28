/**
 * ⚠️ 此文件已废弃 - 临时兼容层
 *
 * Store 架构已重构,睡眠记录数据应直接调用 API 层:
 * import * as sleepApi from '@/api/sleep'
 *
 * 请尽快迁移页面代码!
 */

console.error('[Deprecated] store/sleep.ts 已废弃,请使用 @/api/sleep')

export const getTodayTotalSleepDuration = () => {
  console.error('getTodayTotalSleepDuration() 已废弃,请直接调用 API')
  return 0
}

export const getSleepRecords = () => {
  console.error('getSleepRecords() 已废弃,请直接调用 API')
  return []
}

export const getTodaySleepRecords = () => {
  console.error('getTodaySleepRecords() 已废弃,请直接调用 API')
  return []
}

export const addSleepRecord = async () => {
  console.error('addSleepRecord() 已废弃,请使用 sleepApi.apiCreateSleepRecord()')
  throw new Error('此方法已废弃')
}

export const fetchSleepRecords = async () => {
  console.error('fetchSleepRecords() 已废弃,请使用 sleepApi.apiFetchSleepRecords()')
  return []
}

export const deleteSleepRecord = () => {
  console.error('deleteSleepRecord() 已废弃,请使用 sleepApi.apiDeleteSleepRecord()')
  return false
}

export const getCurrentSleepTimer = () => {
  console.error('getCurrentSleepTimer() 已废弃')
  return null
}

export const startSleepTimer = () => {
  console.error('startSleepTimer() 已废弃')
}

export const stopSleepTimer = () => {
  console.error('stopSleepTimer() 已废弃')
}
