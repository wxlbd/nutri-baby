/**
 * 疫苗管理状态管理 - 按宝宝 ID 隔离存储版本
 *
 * 架构变更 (v2.0):
 * - 疫苗计划按宝宝 ID 独立存储 (vaccine_plans_{babyId})
 * - 每个宝宝有独立的疫苗计划副本,支持个性化调整
 * - 保留全局记录和提醒列表,通过 babyId 过滤
 *
 * 已集成 API:
 * - GET /babies/{babyId}/vaccine-plans (获取疫苗计划)
 * - POST /babies/{babyId}/vaccine-records (创建疫苗接种记录)
 * - GET /babies/{babyId}/vaccine-reminders (获取疫苗提醒列表)
 * - GET /babies/{babyId}/vaccine-statistics (获取疫苗接种统计)
 *
 * 待集成 API (使用本地实现):
 * - PUT /vaccine-records/{recordId} (更新记录) - API 待实现
 * - DELETE /vaccine-records/{recordId} (删除记录) - API 待实现
 */
import { ref, computed } from 'vue'
import type { VaccinePlan, VaccineRecord, VaccineReminder, VaccineType, VaccineReminderStatus } from '@/types'
import { StorageKeys, getStorage, setStorage } from '@/utils/storage'
import { generateId } from '@/utils/common'
// 使用相对路径导入,避免通过 @/store/index.ts 产生循环依赖
import { getCurrentBaby, currentBabyId } from './baby'
import { get, post } from '@/utils/request'

// 默认疫苗计划(国家免疫规划疫苗)
const defaultVaccinePlans: Omit<VaccinePlan, 'id'>[] = [
  // 乙肝疫苗
  { vaccineType: 'HepB', vaccineName: '乙肝疫苗', ageInMonths: 0, doseNumber: 1, isRequired: true, reminderDays: 3, description: '出生24小时内接种' },
  { vaccineType: 'HepB', vaccineName: '乙肝疫苗', ageInMonths: 1, doseNumber: 2, isRequired: true, reminderDays: 7, description: '满1个月接种' },
  { vaccineType: 'HepB', vaccineName: '乙肝疫苗', ageInMonths: 6, doseNumber: 3, isRequired: true, reminderDays: 7, description: '满6个月接种' },

  // 卡介苗
  { vaccineType: 'BCG', vaccineName: '卡介苗', ageInMonths: 0, doseNumber: 1, isRequired: true, reminderDays: 3, description: '出生后尽快接种' },

  // 脊灰疫苗
  { vaccineType: 'OPV', vaccineName: '脊灰疫苗', ageInMonths: 2, doseNumber: 1, isRequired: true, reminderDays: 7, description: '满2个月接种' },
  { vaccineType: 'OPV', vaccineName: '脊灰疫苗', ageInMonths: 3, doseNumber: 2, isRequired: true, reminderDays: 7, description: '满3个月接种' },
  { vaccineType: 'OPV', vaccineName: '脊灰疫苗', ageInMonths: 4, doseNumber: 3, isRequired: true, reminderDays: 7, description: '满4个月接种' },
  { vaccineType: 'OPV', vaccineName: '脊灰疫苗', ageInMonths: 18, doseNumber: 4, isRequired: true, reminderDays: 7, description: '满18个月接种' },

  // 百白破疫苗
  { vaccineType: 'DTaP', vaccineName: '百白破疫苗', ageInMonths: 3, doseNumber: 1, isRequired: true, reminderDays: 7, description: '满3个月接种' },
  { vaccineType: 'DTaP', vaccineName: '百白破疫苗', ageInMonths: 4, doseNumber: 2, isRequired: true, reminderDays: 7, description: '满4个月接种' },
  { vaccineType: 'DTaP', vaccineName: '百白破疫苗', ageInMonths: 5, doseNumber: 3, isRequired: true, reminderDays: 7, description: '满5个月接种' },
  { vaccineType: 'DTaP', vaccineName: '百白破疫苗', ageInMonths: 18, doseNumber: 4, isRequired: true, reminderDays: 7, description: '满18个月接种' },

  // 麻风疫苗/麻腮风疫苗
  { vaccineType: 'MR', vaccineName: '麻风疫苗', ageInMonths: 8, doseNumber: 1, isRequired: true, reminderDays: 7, description: '满8个月接种' },
  { vaccineType: 'MMR', vaccineName: '麻腮风疫苗', ageInMonths: 18, doseNumber: 1, isRequired: true, reminderDays: 7, description: '满18个月接种' },

  // 乙脑疫苗
  { vaccineType: 'JE', vaccineName: '乙脑疫苗', ageInMonths: 8, doseNumber: 1, isRequired: true, reminderDays: 7, description: '满8个月接种' },
  { vaccineType: 'JE', vaccineName: '乙脑疫苗', ageInMonths: 24, doseNumber: 2, isRequired: true, reminderDays: 7, description: '满2岁接种' },

  // 流脑疫苗
  { vaccineType: 'MeningAC', vaccineName: '流脑AC疫苗', ageInMonths: 6, doseNumber: 1, isRequired: true, reminderDays: 7, description: '满6个月接种' },
  { vaccineType: 'MeningAC', vaccineName: '流脑AC疫苗', ageInMonths: 9, doseNumber: 2, isRequired: true, reminderDays: 7, description: '满9个月接种' },
  { vaccineType: 'MeningAC', vaccineName: '流脑AC疫苗', ageInMonths: 36, doseNumber: 3, isRequired: true, reminderDays: 7, description: '满3岁接种' },

  // 甲肝疫苗
  { vaccineType: 'HepA', vaccineName: '甲肝疫苗', ageInMonths: 18, doseNumber: 1, isRequired: true, reminderDays: 7, description: '满18个月接种' },
]

/**
 * 获取指定宝宝的疫苗计划存储键
 */
function getVaccinePlansStorageKey(babyId: string): string {
  return `${StorageKeys.VACCINE_PLANS_PREFIX}${babyId}`
}

/**
 * 获取指定宝宝的疫苗计划
 * @param babyId 宝宝 ID
 * @returns 疫苗计划列表
 */
export function getVaccinePlansByBabyId(babyId: string): VaccinePlan[] {
  const key = getVaccinePlansStorageKey(babyId)
  return getStorage<VaccinePlan[]>(key) || []
}

/**
 * 保存指定宝宝的疫苗计划
 * @param babyId 宝宝 ID
 * @param plans 疫苗计划列表
 */
function saveVaccinePlansForBaby(babyId: string, plans: VaccinePlan[]): void {
  const key = getVaccinePlansStorageKey(babyId)
  setStorage(key, plans)
  console.log(`[Vaccine] 保存宝宝 ${babyId} 的疫苗计划:`, plans.length, '条')
}

// 当前宝宝的疫苗计划 (动态计算,基于 currentBabyId)
const vaccinePlans = computed<VaccinePlan[]>(() => {
  if (!currentBabyId.value) return []
  return getVaccinePlansByBabyId(currentBabyId.value)
})

// 疫苗接种记录
const vaccineRecords = ref<VaccineRecord[]>(
  getStorage<VaccineRecord[]>(StorageKeys.VACCINE_RECORDS) || []
)

// 疫苗提醒列表
const vaccineReminders = ref<VaccineReminder[]>(
  getStorage<VaccineReminder[]>(StorageKeys.VACCINE_REMINDERS) || []
)

/**
 * 从服务器获取疫苗计划
 *
 * API: GET /babies/{babyId}/vaccine-plans
 */
export async function fetchVaccinePlans(babyId: string): Promise<{
  plans: VaccinePlan[]
  total: number
  completed: number
  percentage: number
}> {
  try {
    const response = await get<{
      plans: any[]
      total: number
      completed: number
      percentage: number
    }>(`/babies/${babyId}/vaccine-plans`)

    if (response.code === 0 && response.data) {
      // 映射 API 响应到本地类型
      const plans: VaccinePlan[] = response.data.plans.map((item: any) => ({
        id: item.planId,
        vaccineType: item.vaccineType,
        vaccineName: item.vaccineName,
        description: item.description,
        ageInMonths: item.ageInMonths,
        doseNumber: item.doseNumber,
        isRequired: item.isRequired,
        reminderDays: item.reminderDays,
      }))

      // 保存到指定宝宝的存储位置
      saveVaccinePlansForBaby(babyId, plans)

      return {
        plans,
        total: response.data.total,
        completed: response.data.completed,
        percentage: response.data.percentage,
      }
    } else {
      throw new Error(response.message || '获取疫苗计划失败')
    }
  } catch (error: any) {
    console.error('fetch vaccine plans error:', error)
    // 降级到本地数据
    const plans = getVaccinePlansByBabyId(babyId)
    return {
      plans,
      total: plans.length,
      completed: 0,
      percentage: 0,
    }
  }
}

/**
 * 为指定宝宝初始化疫苗计划
 * @param babyId 宝宝 ID
 * @param force 是否强制重新初始化(覆盖已有计划)
 */
export function initializeVaccinePlans(babyId: string, force: boolean = false) {
  // 首先尝试数据迁移
  migrateLegacyVaccinePlans(babyId)

  const existingPlans = getVaccinePlansByBabyId(babyId)

  if (existingPlans.length > 0 && !force) {
    console.log(`[Vaccine] 宝宝 ${babyId} 已有疫苗计划 (${existingPlans.length} 条),跳过初始化`)
    return
  }

  const plans = defaultVaccinePlans.map(plan => ({
    ...plan,
    id: generateId()
  }))

  saveVaccinePlansForBaby(babyId, plans)
  console.log(`[Vaccine] 已为宝宝 ${babyId} 初始化疫苗计划:`, plans.length, '条')
}

/**
 * 根据宝宝出生日期生成疫苗提醒
 * @param babyId 宝宝 ID
 * @param birthDate 出生日期字符串 (YYYY-MM-DD)
 */
export function generateRemindersForBaby(babyId: string, birthDate: string) {
  console.log('[Vaccine] 为宝宝生成疫苗提醒:', babyId, birthDate)

  // 获取该宝宝的疫苗计划
  const plans = getVaccinePlansByBabyId(babyId)
  if (plans.length === 0) {
    console.warn(`[Vaccine] 宝宝 ${babyId} 没有疫苗计划,无法生成提醒`)
    return
  }

  const birthTime = new Date(birthDate).getTime()
  const reminders: VaccineReminder[] = []

  plans.forEach(plan => {
    // 计算预定接种日期
    const scheduledDate = birthTime + plan.ageInMonths * 30 * 24 * 60 * 60 * 1000

    // 检查是否已接种
    const hasRecord = vaccineRecords.value.some(
      record => record.babyId === babyId &&
                record.planId === plan.id &&
                record.doseNumber === plan.doseNumber
    )

    if (!hasRecord) {
      const reminder: VaccineReminder = {
        id: generateId(),
        babyId,
        planId: plan.id,
        vaccineName: plan.vaccineName,
        doseNumber: plan.doseNumber,
        scheduledDate,
        status: getReminderStatus(scheduledDate),
        reminderSent: false,
        createTime: Date.now()
      }
      reminders.push(reminder)
    }
  })

  // 更新该宝宝的提醒列表
  vaccineReminders.value = [
    ...vaccineReminders.value.filter(r => r.babyId !== babyId),
    ...reminders
  ]
  setStorage(StorageKeys.VACCINE_REMINDERS, vaccineReminders.value)

  console.log(`[Vaccine] 已为宝宝 ${babyId} 生成 ${reminders.length} 条疫苗提醒`)
}

/**
 * 获取提醒状态
 */
function getReminderStatus(scheduledDate: number): VaccineReminderStatus {
  const now = Date.now()
  const daysDiff = (scheduledDate - now) / (24 * 60 * 60 * 1000)

  if (daysDiff > 7) return 'upcoming'       // 7天以上
  if (daysDiff >= 0) return 'due'           // 7天内
  if (daysDiff >= -30) return 'overdue'     // 逾期30天内
  return 'overdue'                          // 逾期超过30天
}

/**
 * 添加疫苗接种记录
 *
 * API: POST /babies/{babyId}/vaccine-records
 */
export async function addVaccineRecord(
  record: Omit<VaccineRecord, 'id' | 'createTime'>
): Promise<VaccineRecord> {
  try {
    const response = await post<any>(`/babies/${record.babyId}/vaccine-records`, {
      planId: record.planId,
      vaccineType: record.vaccineType,
      vaccineName: record.vaccineName,
      doseNumber: record.doseNumber,
      vaccineDate: record.vaccineDate,
      hospital: record.hospital,
      batchNumber: record.batchNumber,
      doctor: record.doctor,
      reaction: record.reaction,
      note: record.note,
    })

    if (response.code === 0 && response.data) {
      const newRecord: VaccineRecord = {
        id: response.data.recordId,
        babyId: response.data.babyId,
        planId: response.data.planId,
        vaccineType: response.data.vaccineType,
        vaccineName: response.data.vaccineName,
        doseNumber: response.data.doseNumber,
        vaccineDate: response.data.vaccineDate,
        hospital: response.data.hospital,
        batchNumber: response.data.batchNumber,
        doctor: response.data.doctor,
        reaction: response.data.reaction,
        note: response.data.note,
        createBy: response.data.createBy,
        createTime: response.data.createTime,
      }

      vaccineRecords.value.unshift(newRecord)
      setStorage(StorageKeys.VACCINE_RECORDS, vaccineRecords.value)

      // 更新对应提醒状态为已完成
      const reminder = vaccineReminders.value.find(
        r => r.planId === record.planId && r.babyId === record.babyId
      )
      if (reminder) {
        reminder.status = 'completed'
        setStorage(StorageKeys.VACCINE_REMINDERS, vaccineReminders.value)
      }

      return newRecord
    } else {
      throw new Error(response.message || '添加疫苗记录失败')
    }
  } catch (error: any) {
    console.error('add vaccine record error:', error)
    // 降级到本地存储
    const newRecord: VaccineRecord = {
      ...record,
      id: generateId(),
      createTime: Date.now(),
    }

    vaccineRecords.value.unshift(newRecord)
    setStorage(StorageKeys.VACCINE_RECORDS, vaccineRecords.value)

    // 更新对应提醒状态为已完成
    const reminder = vaccineReminders.value.find(
      r => r.planId === record.planId && r.babyId === record.babyId
    )
    if (reminder) {
      reminder.status = 'completed'
      setStorage(StorageKeys.VACCINE_REMINDERS, vaccineReminders.value)
    }

    return newRecord
  }
}

/**
 * 从服务器获取疫苗提醒列表
 *
 * API: GET /babies/{babyId}/vaccine-reminders
 */
export async function fetchVaccineReminders(params: {
  babyId: string
  status?: VaccineReminderStatus
  limit?: number
}): Promise<{
  reminders: VaccineReminder[]
  total: number
  upcoming: number
  due: number
  overdue: number
}> {
  try {
    const queryParams: any = {}
    if (params.status) queryParams.status = params.status
    if (params.limit) queryParams.limit = params.limit

    const response = await get<{
      reminders: any[]
      total: number
      upcoming: number
      due: number
      overdue: number
    }>(`/babies/${params.babyId}/vaccine-reminders`, queryParams)

    if (response.code === 0 && response.data) {
      // 映射 API 响应到本地类型
      const reminders: VaccineReminder[] = response.data.reminders.map((item: any) => ({
        id: item.reminderId,
        babyId: item.babyId,
        planId: item.planId,
        vaccineName: item.vaccineName,
        doseNumber: item.doseNumber,
        scheduledDate: item.scheduledDate,
        status: item.status,
        reminderSent: item.reminderSent,
        createTime: item.createTime,
      }))

      vaccineReminders.value = reminders
      setStorage(StorageKeys.VACCINE_REMINDERS, reminders)

      return {
        reminders,
        total: response.data.total,
        upcoming: response.data.upcoming,
        due: response.data.due,
        overdue: response.data.overdue,
      }
    } else {
      throw new Error(response.message || '获取疫苗提醒失败')
    }
  } catch (error: any) {
    console.error('fetch vaccine reminders error:', error)
    // 降级到本地数据
    const reminders = getVaccineRemindersByBabyId(params.babyId)
    return {
      reminders,
      total: reminders.length,
      upcoming: reminders.filter(r => r.status === 'upcoming').length,
      due: reminders.filter(r => r.status === 'due').length,
      overdue: reminders.filter(r => r.status === 'overdue').length,
    }
  }
}

/**
 * 从服务器获取疫苗接种统计
 *
 * API: GET /babies/{babyId}/vaccine-statistics
 */
export async function fetchVaccineStatistics(babyId: string): Promise<{
  total: number
  completed: number
  pending: number
  overdue: number
  percentage: number
  nextVaccine?: {
    vaccineName: string
    doseNumber: number
    scheduledDate: number
    daysUntilDue: number
  }
  recentRecords: VaccineRecord[]
}> {
  try {
    const response = await get<{
      total: number
      completed: number
      pending: number
      overdue: number
      percentage: number
      nextVaccine?: any
      recentRecords: any[]
    }>(`/babies/${babyId}/vaccine-statistics`)

    if (response.code === 0 && response.data) {
      // 映射最近记录
      const recentRecords: VaccineRecord[] = response.data.recentRecords.map((item: any) => ({
        id: item.recordId,
        babyId: item.babyId,
        planId: item.planId,
        vaccineType: item.vaccineType,
        vaccineName: item.vaccineName,
        doseNumber: item.doseNumber,
        vaccineDate: item.vaccineDate,
        hospital: item.hospital,
        batchNumber: item.batchNumber,
        doctor: item.doctor,
        reaction: item.reaction,
        note: item.note,
        createBy: item.createBy,
        createTime: item.createTime,
      }))

      return {
        total: response.data.total,
        completed: response.data.completed,
        pending: response.data.pending,
        overdue: response.data.overdue,
        percentage: response.data.percentage,
        nextVaccine: response.data.nextVaccine,
        recentRecords,
      }
    } else {
      throw new Error(response.message || '获取疫苗统计失败')
    }
  } catch (error: any) {
    console.error('fetch vaccine statistics error:', error)
    // 降级到本地数据
    const stats = getVaccineCompletionStats(babyId)
    const reminders = getVaccineRemindersByBabyId(babyId)
    const records = getVaccineRecordsByBabyId(babyId)

    return {
      total: stats.total,
      completed: stats.completed,
      pending: reminders.filter(r => r.status !== 'completed').length,
      overdue: reminders.filter(r => r.status === 'overdue').length,
      percentage: stats.percentage,
      recentRecords: records.slice(0, 5),
    }
  }
}

/**
 * 获取宝宝的疫苗记录
 */
export function getVaccineRecordsByBabyId(babyId: string): VaccineRecord[] {
  return vaccineRecords.value
    .filter(record => record.babyId === babyId)
    .sort((a, b) => b.vaccineDate - a.vaccineDate)
}

/**
 * 获取宝宝的疫苗提醒
 */
export function getVaccineRemindersByBabyId(babyId: string): VaccineReminder[] {
  return vaccineReminders.value
    .filter(reminder => reminder.babyId === babyId && reminder.status !== 'completed')
    .sort((a, b) => a.scheduledDate - b.scheduledDate)
}

/**
 * 获取即将到期的提醒(7天内)
 */
export function getUpcomingReminders(babyId?: string): VaccineReminder[] {
  const now = Date.now()
  const sevenDaysLater = now + 7 * 24 * 60 * 60 * 1000

  return vaccineReminders.value.filter(reminder => {
    const matchesBaby = !babyId || reminder.babyId === babyId
    const isDueOrUpcoming = reminder.status === 'due' || reminder.status === 'overdue'
    const isWithinRange = reminder.scheduledDate <= sevenDaysLater

    return matchesBaby && isDueOrUpcoming && isWithinRange
  })
}

/**
 * 获取逾期的提醒
 */
export function getOverdueReminders(babyId?: string): VaccineReminder[] {
  return vaccineReminders.value.filter(reminder => {
    const matchesBaby = !babyId || reminder.babyId === babyId
    return matchesBaby && reminder.status === 'overdue'
  })
}

/**
 * 更新疫苗记录
 */
export function updateVaccineRecord(
  id: string,
  data: Partial<Omit<VaccineRecord, 'id' | 'createTime'>>
): boolean {
  const index = vaccineRecords.value.findIndex(r => r.id === id)
  if (index === -1) return false

  vaccineRecords.value[index] = {
    ...vaccineRecords.value[index],
    ...data
  }

  setStorage(StorageKeys.VACCINE_RECORDS, vaccineRecords.value)
  return true
}

/**
 * 删除疫苗记录
 */
export function deleteVaccineRecord(id: string): boolean {
  const index = vaccineRecords.value.findIndex(r => r.id === id)
  if (index === -1) return false

  const record = vaccineRecords.value[index]
  vaccineRecords.value.splice(index, 1)
  setStorage(StorageKeys.VACCINE_RECORDS, vaccineRecords.value)

  // 恢复对应提醒
  const reminder = vaccineReminders.value.find(
    r => r.planId === record.planId && r.babyId === record.babyId
  )
  if (reminder) {
    reminder.status = getReminderStatus(reminder.scheduledDate)
    setStorage(StorageKeys.VACCINE_REMINDERS, vaccineReminders.value)
  }

  return true
}

/**
 * 获取疫苗计划详情
 * @param planId 计划 ID
 * @param babyId 宝宝 ID (可选,未提供时使用当前宝宝)
 */
export function getVaccinePlanById(planId: string, babyId?: string): VaccinePlan | null {
  const targetBabyId = babyId || currentBabyId.value
  if (!targetBabyId) return null

  const plans = getVaccinePlansByBabyId(targetBabyId)
  return plans.find(p => p.id === planId) || null
}

/**
 * 更新所有提醒状态
 */
export function updateAllReminderStatus() {
  let updated = false
  vaccineReminders.value.forEach(reminder => {
    if (reminder.status !== 'completed') {
      const newStatus = getReminderStatus(reminder.scheduledDate)
      if (newStatus !== reminder.status) {
        reminder.status = newStatus
        updated = true
      }
    }
  })

  if (updated) {
    setStorage(StorageKeys.VACCINE_REMINDERS, vaccineReminders.value)
  }
}

/**
 * 获取疫苗完成度统计
 */
export function getVaccineCompletionStats(babyId: string): {
  total: number
  completed: number
  percentage: number
} {
  const reminders = vaccineReminders.value.filter(r => r.babyId === babyId)
  const completed = reminders.filter(r => r.status === 'completed').length
  const total = reminders.length

  return {
    total,
    completed,
    percentage: total > 0 ? Math.round((completed / total) * 100) : 0
  }
}

/**
 * 数据迁移:将旧的全局疫苗计划迁移到按宝宝存储
 * @param babyId 宝宝 ID
 * @deprecated 仅用于一次性数据迁移
 */
export function migrateLegacyVaccinePlans(babyId: string): void {
  // 检查新存储是否已存在
  const existingPlans = getVaccinePlansByBabyId(babyId)
  if (existingPlans.length > 0) {
    console.log(`[Vaccine] 宝宝 ${babyId} 已有疫苗计划,跳过迁移`)
    return
  }

  // 尝试读取旧的全局存储
  const legacyPlans = getStorage<VaccinePlan[]>(StorageKeys.VACCINE_PLANS)
  if (!legacyPlans || legacyPlans.length === 0) {
    console.log('[Vaccine] 未找到旧的疫苗计划数据,跳过迁移')
    return
  }

  // 迁移到新存储
  saveVaccinePlansForBaby(babyId, legacyPlans)
  console.log(`[Vaccine] 已将 ${legacyPlans.length} 条疫苗计划迁移到宝宝 ${babyId}`)
}

/**
 * ==================== 疫苗计划管理 API ====================
 */

/**
 * 从服务器初始化疫苗计划(基于模板)
 * API: POST /babies/{babyId}/vaccine-plans/initialize
 */
export async function initializeVaccinePlansFromServer(babyId: string, force: boolean = false): Promise<boolean> {
  try {
    const response = await post<{
      totalPlans: number
      plans: any[]
      message: string
    }>(`/babies/${babyId}/vaccine-plans/initialize`, { force })

    if (response.code === 0 && response.data) {
      // 映射并保存计划
      const plans: VaccinePlan[] = response.data.plans.map((item: any) => ({
        id: item.planId,
        vaccineType: item.vaccineType,
        vaccineName: item.vaccineName,
        description: item.description,
        ageInMonths: item.ageInMonths,
        doseNumber: item.doseNumber,
        isRequired: item.isRequired,
        reminderDays: item.reminderDays,
      }))

      saveVaccinePlansForBaby(babyId, plans)

      uni.showToast({
        title: response.data.message || '初始化成功',
        icon: 'success'
      })
      return true
    } else {
      throw new Error(response.message || '初始化疫苗计划失败')
    }
  } catch (error: any) {
    console.error('initialize vaccine plans error:', error)
    // 降级到本地初始化
    initializeVaccinePlans(babyId, force)
    return false
  }
}

/**
 * 创建自定义疫苗计划
 * API: POST /babies/{babyId}/vaccine-plans
 */
export async function createVaccinePlan(
  babyId: string,
  plan: Omit<VaccinePlan, 'id'>
): Promise<VaccinePlan | null> {
  try {
    const response = await post<any>(`/babies/${babyId}/vaccine-plans`, {
      vaccineType: plan.vaccineType,
      vaccineName: plan.vaccineName,
      description: plan.description,
      ageInMonths: plan.ageInMonths,
      doseNumber: plan.doseNumber,
      isRequired: plan.isRequired,
      reminderDays: plan.reminderDays,
    })

    if (response.code === 0 && response.data) {
      const newPlan: VaccinePlan = {
        id: response.data.planId,
        vaccineType: response.data.vaccineType,
        vaccineName: response.data.vaccineName,
        description: response.data.description,
        ageInMonths: response.data.ageInMonths,
        doseNumber: response.data.doseNumber,
        isRequired: response.data.isRequired,
        reminderDays: response.data.reminderDays,
      }

      // 更新本地存储
      const plans = getVaccinePlansByBabyId(babyId)
      plans.push(newPlan)
      saveVaccinePlansForBaby(babyId, plans)

      uni.showToast({
        title: '添加成功',
        icon: 'success'
      })
      return newPlan
    } else {
      throw new Error(response.message || '添加疫苗计划失败')
    }
  } catch (error: any) {
    console.error('create vaccine plan error:', error)
    uni.showToast({
      title: error.message || '添加失败',
      icon: 'none'
    })
    return null
  }
}

/**
 * 更新疫苗计划
 * API: PUT /vaccine-plans/{planId}
 */
export async function updateVaccinePlan(
  planId: string,
  updates: Partial<Omit<VaccinePlan, 'id' | 'vaccineType'>>
): Promise<boolean> {
  try {
    const response = await post<any>(`/vaccine-plans/${planId}`, updates, 'PUT')

    if (response.code === 0 && response.data) {
      // 更新本地存储
      if (currentBabyId.value) {
        const plans = getVaccinePlansByBabyId(currentBabyId.value)
        const index = plans.findIndex(p => p.id === planId)
        if (index !== -1) {
          plans[index] = { ...plans[index], ...updates }
          saveVaccinePlansForBaby(currentBabyId.value, plans)
        }
      }

      uni.showToast({
        title: '更新成功',
        icon: 'success'
      })
      return true
    } else {
      throw new Error(response.message || '更新疫苗计划失败')
    }
  } catch (error: any) {
    console.error('update vaccine plan error:', error)
    uni.showToast({
      title: error.message || '更新失败',
      icon: 'none'
    })
    return false
  }
}

/**
 * 删除疫苗计划
 * API: DELETE /vaccine-plans/{planId}
 */
export async function deleteVaccinePlan(planId: string): Promise<boolean> {
  try {
    const response = await post<any>(`/vaccine-plans/${planId}`, {}, 'DELETE')

    if (response.code === 0) {
      // 更新本地存储
      if (currentBabyId.value) {
        const plans = getVaccinePlansByBabyId(currentBabyId.value)
        const filtered = plans.filter(p => p.id !== planId)
        saveVaccinePlansForBaby(currentBabyId.value, filtered)
      }

      uni.showToast({
        title: '删除成功',
        icon: 'success'
      })
      return true
    } else {
      throw new Error(response.message || '删除疫苗计划失败')
    }
  } catch (error: any) {
    console.error('delete vaccine plan error:', error)
    uni.showToast({
      title: error.message || '删除失败',
      icon: 'none'
    })
    return false
  }
}

/**
 * 导出当前宝宝的疫苗计划 (computed)
 * 导出记录和提醒列表 (ref)
 */
export { vaccinePlans, vaccineRecords, vaccineReminders }
