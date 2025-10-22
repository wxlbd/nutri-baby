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
        <canvas
          v-if="!qrcodeUrl"
          canvas-id="qrcode"
          class="qrcode-canvas"
          :style="{ width: qrcodeSize + 'px', height: qrcodeSize + 'px' }"
        />
        <image
          v-else
          :src="qrcodeUrl"
          class="qrcode-image"
          mode="aspectFit"
        />
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
          <text class="value">7天</text>
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
import { ref, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

// 页面参数
const scene = ref('')
const page = ref('')
const babyName = ref('')
const role = ref('')

// 二维码相关
const qrcodeUrl = ref('')
const qrcodeSize = ref(280) // 二维码尺寸

// 角色文本映射
const roleTextMap: Record<string, string> = {
  admin: '管理员',
  editor: '编辑者',
  viewer: '查看者',
}

const roleText = computed(() => roleTextMap[role.value] || '编辑者')

// 页面加载
onLoad((options) => {
  if (options?.scene) {
    scene.value = decodeURIComponent(options.scene)
  }
  if (options?.page) {
    page.value = decodeURIComponent(options.page)
  }
  if (options?.babyName) {
    babyName.value = decodeURIComponent(options.babyName)
  }
  if (options?.role) {
    role.value = options.role
  }
})

// 组件挂载后生成二维码
onMounted(() => {
  generateQRCode()
})

// 生成二维码
async function generateQRCode() {
  // 方式一: 前端生成二维码(使用微信小程序API)
  // @ts-ignore
  if (typeof wx !== 'undefined' && wx.canIUse) {
    try {
      // 调用微信小程序二维码生成API
      // 注意: 这需要后端获取access_token,所以这里我们使用canvas绘制简单的提示
      // 实际项目中应该由后端生成二维码图片URL并返回

      // 暂时显示提示信息
      uni.showModal({
        title: '提示',
        content: '二维码生成功能需要后端支持,请联系管理员配置',
        showCancel: false,
      })

      // TODO: 实际项目中应该调用后端API获取二维码图片URL
      // const response = await get(`/babies/qrcode?scene=${scene.value}&page=${page.value}`)
      // qrcodeUrl.value = response.data.qrcodeUrl
    } catch (error) {
      console.error('generate qrcode error:', error)
      uni.showToast({
        title: '二维码生成失败',
        icon: 'none',
      })
    }
  }

  // 方式二: 使用第三方二维码库在canvas上绘制
  // 这里提供一个简化的示例,实际项目建议使用 uQRCode 等成熟库
  drawQRCodePlaceholder()
}

// 绘制二维码占位符(实际项目中应替换为真实二维码生成库)
function drawQRCodePlaceholder() {
  const ctx = uni.createCanvasContext('qrcode')

  // 绘制白色背景
  ctx.setFillStyle('#ffffff')
  ctx.fillRect(0, 0, qrcodeSize.value, qrcodeSize.value)

  // 绘制边框
  ctx.setStrokeStyle('#000000')
  ctx.setLineWidth(2)
  ctx.strokeRect(10, 10, qrcodeSize.value - 20, qrcodeSize.value - 20)

  // 绘制提示文字
  ctx.setFillStyle('#333333')
  ctx.setFontSize(14)
  ctx.setTextAlign('center')
  ctx.fillText('请扫描此二维码', qrcodeSize.value / 2, qrcodeSize.value / 2 - 10)
  ctx.fillText('加入宝宝协作', qrcodeSize.value / 2, qrcodeSize.value / 2 + 10)

  ctx.draw()
}

// 保存二维码
function saveQRCode() {
  if (qrcodeUrl.value) {
    // 如果有二维码图片URL,直接保存
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
                title: '保存失败',
                icon: 'none',
              })
            },
          })
        }
      },
    })
  } else {
    // 如果是canvas绘制的二维码,需要先转为图片
    uni.canvasToTempFilePath({
      canvasId: 'qrcode',
      success: (res) => {
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
      },
      fail: (err) => {
        console.error('canvas to image error:', err)
        uni.showToast({
          title: '保存失败',
          icon: 'none',
        })
      },
    })
  }
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

  .qrcode-canvas,
  .qrcode-image {
    width: 560rpx;
    height: 560rpx;
    border-radius: 12rpx;
    box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);
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
