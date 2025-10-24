/**
 * 用户信息接口 (去家庭化架构)
 */
export interface UserInfo {
  openid: string
  nickName: string
  avatarUrl: string
  defaultBabyId?: string // 默认宝宝ID
  babies?: BabyProfile[] // 用户可访问的宝宝列表
  createTime: number
}

/**
 * 宝宝档案接口 (去家庭化架构)
 */
export interface BabyProfile {
  babyId: string  // 改为 babyId (与后端一致)
  name: string
  nickname?: string
  birthDate: string // YYYY-MM-DD 格式
  gender: 'male' | 'female'
  avatarUrl?: string
  creatorId: string // 创建者 openid
  familyGroup?: string // 可选的家庭分组名称
  createTime: number
  updateTime: number
}

/**
 * 宝宝协作者角色
 */
export type CollaboratorRole = 'admin' | 'editor' | 'viewer'

/**
 * 访问类型
 */
export type AccessType = 'permanent' | 'temporary'

/**
 * 宝宝协作者接口 (替代 FamilyMember)
 */
export interface BabyCollaborator {
  openid: string
  nickName: string
  avatarUrl: string
  role: CollaboratorRole
  accessType: AccessType
  expiresAt?: number // 临时权限过期时间
  joinTime: number
}

/**
 * 喂养类型
 */
export type FeedingType = 'breast' | 'bottle' | 'food'

/**
 * 母乳喂养记录
 */
export interface BreastFeeding {
  type: 'breast'
  side: 'left' | 'right' | 'both' // 喂养侧
  duration: number // 总时长(秒)
  leftDuration?: number // 左侧时长(秒)
  rightDuration?: number // 右侧时长(秒)
}

/**
 * 奶瓶喂养记录
 */
export interface BottleFeeding {
  type: 'bottle'
  bottleType: 'formula' | 'breast-milk' // 配方奶或母乳
  amount: number // 奶量(ml)
  unit: 'ml' | 'oz'
  remaining?: number // 剩余量
}

/**
 * 辅食记录
 */
export interface FoodFeeding {
  type: 'food'
  foodName: string // 辅食名称
  note?: string // 备注(接受程度、过敏反应等)
}

/**
 * 喂养记录联合类型
 */
export type FeedingDetail = BreastFeeding | BottleFeeding | FoodFeeding

/**
 * 喂养记录
 */
export interface FeedingRecord {
  id: string
  babyId: string
  time: number // 时间戳
  detail: FeedingDetail
  createBy: string // 创建人 openid
  createByName: string // 冗余:创建者昵称
  createByAvatar: string // 冗余:创建者头像
  createTime: number
}

/**
 * 排泄类型
 */
export type DiaperType = 'wet' | 'dirty' | 'both'

/**
 * 大便颜色
 */
export type PoopColor = 'yellow' | 'green' | 'brown' | 'black' | 'red' | 'white'

/**
 * 大便性状
 */
export type PoopTexture = 'watery' | 'loose' | 'paste' | 'soft' | 'formed' | 'hard'

/**
 * 排泄记录
 */
export interface DiaperRecord {
  id: string
  babyId: string
  time: number
  type: DiaperType
  poopColor?: PoopColor // 大便颜色(仅在有大便时)
  poopTexture?: PoopTexture // 大便性状(仅在有大便时)
  note?: string
  createBy: string
  createByName: string // 冗余:创建者昵称
  createByAvatar: string // 冗余:创建者头像
  createTime: number
}

/**
 * 睡眠类型
 */
export type SleepType = 'nap' | 'night'

/**
 * 睡眠记录
 */
export interface SleepRecord {
  id: string
  babyId: string
  startTime: number // 开始时间戳
  endTime?: number // 结束时间戳(进行中时为空)
  duration?: number // 时长(分钟)
  type: SleepType
  createBy: string
  createByName: string // 冗余:创建者昵称
  createByAvatar: string // 冗余:创建者头像
  createTime: number
}

/**
 * 其他记录类型
 */
export type OtherRecordType = 'spit' | 'pump' | 'growth' | 'medicine' | 'temperature' | 'bath' | 'milestone'

/**
 * 吐奶记录
 */
export interface SpitRecord {
  type: 'spit'
  time: number
}

/**
 * 泵奶记录
 */
export interface PumpRecord {
  type: 'pump'
  duration: number // 时长(分钟)
  amount: number // 产量(ml)
}

/**
 * 生长记录
 */
export interface GrowthRecord {
  id: string
  babyId: string
  time: number
  height?: number // 身高(cm)
  weight?: number // 体重(kg)
  headCircumference?: number // 头围(cm)
  createBy: string
  createByName: string // 冗余:创建者昵称
  createByAvatar: string // 冗余:创建者头像
  createTime: number
}

/**
 * 其他事件记录
 */
export interface OtherEventRecord {
  id: string
  babyId: string
  time: number
  recordType: OtherRecordType
  detail: SpitRecord | PumpRecord | GrowthRecord | {
    type: OtherRecordType
    note: string
  }
  createBy: string
  createByName: string // 冗余:创建者昵称
  createByAvatar: string // 冗余:创建者头像
  createTime: number
}

/**
 * 统一的记录类型
 */
export type RecordType = 'feeding' | 'diaper' | 'sleep' | 'other'

/**
 * 所有记录的联合类型
 */
export type Record = FeedingRecord | DiaperRecord | SleepRecord | OtherEventRecord

/**
 * 家庭成员角色
 * @deprecated 此类型已弃用,请使用 CollaboratorRole 替代
 */
export type FamilyRole = 'admin' | 'member'

/**
 * 家庭成员
 * @deprecated 此类型已弃用,请使用 BabyCollaborator 替代
 */
export interface FamilyMember {
  openid: string
  nickName: string
  avatarUrl: string
  role: FamilyRole
  joinTime: number
}

/**
 * 家庭信息
 * @deprecated 此类型已弃用,去家庭化架构已移除家庭概念
 */
export interface FamilyInfo {
  familyId: string
  familyName: string
  description?: string
  creatorId: string
  createTime: number
  members: FamilyMember[]
  babyIds: string[]  // 关联的宝宝ID列表
}

/**
 * 邀请码信息
 * @deprecated 此类型已弃用,请使用基于宝宝的邀请机制
 */
export interface InvitationInfo {
  invitationCode: string
  familyId: string
  familyName: string
  creatorName: string
  qrCodeUrl?: string
  shareUrl?: string
  expiresAt: number
  createTime: number
}

/**
 * 同步状态
 */
export type SyncStatus = 'idle' | 'syncing' | 'success' | 'error'

/**
 * 同步配置
 */
export interface SyncConfig {
  autoSync: boolean       // 是否自动同步
  syncInterval: number    // 同步间隔(毫秒)
  wifiOnly: boolean       // 仅Wi-Fi同步
  lastSyncTime: number    // 上次同步时间
}

/**
 * 疫苗类型
 */
export type VaccineType =
  | 'BCG'              // 卡介苗
  | 'HepB'             // 乙肝疫苗
  | 'OPV'              // 脊灰疫苗(口服)
  | 'IPV'              // 脊灰疫苗(注射)
  | 'DTaP'             // 百白破疫苗
  | 'MR'               // 麻风疫苗
  | 'MMR'              // 麻腮风疫苗
  | 'JE'               // 乙脑疫苗
  | 'MeningB'          // 流脑B疫苗
  | 'MeningAC'         // 流脑AC疫苗
  | 'HepA'             // 甲肝疫苗
  | 'Varicella'        // 水痘疫苗
  | 'Pneumococcal'     // 肺炎疫苗
  | 'Rotavirus'        // 轮状病毒疫苗
  | 'Hib'              // Hib疫苗
  | 'Influenza'        // 流感疫苗
  | 'Other'            // 其他

/**
 * 疫苗计划项
 */
export interface VaccinePlan {
  id: string
  vaccineType: VaccineType
  vaccineName: string      // 疫苗名称
  description?: string     // 说明
  ageInMonths: number      // 接种月龄
  doseNumber: number       // 第几针(1, 2, 3...)
  isRequired: boolean      // 是否必打
  reminderDays: number     // 提前提醒天数
}

/**
 * 疫苗接种记录
 */
export interface VaccineRecord {
  id: string
  babyId: string
  planId: string           // 关联的计划ID
  vaccineType: VaccineType
  vaccineName: string
  doseNumber: number
  vaccineDate: number      // 接种日期(时间戳)
  hospital?: string        // 接种医院
  batchNumber?: string     // 疫苗批号
  doctor?: string          // 接种医生
  reaction?: string        // 不良反应
  note?: string           // 备注
  createBy: string
  createByName: string // 冗余:创建者昵称
  createByAvatar: string // 冗余:创建者头像
  createTime: number
}

/**
 * 疫苗提醒状态
 */
export type VaccineReminderStatus = 'upcoming' | 'due' | 'overdue' | 'completed'

/**
 * 疫苗提醒
 */
export interface VaccineReminder {
  id: string
  babyId: string
  planId: string
  vaccineName: string
  doseNumber: number
  scheduledDate: number    // 预定接种日期
  status: VaccineReminderStatus
  reminderSent: boolean    // 是否已发送提醒
  createTime: number
}

/**
 * API 响应基础接口
 */
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
  timestamp?: number
}