/**
 * ⚠️ 此文件已废弃 - 临时兼容层
 *
 * Store 架构已重构,成长记录数据应直接调用 API 层:
 * import * as growthApi from '@/api/growth'
 *
 * 请尽快迁移页面代码!
 */

console.error('[Deprecated] store/growth.ts 已废弃,请使用 @/api/growth')

export const getGrowthRecords = () => {
  console.error('getGrowthRecords() 已废弃,请直接调用 API')
  return []
}

export const addGrowthRecord = async () => {
  console.error('addGrowthRecord() 已废弃,请使用 growthApi.apiCreateGrowthRecord()')
  throw new Error('此方法已废弃')
}

export const fetchGrowthRecords = async () => {
  console.error('fetchGrowthRecords() 已废弃,请使用 growthApi.apiFetchGrowthRecords()')
  return []
}

export const deleteGrowthRecord = () => {
  console.error('deleteGrowthRecord() 已废弃,请使用 growthApi.apiDeleteGrowthRecord()')
  return false
}

export const getLatestGrowthRecord = () => {
  console.error('getLatestGrowthRecord() 已废弃,请直接调用 API')
  return null
}
