<template>
  <view class="invite-container">
    <!-- 页面标题 -->
    <view class="header">
      <view class="title">邀请协作者</view>
      <view class="subtitle">邀请家人一起记录{{ babyName }}的成长</view>
    </view>

    <!-- 角色选择 -->
    <view class="section">
      <view class="section-title">协作者角色</view>
      <nut-radio-group v-model="selectedRole" direction="horizontal">
        <nut-radio label="admin">管理员</nut-radio>
        <nut-radio label="editor">编辑者</nut-radio>
        <nut-radio label="viewer">查看者</nut-radio>
      </nut-radio-group>
      <view class="role-desc">
        <text v-if="selectedRole === 'admin'">可管理宝宝信息、邀请/移除协作者</text>
        <text v-else-if="selectedRole === 'editor'">可记录和编辑所有数据</text>
        <text v-else>仅可查看数据,不能编辑</text>
      </view>
    </view>

    <!-- 访问权限 -->
    <view class="section">
      <view class="section-title">访问权限</view>
      <nut-radio-group v-model="accessType" direction="horizontal">
        <nut-radio label="permanent">永久</nut-radio>
        <nut-radio label="temporary">临时</nut-radio>
      </nut-radio-group>

      <!-- 临时权限时显示过期时间选择框 -->
      <view v-if="accessType === 'temporary'" class="expire-time">
        <view class="time-selector" @click="showDatetimePickerModal = true">
          <text class="time-label">过期时间</text>
          <text class="time-value">{{ formatDateTime(expiresDate) }}</text>
          <view class="time-icon">
            <text>›</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 生成邀请按钮 -->
    <view class="generate-section">
      <nut-button
        type="primary"
        size="large"
        @click="handleGenerateQRCode"
        :loading="generating"
      >
        {{ generating ? '生成中...' : '生成邀请二维码' }}
      </nut-button>
    </view>

    <!-- 二维码展示区域（生成后显示） -->
    <view v-if="qrcodeUrl" class="qrcode-card">
      <!-- 二维码显示区域 -->
      <view class="qrcode-wrapper">
        <image
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
          <text class="value">{{ validityText }}</text>
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

      <!-- 保存按钮 -->
      <view class="actions">
        <nut-button type="success" size="large" @click="saveQRCode">
          保存二维码到相册
        </nut-button>
      </view>
    </view>

    <!-- 日期时间选择器弹窗 -->
    <nut-popup
      :visible="showDatetimePickerModal"
      position="bottom"
      round
      @update:visible="showDatetimePickerModal = $event"
    >
      <nut-date-picker
        v-model="expiresDate"
        type="datetime"
        title="选择过期时间"
        :min-date="minDate"
        :max-date="maxDate"
        @confirm="onDateTimeConfirm"
        @cancel="onDateTimeCancel"
      ></nut-date-picker>
    </nut-popup>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { inviteCollaborator } from '@/store/collaborator'

// 页面参数
const babyId = ref('')
const babyName = ref('')

// 表单数据
const selectedRole = ref<'admin' | 'editor' | 'viewer'>('editor')
const accessType = ref<'permanent' | 'temporary'>('permanent')
const expiresDate = ref<Date>(new Date(Date.now() + 7 * 24 * 60 * 60 * 1000)) // 默认7天后
const showDatetimePickerModal = ref(false)

// 二维码相关
const qrcodeUrl = ref('')
const generating = ref(false)

// 日期选择器范围
const minDate = new Date() // 最小日期为今天
const maxDate = new Date(Date.now() + 365 * 24 * 60 * 60 * 1000) // 最大1年后

// 角色文本映射
const roleTextMap: Record<string, string> = {
  admin: '管理员',
  editor: '编辑者',
  viewer: '查看者',
}

const roleText = computed(() => roleTextMap[selectedRole.value] || '编辑者')

// 有效期文本
const validityText = computed(() => {
  if (accessType.value === 'permanent') {
    return '永久有效'
  }
  return formatDateTime(expiresDate.value)
})

// 页面加载
onLoad((options) => {
  if (options?.babyId) {
    babyId.value = options.babyId
  }
  if (options?.babyName) {
    babyName.value = decodeURIComponent(options.babyName)
  }
})

// 格式化日期时间
function formatDateTime(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hour = String(date.getHours()).padStart(2, '0')
  const minute = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hour}:${minute}`
}

// 日期时间选择确认
function onDateTimeConfirm() {
  showDatetimePickerModal.value = false
}

// 日期时间选择取消
function onDateTimeCancel() {
  showDatetimePickerModal.value = false
}

// 生成二维码
async function handleGenerateQRCode() {
  if (!babyId.value) {
    uni.showToast({
      title: '宝宝ID不能为空',
      icon: 'none',
    })
    return
  }

  generating.value = true

  try {
    // 计算过期时间戳
    const expiresAt = accessType.value === 'temporary'
      ? expiresDate.value.getTime()
      : undefined

    // 调用API生成邀请（二维码方式）
    const invitationData = await inviteCollaborator(
      babyId.value,
      'qrcode',
      selectedRole.value,
      accessType.value,
      expiresAt
    )

    const { qrcodeParams } = invitationData

    if (!qrcodeParams || !qrcodeParams.qrcodeUrl) {
      uni.showToast({
        title: '二维码生成失败',
        icon: 'none',
      })
      return
    }

    // 显示二维码
    qrcodeUrl.value = qrcodeParams.qrcodeUrl

    uni.showToast({
      title: '二维码生成成功',
      icon: 'success',
    })
  } catch (error: any) {
    console.error('Generate QR code error:', error)
    uni.showToast({
      title: error.message || '生成失败',
      icon: 'none',
    })
  } finally {
    generating.value = false
  }
}

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
.invite-container {
  min-height: 100vh;
  background-color: #f8f8f8;
  padding: 20rpx;
  padding-bottom: 40rpx;
}

.header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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

.section {
  background: white;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;

  .section-title {
    font-size: 32rpx;
    font-weight: bold;
    margin-bottom: 24rpx;
    color: #333;
  }

  .role-desc {
    margin-top: 16rpx;
    font-size: 28rpx;
    color: #999;
  }

  .expire-time {
    margin-top: 20rpx;
  }
}

// 过期时间选择框
.time-selector {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24rpx 28rpx;
  background: #f7f8fa;
  border-radius: 12rpx;
  border: 2rpx solid #e5e5e5;
  transition: all 0.2s;

  &:active {
    background: #f0f1f3;
    border-color: #667eea;
  }

  .time-label {
    font-size: 28rpx;
    color: #666;
  }

  .time-value {
    flex: 1;
    text-align: right;
    font-size: 28rpx;
    color: #667eea;
    font-weight: 500;
    margin: 0 16rpx;
  }

  .time-icon {
    font-size: 32rpx;
    color: #999;
    line-height: 1;
  }
}

// 生成按钮区域
.generate-section {
  margin-bottom: 20rpx;
}

// 二维码卡片
.qrcode-card {
  background: white;
  border-radius: 16rpx;
  padding: 40rpx;
  margin-bottom: 20rpx;
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20rpx);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
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
  padding-top: 20rpx;
}
</style>
