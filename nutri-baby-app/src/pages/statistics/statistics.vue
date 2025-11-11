<template>
  <view class="statistics-page">
    <!-- 时间范围选择 -->
    <view class="time-range">
      <wd-tabs v-model="timeRange">
        <wd-tab title="本周" name="week" />
        <wd-tab title="本月" name="month" />
      </wd-tabs>
    </view>

    <!-- 未登录提示 -->
    <view v-if="!isLoggedIn" class="guest-tip">
      <image src="/static/monitoring_24dp_A2D5AB_FILL0_wght400_GRAD0_opsz24.svg" class="tip-icon" />
      <text class="tip-text">登录后查看数据</text>
    </view>

    <!-- 喂养统计 -->
    <view class="stat-section">
      <view class="section-header">
        <image class="icon-img" src="/static/breastfeeding_24dp_A2D5AB_FILL0_wght400_GRAD0_opsz24.svg" mode="aspectFit" />
        <text class="title">喂养统计</text>
      </view>

      <view class="stat-cards">
        <view class="stat-card">
          <view class="card-label">奶瓶奶量</view>
          <view class="card-value">{{ feedingStats.totalMilk }}ml</view>
        </view>
        <view class="stat-card">
          <view class="card-label">喂养次数</view>
          <view class="card-value">{{ feedingStats.count }}次</view>
        </view>
        <view class="stat-card">
          <view class="card-label">日均奶量</view>
          <view class="card-value">{{ feedingStats.avgMilk }}ml</view>
        </view>
      </view>

      <!-- 每日奶量柱状图 - 使用 uCharts -->
      <view class="daily-chart">
        <view class="chart-title">每日奶瓶奶量趋势</view>
        <UChart
          v-if="feedingChartData.series.length > 0"
          canvas-id="feedingChart"
          chart-type="column"
          :chart-data="feedingChartData"
          :chart-options="feedingChartOpts"
          height="350rpx"
        />
        <view v-else class="empty-chart">暂无数据</view>
      </view>
    </view>

    <!-- 睡眠统计 -->
    <view class="stat-section">
      <view class="section-header">
        <image class="icon-img" src="/static/moon_stars_24dp_A2D5AB_FILL0_wght400_GRAD0_opsz24.svg" mode="aspectFit" />
        <text class="title">睡眠统计</text>
      </view>

      <view class="stat-cards">
        <view class="stat-card">
          <view class="card-label">总时长</view>
          <view class="card-value">{{ sleepStats.totalHours }}h</view>
        </view>
        <view class="stat-card">
          <view class="card-label">睡眠次数</view>
          <view class="card-value">{{ sleepStats.count }}次</view>
        </view>
        <view class="stat-card">
          <view class="card-label">日均时长</view>
          <view class="card-value">{{ sleepStats.avgHours }}h</view>
        </view>
      </view>

      <!-- 睡眠质量分析 -->
      <view class="sleep-quality">
        <view class="quality-title">睡眠质量分析</view>
        <view class="quality-content">
          <view class="quality-item">
            <text class="quality-label">最长单次睡眠:</text>
            <text class="quality-value">{{ sleepStats.longestSleep }}分钟</text>
          </view>
          <view class="quality-item">
            <text class="quality-label">平均单次时长:</text>
            <text class="quality-value">{{ sleepStats.avgSingleSleep }}分钟</text>
          </view>
          <view class="quality-item">
            <text class="quality-label">夜间睡眠:</text>
            <text class="quality-value">{{ sleepStats.nightSleepCount }}次 ({{ sleepStats.nightSleepHours }}h)</text>
          </view>
          <view class="quality-item">
            <text class="quality-label">小睡:</text>
            <text class="quality-value">{{ sleepStats.napCount }}次 ({{ sleepStats.napHours }}h)</text>
          </view>
          <view v-if="sleepStats.recommendation" class="quality-recommendation">
            <image src="/static/moon_stars_24dp_A2D5AB_FILL0_wght400_GRAD0_opsz24.svg" class="recommendation-icon" />
            <text class="recommendation-text">{{ sleepStats.recommendation }}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 排泄统计 -->
    <view class="stat-section">
      <view class="section-header">
        <image class="icon-img" src="/static/baby_changing_station_24dp_A2D5AB_FILL0_wght400_GRAD0_opsz24.svg" mode="aspectFit" />
        <text class="title">排泄统计</text>
      </view>

      <view class="stat-cards">
        <view class="stat-card">
          <view class="card-label">换尿布</view>
          <view class="card-value">{{ diaperStats.total }}次</view>
        </view>
        <view class="stat-card">
          <view class="card-label">小便</view>
          <view class="card-value">{{ diaperStats.wet }}次</view>
        </view>
        <view class="stat-card">
          <view class="card-label">大便</view>
          <view class="card-value">{{ diaperStats.dirty }}次</view>
        </view>
      </view>
    </view>

    <!-- 成长统计 -->
    <view v-if="growthStats.hasData" class="stat-section">
      <view class="section-header">
        <image class="icon-img" src="/static/monitoring_24dp_A2D5AB_FILL0_wght400_GRAD0_opsz24.svg" mode="aspectFit" />
        <text class="title">成长统计</text>
      </view>

      <!-- 最新数据 -->
      <view class="stat-cards">
        <view v-if="growthStats.latestHeight" class="stat-card">
          <view class="card-label">最新身高</view>
          <view class="card-value">{{ growthStats.latestHeight }}cm</view>
        </view>
        <view v-if="growthStats.latestWeight" class="stat-card">
          <view class="card-label">最新体重</view>
          <view class="card-value">{{ growthStats.latestWeight }}kg</view>
        </view>
        <view v-if="growthStats.latestHead" class="stat-card">
          <view class="card-label">最新头围</view>
          <view class="card-value">{{ growthStats.latestHead }}cm</view>
        </view>
      </view>

      <!-- 成长曲线 -->
      <view class="growth-charts">
        <!-- 身高曲线 - 使用 uCharts -->
        <view v-if="growthStats.heightData.length > 0" class="chart-container">
          <view class="chart-title">身高趋势 (cm)</view>
          <UChart
            canvas-id="heightChart"
            chart-type="line"
            :chart-data="heightChartData"
            :chart-options="heightChartOpts"
            height="350rpx"
          />
        </view>

        <!-- 体重曲线 - 使用 uCharts -->
        <view v-if="growthStats.weightData.length > 0" class="chart-container">
          <view class="chart-title">体重趋势 (kg)</view>
          <UChart
            canvas-id="weightChart"
            chart-type="line"
            :chart-data="weightChartData"
            :chart-options="weightChartOpts"
            height="350rpx"
          />
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { isLoggedIn } from '@/store/user'
import { currentBaby } from '@/store/baby'
import { getWeekStart, getMonthStart, formatDate } from '@/utils/date'
import { useUChart, columnChartPreset, lineChartPreset } from '@/composables/useUChart'
import type { ChartData, ChartOptions } from '@/composables/useUChart'
import UChart from '@/components/UChart.vue'

// 直接调用 API 层
import * as feedingApi from '@/api/feeding'
import * as sleepApi from '@/api/sleep'
import * as diaperApi from '@/api/diaper'
import * as growthApi from '@/api/growth'

// 时间范围
const timeRange = ref<string>('week')

// 获取时间范围
const getTimeRange = () => {
  const now = Date.now()
  const start = timeRange.value === 'week' ? getWeekStart() : getMonthStart()
  return { start, end: now }
}

// 记录数据(从 API 获取)
const feedingRecords = ref<feedingApi.FeedingRecordResponse[]>([])
const sleepRecords = ref<sleepApi.SleepRecordResponse[]>([])
const diaperRecords = ref<diaperApi.DiaperRecordResponse[]>([])
const growthRecords = ref<growthApi.GrowthRecordResponse[]>([])

// 加载所有记录
const loadRecords = async () => {
  if (!currentBaby.value) return

  const babyId = currentBaby.value.babyId
  const { start, end } = getTimeRange()

  try {
    const [feedingData, sleepData, diaperData, growthData] = await Promise.all([
      feedingApi.apiFetchFeedingRecords({ babyId, startTime: start, endTime: end, pageSize: 500 }),
      sleepApi.apiFetchSleepRecords({ babyId, startTime: start, endTime: end, pageSize: 500 }),
      diaperApi.apiFetchDiaperRecords({ babyId, startTime: start, endTime: end, pageSize: 500 }),
      growthApi.apiFetchGrowthRecords({ babyId, pageSize: 100 }) // 成长记录不限制时间范围
    ])

    feedingRecords.value = feedingData.records
    sleepRecords.value = sleepData.records
    diaperRecords.value = diaperData.records
    growthRecords.value = growthData.records
  } catch (error) {
    console.error('加载统计数据失败:', error)
    uni.showToast({
      title: '加载数据失败',
      icon: 'none'
    })
  }
}

// 监听时间范围变化,重新加载数据
watch(timeRange, () => {
  loadRecords()
})

// 喂养统计
const feedingStats = computed(() => {
  if (!currentBaby.value) {
    return {
      totalMilk: 0,
      count: 0,
      avgMilk: 0,
      dailyData: [],
      maxDaily: 0,
    }
  }

  let totalMilk = 0
  const dailyMap = new Map<string, number>()

  feedingRecords.value.forEach(record => {
    // 只统计奶瓶喂养的奶量，母乳喂养不计入
    if (record.feedingType === 'bottle') {
      const feedingDetail = record.detail
      const unit = (feedingDetail && feedingDetail.type === 'bottle') ? feedingDetail.unit : 'ml'
      const amount = unit === 'oz'
        ? (record.amount || 0) * 29.5735
        : (record.amount || 0)

      totalMilk += amount

      // 按日期统计
      const date = formatDate(record.feedingTime, 'MM-DD')
      dailyMap.set(date, (dailyMap.get(date) || 0) + amount)
    }
  })

  // 生成每日数据
  const days = timeRange.value === 'week' ? 7 : 30
  const dailyData = []
  let maxDaily = 0

  for (let i = days - 1; i >= 0; i--) {
    const date = new Date(Date.now() - i * 24 * 60 * 60 * 1000)
    const dateStr = formatDate(date.getTime(), 'MM-DD')
    const amount = Math.round(dailyMap.get(dateStr) || 0)

    dailyData.push({
      label: i === 0 ? '今' : formatDate(date.getTime(), 'DD'),
      amount,
    })

    if (amount > maxDaily) maxDaily = amount
  }

  return {
    totalMilk: Math.round(totalMilk),
    count: feedingRecords.value.length,
    avgMilk: feedingRecords.value.length > 0 ? Math.round(totalMilk / days) : 0,
    dailyData,
    maxDaily,
  }
})

// 喂养图表数据
const feedingChartData = computed<ChartData>(() => {
  if (!feedingStats.value.dailyData || feedingStats.value.dailyData.length === 0) {
    return {
      categories: [],
      series: []
    }
  }

  return {
    categories: feedingStats.value.dailyData.map(d => d.label),
    series: [
      {
        name: '奶量(ml)',
        data: feedingStats.value.dailyData.map(d => d.amount)
      }
    ]
  }
})

// 喂养图表配置
const feedingChartOpts = computed<ChartOptions>(() => ({
  ...columnChartPreset(),
  padding: [15, 15, 0, 15],
  extra: {
    column: {
      type: 'group',
      width: 30
    }
  }
}))

// 睡眠统计
const sleepStats = computed(() => {
  if (!currentBaby.value) {
    return {
      totalHours: 0,
      count: 0,
      avgHours: 0,
      longestSleep: 0,
      avgSingleSleep: 0,
      nightSleepCount: 0,
      nightSleepHours: 0,
      napCount: 0,
      napHours: 0,
      recommendation: ''
    }
  }

  const totalMinutes = sleepRecords.value.reduce((sum, r) => sum + (r.duration || 0), 0)
  const days = timeRange.value === 'week' ? 7 : 30

  // 计算最长单次睡眠
  const longestSleep = sleepRecords.value.length > 0
    ? Math.max(...sleepRecords.value.map(r => r.duration || 0))
    : 0

  // 计算平均单次睡眠
  const avgSingleSleep = sleepRecords.value.length > 0
    ? Math.round(totalMinutes / sleepRecords.value.length)
    : 0

  // 统计夜间睡眠和小睡
  let nightSleepMinutes = 0
  let nightSleepCount = 0
  let napMinutes = 0
  let napCount = 0

  sleepRecords.value.forEach(r => {
    if (r.sleepType === 'night') {
      nightSleepMinutes += r.duration || 0
      nightSleepCount++
    } else {
      napMinutes += r.duration || 0
      napCount++
    }
  })

  // 计算宝宝月龄
  const birthDate = new Date(currentBaby.value.birthDate)
  const now = new Date()
  const monthsOld = (now.getFullYear() - birthDate.getFullYear()) * 12 +
                    (now.getMonth() - birthDate.getMonth())

  // 生成建议
  let recommendation = ''
  const dailyHours = totalMinutes / days / 60

  // 根据月龄判断睡眠是否充足
  if (monthsOld < 3) {
    // 0-3个月: 14-17小时
    if (dailyHours < 14) {
      recommendation = '建议增加睡眠时间,新生儿需要14-17小时睡眠'
    } else if (dailyHours > 17) {
      recommendation = '睡眠时间较长,如有异常请咨询医生'
    } else {
      recommendation = '睡眠时间正常,继续保持'
    }
  } else if (monthsOld < 12) {
    // 3-12个月: 12-16小时
    if (dailyHours < 12) {
      recommendation = '建议增加睡眠时间,婴儿需要12-16小时睡眠'
    } else if (dailyHours > 16) {
      recommendation = '睡眠时间较长,注意观察宝宝状态'
    } else {
      recommendation = '睡眠时间正常,继续保持'
    }
  } else {
    // 12个月以上: 11-14小时
    if (dailyHours < 11) {
      recommendation = '建议增加睡眠时间,幼儿需要11-14小时睡眠'
    } else if (dailyHours > 14) {
      recommendation = '睡眠时间较长,可适当增加活动'
    } else {
      recommendation = '睡眠时间正常,继续保持'
    }
  }

  return {
    totalHours: Math.round(totalMinutes / 60 * 10) / 10,
    count: sleepRecords.value.length,
    avgHours: Math.round((totalMinutes / days / 60) * 10) / 10,
    longestSleep,
    avgSingleSleep,
    nightSleepCount,
    nightSleepHours: Math.round(nightSleepMinutes / 60 * 10) / 10,
    napCount,
    napHours: Math.round(napMinutes / 60 * 10) / 10,
    recommendation
  }
})

// 排泄统计
const diaperStats = computed(() => {
  if (!currentBaby.value) {
    return { total: 0, wet: 0, dirty: 0 }
  }

  let wet = 0
  let dirty = 0

  diaperRecords.value.forEach(r => {
    if (r.diaperType === 'pee') wet++
    else if (r.diaperType === 'poo') dirty++
    else {
      wet++
      dirty++
    }
  })

  return {
    total: diaperRecords.value.length,
    wet,
    dirty,
  }
})

// 成长统计
const growthStats = computed(() => {
  if (!currentBaby.value) {
    return {
      hasData: false,
      latestHeight: 0,
      latestWeight: 0,
      latestHead: 0,
      dates: [],
      heightData: [],
      weightData: [],
      headData: [],
      heightMin: 0,
      heightMax: 0,
      weightMin: 0,
      weightMax: 0
    }
  }

  if (growthRecords.value.length === 0) {
    return {
      hasData: false,
      latestHeight: 0,
      latestWeight: 0,
      latestHead: 0,
      dates: [],
      heightData: [],
      weightData: [],
      headData: [],
      heightMin: 0,
      heightMax: 0,
      weightMin: 0,
      weightMax: 0
    }
  }

  // 最新数据
  const latestRecord = growthRecords.value[0]

  if (!latestRecord) {
    return {
      hasData: false,
      latestHeight: 0,
      latestWeight: 0,
      latestHead: 0,
      dates: [],
      heightData: [],
      weightData: [],
      headData: [],
      heightMin: 0,
      heightMax: 0,
      weightMin: 0,
      weightMax: 0
    }
  }

  // 准备曲线数据（按时间正序）
  const sortedRecords = [...growthRecords.value].reverse()
  const dates: string[] = []
  const heightData: number[] = []
  const weightData: number[] = []
  const headData: number[] = []

  sortedRecords.forEach(record => {
    const date = new Date(record.measureTime)
    dates.push(`${date.getMonth() + 1}/${date.getDate()}`)

    if (record.height) heightData.push(record.height)
    if (record.weight) weightData.push(record.weight)
    if (record.headCircumference) headData.push(record.headCircumference)
  })

  // 计算最大最小值
  const heightMin = heightData.length > 0 ? Math.min(...heightData) : 0
  const heightMax = heightData.length > 0 ? Math.max(...heightData) : 0
  const weightMin = weightData.length > 0 ? Math.min(...weightData) : 0
  const weightMax = weightData.length > 0 ? Math.max(...weightData) : 0

  return {
    hasData: true,
    latestHeight: latestRecord.height || 0,
    latestWeight: latestRecord.weight || 0,
    latestHead: latestRecord.headCircumference || 0,
    dates,
    heightData,
    weightData,
    headData,
    heightMin: Math.floor(heightMin - 2),
    heightMax: Math.ceil(heightMax + 2),
    weightMin: Math.floor(weightMin - 0.5),
    weightMax: Math.ceil(weightMax + 0.5)
  }
})

// 身高图表数据
const heightChartData = computed<ChartData>(() => {
  if (!growthStats.value.heightData || growthStats.value.heightData.length === 0) {
    return {
      categories: [],
      series: []
    }
  }

  return {
    categories: growthStats.value.dates,
    series: [
      {
        name: '身高(cm)',
        data: growthStats.value.heightData
      }
    ]
  }
})

// 身高图表配置
const heightChartOpts = computed<ChartOptions>(() => ({
  ...lineChartPreset(),
  color: ['#7dd3a2'],
  extra: {
    line: {
      type: 'curve',
      width: 2
    }
  }
}))

// 体重图表数据
const weightChartData = computed<ChartData>(() => {
  if (!growthStats.value.weightData || growthStats.value.weightData.length === 0) {
    return {
      categories: [],
      series: []
    }
  }

  return {
    categories: growthStats.value.dates,
    series: [
      {
        name: '体重(kg)',
        data: growthStats.value.weightData
      }
    ]
  }
})

// 体重图表配置
const weightChartOpts = computed<ChartOptions>(() => ({
  ...lineChartPreset(),
  color: ['#ff7f50'],
  extra: {
    line: {
      type: 'curve',
      width: 2
    }
  }
}))

// 页面加载
onMounted(() => {
  if (!isLoggedIn.value) {
    return
  }

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

  // 加载数据
  loadRecords()
})
</script>

<style lang="scss" scoped>
.statistics-page {
  min-height: 100vh;
  background: #f6f8f7;
  padding-bottom: 40rpx;
}

.guest-tip {
  background: rgba(125, 211, 162, 0.15);
  color: #333;
  padding: 24rpx 30rpx;
  margin: 20rpx;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  gap: 16rpx;
  border: 1rpx solid rgba(125, 211, 162, 0.3);
}

.tip-icon {
  width: 36rpx;
  height: 36rpx;
}

.tip-text {
  font-size: 28rpx;
  font-weight: 500;
}

.time-range {
  background: white;
}

.stat-section {
  background: white;
  border: 1rpx solid #CAE3D4;
  border-radius: 16rpx;
  margin: 20rpx 20rpx 0;
  padding: 30rpx;
  box-shadow: 0 2rpx 8rpx rgba(125, 211, 162, 0.08);
}

.section-header {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 24rpx;

  .icon {
    font-size: 40rpx;
  }

  .icon-img {
    width: 48rpx;
    height: 48rpx;
  }

  .title {
    font-size: 32rpx;
    font-weight: bold;
  }
}

.stat-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20rpx;
}

.stat-card {
  background: #f6f8f7;
  border-radius: 12rpx;
  padding: 24rpx;
  text-align: center;
}

.card-label {
  font-size: 24rpx;
  color: #999;
  margin-bottom: 12rpx;
}

.card-value {
  font-size: 32rpx;
  font-weight: bold;
  color: #7dd3a2;
}

.sleep-quality {
  margin-top: 30rpx;
  background: #f0f9f6;
  border-radius: 12rpx;
  padding: 24rpx;
  border: 1rpx solid #CAE3D4;
}

.quality-title {
  font-size: 28rpx;
  font-weight: bold;
  margin-bottom: 20rpx;
  color: #333;
}

.quality-content {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.quality-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 26rpx;
}

.quality-label {
  color: #666;
}

.quality-value {
  color: #333;
  font-weight: 500;
}

.quality-recommendation {
  margin-top: 16rpx;
  padding: 20rpx;
  background: rgba(125, 211, 162, 0.15);
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  gap: 12rpx;
  border: 1rpx solid #CAE3D4;
}

.recommendation-icon {
  width: 32rpx;
  height: 32rpx;
}

.recommendation-text {
  flex: 1;
  font-size: 26rpx;
  color: #333;
  line-height: 1.6;
}


.daily-chart {
  margin-top: 30rpx;
}

.chart-title {
  font-size: 28rpx;
  font-weight: bold;
  margin-bottom: 20rpx;
}

.empty-chart {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 350rpx;
  background: #f6f8f7;
  border-radius: 12rpx;
  color: #999;
  font-size: 26rpx;
}

.growth-charts {
  margin-top: 30rpx;
}

.chart-container {
  margin-bottom: 40rpx;

  &:last-child {
    margin-bottom: 0;
  }
}
</style>