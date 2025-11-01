<template>
  <view class="qrcode-container">
    <!-- 页面标题 -->
    <view class="header">
      <view class="title">面对面扫码</view>
      <view class="subtitle">邀请家人加入{{ babyName }}的协作</view>
    </view>

    <!-- 二维码卡片 -->
    <view class="qrcode-card">
      <!-- 二维码显示区域 -->
      <view class="qrcode-wrapper">
        <image
          v-if="qrcodeUrl"
          :src="qrcodeUrl"
          class="qrcode-image"
          mode="aspectFit"
        />
        <view v-else class="qrcode-placeholder">
          <view v-if="loading" class="loading-spinner"></view>
          <text v-else>二维码加载中...</text>
        </view>
      </view>

      <!-- 提示信息 -->
      <view class="qrcode-info">
        <view class="info-item">
          <text class="label">宝宝:</text>
          <text class="value">{{ babyName }}</text>
        </view>
        <view class="info-item">
          <text class="label">角色:</text>
          <text class="value">{{ roleText }}</text>
        </view>
        <view class="info-item">
          <text class="label">有效期:</text>
          <text class="value">永久有效</text>
        </view>
      </view>

      <!-- 操作提示 -->
      <view class="tips">
        <view class="tip-item">
          <text class="tip-icon">📱</text>
          <text class="tip-text">打开微信扫一扫</text>
        </view>
        <view class="tip-item">
          <text class="tip-icon">📷</text>
          <text class="tip-text">扫描上方二维码</text>
        </view>
        <view class="tip-item">
          <text class="tip-icon">✅</text>
          <text class="tip-text">确认加入协作</text>
        </view>
      </view>
    </view>

    <!-- 保存按钮 -->
    <view class="actions">
      <nut-button type="primary" size="large" @click="saveQRCode">
        保存二维码
      </nut-button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

// 页面参数
const babyId = ref('')
const shortCode = ref('')
const babyName = ref('')
const role = ref('')

// 二维码相关
const qrcodeUrl = ref('')
const loading = ref(false)

// 角色文本映射
const roleTextMap: Record<string, string> = {
  admin: '管理员',
  editor: '编辑者',
  viewer: '查看者',
}

const roleText = computed(() => roleTextMap[role.value] || '编辑者')

// 页面加载
onLoad((options) => {
  console.log('QRCode page onLoad with options:', options)

  // 从邀请页面接收二维码URL
  if (options?.qrcodeUrl) {
    qrcodeUrl.value = decodeURIComponent(options.qrcodeUrl)
    console.log('Received QR code URL:', qrcodeUrl.value)
  }
  if (options?.babyName) {
    babyName.value = decodeURIComponent(options.babyName)
  }
  if (options?.role) {
    role.value = options.role
  }
  if (options?.shortCode) {
    shortCode.value = decodeURIComponent(options.shortCode)
  }

  // 如果没有接收到URL，显示错误
  if (!qrcodeUrl.value) {
    uni.showToast({
      title: '二维码URL缺失,请重新生成',
      icon: 'none',
    })
  }
})

// 保存二维码
function saveQRCode() {
  if (!qrcodeUrl.value) {
    uni.showToast({
      title: '二维码未生成',
      icon: 'none',
    })
    return
  }

  // 下载二维码图片
  uni.downloadFile({
    url: qrcodeUrl.value,
    success: (res) => {
      if (res.statusCode === 200) {
        uni.saveImageToPhotosAlbum({
          filePath: res.tempFilePath,
          success: () => {
            uni.showToast({
              title: '保存成功',
              icon: 'success',
            })
          },
          fail: () => {
            uni.showToast({
              title: '保存失败,请授予相册权限',
              icon: 'none',
            })
          },
        })
      }
    },
    fail: (err) => {
      console.error('Download QR code error:', err)
      uni.showToast({
        title: '下载失败',
        icon: 'none',
      })
    },
  })
}
</script>

<style lang="scss" scoped>
.qrcode-container {
  min-height: 100vh;
  background-color: #f8f8f8;
  padding: 20rpx;
}

.header {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  border-radius: 16rpx;
  padding: 40rpx;
  margin-bottom: 20rpx;
  color: white;

  .title {
    font-size: 40rpx;
    font-weight: bold;
    margin-bottom: 12rpx;
  }

  .subtitle {
    font-size: 28rpx;
    opacity: 0.9;
  }
}

.qrcode-card {
  background: white;
  border-radius: 16rpx;
  padding: 40rpx;
  margin-bottom: 20rpx;
}

.qrcode-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40rpx;

  .qrcode-image {
    width: 560rpx;
    height: 560rpx;
    border-radius: 12rpx;
    box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);
  }

  .qrcode-placeholder {
    width: 560rpx;
    height: 560rpx;
    border-radius: 12rpx;
    background-color: #f5f5f5;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    color: #999;
    font-size: 28rpx;

    .loading-spinner {
      width: 80rpx;
      height: 80rpx;
      border: 6rpx solid #e5e5e5;
      border-top-color: #4facfe;
      border-radius: 50%;
      animation: spin 1s linear infinite;
    }
  }
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.qrcode-info {
  padding: 30rpx 0;
  border-top: 1px solid #f0f0f0;
  border-bottom: 1px solid #f0f0f0;

  .info-item {
    display: flex;
    justify-content: space-between;
    padding: 16rpx 0;
    font-size: 28rpx;

    .label {
      color: #999;
    }

    .value {
      color: #333;
      font-weight: 500;
    }
  }
}

.tips {
  padding-top: 30rpx;

  .tip-item {
    display: flex;
    align-items: center;
    padding: 12rpx 0;
    font-size: 28rpx;
    color: #666;

    .tip-icon {
      font-size: 36rpx;
      margin-right: 12rpx;
    }

    .tip-text {
      flex: 1;
    }
  }
}

.actions {
  padding: 20rpx 0;
}
</style>
