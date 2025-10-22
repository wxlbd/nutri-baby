/**
 * 家庭管理状态 - 单家庭模式(重构版)
 *
 * 核心逻辑:
 * - 用户只能属于一个家庭
 * - 首次登录引导用户"创建家庭"或"加入家庭"
 * - 通过邀请码机制实现家庭成员协作
 */
import { ref, computed } from 'vue'
import type { FamilyInfo, FamilyMember, InvitationInfo } from '@/types'
import { StorageKeys, getStorage, setStorage } from '@/utils/storage'
import { get, post, put, del } from '@/utils/request'

// 当前家庭信息(单个)
const currentFamily = ref<FamilyInfo | null>(
  getStorage<FamilyInfo>(StorageKeys.CURRENT_FAMILY_ID) || null
)

// 是否有家庭
const hasFamily = computed(() => currentFamily.value !== null)

// 当前家庭ID
const currentFamilyId = computed(() => currentFamily.value?.familyId || '')

// 当前用户在家庭中的角色
const currentUserRole = computed(() => {
  if (!currentFamily.value) return null
  // 从家庭成员中找到当前用户(需要从 user store 获取 openid)
  // 这里暂时返回 null,具体实现需要和 user.ts 配合
  return null
})

/**
 * 获取当前用户的家庭信息
 *
 * API: GET /families/my
 * 响应: { familyId, familyName, description, role, joinTime, members, babyIds }
 *
 * 注意: 如果用户没有家庭,返回 404 或空数据
 */
export async function fetchMyFamily(): Promise<FamilyInfo | null> {
  try {
    const response = await get<FamilyInfo>('/families/my')

    if (response.code === 0 && response.data) {
      currentFamily.value = response.data
      setStorage(StorageKeys.CURRENT_FAMILY_ID, response.data)
      return response.data
    } else if (response.code === 404) {
      // 用户没有家庭
      currentFamily.value = null
      setStorage(StorageKeys.CURRENT_FAMILY_ID, null)
      return null
    } else {
      throw new Error(response.message || '获取家庭信息失败')
    }
  } catch (error: any) {
    console.error('fetch my family error:', error)
    // 如果是 404 错误,不抛出异常
    if (error.statusCode === 404) {
      currentFamily.value = null
      setStorage(StorageKeys.CURRENT_FAMILY_ID, null)
      return null
    }
    throw error
  }
}

/**
 * 创建家庭
 *
 * API: POST /families
 * 请求: { familyName, description? }
 * 响应: { familyId, familyName, description, role: 'admin', joinTime, members, babyIds }
 *
 * 限制: 用户已有家庭时不允许创建,返回错误
 */
export async function createFamily(
  name: string,
  description?: string
): Promise<FamilyInfo> {
  try {
    // 前置检查:如果已有家庭,禁止创建
    if (currentFamily.value) {
      throw new Error('您已经有家庭了,不能再创建新家庭')
    }

    const response = await post<FamilyInfo>('/families', {
      familyName: name,
      description,
    })

    if (response.code === 0 && response.data) {
      const newFamily = response.data
      currentFamily.value = newFamily
      setStorage(StorageKeys.CURRENT_FAMILY_ID, newFamily)

      uni.showToast({
        title: '家庭创建成功',
        icon: 'success',
      })

      return newFamily
    } else {
      throw new Error(response.message || '创建家庭失败')
    }
  } catch (error: any) {
    console.error('create family error:', error)
    uni.showToast({
      title: error.message || '创建家庭失败',
      icon: 'none',
    })
    throw error
  }
}

/**
 * 更新家庭信息
 *
 * API: PUT /families/{familyId}
 * 请求: { familyName, description? }
 * 响应: { code, message, data: null }
 *
 * 权限: 仅管理员可操作
 */
export async function updateFamily(
  data: { familyName: string; description?: string }
): Promise<boolean> {
  try {
    if (!currentFamily.value) {
      throw new Error('未找到当前家庭')
    }

    const response = await put(`/families/${currentFamily.value.familyId}`, data)

    if (response.code === 0) {
      // 更新本地数据
      currentFamily.value = {
        ...currentFamily.value,
        ...data,
      }
      setStorage(StorageKeys.CURRENT_FAMILY_ID, currentFamily.value)

      uni.showToast({
        title: '更新成功',
        icon: 'success',
      })

      return true
    } else {
      throw new Error(response.message || '更新家庭失败')
    }
  } catch (error: any) {
    console.error('update family error:', error)
    uni.showToast({
      title: error.message || '更新失败',
      icon: 'none',
    })
    throw error
  }
}

/**
 * 退出当前家庭
 *
 * API: POST /families/{familyId}/leave
 * 响应: { code, message, data: null }
 *
 * 注意事项:
 * - 如果是管理员且家庭中有其他成员,需先转让管理员权限
 * - 退出后可以创建新家庭或加入其他家庭
 */
export async function leaveFamily(): Promise<boolean> {
  try {
    if (!currentFamily.value) {
      throw new Error('未找到当前家庭')
    }

    const response = await post(`/families/${currentFamily.value.familyId}/leave`)

    if (response.code === 0) {
      // 清除本地数据
      currentFamily.value = null
      setStorage(StorageKeys.CURRENT_FAMILY_ID, null)

      uni.showToast({
        title: '已退出家庭',
        icon: 'success',
      })

      return true
    } else {
      throw new Error(response.message || '退出家庭失败')
    }
  } catch (error: any) {
    console.error('leave family error:', error)
    uni.showToast({
      title: error.message || '退出失败',
      icon: 'none',
    })
    throw error
  }
}

/**
 * 解散当前家庭(仅管理员)
 *
 * API: DELETE /families/{familyId}
 * 权限: 仅管理员可操作
 * 响应: { code, message, data: null }
 *
 * 注意: 解散后所有成员都会被移除
 */
export async function dissolveFamily(): Promise<boolean> {
  try {
    if (!currentFamily.value) {
      throw new Error('未找到当前家庭')
    }

    const response = await del(`/families/${currentFamily.value.familyId}`)

    if (response.code === 0) {
      // 清除本地数据
      currentFamily.value = null
      setStorage(StorageKeys.CURRENT_FAMILY_ID, null)

      uni.showToast({
        title: '家庭已解散',
        icon: 'success',
      })

      return true
    } else {
      throw new Error(response.message || '解散家庭失败')
    }
  } catch (error: any) {
    console.error('dissolve family error:', error)
    uni.showToast({
      title: error.message || '解散失败',
      icon: 'none',
    })
    throw error
  }
}

/**
 * 生成邀请码
 *
 * API: POST /families/invitations
 * 请求: { familyId, role, expireDays? }
 * 响应: { invitationCode, familyId, familyName, role, createBy, createTime, expireTime }
 */
export async function generateInvitation(
  role: 'admin' | 'member' = 'member',
  expireDays: number = 7
): Promise<InvitationInfo> {
  try {
    if (!currentFamily.value) {
      throw new Error('未找到当前家庭')
    }

    const response = await post<InvitationInfo>('/families/invitations', {
      familyId: currentFamily.value.familyId,
      role,
      expireDays,
    })

    if (response.code === 0 && response.data) {
      uni.showToast({
        title: '邀请码生成成功',
        icon: 'success',
      })

      return response.data
    } else {
      throw new Error(response.message || '生成邀请码失败')
    }
  } catch (error: any) {
    console.error('generate invitation error:', error)
    uni.showToast({
      title: error.message || '生成邀请码失败',
      icon: 'none',
    })
    throw error
  }
}

/**
 * 通过邀请码加入家庭
 *
 * API: POST /families/join
 * 请求: { invitationCode }
 * 响应: { familyId, familyName, description, role, joinTime, members, babyIds }
 *
 * 限制: 用户已有家庭时不允许加入,返回错误
 */
export async function joinFamilyByCode(code: string): Promise<FamilyInfo> {
  try {
    // 前置检查:如果已有家庭,禁止加入
    if (currentFamily.value) {
      throw new Error('您已经有家庭了,请先退出当前家庭')
    }

    const response = await post<FamilyInfo>('/families/join', {
      invitationCode: code,
    })

    if (response.code === 0 && response.data) {
      const newFamily = response.data
      currentFamily.value = newFamily
      setStorage(StorageKeys.CURRENT_FAMILY_ID, newFamily)

      uni.showToast({
        title: '加入家庭成功',
        icon: 'success',
      })

      return newFamily
    } else {
      throw new Error(response.message || '加入家庭失败')
    }
  } catch (error: any) {
    console.error('join family error:', error)
    uni.showToast({
      title: error.message || '加入家庭失败',
      icon: 'none',
    })
    throw error
  }
}

/**
 * 移除家庭成员(仅管理员)
 *
 * API: DELETE /families/{familyId}/members/{memberId}
 * 权限: 仅管理员可操作
 * 响应: { code, message, data: null }
 */
export async function removeFamilyMember(memberId: string): Promise<boolean> {
  try {
    if (!currentFamily.value) {
      throw new Error('未找到当前家庭')
    }

    const response = await del(
      `/families/${currentFamily.value.familyId}/members/${memberId}`
    )

    if (response.code === 0) {
      // 从本地数据中移除成员
      if (currentFamily.value.members) {
        const index = currentFamily.value.members.findIndex(
          m => m.openid === memberId
        )
        if (index !== -1) {
          currentFamily.value.members.splice(index, 1)
          setStorage(StorageKeys.CURRENT_FAMILY_ID, currentFamily.value)
        }
      }

      uni.showToast({
        title: '移除成功',
        icon: 'success',
      })

      return true
    } else {
      throw new Error(response.message || '移除成员失败')
    }
  } catch (error: any) {
    console.error('remove family member error:', error)
    uni.showToast({
      title: error.message || '移除失败',
      icon: 'none',
    })
    throw error
  }
}

/**
 * 转让管理员权限(仅管理员)
 *
 * API: POST /families/{familyId}/transfer
 * 请求: { newAdminId }
 * 响应: { code, message, data: null }
 */
export async function transferAdmin(newAdminId: string): Promise<boolean> {
  try {
    if (!currentFamily.value) {
      throw new Error('未找到当前家庭')
    }

    const response = await post(
      `/families/${currentFamily.value.familyId}/transfer`,
      { newAdminId }
    )

    if (response.code === 0) {
      // 刷新家庭信息
      await fetchMyFamily()

      uni.showToast({
        title: '转让成功',
        icon: 'success',
      })

      return true
    } else {
      throw new Error(response.message || '转让管理员失败')
    }
  } catch (error: any) {
    console.error('transfer admin error:', error)
    uni.showToast({
      title: error.message || '转让失败',
      icon: 'none',
    })
    throw error
  }
}

/**
 * 获取家庭成员列表
 */
export function getFamilyMembers(): FamilyMember[] {
  return currentFamily.value?.members || []
}

/**
 * 清除本地家庭数据(用于登出)
 */
export function clearFamilyData() {
  currentFamily.value = null
  setStorage(StorageKeys.CURRENT_FAMILY_ID, null)
}

export {
  currentFamily,
  hasFamily,
  currentFamilyId,
  currentUserRole,
}
