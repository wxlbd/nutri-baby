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
              :style="{ width: completionStats.completionRate + '%' }"
            ></view>
          </view>
          <text class="progress-text">
            {{ completionStats.completed + completionStats.skipped }} / {{ completionStats.total }} ({{
              completionStats.completionRate
            }}%)
          </text>
        </view>
      </view>
    </view>

    <!-- 即将到期提醒 (基于待接种的日程计算) -->
    <view
      v-if="upcomingSchedules && upcomingSchedules.length > 0"
      class="reminders-section"
    >
      <view class="section-title"
        >⏰ 近期待接种 ({{ upcomingSchedules.length }})</view
      >

      <view class="reminder-list">
        <view
          v-for="schedule in upcomingSchedules"
          :key="schedule.scheduleId"
          class="reminder-item"
        >
          <view class="reminder-content">
            <view class="vaccine-name">
              {{ schedule.vaccineName }} (第{{ schedule.doseNumber }}针)
            </view>
            <view class="vaccine-date">
              建议月龄: {{ schedule.ageInMonths }}个月
            </view>
          </view>
          <view class="reminder-action">
            <wd-button
              size="small"
              type="primary"
              @click.stop="handleRecordVaccine(schedule)"
            >
              记录接种
            </wd-button>
          </view>
        </view>
      </view>
    </view>

    <!-- 疫苗计划列表 -->
    <view class="plan-section">
      <view class="section-header">
        <text class="section-title">📋 疫苗日程</text>
        <wd-button size="small" @click="goToManage"> 管理计划 </wd-button>
      </view>

      <wd-tabs v-model="activeTab">
        <wd-tab title="全部" pane-key="all" />
        <wd-tab title="已完成" pane-key="completed" />
        <wd-tab title="未完成" pane-key="pending" />
      </wd-tabs>

      <view class="plan-list">
        <view
          v-for="schedule in filteredSchedules"
          :key="schedule.scheduleId"
          class="plan-item"
          :class="{ completed: schedule.vaccinationStatus === 'completed' }"
        >
          <view class="plan-header">
            <view class="plan-name">
              <text class="required-badge" v-if="schedule.isRequired"
                >必打</text
              >
              {{ schedule.vaccineName }}
            </view>
            <text class="plan-age">{{ schedule.ageInMonths }}个月</text>
          </view>

          <view class="plan-detail">
            <text class="plan-dose">第{{ schedule.doseNumber }}针</text>
            <text v-if="schedule.description" class="plan-desc">{{
              schedule.description
            }}</text>
          </view>

          <view
            v-if="schedule.vaccinationStatus === 'completed'"
            class="plan-record"
          >
            <text class="completed-icon">✓</text>
            <text class="completed-text">已接种</text>
            <text class="completed-date">
              {{ formatDate(schedule.vaccineDate || 0, "YYYY-MM-DD") }}
            </text>
            <text v-if="schedule.hospital" class="hospital-info">
              {{ schedule.hospital }}
            </text>
          </view>

          <view
            v-else-if="schedule.vaccinationStatus === 'skipped'"
            class="plan-record"
          >
            <text class="skipped-icon">⊘</text>
            <text class="skipped-text">已跳过</text>
          </view>

          <view v-else class="plan-action">
            <wd-button
              size="small"
              type="primary"
              @click="handleRecordBySchedule(schedule)"
            >
              记录接种
            </wd-button>
            <wd-button
              size="small"
              type="info"
              @click="handleSkipSchedule(schedule)"
            >
              跳过
            </wd-button>
          </view>
        </view>
      </view>
    </view>

    <!-- 接种记录对话框 -->
    <wd-popup
      v-model="showRecordDialog"
      position="bottom"
      custom-style="height: 75%"
      round
      closeable
    >
      <wd-cell-group title="记录疫苗接种" border>
        <wd-input
          v-model="recordForm.vaccineName"
          placeholder="疫苗名称"
          readonly
          label="疫苗名称"
        />
        <wd-datetime-picker
          v-model="recordForm.vaccineDate"
          type="date"
          label="接种日期"
        />
        <wd-input
          v-model="recordForm.hospital"
          placeholder="接种医院"
          label="接种医院*"
          required
        />
        <wd-input
          v-model="recordForm.batchNumber"
          placeholder="疫苗批号"
          label="疫苗批号"
        />
        <wd-input
          v-model="recordForm.doctor"
          placeholder="接种医生"
          label="接种医生"
        />
        <wd-textarea
          v-model="recordForm.reaction"
          placeholder="不良反应"
          label="不良反应"
          auto-height
        />
        <wd-textarea
          v-model="recordForm.note"
          placeholder="备注"
          label="备注"
          auto-height
        />
      </wd-cell-group>
      <view class="dialog-footer">
        <wd-button type="primary" size="large" @click="handleSaveRecord">
          保存
        </wd-button>
        <wd-button type="info" size="large" @click="showRecordDialog = false">
          取消
        </wd-button>
      </view>
    </wd-popup>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { currentBaby } from "@/store/baby";
import { userInfo } from "@/store/user";
import { formatDate } from "@/utils/date";
import { shouldShowGuide } from "@/store/subscribe";

// 直接调用 API 层 (使用新架构)
import * as vaccineApi from "@/api/vaccine";

// Tab状态
const activeTab = ref("all");

// 对话框状态
const showRecordDialog = ref(false);

// 订阅消息引导状态
const showVaccineGuide = ref(false);

// 疫苗日程数据 (新架构 - 合并计划和记录)
const vaccineSchedules = ref<vaccineApi.VaccineScheduleResponse[]>([]);
const vaccineStats = ref<vaccineApi.VaccineScheduleListResponse["statistics"]>(
  {
    total: 0,
    completed: 0,
    pending: 0,
    skipped: 0,
    completionRate: 0,
  }
);

// 表单数据 (新架构)
const recordForm = ref({
  scheduleId: "",
  vaccineName: "",
  vaccineDate: Date.now(),
  hospital: "",
  batchNumber: "",
  doctor: "",
  reaction: "",
  note: "",
});

// 加载疫苗数据 (新架构)
const loadVaccineData = async () => {
  if (!currentBaby.value) return;

  const babyId = currentBaby.value.babyId;
  console.log("加载疫苗日程数据", babyId);
  try {
    const data = await vaccineApi.apiFetchVaccineSchedules(babyId);

    vaccineSchedules.value = data.schedules || [];
    vaccineStats.value = data.statistics || {
      total: 0,
      completed: 0,
      pending: 0,
      skipped: 0,
      completionRate: 0,
    };

    console.log("疫苗数据加载成功", {
      schedules: vaccineSchedules.value.length,
      stats: vaccineStats.value,
    });
  } catch (error) {
    console.error("加载疫苗数据失败:", error);
    // 确保即使出错也初始化为空数组
    vaccineSchedules.value = [];
    vaccineStats.value = {
      total: 0,
      completed: 0,
      pending: 0,
      skipped: 0,
      completionRate: 0,
    };

    uni.showToast({
      title: "加载数据失败",
      icon: "none",
    });
  }
};

// 完成度统计 - 直接使用后端返回的数据
const completionStats = computed(() => {
  if (!currentBaby.value) {
    return { total: 0, completed: 0, pending: 0, skipped: 0, completionRate: 0 };
  }

  return vaccineStats.value;
});

// 近期待接种的日程 (pending状态，按月龄排序，取前3个)
const upcomingSchedules = computed(() => {
  if (!currentBaby.value || !vaccineSchedules.value) {
    return [];
  }

  return vaccineSchedules.value
    .filter((s) => s.vaccinationStatus === "pending")
    .sort((a, b) => a.ageInMonths - b.ageInMonths)
    .slice(0, 3); // 只显示前3个
});

// 过滤后的日程列表
const filteredSchedules = computed(() => {
  let schedules = vaccineSchedules.value || [];

  if (!Array.isArray(schedules)) {
    return [];
  }

  if (activeTab.value === "completed") {
    schedules = schedules.filter((s) => s.vaccinationStatus === "completed");
  } else if (activeTab.value === "pending") {
    schedules = schedules.filter((s) => s.vaccinationStatus === "pending");
  }

  return schedules.sort((a, b) => a.ageInMonths - b.ageInMonths);
});

// 处理记录接种(通过日程)
const handleRecordVaccine = (schedule: vaccineApi.VaccineScheduleResponse) => {
  recordForm.value = {
    scheduleId: schedule.scheduleId,
    vaccineName: schedule.vaccineName,
    vaccineDate: Date.now(),
    hospital: "",
    batchNumber: "",
    doctor: "",
    reaction: "",
    note: "",
  };
  showRecordDialog.value = true;
};

// 处理记录接种(通过日程 - 别名方法)
const handleRecordBySchedule = (
  schedule: vaccineApi.VaccineScheduleResponse
) => {
  handleRecordVaccine(schedule);
};

// 处理跳过接种
const handleSkipSchedule = async (
  schedule: vaccineApi.VaccineScheduleResponse
) => {
  if (!currentBaby.value) return;

  uni.showModal({
    title: "跳过接种",
    content: `确定要跳过「${schedule.vaccineName}」吗？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await vaccineApi.apiUpdateVaccineSchedule(
            currentBaby.value!.babyId,
            schedule.scheduleId,
            {
              vaccinationStatus: "skipped",
            }
          );

          uni.showToast({
            title: "已标记为跳过",
            icon: "success",
          });

          // 重新加载数据
          await loadVaccineData();
        } catch (error: any) {
          uni.showToast({
            title: error.message || "操作失败",
            icon: "none",
          });
        }
      }
    },
  });
};

// 保存接种记录 (新架构)
const handleSaveRecord = async () => {
  console.log("handleSaveRecord", recordForm.value);
  if (!currentBaby.value || !userInfo.value) {
    uni.showToast({
      title: "请先登录",
      icon: "none",
    });
    return;
  }

  if (!recordForm.value.hospital.trim()) {
    uni.showToast({
      title: "请输入接种医院",
      icon: "none",
    });
    return;
  }

  // 保存前记录当前完成数
  const completedBefore = vaccineStats.value.completed;

  try {
    await vaccineApi.apiUpdateVaccineSchedule(
      currentBaby.value.babyId,
      recordForm.value.scheduleId,
      {
        vaccinationStatus: "completed",
        vaccineDate: recordForm.value.vaccineDate,
        hospital: recordForm.value.hospital.trim(),
        batchNumber: recordForm.value.batchNumber.trim() || undefined,
        doctor: recordForm.value.doctor.trim() || undefined,
        reaction: recordForm.value.reaction.trim() || undefined,
        note: recordForm.value.note.trim() || undefined,
      }
    );

    uni.showToast({
      title: "记录成功",
      icon: "success",
    });

    showRecordDialog.value = false;

    // 重新加载数据
    await loadVaccineData();

    // 检查是否是首次添加疫苗记录
    const isFirstRecord = completedBefore === 0;

    // 首次记录后,延迟显示订阅引导
    if (isFirstRecord && shouldShowGuide("vaccine_reminder")) {
      setTimeout(() => {
        showVaccineGuide.value = true;
      }, 1500); // 延迟1.5秒,让用户看到成功提示
    }
  } catch (error: any) {
    uni.showToast({
      title: error.message || "保存失败",
      icon: "none",
    });
  }
};

// 处理订阅消息结果
const handleSubscribeResult = (result: "accept" | "reject") => {
  if (result === "accept") {
    console.log("用户同意订阅疫苗提醒");
  }
};

// 页面加载
onMounted(async () => {
  if (!currentBaby.value) {
    uni.showToast({
      title: "请先选择宝宝",
      icon: "none",
    });
    setTimeout(() => {
      uni.navigateBack();
    }, 1500);
    return;
  }

  // 加载疫苗数据
  await loadVaccineData();
});

// 跳转到疫苗计划管理页面
const goToManage = () => {
  uni.navigateTo({
    url: "/pages/vaccine/manage/manage",
  });
};
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

.reminders-section,
.plan-section {
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

.reminder-list,
.plan-list {
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

.plan-record,
.plan-action {
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

.hospital-info {
  font-size: 22rpx;
  color: #999;
  margin-left: 8rpx;
}

.skipped-icon {
  font-size: 32rpx;
  color: #999;
}

.skipped-text {
  font-size: 26rpx;
  color: #999;
  font-weight: bold;
}

.dialog-footer {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  padding: 20rpx 30rpx 30rpx 30rpx;
  background: #fff;
  border-top: 1rpx solid #f0f0f0;
  box-shadow: 0 -2rpx 8rpx rgba(0, 0, 0, 0.05);
}

.dialog-footer .wd-button {
  width: 100%;
}
</style>
