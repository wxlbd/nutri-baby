/**
 * ⚠️ 此文件已废弃 - 临时兼容层
 *
 * Store 架构已重构,换尿布记录数据应直接调用 API 层:
 * import * as diaperApi from '@/api/diaper'
 *
 * 请尽快迁移页面代码!
 */

console.error('[Deprecated] store/diaper.ts 已废弃,请使用 @/api/diaper')

export const getTodayDiaperCount = () => {
  console.error('getTodayDiaperCount() 已废弃,请直接调用 API')
  return 0
}

export const getDiaperRecords = () => {
  console.error('getDiaperRecords() 已废弃,请直接调用 API')
  return []
}

export const getTodayDiaperRecords = () => {
  console.error('getTodayDiaperRecords() 已废弃,请直接调用 API')
  return []
}

export const addDiaperRecord = async () => {
  console.error('addDiaperRecord() 已废弃,请使用 diaperApi.apiCreateDiaperRecord()')
  throw new Error('此方法已废弃')
}

export const fetchDiaperRecords = async () => {
  console.error('fetchDiaperRecords() 已废弃,请使用 diaperApi.apiFetchDiaperRecords()')
  return []
}

export const deleteDiaperRecord = () => {
  console.error('deleteDiaperRecord() 已废弃,请使用 diaperApi.apiDeleteDiaperRecord()')
  return false
}
