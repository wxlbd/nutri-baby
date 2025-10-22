<template>
  <view class="sleep-page">
    <!-- 当前状态 -->
    <view class="status-card">
      <view v-if="ongoingRecord" class="sleeping">
        <view class="status-icon">💤</view>
        <view class="status-text">宝宝正在睡觉</view>
        <view class="sleep-duration">
          <text class="duration">{{ sleepDuration }}</text>
          <text class="label">已睡眠</text>
        </view>
      </view>
      <view v-else class="awake">
        <view class="status-icon">👀</view>
        <view class="status-text">宝宝醒着</view>
      </view>
    </view>

    <!-- 睡眠类型选择 -->
    <view v-if="!ongoingRecord" class="sleep-type">
      <view class="section-title">睡眠类型</view>
      <nut-radio-group v-model="sleepType" direction="horizontal">
        <nut-radio label="nap">小睡</nut-radio>
        <nut-radio label="night">夜间长睡</nut-radio>
      </nut-radio-group>
    </view>

    <!-- 操作按钮 -->
    <view class="action-buttons">
      <nut-button
        v-if="!ongoingRecord"
        type="primary"
        size="large"
        block
        @click="startSleep"
      >
        <view class="button-content">
          <text class="icon">💤</text>
          <text>开始睡觉</text>
        </view>
      </nut-button>

      <nut-button
        v-else
        type="success"
        size="large"
        block
        @click="endSleep"
      >
        <view class="button-content">
          <text class="icon">🌟</text>
          <text>宝宝醒了</text>
        </view>
      </nut-button>
    </view>

    <!-- 最近记录 -->
    <view v-if="lastRecord && !ongoingRecord" class="last-record">
      <view class="section-title">上次睡眠</view>
      <nut-cell-group>
        <nut-cell
          :title="lastRecord.type === 'nap' ? '小睡' : '夜间长睡'"
          :desc="formatRecordTime(lastRecord)"
        >
          <template #link>
            <text class="duration-text">{{ formatDuration(lastRecord.duration) }}</text>
          </template>
        </nut-cell>
      </nut-cell-group>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { currentBabyId } from '@/store/baby'
import { startSleepRecord, endSleepRecord, getOngoingSleepRecord, getLastSleepRecord } from '@/store/sleep'
import { getUserInfo } from '@/store/user'
import { formatDate, formatDuration } from '@/utils/date'
import { padZero } from '@/utils/common'
import type { SleepRecord } from '@/types'

// 睡眠类型
const sleepType = ref<'nap' | 'night'>('nap')

// 进行中的睡眠记录
const ongoingRecord = ref<SleepRecord | null>(null)

// 最后一次睡眠记录
const lastRecord = ref<SleepRecord | null>(null)

// 睡眠时长(实时)
const sleepDuration = ref('00:00')

// 定时器
let durationTimer: number | null = null

// 更新睡眠时长
const updateDuration = () => {
  if (!ongoingRecord.value) return

  const now = Date.now()
  const duration = Math.floor((now - ongoingRecord.value.startTime) / 1000)
  const hours = Math.floor(duration / 3600)
  const minutes = Math.floor((duration % 3600) / 60)

  sleepDuration.value = `${padZero(hours)}:${padZero(minutes)}`
}

// 页面加载
onMounted(() => {
  if (!currentBaby.value) {
    uni.showToast({
      title: '请先选择宝宝',
      icon: 'none'
    })
    setTimeout(() => {
      uni.navigateBack()
    }, 1500)
    return
  }

  // 检查是否有进行中的睡眠
  ongoingRecord.value = getOngoingSleepRecord(currentBaby.value.babyId)

  if (ongoingRecord.value) {
    // 启动定时器更新时长
    updateDuration()
    durationTimer = setInterval(updateDuration, 1000) as unknown as number
  } else {
    // 获取最后一次睡眠记录
    lastRecord.value = getLastSleepRecord(currentBaby.value.babyId)
  }
})

// 组件卸载
onUnmounted(() => {
  if (durationTimer) {
    clearInterval(durationTimer)
  }
})

// 开始睡觉
const startSleep = () => {
  const user = getUserInfo()
  if (!user) {
    uni.showToast({
      title: '请先登录',
      icon: 'none'
    })
    return
  }

  try {
    const record = startSleepRecord(
      currentBabyId.value,
      sleepType.value,
      user.openid
    )

    ongoingRecord.value = record

    uni.showToast({
      title: '开始记录睡眠',
      icon: 'success'
    })

    // 启动定时器
    updateDuration()
    durationTimer = setInterval(updateDuration, 1000) as unknown as number

  } catch (error: any) {
    uni.showToast({
      title: error.message || '开始失败',
      icon: 'none'
    })
  }
}

// 结束睡觉
const endSleep = () => {
  if (!ongoingRecord.value) return

  const success = endSleepRecord(ongoingRecord.value.id)

  if (success) {
    uni.showToast({
      title: '睡眠记录已保存',
      icon: 'success'
    })

    // 清除定时器
    if (durationTimer) {
      clearInterval(durationTimer)
      durationTimer = null
    }

    setTimeout(() => {
      uni.navigateBack()
    }, 1000)
  }
}

// 格式化记录时间
const formatRecordTime = (record: SleepRecord) => {
  return formatDate(record.startTime, 'MM-DD HH:mm')
}
</script>

<style lang="scss" scoped>
.sleep-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 20rpx;
}

.status-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16rpx;
  padding: 60rpx 30rpx;
  margin-bottom: 20rpx;
  text-align: center;
  color: white;
}

.status-icon {
  font-size: 100rpx;
  margin-bottom: 20rpx;
}

.status-text {
  font-size: 36rpx;
  font-weight: bold;
  margin-bottom: 30rpx;
}

.sleep-duration {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.duration {
  font-size: 64rpx;
  font-weight: bold;
}

.label {
  font-size: 28rpx;
  opacity: 0.9;
}

.sleep-type {
  background: white;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: bold;
  margin-bottom: 24rpx;
}

.action-buttons {
  margin-bottom: 20rpx;
}

.button-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;

  .icon {
    font-size: 36rpx;
  }
}

.last-record {
  background: white;
  border-radius: 16rpx;
  padding: 30rpx;
}

.duration-text {
  color: #fa2c19;
  font-weight: bold;
}
</style>