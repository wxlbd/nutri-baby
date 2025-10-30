<template>
  <view class="diaper-page">
    <!-- 排泄类型快捷按钮 -->
    <view class="quick-buttons">
      <view class="button-row">
        <nut-button
          type="primary"
          size="large"
          class="type-button"
          @click="quickRecord('pee')"
        >
          <view class="button-content">
            <text class="icon">💧</text>
            <text>小便</text>
          </view>
        </nut-button>

        <nut-button
          type="warning"
          size="large"
          class="type-button"
          @click="quickRecord('poop')"
        >
          <view class="button-content">
            <text class="icon">💩</text>
            <text>大便</text>
          </view>
        </nut-button>
      </view>

      <nut-button
        type="success"
        size="large"
        block
        @click="quickRecord('both')"
      >
        <view class="button-content">
          <text class="icon">💧💩</text>
          <text>小便+大便</text>
        </view>
      </nut-button>
    </view>

    <!-- 大便详情 -->
    <view v-if="showDetails" class="details-section">
      <view class="section-title">大便详情(可选)</view>

      <!-- 记录时间选择 -->
      <nut-cell-group>
        <nut-cell title="记录时间">
          <template #link>
            <view class="time-display" @click="showDatePicker">
              <text>{{ formatRecordTime(recordDateTime) }}</text>
              <nut-icon name="right"></nut-icon>
            </view>
          </template>
        </nut-cell>
      </nut-cell-group>

      <nut-cell-group>
        <!-- 大便颜色 -->
        <nut-cell title="颜色">
          <view class="color-selector">
            <view
              v-for="color in poopColors"
              :key="color.value"
              class="color-item"
              :class="{ active: form.poopColor === color.value }"
              @click="form.poopColor = color.value"
            >
              <view class="color-circle" :style="{ background: color.color }"></view>
              <text class="color-label">{{ color.label }}</text>
            </view>
          </view>
        </nut-cell>

        <!-- 大便性状 -->
        <nut-cell title="性状">
          <nut-radio-group v-model="form.poopTexture">
            <view class="texture-list">
              <nut-radio
                v-for="texture in poopTextures"
                :key="texture.value"
                :label="texture.value"
              >
                {{ texture.label }}
              </nut-radio>
            </view>
          </nut-radio-group>
        </nut-cell>

        <!-- 备注 -->
        <nut-cell title="备注">
          <nut-textarea
            v-model="form.note"
            placeholder="有什么需要记录的吗?"
            :max-length="200"
            :rows="3"
          />
        </nut-cell>
      </nut-cell-group>

      <!-- 提交按钮 -->
      <view class="submit-button">
        <nut-button
          type="primary"
          size="large"
          block
          @click="handleSubmit"
        >
          保存记录
        </nut-button>
      </view>
    </view>

    <!-- 日期选择器 -->
    <nut-date-picker
        v-model="recordDateTime"
        type="datetime"
        :min-date="minDateTime"
        :max-date="maxDateTime"
        @confirm="onDateTimeConfirm"
        @cancel="onDateTimeCancel"
        :visible="showDatetimePickerModal"
    ></nut-date-picker>

    <!-- 快速记录时间选择弹窗 -->
    <nut-dialog v-model:visible="showQuickTimePickerModal" title="选择记录时间" @confirm="onQuickTimeConfirm" @cancel="onQuickTimeCancel">
      <view class="quick-time-picker">
        <nut-date-picker
            v-model="quickRecordDateTime"
            type="datetime"
            :min-date="minDateTime"
            :max-date="maxDateTime"
        ></nut-date-picker>
      </view>
    </nut-dialog>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { currentBabyId, getCurrentBaby } from '@/store/baby'
import { getUserInfo } from '@/store/user'
import type { DiaperType, PoopColor, PoopTexture } from '@/types'

// 直接调用 API 层
import * as diaperApi from '@/api/diaper'

// 表单数据
const form = ref({
  type: 'pee' as DiaperType,
  poopColor: undefined as PoopColor | undefined,
  poopTexture: undefined as PoopTexture | undefined,
  note: '',
})

// 是否显示详情
const showDetails = ref(false)

// 日期时间选择器
const recordDateTime = ref<Date>(new Date()) // 记录时间,初始为当前时间
const showDatetimePickerModal = ref(false)
const quickRecordDateTime = ref<Date>(new Date())
const showQuickTimePickerModal = ref(false)
const minDateTime = ref<Date>(new Date(Date.now() - 30 * 24 * 60 * 60 * 1000)) // 最小: 30天前
const maxDateTime = ref<Date>(new Date()) // 最大: 当前时间

// 显示日期时间选择器
const showDatePicker = () => {
    showDatetimePickerModal.value = true
}

// 确认日期时间选择
const onDateTimeConfirm = (value: Date) => {
    recordDateTime.value = value
    showDatetimePickerModal.value = false
    console.log('[Diaper] 记录时间已更改为:', value)
}

// 取消日期时间选择
const onDateTimeCancel = () => {
    showDatetimePickerModal.value = false
}

// 快速记录时间确认
const onQuickTimeConfirm = () => {
    showQuickTimePickerModal.value = false
    saveRecord(quickRecordDateTime.value.getTime())
}

// 快速记录时间取消
const onQuickTimeCancel = () => {
    showQuickTimePickerModal.value = false
}

// 格式化记录时间显示
const formatRecordTime = (date: Date): string => {
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    return `${year}-${month}-${day} ${hours}:${minutes}`
}

// 大便颜色选项
const poopColors = [
  { value: 'yellow', label: '黄色', color: '#FFD700' },
  { value: 'green', label: '绿色', color: '#90EE90' },
  { value: 'brown', label: '棕色', color: '#8B4513' },
  { value: 'black', label: '黑色', color: '#000000' },
  { value: 'red', label: '红色', color: '#FF6347' },
  { value: 'white', label: '白色', color: '#F0F0F0' },
] as const

// 大便性状选项
const poopTextures = [
  { value: 'watery', label: '稀水状' },
  { value: 'loose', label: '稀软' },
  { value: 'paste', label: '糊状' },
  { value: 'soft', label: '软便' },
  { value: 'formed', label: '成形' },
  { value: 'hard', label: '硬结' },
] as const

// 快速记录
const quickRecord = (type: DiaperType) => {
  const currentBaby = getCurrentBaby()
  if (!currentBaby) {
    uni.showToast({
      title: '请先选择宝宝',
      icon: 'none'
    })
    return
  }

  form.value.type = type

  // 如果包含大便,显示详情填写
  if (type === 'poop' || type === 'both') {
    // 重置时间为当前时间
    recordDateTime.value = new Date()
    showDetails.value = true
    return
  }

  // 小便显示时间选择弹窗
  quickRecordDateTime.value = new Date()
  showQuickTimePickerModal.value = true
}

// 保存记录
const saveRecord = async (changeTime?: number) => {
  const user = getUserInfo()
  if (!user) {
    uni.showToast({
      title: '请先登录',
      icon: 'none'
    })
    return
  }

  try {
    // 使用传入的时间或当前表单中的时间
    const finalChangeTime = changeTime ?? recordDateTime.value.getTime()

    // 直接调用 API 层创建记录
    await diaperApi.apiCreateDiaperRecord({
      babyId: currentBabyId.value,
      diaperType: form.value.type,
      pooColor: form.value.poopColor,
      pooTexture: form.value.poopTexture,
      note: form.value.note || undefined,
      changeTime: finalChangeTime
    })

    uni.showToast({
      title: '保存成功',
      icon: 'success'
    })

    setTimeout(() => {
      uni.navigateBack()
    }, 1000)
  } catch (error: any) {
    console.error('[Diaper] 保存换尿布记录失败:', error)
    uni.showToast({
      title: error.message || '保存失败',
      icon: 'none'
    })
  }
}

// 提交记录
const handleSubmit = () => {
  saveRecord()
}
</script>

<style lang="scss" scoped>
.diaper-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 20rpx;
}

.quick-buttons {
  background: white;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.button-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20rpx;
  margin-bottom: 20rpx;
}

.type-button {
  flex: 1;
}

.button-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
  padding: 8rpx 0;

  .icon {
    font-size: 32rpx;
  }
}

.details-section {
  background: white;
  border-radius: 16rpx;
  padding: 30rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: bold;
  margin-bottom: 24rpx;
}

.time-display {
  display: flex;
  align-items: center;
  gap: 10rpx;
  color: #333;
  font-size: 28rpx;
}

.color-selector {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20rpx;
  padding: 20rpx 0;
}

.color-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12rpx;
  padding: 16rpx;
  border-radius: 12rpx;
  border: 2rpx solid transparent;
  transition: all 0.3s;

  &.active {
    border-color: #fa2c19;
    background: rgba(250, 44, 25, 0.05);
  }
}

.color-circle {
  width: 60rpx;
  height: 60rpx;
  border-radius: 50%;
  border: 2rpx solid #ddd;
}

.color-label {
  font-size: 24rpx;
  color: #666;
}

.texture-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  padding: 20rpx 0;
}

.submit-button {
  margin-top: 40rpx;
}

.quick-time-picker {
  height: 400rpx;
}
</style>