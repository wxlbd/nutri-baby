<template>
  <view class="index-page">
    <!-- 自定义导航栏 -->
    <custom-navbar ref="navbarRef" title="今日概览" />

    <!-- 页面内容 -->
    <view class="page-content" :style="{ paddingTop: pageContentPaddingTop }">
      <!-- 今日数据概览 -->
    <view class="today-stats">
      <view class="stats-title">今日数据</view>
      <view class="stats-grid">
        <view class="stat-item">
          <view class="stat-icon">🍼</view>
          <view class="stat-value">{{ todayStats.totalMilk }}ml</view>
          <view class="stat-label">奶瓶奶量</view>
        </view>
        <view class="stat-item">
          <view class="stat-icon">🤱</view>
          <view class="stat-value">{{ todayStats.breastfeedingCount }}次</view>
          <view class="stat-label">母乳喂养</view>
        </view>
        <view class="stat-item">
          <view class="stat-icon">💤</view>
          <view class="stat-value">{{ formatDuration(todayStats.sleepDuration) }}</view>
          <view class="stat-label">睡眠时长</view>
        </view>
        <view class="stat-item">
          <view class="stat-icon">🧷</view>
          <view class="stat-value">{{ todayStats.diaperCount }}</view>
          <view class="stat-label">换尿布</view>
        </view>
      </view>
    </view>

    <!-- 距上次喂奶时间 -->
    <view class="last-feeding">
      <view class="time-info">
        <text class="label">距上次喂奶</text>
        <text class="time">{{ lastFeedingTime }}</text>
        <text v-if="nextFeedingTime" class="next-time">
          {{ nextFeedingTime }}
        </text>
      </view>
    </view>

    <!-- 疫苗提醒 -->
    <view v-if="upcomingVaccines.length > 0" class="vaccine-reminder" @click="goToVaccine">
      <view class="reminder-header">
        <view class="header-left">
          <text class="reminder-icon">💉</text>
          <text class="reminder-title">疫苗提醒</text>
        </view>
        <view class="header-right">
          <text class="view-all">查看全部</text>
          <nut-icon name="right" size="14" />
        </view>
      </view>
      <view class="vaccine-list">
        <view
          v-for="vaccine in upcomingVaccines"
          :key="vaccine.id"
          class="vaccine-item"
          :class="`status-${vaccine.status}`"
        >
          <view class="vaccine-info">
            <text class="vaccine-name">{{ vaccine.vaccineName }} (第{{ vaccine.doseNumber }}针)</text>
            <text class="vaccine-date">{{ formatVaccineDate(vaccine.scheduledDate) }}</text>
          </view>
          <view class="vaccine-badge" :class="vaccine.status">
            {{ vaccine.status === 'due' ? '即将到期' : '已逾期' }}
          </view>
        </view>
      </view>
    </view>

    <!-- 快捷操作 -->
    <view class="quick-actions">
      <view class="action-title">快捷记录</view>
      <view class="action-buttons">
        <view class="button-row">
          <nut-button
            type="primary"
            size="large"
            @click="handleFeeding"
          >
            <view class="button-content">
              <text class="icon">🍼</text>
              <text>喂养</text>
            </view>
          </nut-button>
          <nut-button
            type="success"
            size="large"
            @click="handleDiaper"
          >
            <view class="button-content">
              <text class="icon">🧷</text>
              <text>换尿布</text>
            </view>
          </nut-button>
        </view>
        <view class="button-row">
          <nut-button
            type="info"
            size="large"
            @click="handleSleep"
          >
            <view class="button-content">
              <text class="icon">💤</text>
              <text>睡觉</text>
            </view>
          </nut-button>
          <nut-button
            type="warning"
            size="large"
            @click="handleGrowth"
          >
            <view class="button-content">
              <text class="icon">📏</text>
              <text>成长</text>
            </view>
          </nut-button>
        </view>
      </view>
    </view>

    <!-- 底部提示 -->
    <view v-if="!isLoggedIn" class="login-tip">
      <nut-button type="primary" size="small" @click="goToLogin">
        请先登录
      </nut-button>
    </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { isLoggedIn, fetchUserInfo } from '@/store/user'
import { currentBaby, fetchBabyList } from '@/store/baby'
import { getTodayTotalMilk, getLastFeedingRecord, getTodayBreastfeedingStats } from '@/store/feeding'
import { getTodayDiaperCount } from '@/store/diaper'
import { getTodayTotalSleepDuration } from '@/store/sleep'
import { getUpcomingReminders, initializeVaccinePlansFromServer, generateRemindersForBaby } from '@/store/vaccine'
import { formatRelativeTime, formatDuration, formatDate } from '@/utils/date'

// 导航栏引用
const navbarRef = ref<any>(null)
// 页面内容区域的 padding-top
const pageContentPaddingTop = ref('152rpx') // 默认值（状态栏44px + 内容88rpx + 间距20rpx）

// 今日数据统计
const todayStats = computed(() => {
  if (!currentBaby.value) {
    return {
      totalMilk: 0,
      breastfeedingCount: 0,
      sleepDuration: 0,
      diaperCount: 0
    }
  }

  const breastfeedingStats = getTodayBreastfeedingStats(currentBaby.value.babyId)

  return {
    totalMilk: getTodayTotalMilk(currentBaby.value.babyId),
    breastfeedingCount: breastfeedingStats.count,
    sleepDuration: getTodayTotalSleepDuration(currentBaby.value.babyId),
    diaperCount: getTodayDiaperCount(currentBaby.value.babyId)
  }
})

// 距上次喂奶时间
const lastFeedingTime = computed(() => {
  if (!currentBaby.value) return '-'

  const lastRecord = getLastFeedingRecord(currentBaby.value.babyId)
  if (!lastRecord) return '-'

  return formatRelativeTime(lastRecord.time)
})

// 下次喂奶建议时间
const nextFeedingTime = computed(() => {
  if (!currentBaby.value) return ''

  const lastRecord = getLastFeedingRecord(currentBaby.value.babyId)
  if (!lastRecord) return ''

  // 计算宝宝月龄
  const birthDate = new Date(currentBaby.value.birthDate)
  const now = new Date()
  const monthsOld = (now.getFullYear() - birthDate.getFullYear()) * 12 +
                    (now.getMonth() - birthDate.getMonth())

  // 根据月龄确定建议喂奶间隔(分钟)
  let intervalMinutes = 180 // 默认3小时
  if (monthsOld < 1) {
    intervalMinutes = 120 // 新生儿: 2小时
  } else if (monthsOld < 3) {
    intervalMinutes = 150 // 1-3个月: 2.5小时
  } else if (monthsOld < 6) {
    intervalMinutes = 180 // 3-6个月: 3小时
  } else {
    intervalMinutes = 240 // 6个月以上: 4小时
  }

  const nextTime = lastRecord.time + intervalMinutes * 60 * 1000
  const timeDiff = nextTime - Date.now()

  if (timeDiff <= 0) {
    return '建议现在喂奶'
  }

  const hours = Math.floor(timeDiff / (60 * 60 * 1000))
  const minutes = Math.floor((timeDiff % (60 * 60 * 1000)) / (60 * 1000))

  if (hours > 0) {
    return `建议 ${hours}小时${minutes}分钟后喂奶`
  } else {
    return `建议 ${minutes}分钟后喂奶`
  }
})


// 即将到期的疫苗(最多显示2个)
const upcomingVaccines = computed(() => {
  if (!currentBaby.value) return []
  return getUpcomingReminders(currentBaby.value.babyId).slice(0, 2)
})

// 格式化疫苗日期
const formatVaccineDate = (timestamp: number): string => {
  return formatDate(timestamp, 'MM-DD')
}

// 页面加载
onMounted(async () => {
  await checkLoginAndBaby()
  initializeVaccines()

  // 计算页面内容区域的 padding-top
  calculatePagePadding()
})

// 计算页面内容的 padding-top
const calculatePagePadding = () => {
  if (navbarRef.value && navbarRef.value.navbarTotalHeight) {
    const totalHeight = navbarRef.value.navbarTotalHeight
    // 导航栏总高度 + 间距 20rpx
    pageContentPaddingTop.value = `${totalHeight + 20}rpx`
  }
}

// 初始化疫苗计划
const initializeVaccines = async () => {
  if (!currentBaby.value) return

  // 为当前宝宝从服务器初始化疫苗计划
  await initializeVaccinePlansFromServer(currentBaby.value.babyId)

  // 生成提醒
  generateRemindersForBaby(currentBaby.value.babyId, currentBaby.value.birthDate)
}

// 检查登录和宝宝信息
const checkLoginAndBaby = async () => {
  // 1. 检查登录状态
  if (!isLoggedIn.value) {
    // 未登录,跳转到登录页
    uni.reLaunch({
      url: '/pages/user/login'
    })
    return
  }

  try {
    // 2. 获取用户信息
    await fetchUserInfo()

    // 3. 获取宝宝列表
    await fetchBabyList()

    // 4. 检查是否有宝宝
    if (!currentBaby.value) {
      // 没有宝宝,跳转到添加宝宝页面
      uni.showModal({
        title: '提示',
        content: '请先添加宝宝信息',
        showCancel: false,
        success: () => {
          uni.navigateTo({
            url: '/pages/baby/edit/edit'
          })
        }
      })
    }
  } catch (error) {
    console.error('[Index] 获取用户/宝宝信息失败:', error)
    uni.showToast({
      title: '加载数据失败',
      icon: 'none'
    })
  }
}

// 跳转到登录
const goToLogin = () => {
  uni.navigateTo({
    url: '/pages/user/login'
  })
}

// 跳转到疫苗提醒
const goToVaccine = () => {
  if (!currentBaby.value) {
    uni.showToast({
      title: '请先添加宝宝',
      icon: 'none'
    })
    return
  }
  uni.navigateTo({
    url: '/pages/vaccine/vaccine'
  })
}

// 喂养记录
const handleFeeding = () => {
  if (!currentBaby.value) {
    uni.showToast({
      title: '请先添加宝宝',
      icon: 'none'
    })
    return
  }
  uni.navigateTo({
    url: '/pages/record/feeding/feeding'
  })
}

// 换尿布记录
const handleDiaper = () => {
  if (!currentBaby.value) {
    uni.showToast({
      title: '请先添加宝宝',
      icon: 'none'
    })
    return
  }
  uni.navigateTo({
    url: '/pages/record/diaper/diaper'
  })
}

// 睡眠记录
const handleSleep = () => {
  if (!currentBaby.value) {
    uni.showToast({
      title: '请先添加宝宝',
      icon: 'none'
    })
    return
  }
  uni.navigateTo({
    url: '/pages/record/sleep/sleep'
  })
}

// 成长记录
const handleGrowth = () => {
  if (!currentBaby.value) {
    uni.showToast({
      title: '请先添加宝宝',
      icon: 'none'
    })
    return
  }
  uni.navigateTo({
    url: '/pages/record/growth/growth'
  })
}
</script>

<style lang="scss" scoped>
// ===== 设计系统变量 =====
$spacing: 20rpx;  // 统一间距

.index-page {
  min-height: 100vh;
  background: #f5f5f5;
}

// 页面内容区域 - 修复布局
.page-content {
  // 顶部由内联样式动态设置 (导航栏总高度 + 间距)
  padding-left: $spacing;
  padding-right: $spacing;
  padding-bottom: $spacing;

  // 为 tabBar 预留空间（env(safe-area-inset-bottom) 处理全面屏底部安全区）
  margin-bottom: calc(100rpx + env(safe-area-inset-bottom));
}

// 今日数据卡片
.today-stats {
  background: white;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: $spacing;
}

.stats-title {
  font-size: 32rpx;
  font-weight: bold;
  margin-bottom: 24rpx;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20rpx;
}

.stat-item {
  text-align: center;
  padding: 20rpx;
  background: #f5f5f5;
  border-radius: 12rpx;
}

.stat-icon {
  font-size: 40rpx;
  margin-bottom: 12rpx;
}

.stat-value {
  font-size: 32rpx;
  font-weight: bold;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.stat-label {
  font-size: 24rpx;
  color: #808080;
}

.last-feeding {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: $spacing;
  color: white;
  text-align: center;
}

.time-info {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.label {
  font-size: 28rpx;
  opacity: 0.9;
}

.time {
  font-size: 48rpx;
  font-weight: bold;
}

.next-time {
  font-size: 24rpx;
  opacity: 0.8;
  text-align: center;
}

.vaccine-reminder {
  background: white;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: $spacing;
}

.reminder-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.reminder-icon {
  font-size: 32rpx;
}

.reminder-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #1a1a1a;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8rpx;
  color: #999;
  font-size: 24rpx;
}

.vaccine-list {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.vaccine-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx;
  background: #f8f9fa;
  border-radius: 12rpx;
  border-left: 6rpx solid #fa2c19;

  &.status-due {
    border-left-color: #fa2c19;
    background: #fff7f0;
  }

  &.status-overdue {
    border-left-color: #ff4d4f;
    background: #fff1f0;
  }
}

.vaccine-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.vaccine-name {
  font-size: 28rpx;
  font-weight: bold;
  color: #1a1a1a;
}

.vaccine-date {
  font-size: 24rpx;
  color: #666;
}

.vaccine-badge {
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
  font-size: 22rpx;
  color: white;

  &.due {
    background: #fa2c19;
  }

  &.overdue {
    background: #ff4d4f;
  }
}

.quick-actions {
  background: white;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: $spacing;
}

.action-title {
  font-size: 32rpx;
  font-weight: bold;
  margin-bottom: 24rpx;
}

.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.button-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16rpx;
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

.login-tip {
  text-align: center;
  padding: 40rpx 0;
}
</style>