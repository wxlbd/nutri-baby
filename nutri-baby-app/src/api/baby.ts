/**
 * 宝宝管理 API 接口
 * 职责: 纯 API 调用,无状态,无副作用
 */
import { get, post, put, del } from '@/utils/request'

// ============ 类型定义 ============

/**
 * API 响应: 宝宝详情
 */
export interface BabyResponse {
  babyId: string
  babyName: string
  nickname?: string
  gender: 'male' | 'female'
  birthDate: string
  avatarUrl?: string
  creatorId: string
  familyGroup?: string
  height?: number
  weight?: number
  createTime: number
  updateTime: number
}

/**
 * API 请求: 创建宝宝
 */
export interface CreateBabyRequest {
  babyName: string
  gender: 'male' | 'female'
  birthDate: string
  nickname?: string
  avatarUrl?: string
  familyGroup?: string
  copyCollaboratorsFrom?: string
}

/**
 * API 请求: 更新宝宝
 */
export interface UpdateBabyRequest {
  babyName?: string
  nickname?: string
  gender?: 'male' | 'female'
  birthDate?: string
  avatarUrl?: string
  familyGroup?: string
  height?: number
  weight?: number
}

// ============ API 函数 ============

/**
 * 获取用户可访问的宝宝列表
 *
 * @returns Promise<BabyResponse[]>
 */
export async function apiFetchBabyList(): Promise<BabyResponse[]> {
  const response = await get<BabyResponse[]>('/babies')
  return response.data || []
}

/**
 * 获取宝宝详情
 *
 * @param babyId 宝宝ID
 * @returns Promise<BabyResponse>
 */
export async function apiFetchBabyDetail(babyId: string): Promise<BabyResponse> {
  const response = await get<BabyResponse>(`/babies/${babyId}`)
  if (!response.data) {
    throw new Error(response.message || '获取宝宝详情失败')
  }
  return response.data
}

/**
 * 创建宝宝
 *
 * @param data 创建请求数据
 * @returns Promise<BabyResponse>
 */
export async function apiCreateBaby(data: CreateBabyRequest): Promise<BabyResponse> {
  const response = await post<BabyResponse>('/babies', data)
  if (!response.data) {
    throw new Error(response.message || '添加宝宝失败')
  }
  return response.data
}

/**
 * 更新宝宝信息
 *
 * @param babyId 宝宝ID
 * @param data 更新数据
 * @returns Promise<void>
 */
export async function apiUpdateBaby(
  babyId: string,
  data: UpdateBabyRequest
): Promise<void> {
  const response = await put(`/babies/${babyId}`, data)
  if (response.code !== 0) {
    throw new Error(response.message || '更新宝宝信息失败')
  }
}

/**
 * 删除宝宝
 *
 * @param babyId 宝宝ID
 * @returns Promise<void>
 */
export async function apiDeleteBaby(babyId: string): Promise<void> {
  const response = await del(`/babies/${babyId}`)
  if (response.code !== 0) {
    throw new Error(response.message || '删除宝宝失败')
  }
}
