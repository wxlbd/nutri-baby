/**
 * 用户状态管理 - 去家庭化架构
 */
import { ref } from 'vue'
import type { UserInfo } from '@/types'
import { StorageKeys, getStorage, setStorage, removeStorage } from '@/utils/storage'
import { post, get, put } from '@/utils/request'

// 用户信息 - 延迟初始化
const userInfo = ref<UserInfo | null>(null)

// Token - 延迟初始化
const token = ref<string | null>(null)

// 是否已登录
const isLoggedIn = ref<boolean>(false)

// 是否为新用户 (首次登录且无宝宝)
const isNewUser = ref<boolean>(false)

// 初始化标记
let isInitialized = false

// 延迟初始化 - 仅在首次访问时从存储读取
function initializeIfNeeded() {
  if (!isInitialized) {
    userInfo.value = getStorage<UserInfo>(StorageKeys.USER_INFO) || null
    token.value = getStorage<string>(StorageKeys.TOKEN) || null
    isLoggedIn.value = !!userInfo.value
    isInitialized = true
  }
}

/**
 * 设置用户信息
 */
export function setUserInfo(info: UserInfo) {
  userInfo.value = info
  isLoggedIn.value = true
  setStorage(StorageKeys.USER_INFO, info)
}

/**
 * 设置 Token
 */
export function setToken(newToken: string) {
  token.value = newToken
  setStorage(StorageKeys.TOKEN, newToken)
}

/**
 * 获取用户信息
 */
export function getUserInfo() {
  initializeIfNeeded()
  return userInfo.value
}

/**
 * 获取 Token
 */
export function getToken() {
  return token.value
}

/**
 * 清除用户信息(退出登录)
 */
export function clearUserInfo() {
  userInfo.value = null
  isLoggedIn.value = false
  token.value = null
  removeStorage(StorageKeys.USER_INFO)
  removeStorage(StorageKeys.TOKEN)
}

/**
 * 检查登录状态
 */
export function checkLoginStatus(): boolean {
  return isLoggedIn.value && !!token.value
}

/**
 * 微信登录 - 集成后端 API (去家庭化架构)
 *
 * API: POST /auth/wechat-login
 * 请求: { code, nickName?, avatarUrl? }
 * 响应: { token, userInfo, isNewUser }
 */
export async function wxLogin(): Promise<UserInfo> {
  return new Promise((resolve, reject) => {
    // 1. 调用 uni.login 获取 code
    uni.login({
      provider: 'weixin',
      success: async (loginRes) => {
        console.log('uni.login success:', loginRes)

        try {
          // 2. 调用后端登录接口
          const response = await post<{
            token: string
            userInfo: UserInfo
            isNewUser: boolean
          }>('/auth/wechat-login', {
            code: loginRes.code,
            // 注意: 由于 getUserProfile 已废弃,nickName 和 avatarUrl 可以不传
            // 或者使用头像昵称填写组件获取
          })

          console.log('API login response:', response)

          if (response.code === 0 && response.data) {
            const { token: newToken, userInfo: user, isNewUser: newUser } = response.data

            // 3. 保存 Token 和用户信息
            setToken(newToken)
            setUserInfo(user)
            isNewUser.value = newUser

            uni.showToast({
              title: '登录成功',
              icon: 'success',
              duration: 1500,
            })

            resolve(user)
          } else {
            throw new Error(response.message || '登录失败')
          }
        } catch (error: any) {
          console.error('login API error:', error)
          uni.showToast({
            title: error.message || '登录失败',
            icon: 'none',
          })
          reject(error)
        }
      },
      fail: (err) => {
        console.error('uni.login fail:', err)
        uni.showToast({
          title: '微信登录失败',
          icon: 'none',
        })
        reject(err)
      },
    })
  })
}

/**
 * 刷新 Token
 *
 * API: POST /auth/refresh-token
 * Headers: Authorization: Bearer {token}
 * 响应: { token, expiresIn }
 */
export async function refreshToken(): Promise<string> {
  try {
    const response = await post<{
      token: string
      expiresIn: number
    }>('/auth/refresh-token')

    if (response.code === 0 && response.data) {
      const newToken = response.data.token
      setToken(newToken)
      return newToken
    } else {
      throw new Error(response.message || 'Token 刷新失败')
    }
  } catch (error: any) {
    console.error('refresh token error:', error)
    // Token 刷新失败,清除登录状态
    clearUserInfo()
    throw error
  }
}

/**
 * 从服务器获取用户信息
 *
 * API: GET /auth/user-info
 * Headers: Authorization: Bearer {token}
 * 响应: { openid, nickName, avatarUrl, defaultBabyId, createTime, lastLoginTime }
 */
export async function fetchUserInfo(): Promise<UserInfo> {
  try {
    const response = await get<UserInfo>('/auth/user-info')

    if (response.code === 0 && response.data) {
      setUserInfo(response.data)
      return response.data
    } else {
      throw new Error(response.message || '获取用户信息失败')
    }
  } catch (error: any) {
    console.error('fetch user info error:', error)
    throw error
  }
}

/**
 * 设置默认宝宝
 *
 * API: PUT /auth/default-baby
 * Headers: Authorization: Bearer {token}
 * 请求: { babyId }
 * 响应: { code: 0, message: "success", data: null }
 */
export async function setDefaultBaby(babyId: string): Promise<void> {
  try {
    const response = await put('/auth/default-baby', { babyId })

    if (response.code === 0) {
      // 更新本地用户信息中的默认宝宝ID
      if (userInfo.value) {
        userInfo.value.defaultBabyId = babyId
        setStorage(StorageKeys.USER_INFO, userInfo.value)
      }

      uni.showToast({
        title: '设置成功',
        icon: 'success',
      })
    } else {
      throw new Error(response.message || '设置默认宝宝失败')
    }
  } catch (error: any) {
    console.error('set default baby error:', error)
    uni.showToast({
      title: error.message || '设置失败',
      icon: 'none',
    })
    throw error
  }
}

/**
 * 退出登录
 */
export async function logout() {
  // 清除本地存储的用户信息和 Token
  clearUserInfo()

  // 跳转到登录页
  uni.reLaunch({
    url: '/pages/user/login',
  })

  uni.showToast({
    title: '已退出登录',
    icon: 'success',
  })
}

/**
 * 获取是否为新用户
 */
export function getIsNewUser() {
  return isNewUser.value
}

/**
 * 设置是否为新用户
 */
export function setIsNewUser(value: boolean) {
  isNewUser.value = value
}

export { userInfo, isLoggedIn, token, isNewUser }

// 确保导出时也触发初始化检查
if (typeof window !== 'undefined') {
  // 浏览器环境下,在模块加载完成后立即初始化
  setTimeout(() => initializeIfNeeded(), 0)
}
