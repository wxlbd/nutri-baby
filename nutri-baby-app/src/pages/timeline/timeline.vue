<template>
  <view class="timeline-page">
    <!-- 日期筛选 -->
    <view class="date-filter">
      <wd-button size="small" @click="filterDate('today')">今天</wd-button>
      <wd-button size="small" @click="filterDate('week')">本周</wd-button>
      <wd-button size="small" @click="filterDate('month')">本月</wd-button>
      <!-- 使用 Wot UI 日期选择器 -->
      <wd-datetime-picker
        v-model="selectedDateTimestamp"
        @confirm="onDateConfirm"
      >
        <wd-button size="small" type="primary"> 自定义 </wd-button>
      </wd-datetime-picker>
    </view>

    <!-- 记录列表 -->
    <view class="timeline-list">
      <view v-if="groupedRecords.length === 0" class="empty-state">
        <wd-status-tip :description="emptyDescription" />
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

            <!-- 记录内容 使用 WotUI Card -->
            <wd-card custom-class="record-card">
              <template #title>
                <view class="record-header">
                  <view class="record-type">
                    <text class="type-icon">{{ record.icon }}</text>
                    <text class="type-name">{{ record.typeName }}</text>
                  </view>
                  <text class="record-time">{{ record.timeText }}</text>
                </view>
              </template>

              <view class="record-detail">{{ record.detail }}</view>

              <template #footer>
                <view class="record-actions">
                  <wd-button
                    size="small"
                    type="default"
                    @click="deleteRecord(record)"
                  >
                    删除
                  </wd-button>
                </view>
              </template>
            </wd-card>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { isLoggedIn } from "@/store/user";
import { currentBaby } from "@/store/baby";
import {
  formatDate,
  isToday,
  getTodayStart,
  getWeekStart,
  getMonthStart,
} from "@/utils/date";
import { formatDuration } from "@/utils/common";

// 使用新的时间线聚合 API
import * as timelineApi from "@/api/timeline";
import type { TimelineItem } from "@/api/timeline";
import * as feedingApi from "@/api/feeding";
import * as diaperApi from "@/api/diaper";
import * as sleepApi from "@/api/sleep";

// 日期筛选
const filterType = ref<"today" | "week" | "month" | "custom">("today");
const customStartDate = ref(getTodayStart());
const customEndDate = ref(Date.now());

// Wot UI 日期选择器相关
const selectedDateTimestamp = ref<number[]>([]);

// 时间线数据(从聚合 API 获取)
const timelineItems = ref<TimelineItem[]>([]);
const totalRecords = ref(0);

// 展示用的记录接口
interface TimelineRecord {
  id: string;
  type: "feeding" | "diaper" | "sleep" | "growth";
  time: number;
  icon: string;
  typeName: string;
  timeText: string;
  detail: string;
  originalRecord: any;
}

// 转换时间线数据为展示格式
const allRecords = computed<TimelineRecord[]>(() => {
  if (!currentBaby.value) return [];

  const records: TimelineRecord[] = [];

  timelineItems.value.forEach((item) => {
    let icon = "";
    let typeName = "";
    let detail = "";

    if (item.recordType === "feeding") {
      const record = item.detail as feedingApi.FeedingRecordResponse;
      icon = "🍼";
      typeName = "喂养";

      if (record.feedingType === "breast") {
        detail = `母乳喂养 ${formatDuration(record.duration || 0)}`;
        const breastSide = record.detail?.breastSide;
        if (breastSide === "left") detail += " (左侧)";
        else if (breastSide === "right") detail += " (右侧)";
        else if (breastSide === "both") detail += " (双侧)";
      } else if (record.feedingType === "bottle") {
        detail = `奶瓶喂养 ${record.amount}${record.detail?.unit || "ml"}`;
        detail +=
          record.detail?.bottleType === "formula" ? " (配方奶)" : " (母乳)";
      } else {
        detail = `辅食: ${record.detail?.foodName || "未知"}`;
      }
    } else if (item.recordType === "diaper") {
      const record = item.detail as diaperApi.DiaperRecordResponse;
      icon = "🧷";
      typeName = "换尿布";

      if (record.diaperType === "pee") detail = "小便";
      else if (record.diaperType === "poo") detail = "大便";
      else detail = "小便+大便";

      if (record.pooColor) detail += ` (${record.pooColor})`;
    } else if (item.recordType === "sleep") {
      const record = item.detail as sleepApi.SleepRecordResponse;
      icon = "💤";
      typeName = "睡眠";

      const duration = record.duration || 0;
      detail = `${
        record.sleepType === "nap" ? "小睡" : "夜间睡眠"
      } ${formatDuration(duration)}`;
    } else if (item.recordType === "growth") {
      icon = "📏";
      typeName = "成长";
      const record = item.detail as any;
      const parts: string[] = [];
      if (record.height) parts.push(`身高 ${record.height}cm`);
      if (record.weight) parts.push(`体重 ${record.weight}kg`);
      if (record.headCircumference) parts.push(`头围 ${record.headCircumference}cm`);
      detail = parts.join(", ");
    }

    records.push({
      id: item.recordId,
      type: item.recordType,
      time: item.eventTime,
      icon,
      typeName,
      timeText: formatDate(item.eventTime, "HH:mm"),
      detail,
      originalRecord: item.detail,
    });
  });

  return records;
});

// 按日期分组
const groupedRecords = computed(() => {
  const groups: {
    date: string;
    dateText: string;
    records: TimelineRecord[];
  }[] = [];

  allRecords.value.forEach((record) => {
    const date = formatDate(record.time, "YYYY-MM-DD");
    let group = groups.find((g) => g.date === date);

    if (!group) {
      let dateText = formatDate(record.time, "MM月DD日");
      if (isToday(record.time)) {
        dateText = "今天 " + dateText;
      }

      group = { date, dateText, records: [] };
      groups.push(group);
    }

    group.records.push(record);
  });

  return groups;
});

// 空状态描述
const emptyDescription = computed(() => {
  return !isLoggedIn.value ? "登录后查看记录" : "暂无记录";
});

// 加载时间线记录 (使用新的聚合 API)
const loadRecords = async () => {
  if (!currentBaby.value) return;

  const babyId = currentBaby.value.babyId;

  // 计算时间范围
  let startTime = 0;
  let endTime = Date.now();

  if (filterType.value === "today") {
    startTime = getTodayStart();
  } else if (filterType.value === "week") {
    startTime = getWeekStart();
  } else if (filterType.value === "month") {
    startTime = getMonthStart();
  } else if (filterType.value === "custom") {
    startTime = customStartDate.value;
    endTime = customEndDate.value;
  }

  try {
    const response = await timelineApi.apiFetchTimeline({
      babyId,
      startTime,
      endTime,
      pageSize: 200,
    });

    timelineItems.value = response.data.items;
    totalRecords.value = response.data.total;
  } catch (error) {
    console.error("加载时间线失败:", error);
    uni.showToast({
      title: "加载数据失败",
      icon: "none",
    });
  }
};

// 页面加载
onMounted(() => {
  if (isLoggedIn.value) {
    loadRecords();
  }
});

// 筛选日期
const filterDate = (type: "today" | "week" | "month") => {
  filterType.value = type;
  loadRecords(); // 重新加载数据
};

// Wot UI 日期选择器的 confirm 事件处理
const onDateConfirm = ({ value }: { value: number[] }) => {
  console.log("[Timeline] 选择的日期时间戳范围:", value);

  if (!value || value.length === 0 || !value[0]) return;

  // value 是时间戳数组
  const timestamp = value[0];
  const endTimestamp = value[1] || timestamp;
  // 更新时间戳
  selectedDateTimestamp.value = value;

  // 设置当天的起止时间
  customStartDate.value = new Date(timestamp).setHours(0, 0, 0, 0);
  customEndDate.value = new Date(endTimestamp).setHours(23, 59, 59, 999);
  filterType.value = "custom";

  // 重新加载数据
  loadRecords();
};

// 删除记录
const deleteRecord = async (record: TimelineRecord) => {
  uni.showModal({
    title: "确认删除",
    content: "确定要删除这条记录吗?",
    success: async (res) => {
      if (res.confirm) {
        try {
          if (record.type === "feeding") {
            await feedingApi.apiDeleteFeedingRecord(record.id);
          } else if (record.type === "diaper") {
            await diaperApi.apiDeleteDiaperRecord(record.id);
          } else if (record.type === "sleep") {
            await sleepApi.apiDeleteSleepRecord(record.id);
          }

          uni.showToast({
            title: "删除成功",
            icon: "success",
          });

          // 重新加载记录
          await loadRecords();
        } catch (error: any) {
          uni.showToast({
            title: error.message || "删除失败",
            icon: "none",
          });
        }
      }
    },
  });
};
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

// WotUI Card 组件自定义样式
:deep(.record-card) {
  border-radius: 12rpx;
}

.record-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
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
}

.record-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 16rpx;
}
</style>
