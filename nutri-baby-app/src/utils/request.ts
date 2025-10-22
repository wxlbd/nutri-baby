/**
 * HTTP 请求工具
 */

import type { ApiResponse } from '@/types'
import { StorageKeys, getStorage } from '@/utils/storage'

// API 基础配置
const BASE_URL = import.meta.env.VITE_API_BASE_URL || 'https://api.example.com'
const TIMEOUT = 30000

/**
 * 请求配置接口
 */
interface RequestConfig {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: any
  header?: Record<string, string>
  timeout?: number
}

/**
 * 获取请求头
 */
function getHeaders(): Record<string, string> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  // 从本地存储获取 token (使用 getStorage 函数确保正确解析 JSON)
  const token = getStorage<string>(StorageKeys.TOKEN)
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  return headers
}

/**
 * 封装的请求方法
 */
export function request<T = any>(config: RequestConfig): Promise<ApiResponse<T>> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}${config.url}`,
      method: config.method || 'GET',
      data: config.data,
      header: {
        ...getHeaders(),
        ...config.header,
      },
      timeout: config.timeout || TIMEOUT,
      success: (res) => {
        const data = res.data as ApiResponse<T>

        // 请求成功
        if (res.statusCode === 200) {
          if (data.code === 0) {
            resolve(data)
          } else {
            // 业务错误
            uni.showToast({
              title: data.message || '请求失败',
              icon: 'none',
              duration: 2000,
            })
            reject(data)
          }
        } else if (res.statusCode === 401) {
          // token 过期,跳转登录
          uni.showToast({
            title: '登录已过期,请重新登录',
            icon: 'none',
          })
          // 清除 token (使用 StorageKeys 常量)
          uni.removeStorageSync(StorageKeys.TOKEN)
          // 跳转到登录页
          setTimeout(() => {
            uni.reLaunch({
              url: '/pages/user/login',
            })
          }, 1500)
          reject(data)
        } else {
          // HTTP 错误
          uni.showToast({
            title: `请求失败: ${res.statusCode}`,
            icon: 'none',
          })
          reject(data)
        }
      },
      fail: (err) => {
        console.error('request error:', err)
        uni.showToast({
          title: '网络请求失败',
          icon: 'none',
        })
        reject(err)
      },
    })
  })
}

/**
 * GET 请求
 */
export function get<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
  return request<T>({
    url,
    method: 'GET',
    data,
  })
}

/**
 * POST 请求
 */
export function post<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
  return request<T>({
    url,
    method: 'POST',
    data,
  })
}

/**
 * PUT 请求
 */
export function put<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
  return request<T>({
    url,
    method: 'PUT',
    data,
  })
}

/**
 * DELETE 请求
 */
export function del<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
  return request<T>({
    url,
    method: 'DELETE',
    data,
  })
}

/**
 * 文件上传
 */
export function uploadFile(filePath: string, name: string = 'file'): Promise<any> {
  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: `${BASE_URL}/upload`,
      filePath,
      name,
      header: getHeaders(),
      success: (res) => {
        if (res.statusCode === 200) {
          const data = JSON.parse(res.data)
          resolve(data)
        } else {
          reject(res)
        }
      },
      fail: reject,
    })
  })
}