/**
 * 认证 API 接口
 * 职责: 纯 API 调用,无状态,无副作用
 */
import { get, post, put } from '@/utils/request'
import type { UserInfo } from '@/types'

// ============ 类型定义 ============

/**
 * API 请求: 微信登录
 */
export interface WechatLoginRequest {
  code: string
  nickName?: string
  avatarUrl?: string
}

/**
 * API 响应: 微信登录
 */
export interface WechatLoginResponse {
  token: string
  userInfo: UserInfo
  isNewUser: boolean
}

/**
 * API 响应: 刷新 Token
 */
export interface RefreshTokenResponse {
  token: string
  expiresIn: number
}

// ============ API 函数 ============

/**
 * 微信登录
 *
 * @param data 登录请求数据
 * @returns Promise<WechatLoginResponse>
 */
export async function apiWechatLogin(data: WechatLoginRequest): Promise<WechatLoginResponse> {
  const response = await post<WechatLoginResponse>('/auth/wechat-login', data)
  if (!response.data) {
    throw new Error(response.message || '登录失败')
  }
  return response.data
}

/**
 * 刷新 Token
 *
 * @returns Promise<RefreshTokenResponse>
 */
export async function apiRefreshToken(): Promise<RefreshTokenResponse> {
  const response = await post<RefreshTokenResponse>('/auth/refresh-token')
  if (!response.data) {
    throw new Error(response.message || 'Token 刷新失败')
  }
  return response.data
}

/**
 * 获取用户信息
 *
 * @returns Promise<UserInfo>
 */
export async function apiFetchUserInfo(): Promise<UserInfo> {
  const response = await get<UserInfo>('/auth/user-info')
  if (!response.data) {
    throw new Error(response.message || '获取用户信息失败')
  }
  return response.data
}

/**
 * 设置默认宝宝
 *
 * @param babyId 宝宝ID
 * @returns Promise<void>
 */
export async function apiSetDefaultBaby(babyId: string): Promise<void> {
  const response = await put('/auth/default-baby', { babyId })
  if (response.code !== 0) {
    throw new Error(response.message || '设置默认宝宝失败')
  }
}
