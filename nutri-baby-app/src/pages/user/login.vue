<template>
  <view class="login-container">
    <view class="login-content">
      <!-- Logo -->
      <view class="logo-section">
        <image class="logo" src="/static/logo.png" mode="aspectFit" />
        <text class="app-name">宝宝喂养日志</text>
        <text class="app-desc">科学记录 · 智能分析 · 健康成长</text>
      </view>

      <!-- 登录按钮 -->
      <view class="login-actions">
        <nut-button
          type="primary"
          size="large"
          block
          :loading="loading"
          @click="handleLogin"
        >
          <text v-if="!loading">微信一键登录</text>
          <text v-else>登录中...</text>
        </nut-button>

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
import { ref } from 'vue'
import { wxLogin } from '@/store/user'

const loading = ref(false)

/**
 * 登录后重定向逻辑
 *
 * 流程: 登录成功 -> 首页
 */
const redirectAfterLogin = () => {
  console.log('[Login] 登录成功,跳转到首页')
  uni.reLaunch({
    url: '/pages/index/index'
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
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40rpx;
}

.login-content {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.logo-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 120rpx;
}

.logo {
  width: 200rpx;
  height: 200rpx;
  margin-bottom: 40rpx;
  border-radius: 40rpx;
  background: white;
  padding: 20rpx;
}

.app-name {
  font-size: 48rpx;
  font-weight: bold;
  color: white;
  margin-bottom: 20rpx;
}

.app-desc {
  font-size: 28rpx;
  color: rgba(255, 255, 255, 0.8);
}

.login-actions {
  width: 100%;
}

.privacy-tips {
  margin-top: 40rpx;
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.7);
  text-align: center;
  line-height: 40rpx;
}

.link {
  color: white;
  text-decoration: underline;
}
</style>