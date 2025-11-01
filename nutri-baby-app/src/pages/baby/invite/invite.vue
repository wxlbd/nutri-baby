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

      <!-- 临时权限时显示过期时间选择 -->
      <view v-if="accessType === 'temporary'" class="expire-time">
        <nut-cell title="过期时间" :desc="expiresAtText" @click="showDatePicker = true" />
      </view>
    </view>

    <!-- 邀请方式选择 -->
    <view class="section">
      <view class="section-title">邀请方式</view>
      <view class="invite-methods">
        <view class="method-card" @click="handleInvite('share')">
          <view class="method-icon">📱</view>
          <view class="method-title">微信分享</view>
          <view class="method-desc">分享给微信好友或群</view>
        </view>
        <view class="method-card" @click="handleInvite('qrcode')">
          <view class="method-icon">📷</view>
          <view class="method-title">面对面扫码</view>
          <view class="method-desc">生成二维码供扫描</view>
        </view>
      </view>
    </view>

    <!-- 日期时间选择器 -->
    <nut-date-picker
      v-model:visible="showDatePicker"
      v-model="expiresDate"
      type="datetime"
      title="选择过期时间"
      :min-date="minDate"
      :max-date="maxDate"
      @confirm="onDateConfirm"
    />
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { inviteCollaborator } from '@/store/collaborator'

// 页面参数
const babyId = ref('')
const babyName = ref('')

// 表单数据
const selectedRole = ref<'admin' | 'editor' | 'viewer'>('editor')
const accessType = ref<'permanent' | 'temporary'>('permanent')
const expiresDate = ref<Date>(new Date(Date.now() + 7 * 24 * 60 * 60 * 1000)) // 默认7天后
const showDatePicker = ref(false)

// 日期选择器范围
const minDate = new Date() // 最小日期为今天
const maxDate = new Date(Date.now() + 365 * 24 * 60 * 60 * 1000) // 最大1年后

// 计算过期时间文本
const expiresAtText = computed(() => {
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

// 日期选择确认
function onDateConfirm() {
  showDatePicker.value = false
}

// 处理邀请
async function handleInvite(inviteType: 'share' | 'qrcode') {
  if (!babyId.value) {
    uni.showToast({
      title: '宝宝ID不能为空',
      icon: 'none',
    })
    return
  }

  uni.showLoading({
    title: '生成邀请中...',
  })

  try {
    // 计算过期时间戳
    const expiresAt = accessType.value === 'temporary'
      ? expiresDate.value.getTime()
      : undefined

    // 调用API生成邀请
    const invitationData = await inviteCollaborator(
      babyId.value,
      inviteType,
      selectedRole.value,
      accessType.value,
      expiresAt
    )

    uni.hideLoading()

    // 根据邀请类型跳转到不同页面
    if (inviteType === 'share') {
      // 微信分享
      handleWechatShare(invitationData)
    } else {
      // 二维码
      handleQRCode(invitationData)
    }
  } catch (error: any) {
    uni.hideLoading()
    console.error('invite error:', error)
  }
}

// 处理微信分享
function handleWechatShare(invitationData: any) {
  const { shareParams } = invitationData

  if (!shareParams) {
    uni.showToast({
      title: '分享参数缺失',
      icon: 'none',
    })
    return
  }

  // 使用微信分享API
  // @ts-ignore
  if (typeof wx !== 'undefined') {
    // @ts-ignore
    wx.shareAppMessage({
      title: shareParams.title,
      path: shareParams.path,
      imageUrl: shareParams.imageUrl,
      success: () => {
        uni.showToast({
          title: '分享成功',
          icon: 'success',
        })
        // 返回上一页
        setTimeout(() => {
          uni.navigateBack()
        }, 1500)
      },
      fail: (err: any) => {
        console.error('share error:', err)
        uni.showToast({
          title: '分享失败',
          icon: 'none',
        })
      },
    })
  } else {
    // 非微信环境,显示提示
    uni.showModal({
      title: '提示',
      content: '请在微信小程序中使用分享功能',
      showCancel: false,
    })
  }
}

// 处理二维码
function handleQRCode(invitationData: any) {
  const { qrcodeParams, babyId: returnedBabyId, shortCode } = invitationData

  if (!qrcodeParams || !qrcodeParams.qrcodeUrl) {
    uni.showToast({
      title: '二维码生成失败',
      icon: 'none',
    })
    return
  }

  // 跳转到二维码显示页面，传递完整的二维码URL
  uni.navigateTo({
    url: `/pages/baby/qrcode/qrcode?qrcodeUrl=${encodeURIComponent(qrcodeParams.qrcodeUrl)}&babyName=${encodeURIComponent(babyName.value)}&role=${selectedRole.value}&shortCode=${encodeURIComponent(shortCode || '')}`,
  })
}
</script>

<style lang="scss" scoped>
.invite-container {
  min-height: 100vh;
  background-color: #f8f8f8;
  padding: 20rpx;
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

.invite-methods {
  display: flex;
  gap: 20rpx;

  .method-card {
    flex: 1;
    background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
    border-radius: 16rpx;
    padding: 40rpx 20rpx;
    text-align: center;
    color: white;
    box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);
    transition: transform 0.2s;

    &:active {
      transform: scale(0.98);
    }

    .method-icon {
      font-size: 64rpx;
      margin-bottom: 12rpx;
    }

    .method-title {
      font-size: 32rpx;
      font-weight: bold;
      margin-bottom: 8rpx;
    }

    .method-desc {
      font-size: 24rpx;
      opacity: 0.9;
    }
  }

  .method-card:last-child {
    background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  }
}
</style>
