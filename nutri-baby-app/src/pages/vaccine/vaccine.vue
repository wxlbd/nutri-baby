<template>
  <view class="vaccine-page">
    <!-- 疫苗完成度 -->
    <view v-if="currentBaby" class="progress-card">
      <view class="card-header">
        <text class="header-icon">💉</text>
        <text class="header-title">疫苗接种进度</text>
      </view>

      <view class="progress-info">
        <view class="progress-bar-container">
          <view class="progress-bar">
            <view
              class="progress-fill"
              :style="{ width: completionStats.percentage + '%' }"
            ></view>
          </view>
          <text class="progress-text">
            {{ completionStats.completed }} / {{ completionStats.total }}
            ({{ completionStats.percentage }}%)
          </text>
        </view>
      </view>
    </view>

    <!-- 即将到期提醒 -->
    <view v-if="upcomingReminders.length > 0" class="reminders-section">
      <view class="section-title">⏰ 近期待接种 ({{ upcomingReminders.length }})</view>

      <view class="reminder-list">
        <view
          v-for="reminder in upcomingReminders"
          :key="reminder.id"
          class="reminder-item"
          :class="`status-${reminder.status}`"
          @click="handleRecordVaccine(reminder)"
        >
          <view class="reminder-content">
            <view class="vaccine-name">
              {{ reminder.vaccineName }} (第{{ reminder.doseNumber }}针)
            </view>
            <view class="vaccine-date">
              预定时间: {{ formatDate(reminder.scheduledDate, 'YYYY-MM-DD') }}
            </view>
            <view class="vaccine-status">
              <text v-if="reminder.status === 'due'" class="status-badge due">即将到期</text>
              <text v-if="reminder.status === 'overdue'" class="status-badge overdue">已逾期</text>
            </view>
          </view>
          <view class="reminder-action">
            <nut-button size="small" type="primary">记录接种</nut-button>
          </view>
        </view>
      </view>
    </view>

    <!-- 疫苗计划列表 -->
    <view class="plan-section">
      <view class="section-header">
        <text class="section-title">📋 疫苗计划</text>
        <nut-button size="small" @click="goToManage">
          管理计划
        </nut-button>
      </view>

      <nut-tabs v-model="activeTab">
        <nut-tab-pane title="全部" pane-key="all" />
        <nut-tab-pane title="已完成" pane-key="completed" />
        <nut-tab-pane title="未完成" pane-key="pending" />
      </nut-tabs>

      <view class="plan-list">
        <view
          v-for="plan in filteredPlans"
          :key="plan.id"
          class="plan-item"
          :class="{ completed: isPlanCompleted(plan.id) }"
        >
          <view class="plan-header">
            <view class="plan-name">
              <text class="required-badge" v-if="plan.isRequired">必打</text>
              {{ plan.vaccineName }}
            </view>
            <text class="plan-age">{{ plan.ageInMonths }}个月</text>
          </view>

          <view class="plan-detail">
            <text class="plan-dose">第{{ plan.doseNumber }}针</text>
            <text v-if="plan.description" class="plan-desc">{{ plan.description }}</text>
          </view>

          <view v-if="isPlanCompleted(plan.id)" class="plan-record">
            <text class="completed-icon">✓</text>
            <text class="completed-text">已接种</text>
            <text class="completed-date">
              {{ getRecordDate(plan.id) }}
            </text>
          </view>

          <view v-else class="plan-action">
            <nut-button
              size="small"
              type="primary"
              @click="handleRecordByPlan(plan)"
            >
              记录接种
            </nut-button>
          </view>
        </view>
      </view>
    </view>

    <!-- 接种记录对话框 -->
    <nut-popup
      v-model:visible="showRecordDialog"
      position="bottom"
      :style="{ height: '70%' }"
      round
      closeable
    >
      <view class="dialog-content">
        <view class="dialog-title">记录疫苗接种</view>

        <view class="form-section">
          <view class="form-item">
            <view class="form-label">疫苗名称</view>
            <nut-input
              v-model="recordForm.vaccineName"
              placeholder="疫苗名称"
              readonly
            />
          </view>

          <view class="form-item">
            <view class="form-label">接种日期</view>
            <nut-input
              :model-value="formatDate(recordForm.vaccineDate, 'YYYY-MM-DD')"
              readonly
              @click="showDatePicker = true"
            />
          </view>

          <view class="form-item">
            <view class="form-label">接种医院</view>
            <nut-input
              v-model="recordForm.hospital"
              placeholder="请输入医院名称"
              clearable
            />
          </view>

          <view class="form-item">
            <view class="form-label">疫苗批号</view>
            <nut-input
              v-model="recordForm.batchNumber"
              placeholder="请输入疫苗批号(可选)"
              clearable
            />
          </view>

          <view class="form-item">
            <view class="form-label">不良反应</view>
            <nut-textarea
              v-model="recordForm.reaction"
              placeholder="如有不良反应请记录(可选)"
              :max-length="200"
            />
          </view>

          <view class="form-item">
            <view class="form-label">备注</view>
            <nut-textarea
              v-model="recordForm.note"
              placeholder="其他备注信息(可选)"
              :max-length="200"
            />
          </view>
        </view>

        <view class="dialog-footer">
          <nut-button
            type="default"
            size="large"
            block
            @click="showRecordDialog = false"
          >
            取消
          </nut-button>
          <nut-button
            type="primary"
            size="large"
            block
            @click="handleSaveRecord"
          >
            保存
          </nut-button>
        </view>
      </view>
    </nut-popup>

    <!-- 日期选择器 -->
    <nut-date-picker
      v-model:visible="showDatePicker"
      v-model="selectedDate"
      type="date"
      title="选择接种日期"
      @confirm="handleDateConfirm"
    />
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { currentBaby, currentBabyId } from '@/store/baby'
import { userInfo } from '@/store/user'
import {
  vaccinePlans,
  vaccineRecords,
  initializeVaccinePlansFromServer,
  generateRemindersForBaby,
  getVaccineRemindersByBabyId,
  getUpcomingReminders,
  getVaccineCompletionStats,
  addVaccineRecord,
  getVaccinePlanById
} from '@/store/vaccine'
import { formatDate } from '@/utils/date'
import type { VaccinePlan, VaccineReminder } from '@/types'

// Tab状态
const activeTab = ref('all')

// 对话框状态
const showRecordDialog = ref(false)
const showDatePicker = ref(false)
const selectedDate = ref(new Date())

// 表单数据
const recordForm = ref({
  planId: '',
  vaccineType: '',
  vaccineName: '',
  doseNumber: 1,
  vaccineDate: Date.now(),
  hospital: '',
  batchNumber: '',
  reaction: '',
  note: ''
})

// 完成度统计
const completionStats = computed(() => {
  if (!currentBaby.value) return { total: 0, completed: 0, percentage: 0 }
  return getVaccineCompletionStats(currentBaby.value.babyId)
})

// 即将到期的提醒
const upcomingReminders = computed(() => {
  if (!currentBaby.value) return []
  return getUpcomingReminders(currentBaby.value.babyId).slice(0, 3) // 只显示前3个
})

// 过滤后的计划列表
const filteredPlans = computed(() => {
  let plans = vaccinePlans.value

  if (activeTab.value === 'completed') {
    plans = plans.filter(plan => isPlanCompleted(plan.id))
  } else if (activeTab.value === 'pending') {
    plans = plans.filter(plan => !isPlanCompleted(plan.id))
  }

  return plans.sort((a, b) => a.ageInMonths - b.ageInMonths)
})

// 判断计划是否已完成
const isPlanCompleted = (planId: string): boolean => {
  if (!currentBabyId.value) return false
  return vaccineRecords.value.some(
    record => record.babyId === currentBabyId.value && record.planId === planId
  )
}

// 获取接种记录日期
const getRecordDate = (planId: string): string => {
  if (!currentBabyId.value) return ''
  const record = vaccineRecords.value.find(
    r => r.babyId === currentBabyId.value && r.planId === planId
  )
  return record ? formatDate(record.vaccineDate, 'YYYY-MM-DD') : ''
}

// 处理记录接种(通过提醒)
const handleRecordVaccine = (reminder: VaccineReminder) => {
  const plan = getVaccinePlanById(reminder.planId)
  if (!plan) return

  recordForm.value = {
    planId: plan.id,
    vaccineType: plan.vaccineType,
    vaccineName: plan.vaccineName,
    doseNumber: plan.doseNumber,
    vaccineDate: Date.now(),
    hospital: '',
    batchNumber: '',
    reaction: '',
    note: ''
  }

  showRecordDialog.value = true
}

// 处理记录接种(通过计划)
const handleRecordByPlan = (plan: VaccinePlan) => {
  recordForm.value = {
    planId: plan.id,
    vaccineType: plan.vaccineType,
    vaccineName: plan.vaccineName,
    doseNumber: plan.doseNumber,
    vaccineDate: Date.now(),
    hospital: '',
    batchNumber: '',
    reaction: '',
    note: ''
  }

  showRecordDialog.value = true
}

// 日期选择确认
const handleDateConfirm = ({ selectedValue }: any) => {
  const date = new Date(selectedValue.join('-'))
  recordForm.value.vaccineDate = date.getTime()
  showDatePicker.value = false
}

// 保存接种记录
const handleSaveRecord = () => {
  if (!currentBaby.value || !userInfo.value) {
    uni.showToast({
      title: '请先登录',
      icon: 'none'
    })
    return
  }

  if (!recordForm.value.hospital.trim()) {
    uni.showToast({
      title: '请输入接种医院',
      icon: 'none'
    })
    return
  }

  addVaccineRecord({
    babyId: currentBaby.value.babyId,
    planId: recordForm.value.planId,
    vaccineType: recordForm.value.vaccineType,
    vaccineName: recordForm.value.vaccineName,
    doseNumber: recordForm.value.doseNumber,
    vaccineDate: recordForm.value.vaccineDate,
    hospital: recordForm.value.hospital.trim(),
    batchNumber: recordForm.value.batchNumber.trim() || undefined,
    reaction: recordForm.value.reaction.trim() || undefined,
    note: recordForm.value.note.trim() || undefined,
    createBy: userInfo.value.openid
  })

  uni.showToast({
    title: '记录成功',
    icon: 'success'
  })

  showRecordDialog.value = false
}

// 页面加载
onMounted(async () => {
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

  // 为当前宝宝初始化疫苗计划（使用服务器API）
  await initializeVaccinePlansFromServer(currentBaby.value.babyId)

  // 生成提醒
  generateRemindersForBaby(currentBaby.value.babyId, currentBaby.value.birthDate)
})

// 跳转到疫苗计划管理页面
const goToManage = () => {
  uni.navigateTo({
    url: '/pages/vaccine/manage/manage'
  })
}
</script>

<style lang="scss" scoped>
.vaccine-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 20rpx;
  padding-bottom: 40rpx;
}

.progress-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
  color: white;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 20rpx;
}

.header-icon {
  font-size: 40rpx;
}

.header-title {
  font-size: 32rpx;
  font-weight: bold;
}

.progress-bar-container {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.progress-bar {
  height: 16rpx;
  background: rgba(255, 255, 255, 0.3);
  border-radius: 8rpx;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: white;
  transition: width 0.3s;
}

.progress-text {
  font-size: 28rpx;
  text-align: right;
}

.reminders-section, .plan-section {
  background: white;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: bold;
}

.reminder-list, .plan-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.reminder-item {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 20rpx;
  background: #f8f9fa;
  border-radius: 12rpx;
  border-left: 6rpx solid #fa2c19;
}

.reminder-item.status-due {
  border-left-color: #fa2c19;
}

.reminder-item.status-overdue {
  border-left-color: #ff4d4f;
  background: #fff1f0;
}

.reminder-content {
  flex: 1;
}

.vaccine-name {
  font-size: 28rpx;
  font-weight: bold;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.vaccine-date {
  font-size: 24rpx;
  color: #666;
  margin-bottom: 8rpx;
}

.status-badge {
  display: inline-block;
  padding: 4rpx 12rpx;
  border-radius: 4rpx;
  font-size: 20rpx;
  color: white;

  &.due {
    background: #fa2c19;
  }

  &.overdue {
    background: #ff4d4f;
  }
}

.plan-item {
  padding: 24rpx;
  background: #f8f9fa;
  border-radius: 12rpx;

  &.completed {
    opacity: 0.6;
  }
}

.plan-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}

.plan-name {
  font-size: 28rpx;
  font-weight: bold;
  color: #1a1a1a;
}

.required-badge {
  display: inline-block;
  padding: 4rpx 8rpx;
  background: #fa2c19;
  color: white;
  font-size: 20rpx;
  border-radius: 4rpx;
  margin-right: 8rpx;
}

.plan-age {
  font-size: 24rpx;
  color: #fa2c19;
  font-weight: bold;
}

.plan-detail {
  display: flex;
  gap: 20rpx;
  margin-bottom: 12rpx;
  font-size: 24rpx;
  color: #666;
}

.plan-record, .plan-action {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-top: 12rpx;
}

.completed-icon {
  font-size: 32rpx;
  color: #52c41a;
}

.completed-text {
  font-size: 26rpx;
  color: #52c41a;
  font-weight: bold;
}

.completed-date {
  font-size: 24rpx;
  color: #999;
}

.dialog-content {
  padding: 30rpx;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.dialog-title {
  font-size: 36rpx;
  font-weight: bold;
  text-align: center;
  margin-bottom: 30rpx;
}

.form-section {
  flex: 1;
  overflow-y: auto;
}

.form-item {
  margin-bottom: 30rpx;
}

.form-label {
  font-size: 28rpx;
  font-weight: bold;
  margin-bottom: 12rpx;
  color: #333;
}

.dialog-footer {
  display: flex;
  gap: 20rpx;
  margin-top: 20rpx;
}
</style>
