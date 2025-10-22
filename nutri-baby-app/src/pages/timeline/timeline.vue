<template>
  <view class="timeline-page">
    <!-- 日期筛选 -->
    <view class="date-filter">
      <nut-button size="small" @click="filterDate('today')">今天</nut-button>
      <nut-button size="small" @click="filterDate('week')">本周</nut-button>
      <nut-button size="small" @click="filterDate('month')">本月</nut-button>
      <nut-button size="small" @click="showDatePicker = true">自定义</nut-button>
    </view>

    <!-- 记录列表 -->
    <view class="timeline-list">
      <view v-if="groupedRecords.length === 0" class="empty-state">
        <nut-empty description="暂无记录" />
      </view>

      <view v-else>
        <view
          v-for="group in groupedRecords"
          :key="group.date"
          class="date-group"
        >
          <!-- 日期标题 -->
          <view class="date-header">{{ group.dateText }}</view>

          <!-- 记录列表 -->
          <view
            v-for="record in group.records"
            :key="record.id"
            class="record-item"
            :class="`record-${record.type}`"
          >
            <!-- 时间轴圆点 -->
            <view class="timeline-dot" :class="`dot-${record.type}`"></view>
            <view class="timeline-line"></view>

            <!-- 记录内容 -->
            <view class="record-content">
              <view class="record-header">
                <view class="record-type">
                  <text class="type-icon">{{ record.icon }}</text>
                  <text class="type-name">{{ record.typeName }}</text>
                </view>
                <text class="record-time">{{ record.timeText }}</text>
              </view>

              <view class="record-detail">{{ record.detail }}</view>

              <!-- 操作按钮 -->
              <view class="record-actions">
                <nut-button
                  size="small"
                  type="default"
                  @click="deleteRecord(record)"
                >
                  删除
                </nut-button>
              </view>
            </view>
          </view>
        </view>
      </view>
    </view>

    <!-- 日期选择器 -->
    <nut-date-picker
      v-model:visible="showDatePicker"
      v-model="selectedDate"
      type="date"
      title="选择日期"
      @confirm="handleDateConfirm"
    />
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { currentBaby } from '@/store/baby'
import { getFeedingRecordsByBabyId, deleteFeedingRecord } from '@/store/feeding'
import { getDiaperRecordsByBabyId, deleteDiaperRecord } from '@/store/diaper'
import { getSleepRecordsByBabyId, deleteSleepRecord } from '@/store/sleep'
import { formatDate, isToday, getTodayStart, getWeekStart, getMonthStart } from '@/utils/date'
import type { FeedingRecord, DiaperRecord, SleepRecord } from '@/types'

// 日期筛选
const filterType = ref<'today' | 'week' | 'month' | 'custom'>('today')
const customStartDate = ref(getTodayStart())
const customEndDate = ref(Date.now())
const showDatePicker = ref(false)
const selectedDate = ref(new Date())

// 所有记录
interface TimelineRecord {
  id: string
  type: 'feeding' | 'diaper' | 'sleep'
  time: number
  icon: string
  typeName: string
  timeText: string
  detail: string
  originalRecord: FeedingRecord | DiaperRecord | SleepRecord
}

// 获取所有记录
const allRecords = computed<TimelineRecord[]>(() => {
  if (!currentBaby.value) return []

  const records: TimelineRecord[] = []

  // 喂养记录
  const feedingRecords = getFeedingRecordsByBabyId(currentBaby.value.babyId)
  feedingRecords.forEach(record => {
    let detail = ''
    if (record.detail.type === 'breast') {
      detail = `母乳喂养 ${record.detail.duration}分钟`
      if (record.detail.side === 'left') detail += ' (左侧)'
      else if (record.detail.side === 'right') detail += ' (右侧)'
      else detail += ' (双侧)'
    } else if (record.detail.type === 'bottle') {
      detail = `奶瓶喂养 ${record.detail.amount}${record.detail.unit}`
      detail += record.detail.bottleType === 'formula' ? ' (配方奶)' : ' (母乳)'
    } else {
      detail = `辅食: ${record.detail.foodName}`
    }

    records.push({
      id: record.id,
      type: 'feeding',
      time: record.time,
      icon: '🍼',
      typeName: '喂养',
      timeText: formatDate(record.time, 'HH:mm'),
      detail,
      originalRecord: record,
    })
  })

  // 排泄记录
  const diaperRecords = getDiaperRecordsByBabyId(currentBaby.value.babyId)
  diaperRecords.forEach(record => {
    let detail = ''
    if (record.type === 'wet') detail = '小便'
    else if (record.type === 'dirty') detail = '大便'
    else detail = '小便+大便'

    if (record.poopColor) detail += ` (${record.poopColor})`

    records.push({
      id: record.id,
      type: 'diaper',
      time: record.time,
      icon: '🧷',
      typeName: '换尿布',
      timeText: formatDate(record.time, 'HH:mm'),
      detail,
      originalRecord: record,
    })
  })

  // 睡眠记录
  const sleepRecords = getSleepRecordsByBabyId(currentBaby.value.babyId)
  sleepRecords.forEach(record => {
    let detail = record.type === 'nap' ? '小睡' : '夜间长睡'
    if (record.duration) {
      const hours = Math.floor(record.duration / 60)
      const minutes = record.duration % 60
      detail += ` ${hours}小时${minutes}分钟`
    } else {
      detail += ' (进行中)'
    }

    records.push({
      id: record.id,
      type: 'sleep',
      time: record.startTime,
      icon: '💤',
      typeName: '睡眠',
      timeText: formatDate(record.startTime, 'HH:mm'),
      detail,
      originalRecord: record,
    })
  })

  // 按时间倒序排序
  return records.sort((a, b) => b.time - a.time)
})

// 筛选后的记录
const filteredRecords = computed(() => {
  let startTime = 0
  let endTime = Date.now()

  if (filterType.value === 'today') {
    startTime = getTodayStart()
  } else if (filterType.value === 'week') {
    startTime = getWeekStart()
  } else if (filterType.value === 'month') {
    startTime = getMonthStart()
  } else {
    startTime = customStartDate.value
    endTime = customEndDate.value
  }

  return allRecords.value.filter(
    record => record.time >= startTime && record.time <= endTime
  )
})

// 按日期分组
const groupedRecords = computed(() => {
  const groups: { date: string; dateText: string; records: TimelineRecord[] }[] = []

  filteredRecords.value.forEach(record => {
    const date = formatDate(record.time, 'YYYY-MM-DD')
    let group = groups.find(g => g.date === date)

    if (!group) {
      let dateText = formatDate(record.time, 'MM月DD日')
      if (isToday(record.time)) {
        dateText = '今天 ' + dateText
      }

      group = { date, dateText, records: [] }
      groups.push(group)
    }

    group.records.push(record)
  })

  return groups
})

// 筛选日期
const filterDate = (type: 'today' | 'week' | 'month') => {
  filterType.value = type
}

// 日期选择确认
const handleDateConfirm = ({ selectedValue }: any) => {
  const date = new Date(selectedValue.join('-'))
  customStartDate.value = date.setHours(0, 0, 0, 0)
  customEndDate.value = date.setHours(23, 59, 59, 999)
  filterType.value = 'custom'
  showDatePicker.value = false
}

// 删除记录
const deleteRecord = (record: TimelineRecord) => {
  uni.showModal({
    title: '确认删除',
    content: '确定要删除这条记录吗?',
    success: (res) => {
      if (res.confirm) {
        let success = false

        if (record.type === 'feeding') {
          success = deleteFeedingRecord(record.id)
        } else if (record.type === 'diaper') {
          success = deleteDiaperRecord(record.id)
        } else if (record.type === 'sleep') {
          success = deleteSleepRecord(record.id)
        }

        if (success) {
          uni.showToast({
            title: '删除成功',
            icon: 'success'
          })
        }
      }
    }
  })
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
  }
})
</script>

<style lang="scss" scoped>
.timeline-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 40rpx;
}

.date-filter {
  background: white;
  padding: 20rpx;
  display: flex;
  gap: 12rpx;
  position: sticky;
  top: 0;
  z-index: 10;
}

.timeline-list {
  padding: 20rpx;
}

.empty-state {
  padding: 100rpx 0;
}

.date-group {
  margin-bottom: 40rpx;
}

.date-header {
  font-size: 28rpx;
  font-weight: bold;
  color: #666;
  padding: 20rpx 0;
  position: sticky;
  top: 100rpx;
  background: #f5f5f5;
  z-index: 5;
}

.record-item {
  position: relative;
  padding-left: 60rpx;
  margin-bottom: 20rpx;

  &:last-child .timeline-line {
    display: none;
  }
}

.timeline-dot {
  position: absolute;
  left: 10rpx;
  top: 8rpx;
  width: 24rpx;
  height: 24rpx;
  border-radius: 50%;
  border: 4rpx solid;
  background: white;
  z-index: 2;

  &.dot-feeding {
    border-color: #fa2c19;
  }

  &.dot-diaper {
    border-color: #52c41a;
  }

  &.dot-sleep {
    border-color: #1890ff;
  }
}

.timeline-line {
  position: absolute;
  left: 18rpx;
  top: 32rpx;
  bottom: -20rpx;
  width: 2rpx;
  background: #e8e8e8;
  z-index: 1;
}

.record-content {
  background: white;
  border-radius: 12rpx;
  padding: 24rpx;
}

.record-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}

.record-type {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.type-icon {
  font-size: 32rpx;
}

.type-name {
  font-size: 28rpx;
  font-weight: bold;
  color: #1a1a1a;
}

.record-time {
  font-size: 24rpx;
  color: #999;
}

.record-detail {
  font-size: 26rpx;
  color: #666;
  line-height: 1.6;
  margin-bottom: 16rpx;
}

.record-actions {
  display: flex;
  justify-content: flex-end;
}
</style>