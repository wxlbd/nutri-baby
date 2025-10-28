<template>
  <view class="custom-navbar" :style="{ paddingTop: statusBarHeight * 2 + 'rpx' }">
    <view class="navbar-content" :style="{ height: navbarContentHeight + 'rpx' }">
      <!-- 左侧宝宝信息 -->
      <view class="baby-info" @click="goToBabyList">
        <view v-if="currentBaby" class="baby-content">
          <view class="baby-avatar">
            <image
              v-if="currentBaby.avatarUrl"
              :src="currentBaby.avatarUrl"
              mode="aspectFill"
              class="avatar-img"
            />
            <view v-else class="avatar-placeholder">
              {{ currentBaby.name ? currentBaby.name.charAt(0) : '宝' }}
            </view>
          </view>
          <view class="baby-text">
            <text class="baby-name">{{ currentBaby.name || '宝宝' }}</text>
            <text class="baby-age">{{ babyAge }}</text>
          </view>
          <nut-icon name="right" size="12" color="#999" class="arrow-icon" />
        </view>
        <view v-else class="add-baby-hint">
          <text>添加宝宝</text>
        </view>
      </view>

      <!-- 中间标题 -->
      <view class="navbar-title">
        <text>{{ title }}</text>
      </view>

      <!-- 右侧操作区（预留） -->
      <view class="navbar-actions">
        <slot name="right"></slot>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { babyList } from '@/store/baby'
import { getUserInfo } from '@/store/user'
import { calculateAge } from '@/utils/date'

interface Props {
  title?: string
}

const props = withDefaults(defineProps<Props>(), {
  title: '今日概览'
})

// 状态栏高度（px）
const statusBarHeight = ref(0)
// 导航栏内容高度（rpx）
const navbarContentHeight = ref(88)

// 获取当前宝宝 - 通过 defaultBabyId 从列表中匹配
// 注意: 这里使用 defaultBabyId 而不是 currentBabyId
// 原因: 导航栏应该显示用户设置的默认宝宝,而不是当前操作的宝宝
// currentBabyId 用于记录页面等地方记录数据时使用,可能与 defaultBabyId 不同
const currentBaby = computed(() => {
  const userInfo = getUserInfo()
  const defaultBabyId = userInfo?.defaultBabyId

  if (!defaultBabyId || !babyList.value) {
    return null
  }

  return babyList.value.find(baby => baby.babyId === defaultBabyId) || null
})

// 宝宝年龄
const babyAge = computed(() => {
  if (!currentBaby.value) return ''
  return calculateAge(currentBaby.value.birthDate)
})

// 跳转到宝宝列表
const goToBabyList = () => {
  uni.navigateTo({
    url: '/pages/baby/list/list'
  })
}

onMounted(() => {
  // 获取系统信息
  const systemInfo = uni.getSystemInfoSync()
  statusBarHeight.value = systemInfo.statusBarHeight || 0

  // 获取胶囊按钮信息（仅微信小程序）
  // #ifdef MP-WEIXIN
  try {
    const menuButton = uni.getMenuButtonBoundingClientRect()
    if (menuButton) {
      // 计算导航栏内容高度 = (胶囊底部 - 状态栏高度) + (胶囊顶部 - 状态栏高度)
      // 即：胶囊高度 + 胶囊上下边距
      const menuButtonHeight = menuButton.height
      const menuButtonTop = menuButton.top
      const contentHeight = (menuButtonTop - statusBarHeight.value) * 2 + menuButtonHeight
      // 转换为 rpx (px * 2)
      navbarContentHeight.value = Math.ceil(contentHeight * 2)
    }
  } catch (e) {
    console.warn('[CustomNavbar] 获取胶囊信息失败，使用默认高度', e)
  }
  // #endif
})

// 导出导航栏总高度供父组件使用
defineExpose({
  navbarTotalHeight: computed(() => statusBarHeight.value * 2 + navbarContentHeight.value)
})
</script>

<style lang="scss" scoped>
.custom-navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  background: #ffffff;
  z-index: 9999;
}

.navbar-content {
  // 高度由内联样式动态设置
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4rpx 32rpx; // 左右边距
}

// 左侧宝宝信息
.baby-info {
  flex-shrink: 0;
}

.baby-content {
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 8rpx 24rpx 8rpx 8rpx;
  background: #f5f7fa;
  border-radius: 40rpx;
  min-width: 200rpx;
  max-width: 280rpx;
}

.baby-avatar {
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
}

.avatar-img {
  width: 100%;
  height: 100%;
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 30rpx;
  font-weight: bold;
}

.baby-text {
  display: flex;
  flex-direction: column;
  gap: 2rpx;
  flex: 1;
  min-width: 0;
}

.baby-name {
  font-size: 22rpx;
  font-weight: 500;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.2;
}

.baby-age {
  font-size: 20rpx;
  color: #999;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.2;
}

.arrow-icon {
  flex-shrink: 0;
  margin-left: 4rpx;
}

.add-baby-hint {
  padding: 16rpx 32rpx;
  background: #f5f7fa;
  border-radius: 40rpx;
  font-size: 24rpx;
  color: #999;
}

// 中间标题
.navbar-title {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  font-size: 34rpx; // 标准导航栏标题大小 (17px = 34rpx)
  font-weight: 600;
  color: #000;
  pointer-events: none;
}

// 右侧操作区
.navbar-actions {
  display: flex;
  align-items: center;
  gap: 24rpx;
}
</style>
