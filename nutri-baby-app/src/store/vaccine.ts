/**
 * ⚠️ 此文件已废弃 - 临时兼容层
 *
 * Store 架构已重构,疫苗管理数据应直接调用 API 层:
 * import * as vaccineApi from '@/api/vaccine'
 *
 * 请尽快迁移页面代码!
 */

console.error('[Deprecated] store/vaccine.ts 已废弃,请使用 @/api/vaccine')

export const getVaccinePlans = () => {
  console.error('getVaccinePlans() 已废弃,请直接调用 API')
  return []
}

export const getVaccineRecords = () => {
  console.error('getVaccineRecords() 已废弃,请直接调用 API')
  return []
}

export const getUpcomingReminders = () => {
  console.error('getUpcomingReminders() 已废弃,请直接调用 API')
  return []
}

export const getVaccineStats = () => {
  console.error('getVaccineStats() 已废弃,请直接调用 API')
  return {
    totalPlans: 0,
    completedCount: 0,
    upcomingCount: 0,
    overdueCount: 0,
    completionRate: 0
  }
}

export const fetchVaccinePlans = async () => {
  console.error('fetchVaccinePlans() 已废弃,请使用 vaccineApi.apiFetchVaccinePlans()')
  return []
}

export const fetchVaccineRecords = async () => {
  console.error('fetchVaccineRecords() 已废弃,请使用 vaccineApi.apiFetchVaccineRecords()')
  return []
}

export const fetchVaccineReminders = async () => {
  console.error('fetchVaccineReminders() 已废弃,请使用 vaccineApi.apiFetchVaccineReminders()')
  return []
}

export const addVaccineRecord = async () => {
  console.error('addVaccineRecord() 已废弃,请使用 vaccineApi.apiCreateVaccineRecord()')
  throw new Error('此方法已废弃')
}

export const deleteVaccineRecord = () => {
  console.error('deleteVaccineRecord() 已废弃,请使用 vaccineApi.apiDeleteVaccineRecord()')
  return false
}

export const markReminderSent = async () => {
  console.error('markReminderSent() 已废弃,请使用 vaccineApi.apiMarkReminderSent()')
  throw new Error('此方法已废弃')
}

// 特殊函数 - 暂时保留占位,但标记为废弃
export const initializeVaccinePlansFromServer = async () => {
  console.error('initializeVaccinePlansFromServer() 已废弃')
  console.error('TODO: 直接在页面中调用 vaccineApi.apiFetchVaccinePlans()')
  return []
}

export const generateRemindersForBaby = async () => {
  console.error('generateRemindersForBaby() 已废弃')
  console.error('TODO: 直接在页面中调用 vaccineApi.apiFetchVaccineReminders()')
  return []
}
