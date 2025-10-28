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
import { currentBabyId, currentBaby } from '@/store/baby'
import { getUserInfo } from '@/store/user'
import { formatDate, formatDuration } from '@/utils/date'
import { padZero } from '@/utils/common'
import type { SleepRecord } from '@/types'

// 直接调用 API 层
import * as sleepApi from '@/api/sleep'

// ⚠️ 注意: 睡眠计时器功能需要本地状态管理,暂时简化为手动输入时间
// TODO: 后续可以考虑使用 localStorage 或独立的计时器状态管理

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

  // TODO: 加载进行中的睡眠记录(需要从后端或 localStorage 获取)
  // ongoingRecord.value = ...
})

// 组件卸载
onUnmounted(() => {
  if (durationTimer) {
    clearInterval(durationTimer)
  }
})

// 开始睡觉
const startSleep = async () => {
  const user = getUserInfo()
  if (!user) {
    uni.showToast({
      title: '请先登录',
      icon: 'none'
    })
    return
  }

  try {
    // 直接创建睡眠记录(开始时间)
    const record = await sleepApi.apiCreateSleepRecord({
      babyId: currentBabyId.value,
      sleepType: sleepType.value,
      startTime: Date.now()
      // endTime 在结束时更新
    })

    uni.showToast({
      title: '开始记录睡眠',
      icon: 'success'
    })

    // TODO: 保存进行中的记录到 localStorage
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
const endSleep = async () => {
  if (!ongoingRecord.value) return

  try {
    // 更新睡眠记录(结束时间)
    await sleepApi.apiUpdateSleepRecord(ongoingRecord.value.id, {
      endTime: Date.now()
    })

    uni.showToast({
      title: '保存成功',
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

  } catch (error: any) {
    uni.showToast({
      title: error.message || '保存失败',
      icon: 'none'
    })
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