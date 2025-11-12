<template>
  <view class="login-container">
    <!-- 背景图片 -->
    <image class="login-bg-image" src="/static/login-background.png" mode="aspectFill" />

    <view class="login-content">
      <!-- Logo 和标题 -->
      <view class="logo-section">
        <view class="logo-wrapper">
          <image class="logo" src="/static/logo.png" mode="aspectFit" />
        </view>
        <text class="app-name">宝宝喂养时刻</text>
        <text class="app-desc">记录宝宝的每一刻</text>
      </view>

      <!-- 登录按钮 -->
      <view class="login-actions">
        <wd-button
          type="primary"
          size="large"
          block
          :loading="loading"
          @click="handleLogin"
          class="login-btn"
        >
          <text v-if="!loading">微信一键登录</text>
          <text v-else>登录中...</text>
        </wd-button>

        <view class="privacy-tips">
          登录即表示同意
          <text class="link">《用户协议》</text>
          和
          <text class="link">《隐私政策》</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wxLogin, isLoggedIn } from '@/store/user'
import { StorageKeys } from '@/utils/storage'

const loading = ref(false)


/**
 * 登录后重定向逻辑
 *
 * 流程:
 * 1. 检查是否有待处理的邀请码 -> 跳转到加入页面
 * 2. 否则 -> 跳转到首页
 *
 * 防止无限重定向:
 * 1. 使用 switchTab 代替 reLaunch (因为首页是 tabBar 页面)
 * 2. 设置延迟确保登录状态完全保存
 */
const redirectAfterLogin = () => {
  console.log('[Login] 登录成功,检查重定向目标')

  // 检查是否有待处理的邀请码（从扫码进入但未登录的场景）
  const pendingCode = uni.getStorageSync(StorageKeys.PENDING_INVITE_CODE)
  console.log('[Login] Checking PENDING_INVITE_CODE:', pendingCode)

  if (pendingCode) {
    console.log('[Login] 检测到待处理的邀请码,跳转到加入页面:', pendingCode)

    // 清除缓存（避免重复跳转）
    uni.removeStorageSync(StorageKeys.PENDING_INVITE_CODE)

    // 跳转回加入页面
    uni.reLaunch({
      url: `/pages/baby/join/join?code=${pendingCode}`,
    })

    return
  }

  // 正常跳转到首页
  console.log('[Login] 登录成功,跳转到首页')

  // 对于 tabBar 页面,应该使用 switchTab 而不是 reLaunch
  // 这样可以避免页面重新加载和生命周期冲突
  uni.switchTab({
    url: '/pages/index/index',
    fail: (err) => {
      console.error('[Login] 跳转失败,使用 reLaunch 降级:', err)
      // 如果 switchTab 失败,降级使用 reLaunch
      uni.reLaunch({
        url: '/pages/index/index'
      })
    }
  })
}

// 处理登录 - 必须在用户点击事件中调用
const handleLogin = async () => {
  if (loading.value) return

  loading.value = true

  try {
    // 微信登录
    await wxLogin()

    // 延迟跳转,让用户看到成功提示
    setTimeout(() => {
      redirectAfterLogin()
    }, 1500)

  } catch (error) {
    console.error('登录失败:', error)
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.login-container {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40rpx 20rpx;
  overflow: hidden;
}

.login-bg-image {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 0;
}

.login-content {
  position: relative;
  z-index: 10;
  width: 100%;
  max-width: 420rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
}

// Logo 部分
.logo-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 120rpx;
  text-align: center;
}

.logo-wrapper {
  position: relative;
  width: 200rpx;
  height: 200rpx;
  margin-bottom: 40rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo {
  width: 200rpx;
  height: 200rpx;
  border-radius: 100%;
  background: white;
  object-fit: contain;
  box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, 0.15);
}

.app-name {
  font-size: 48rpx;
  font-weight: bold;
  color: white;
  margin-bottom: 16rpx;
  letter-spacing: 2rpx;
}

.app-desc {
  font-size: 28rpx;
  color: rgba(255, 255, 255, 0.9);
  font-weight: 500;
}

// 登录操作区域
.login-actions {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.login-btn {
  :deep(.wd-button) {
    background: white !important;
    color: #1a9e4b !important;
    font-weight: 600;
    font-size: 32rpx;
    border-radius: 16rpx;
    height: 88rpx;
    box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, 0.15);

    &:active {
      background: rgba(255, 255, 255, 0.9) !important;
    }
  }
}

.privacy-tips {
  margin-top: 24rpx;
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.8);
  text-align: center;
  line-height: 36rpx;
  padding: 0 20rpx;

  .link {
    color: white;
    text-decoration: underline;
    font-weight: 500;
  }
}
</style>