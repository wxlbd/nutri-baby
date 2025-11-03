<template>
  <view class="feeding-page">
    <!-- 喂养记录主表单 -->
    <view class="form-wrapper">
      <!-- 喂养类型选择 -->
      <view class="form-section">
        <view class="section-title">喂养类型</view>
        <view class="radio-group-custom">
          <label
            v-for="type in feedingTypes"
            :key="type.value"
            class="radio-item"
            :class="{ active: feedingType === type.value }"
            @click="feedingType = type.value"
          >
            <view class="radio-circle"></view>
            <text>{{ type.label }}</text>
          </label>
        </view>
      </view>

      <!-- 母乳喂养 -->
      <view v-if="feedingType === 'breast'" class="form-section">
        <view class="section-title">喂养侧</view>
        <view class="radio-group-custom">
          <label
            v-for="side in breastSides"
            :key="side.value"
            class="radio-item"
            :class="{ active: breastForm.side === side.value }"
            @click="breastForm.side = side.value"
          >
            <view class="radio-circle"></view>
            <text>{{ side.label }}</text>
          </label>
        </view>

        <!-- 计时器 - 独立高亮块 -->
        <view class="timer-card">
          <view class="timer-display">
            <text class="timer-time">{{ formattedTime }}</text>
            <text class="timer-status">{{
              timerRunning ? "进行中" : "未开始"
            }}</text>
          </view>
          <wd-button
            v-if="!timerRunning"
            type="primary"
            size="large"
            block
            @click="startTimer"
          >
            开始计时
          </wd-button>
          <wd-button
            v-else
            type="success"
            size="large"
            block
            @click="stopTimer"
          >
            停止计时
          </wd-button>
        </view>
      </view>

      <!-- 奶瓶喂养 -->
      <view v-if="feedingType === 'bottle'" class="form-section">
        <view class="section-title">奶类型</view>
        <view class="radio-group-custom">
          <label
            v-for="type in bottleTypes"
            :key="type.value"
            class="radio-item"
            :class="{ active: bottleForm.bottleType === type.value }"
            @click="bottleForm.bottleType = type.value"
          >
            <view class="radio-circle"></view>
            <text>{{ type.label }}</text>
          </label>
        </view>

        <view class="form-row">
          <view class="form-group">
            <label class="form-label">单位</label>
            <view class="unit-selector">
              <label
                v-for="unit in units"
                :key="unit.value"
                class="unit-item"
                :class="{ active: bottleForm.unit === unit.value }"
                @click="bottleForm.unit = unit.value"
              >
                {{ unit.label }}
              </label>
            </view>
          </view>
          <view class="form-group">
            <label class="form-label">喂养量</label>
            <view class="input-group">
              <button
                class="input-btn"
                @click="bottleForm.amount = Math.max(0, bottleForm.amount - 10)"
              >
                −
              </button>
              <text class="input-value">{{ bottleForm.amount }}</text>
              <button
                class="input-btn"
                @click="
                  bottleForm.amount = Math.min(500, bottleForm.amount + 10)
                "
              >
                +
              </button>
            </view>
          </view>
        </view>

        <view v-if="bottleForm.amount > 0" class="form-group">
          <label class="form-label">剩余量（可选）</label>
          <view class="input-group">
            <button
              class="input-btn"
              @click="
                bottleForm.remaining = Math.max(0, bottleForm.remaining - 5)
              "
            >
              −
            </button>
            <text class="input-value">{{ bottleForm.remaining }}</text>
            <button
              class="input-btn"
              @click="
                bottleForm.remaining = Math.min(
                  bottleForm.amount,
                  bottleForm.remaining + 5
                )
              "
            >
              +
            </button>
          </view>
        </view>
      </view>

      <!-- 辅食 -->
      <view v-if="feedingType === 'food'" class="form-section">
        <view class="section-title">辅食名称</view>
        <input
          v-model="foodForm.foodName"
          type="text"
          placeholder="如：米粉、苹果泥等"
          class="text-input"
        />

        <view class="form-group" style="margin-top: 20rpx">
          <label class="form-label">备注（可选）</label>
          <textarea
            v-model="foodForm.note"
            placeholder="记录宝宝的接受程度、有无过敏反应等"
            class="textarea-input"
            maxlength="200"
          ></textarea>
        </view>
      </view>
    </view>

    <!-- 时间和提醒 -->
    <view class="form-wrapper" style="margin-top: 16rpx">
      <!-- 记录时间 -->
      <view class="form-section">
        <view class="section-title">记录时间</view>

        <!-- 日期选择器 -->
        <wd-datetime-picker
          v-model="recordDateTime"
          type="datetime"
          :min-date="minDateTime"
          :max-date="maxDateTime"
          @confirm="onDateTimeConfirm"
          @cancel="onDateTimeCancel"
        />
      </view>

      <!-- 提醒设置 -->
      <view class="form-section">
        <view class="section-title-with-toggle">
          <text>下次提醒</text>
          <view
            class="toggle-switch"
            :class="{ active: reminderEnabled }"
            @click="reminderEnabled = !reminderEnabled"
          >
            <view class="switch-slider"></view>
          </view>
        </view>

        <view v-if="reminderEnabled" class="reminder-settings">
          <view class="reminder-time">
            <text class="time-label">预计提醒时间</text>
            <text class="time-display">{{ formatNextReminderTime }}</text>
          </view>

          <view class="reminder-interval">
            <view class="interval-label">提醒间隔</view>
            <view class="interval-buttons">
              <button
                v-for="option in quickReminderOptions"
                :key="option.value"
                class="interval-btn"
                :class="{ active: reminderInterval === option.value }"
                @click="reminderInterval = option.value"
              >
                {{ option.label }}
              </button>
            </view>
            <view class="custom-interval">
              <text class="custom-label">自定义(分钟)</text>
              <view class="input-group">
                <button
                  class="input-btn"
                  @click="reminderInterval = Math.max(1, reminderInterval - 15)"
                >
                  −
                </button>
                <text class="input-value">{{ reminderInterval }}</text>
                <button
                  class="input-btn"
                  @click="
                    reminderInterval = Math.min(1440, reminderInterval + 15)
                  "
                >
                  +
                </button>
              </view>
            </view>
          </view>
        </view>

        <view v-else class="reminder-disabled">
          <text>不设置提醒</text>
        </view>
      </view>
    </view>

    <!-- 提交按钮 -->
    <view class="submit-section">
      <wd-button type="primary" size="large" block @click="handleSubmit">
        保存记录
      </wd-button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { onShow } from "@dcloudio/uni-app";
import { currentBaby, currentBabyId } from "@/store/baby";
import { getUserInfo } from "@/store/user";
import { StorageKeys, getStorage, removeStorage, setStorage } from "@/utils/storage";
import type { FeedingDetail } from "@/types";

// 直接调用 API 层
import * as feedingApi from "@/api/feeding";

// 喂养类型选项
const feedingTypes: Array<{
  label: string;
  value: "breast" | "bottle" | "food";
}> = [
  { label: "母乳喂养", value: "breast" },
  { label: "奶瓶喂养", value: "bottle" },
  { label: "辅食", value: "food" },
];

const breastSides: Array<{ label: string; value: "left" | "right" | "both" }> =
  [
    { label: "左侧", value: "left" },
    { label: "右侧", value: "right" },
    { label: "两侧", value: "both" },
  ];

const bottleTypes: Array<{ label: string; value: "formula" | "breast-milk" }> =
  [
    { label: "配方奶", value: "formula" },
    { label: "母乳/冻奶", value: "breast-milk" },
  ];

const units: Array<{ label: string; value: "ml" | "oz" }> = [
  { label: "ml", value: "ml" },
  { label: "oz", value: "oz" },
];

// 喂养类型
const feedingType = ref<"breast" | "bottle" | "food">("breast");

// 母乳喂养表单
const breastForm = ref({
  side: "left" as "left" | "right" | "both",
  leftDuration: 0,
  rightDuration: 0,
});

// 奶瓶喂养表单
const bottleForm = ref({
  bottleType: "formula" as "formula" | "breast-milk",
  amount: 60,
  unit: "ml" as "ml" | "oz",
  remaining: 0,
});

// 辅食表单
const foodForm = ref({
  foodName: "",
  note: "",
});

// ============ 计时器管理 (简单实现 + 持久化) ============
const startTime = ref(0); // 开始时间戳(毫秒)
const timerRunning = ref(false); // 计时器运行状态
const elapsedSeconds = ref(0); // 已经过的秒数
let timerInterval: number | null = null; // 定时器ID

// 临时记录数据结构
interface TempTimerRecord {
  babyId: string;
  startTime: number;
  side: "left" | "right" | "both";
}

// 格式化时间显示
const formattedTime = computed(() => {
  const minutes = Math.floor(elapsedSeconds.value / 60);
  const seconds = elapsedSeconds.value % 60;
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
});

// 保存临时记录到本地
const saveTempRecord = () => {
  if (startTime.value > 0) {
    const tempRecord: TempTimerRecord = {
      babyId: currentBabyId.value,
      startTime: startTime.value,
      side: breastForm.value.side,
    };
    console.log("[Feeding] 保存临时记录:", tempRecord);
    setStorage(StorageKeys.TEMP_BREAST_FEEDING, tempRecord);
  } else {
    console.warn("[Feeding] startTime 无效,跳过保存:", startTime.value);
  }
};

// 清除临时记录
const clearTempRecord = () => {
  removeStorage(StorageKeys.TEMP_BREAST_FEEDING);
};

// 恢复临时记录
const restoreTempRecord = () => {
  const tempRecord = getStorage<TempTimerRecord>(StorageKeys.TEMP_BREAST_FEEDING);

  console.log("[Feeding] 读取到的临时记录:", tempRecord);

  if (!tempRecord) {
    console.log("[Feeding] 没有临时记录");
    return;
  }

  // 验证数据完整性
  if (!tempRecord.startTime || !tempRecord.side || !tempRecord.babyId) {
    console.warn("[Feeding] 临时记录数据不完整,已清除:", tempRecord);
    clearTempRecord();
    return;
  }

  // 检查是否属于当前宝宝
  if (tempRecord.babyId !== currentBabyId.value) {
    console.log("[Feeding] 临时记录不属于当前宝宝,已忽略");
    return;
  }

  // 计算已经过的时长
  const now = Date.now();
  const elapsed = Math.floor((now - tempRecord.startTime) / 1000);

  console.log("[Feeding] 计算时长:", {
    now,
    startTime: tempRecord.startTime,
    diff: now - tempRecord.startTime,
    elapsed
  });

  // 验证时长是否合理
  if (isNaN(elapsed) || elapsed < 0) {
    console.warn("[Feeding] 计算出的时长无效,已清除记录");
    clearTempRecord();
    return;
  }

  // 弹窗询问用户
  uni.showModal({
    title: "检测到未完成的记录",
    content: `您有一个未完成的母乳喂养记录(${
      tempRecord.side === "left" ? "左侧" :
      tempRecord.side === "right" ? "右侧" : "两侧"
    }), 已过 ${Math.floor(elapsed / 60)} 分 ${elapsed % 60} 秒，是否继续？`,
    confirmText: "继续",
    cancelText: "重新开始",
    success: (res) => {
      if (res.confirm) {
        // 继续计时
        console.log("[Feeding] 用户选择继续计时");
        startTime.value = tempRecord.startTime;
        breastForm.value.side = tempRecord.side;
        elapsedSeconds.value = elapsed;
        timerRunning.value = true;

        // 启动定时器
        timerInterval = setInterval(() => {
          elapsedSeconds.value = Math.floor((Date.now() - startTime.value) / 1000);
          // 每10秒保存一次
          if (elapsedSeconds.value % 10 === 0) {
            saveTempRecord();
          }
        }, 1000) as unknown as number;

        console.log("[Feeding] 计时器已恢复");
      } else {
        // 重新开始
        console.log("[Feeding] 用户选择重新开始");
        clearTempRecord();
      }
    }
  });
};

// 开始计时
const startTimer = () => {
  if (timerRunning.value) {
    console.log("[Feeding] 计时器已在运行");
    return;
  }

  startTime.value = Date.now();
  timerRunning.value = true;
  elapsedSeconds.value = 0;

  // 保存临时记录
  saveTempRecord();

  // 每秒更新一次
  timerInterval = setInterval(() => {
    elapsedSeconds.value = Math.floor((Date.now() - startTime.value) / 1000);
    // 每10秒保存一次临时记录
    if (elapsedSeconds.value % 10 === 0) {
      saveTempRecord();
    }
  }, 1000) as unknown as number;

  console.log("[Feeding] 计时器已启动");
};

// 停止计时
const stopTimer = () => {
  if (!timerRunning.value) {
    console.log("[Feeding] 计时器未运行");
    return;
  }

  if (timerInterval) {
    clearInterval(timerInterval);
    timerInterval = null;
  }

  timerRunning.value = false;

  // 最后再计算一次确保准确
  if (startTime.value > 0) {
    elapsedSeconds.value = Math.floor((Date.now() - startTime.value) / 1000);
  }

  // 计算总时长并分配到左右侧,确保是有效数字
  const totalDuration = Math.max(0, elapsedSeconds.value || 0);
  if (breastForm.value.side === "both") {
    // 两侧平均分配
    breastForm.value.leftDuration = Math.floor(totalDuration / 2);
    breastForm.value.rightDuration = totalDuration - breastForm.value.leftDuration;
  } else if (breastForm.value.side === "left") {
    breastForm.value.leftDuration = totalDuration;
    breastForm.value.rightDuration = 0;
  } else {
    breastForm.value.leftDuration = 0;
    breastForm.value.rightDuration = totalDuration;
  }

  console.log("[Feeding] 计时器已停止", {
    totalDuration,
    left: breastForm.value.leftDuration,
    right: breastForm.value.rightDuration
  });
};

// 日期时间选择器
const recordDateTime = ref(new Date().getTime()); // 记录时间,初始为当前时间戳
const showDatetimePickerModal = ref(false);
const minDateTime = ref(
  new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).getTime()
); // 最小: 30天前
const maxDateTime = ref(new Date().getTime()); // 最大: 当前时间

// 提醒设置相关
const reminderEnabled = ref(true);
const reminderInterval = ref(180); // 默认3小时(分钟)

// 提醒间隔快捷选项（预设）
const quickReminderOptions = [
  { label: "1h", value: 60 },
  { label: "2h", value: 120 },
  { label: "3h", value: 180 },
  { label: "4h", value: 240 },
];

// 计算下次提醒时间显示
const formatNextReminderTime = computed(() => {
  if (!reminderEnabled.value) return "不提醒";

  const nextTime = recordDateTime.value + reminderInterval.value * 60 * 1000;
  return formatRecordTime(nextTime);
});

// 确认日期时间选择
const onDateTimeConfirm = ({ value }: { value: number }) => {
  recordDateTime.value = value;
  showDatetimePickerModal.value = false;
  console.log("[Feeding] 记录时间已更改为:", new Date(value));
};

// 取消日期时间选择
const onDateTimeCancel = () => {
  showDatetimePickerModal.value = false;
};

// 格式化记录时间显示
const formatRecordTime = (timestamp: number): string => {
  const date = new Date(timestamp);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  const hours = String(date.getHours()).padStart(2, "0");
  const minutes = String(date.getMinutes()).padStart(2, "0");
  return `${year}-${month}-${day} ${hours}:${minutes}`;
};

// 组件挂载时加载偏好和恢复临时记录
onMounted(() => {
  loadReminderPreferences();
  // 检查是否有未完成的母乳喂养记录
  if (feedingType.value === 'breast') {
    restoreTempRecord();
  }
});

// 组件卸载时清除计时器
onUnmounted(() => {
  if (timerInterval) {
    clearInterval(timerInterval);
    timerInterval = null;
  }
});

// 页面显示时加载偏好
onShow(() => {
  loadReminderPreferences();
});

// 监听喂养类型变化,加载对应的提醒偏好
watch(
  () => feedingType.value,
  () => {
    loadReminderPreferences();
    console.log(
      "[Feeding] 喂养类型已变更,提醒间隔已更新:",
      reminderInterval.value
    );
  }
);

// 加载用户喂养提醒偏好
const loadReminderPreferences = () => {
  const prefs = getStorage<any>(StorageKeys.FEEDING_REMINDER_PREFERENCES);
  if (prefs && prefs[feedingType.value]) {
    reminderInterval.value = prefs[feedingType.value];
    console.log(
      "[Feeding] 已加载用户偏好 - 喂养类型:",
      feedingType.value,
      "间隔:",
      reminderInterval.value
    );
  } else {
    // 使用默认值
    const defaults = { breast: 180, bottle: 180, food: 240 };
    reminderInterval.value =
      defaults[feedingType.value as "breast" | "bottle" | "food"] || 180;
    console.log("[Feeding] 使用默认提醒间隔:", reminderInterval.value);
  }
};

// 表单验证
const validateForm = (): boolean => {
  if (!currentBaby.value) {
    uni.showToast({
      title: "请先选择宝宝",
      icon: "none",
    });
    return false;
  }

  if (feedingType.value === "breast") {
    const totalDuration =
      breastForm.value.leftDuration + breastForm.value.rightDuration;
    console.log(
      "[Feeding] 验证母乳喂养,左侧:",
      breastForm.value.leftDuration,
      "右侧:",
      breastForm.value.rightDuration,
      "总时长:",
      totalDuration
    );
    if (totalDuration === 0) {
      uni.showToast({
        title: "请记录喂养时长",
        icon: "none",
      });
      return false;
    }
  } else if (feedingType.value === "bottle") {
    if (bottleForm.value.amount <= 0) {
      uni.showToast({
        title: "请输入喂养量",
        icon: "none",
      });
      return false;
    }
  } else if (feedingType.value === "food") {
    if (!foodForm.value.foodName.trim()) {
      uni.showToast({
        title: "请输入辅食名称",
        icon: "none",
      });
      return false;
    }
  }

  return true;
};

// 提交记录
const handleSubmit = async () => {
  // 如果还在计时中，先停止计时以获得准确的时长
  if (timerRunning.value && feedingType.value === "breast") {
    console.log("[Feeding] 保存前检测到仍在计时,自动停止计时");
    stopTimer();
  }

  if (!validateForm()) {
    return;
  }

  const user = getUserInfo();
  if (!user) {
    uni.showToast({
      title: "请先登录",
      icon: "none",
    });
    return;
  }

  let detail: FeedingDetail;

  if (feedingType.value === "breast") {
    const totalDuration =
      breastForm.value.leftDuration + breastForm.value.rightDuration;
    detail = {
      type: "breast",
      side: breastForm.value.side,
      duration: totalDuration, // 总时长(秒)
      leftDuration: breastForm.value.leftDuration, // 左侧时长(秒)
      rightDuration: breastForm.value.rightDuration, // 右侧时长(秒)
    };
  } else if (feedingType.value === "bottle") {
    detail = {
      type: "bottle",
      bottleType: bottleForm.value.bottleType,
      amount: bottleForm.value.amount,
      unit: bottleForm.value.unit,
      remaining: bottleForm.value.remaining || undefined,
    };
  } else {
    detail = {
      type: "food",
      foodName: foodForm.value.foodName,
      note: foodForm.value.note || undefined,
    };
  }

  try {
    console.log("[Feeding] 开始保存喂养记录...");

    // 直接调用 API 层创建记录
    const requestData: feedingApi.CreateFeedingRecordRequest = {
      babyId: currentBabyId.value,
      feedingType: detail.type,
      feedingTime: recordDateTime.value,
      detail: detail, // 直接使用强类型的 detail
    };

    // 根据类型填充额外字段
    if (detail.type === "breast") {
      requestData.duration = detail.duration;
    } else if (detail.type === "bottle") {
      requestData.amount = detail.amount;
    }

    // 添加提醒间隔（如果启用了提醒）
    if (reminderEnabled.value) {
      requestData.reminderInterval = reminderInterval.value;
      console.log("[Feeding] 已设置提醒间隔:", reminderInterval.value, "分钟");
    }

    // 添加实际完成时间（如果有）- 用于准确计算提醒时间
    // 对于母乳喂养,如果用户使用了计时器并停止,则记录实际完成时间
    if (feedingType.value === "breast" && startTime.value > 0) {
      const actualTime = startTime.value + (elapsedSeconds.value * 1000);
      requestData.actualCompleteTime = actualTime;
      console.log("[Feeding] 已记录实际完成时间:", actualTime);
    }

    await feedingApi.apiCreateFeedingRecord(requestData);
    console.log("[Feeding] 喂养记录保存成功");

    // 清除临时记录
    clearTempRecord();

    uni.showToast({
      title: "记录成功",
      icon: "success",
    });

    // 重置计时器
    startTime.value = 0;
    elapsedSeconds.value = 0;
    timerRunning.value = false;
    if (timerInterval) {
      clearInterval(timerInterval);
      timerInterval = null;
    }

    // 重置表单字段
    feedingType.value = "breast";
    breastForm.value = { side: "left", leftDuration: 0, rightDuration: 0 };
    bottleForm.value = {
      bottleType: "formula",
      amount: 60,
      unit: "ml",
      remaining: 0,
    };
    foodForm.value = { foodName: "", note: "" };
    recordDateTime.value = new Date().getTime();
    reminderEnabled.value = true;
    reminderInterval.value = 180;

    // 延迟返回上一页，让用户看到成功提示
    setTimeout(() => {
      uni.navigateBack({
        fail: () => {
          // navigateBack 失败时（比如在首页），跳转到首页
          console.log("[Feeding] navigateBack 失败，可能在首页，跳转到首页");
          uni.switchTab({
            url: "/pages/index/index",
          });
        },
      });
    }, 1500);
  } catch (error: any) {
    console.error("[Feeding] 保存喂养记录失败:", error);
    uni.showToast({
      title: error.message || "记录失败",
      icon: "none",
    });
  }
};
</script>

<style lang="scss" scoped>
.feeding-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 16rpx 0;
  padding-bottom: 140rpx;
}

// 表单包装器
.form-wrapper {
  background: #ffffff;
  margin: 0 16rpx;
  border-radius: 12rpx;
  overflow: hidden;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.04);
}

// 表单分组
.form-section {
  padding: 24rpx;
  border-bottom: 1rpx solid #f5f5f5;

  &:last-child {
    border-bottom: none;
  }
}

// 区域标题
.section-title {
  font-size: 28rpx;
  font-weight: 500;
  color: #262626;
  margin-bottom: 16rpx;
  display: block;
}

.section-title-with-toggle {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 28rpx;
  font-weight: 500;
  color: #262626;
  margin-bottom: 16rpx;
}

// 自定义单选框组
.radio-group-custom {
  display: flex;
  gap: 12rpx;
  flex-wrap: wrap;
}

.radio-item {
  display: flex;
  align-items: center;
  gap: 8rpx;
  padding: 12rpx 16rpx;
  border: 1rpx solid #e5e5e5;
  border-radius: 8rpx;
  background: #fafafa;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 26rpx;
  color: #666;

  &.active {
    border-color: #fa2c19;
    background: #fff5f3;
    color: #fa2c19;

    .radio-circle {
      border-color: #fa2c19;
      background: #fa2c19;
    }
  }

  &:active {
    background: #f0f0f0;
  }
}

.radio-circle {
  width: 16rpx;
  height: 16rpx;
  border: 2rpx solid #e5e5e5;
  border-radius: 50%;
  transition: all 0.2s ease;
}

// 单位选择器
.unit-selector {
  display: flex;
  gap: 12rpx;
  margin-top: 8rpx;
}

.unit-item {
  flex: 1;
  padding: 12rpx;
  text-align: center;
  border: 1rpx solid #e5e5e5;
  border-radius: 8rpx;
  background: #fafafa;
  cursor: pointer;
  font-size: 26rpx;
  color: #666;
  transition: all 0.2s ease;

  &.active {
    border-color: #fa2c19;
    background: #fff5f3;
    color: #fa2c19;
    font-weight: 500;
  }
}

// 表单行布局
.form-row {
  display: flex;
  gap: 16rpx;
  margin-top: 16rpx;

  .form-group {
    flex: 1;
  }
}

// 表单分组
.form-group {
  margin-top: 16rpx;

  &:first-child {
    margin-top: 0;
  }
}

.form-label {
  display: block;
  font-size: 26rpx;
  color: #666;
  margin-bottom: 8rpx;
}

// 输入框组（用于数字增减）
.input-group {
  display: flex;
  align-items: center;
  border: 1rpx solid #e5e5e5;
  border-radius: 8rpx;
  background: #fafafa;
  overflow: hidden;

  .input-btn {
    width: 56rpx;
    height: 56rpx;
    border: none;
    background: transparent;
    font-size: 32rpx;
    color: #fa2c19;
    cursor: pointer;
    transition: background 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0;

    &:active {
      background: rgba(0, 0, 0, 0.05);
    }
  }

  .input-value {
    flex: 1;
    text-align: center;
    font-size: 28rpx;
    color: #262626;
    font-weight: 500;
    min-width: 0;
  }
}

// 文本输入
.text-input {
  width: 100%;
  padding: 12rpx 16rpx;
  border: 1rpx solid #e5e5e5;
  border-radius: 8rpx;
  font-size: 28rpx;
  background: #fafafa;
  box-sizing: border-box;
  color: #262626;
  height: 56rpx;
  line-height: 32rpx;

  &:focus {
    border-color: #fa2c19;
    background: #ffffff;
  }
}

// 文本域
.textarea-input {
  width: 100%;
  padding: 12rpx 16rpx;
  border: 1rpx solid #e5e5e5;
  border-radius: 8rpx;
  font-size: 26rpx;
  background: #fafafa;
  box-sizing: border-box;
  color: #262626;
  min-height: 100rpx;
  font-family: inherit;

  &:focus {
    border-color: #fa2c19;
    background: #ffffff;
  }
}

// 计时器卡片
.timer-card {
  background: linear-gradient(135deg, #fff7f0 0%, #fff9f7 100%);
  border: 1rpx solid #ffe0cc;
  border-radius: 12rpx;
  padding: 28rpx;
  text-align: center;
  margin-top: 16rpx;
}

.timer-display {
  margin-bottom: 24rpx;
}

.timer-time {
  display: block;
  font-size: 80rpx;
  font-weight: bold;
  color: #fa2c19;
  margin-bottom: 8rpx;
  line-height: 1;
  letter-spacing: -2rpx;
}

.timer-status {
  display: block;
  font-size: 26rpx;
  color: #999;
}

// 时间选择器
.time-selector {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12rpx 16rpx;
  border: 1rpx solid #e5e5e5;
  border-radius: 8rpx;
  background: #fafafa;
  cursor: pointer;
  transition: all 0.2s ease;

  &:active {
    background: #f0f0f0;
  }

  .time-value {
    font-size: 28rpx;
    color: #fa2c19;
    font-weight: 500;
  }

  .time-icon {
    font-size: 32rpx;
    color: #ccc;
  }
}

// 提醒设置
.reminder-settings {
  margin-top: 16rpx;
}

.reminder-time {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12rpx 16rpx;
  background: #fafafa;
  border-radius: 8rpx;
  margin-bottom: 16rpx;

  .time-label {
    font-size: 26rpx;
    color: #666;
  }

  .time-display {
    font-size: 28rpx;
    color: #fa2c19;
    font-weight: 500;
  }
}

.reminder-interval {
  .interval-label {
    font-size: 26rpx;
    color: #666;
    margin-bottom: 8rpx;
    display: block;
  }

  .interval-buttons {
    display: flex;
    gap: 8rpx;
    margin-bottom: 16rpx;

    .interval-btn {
      flex: 1;
      padding: 10rpx 12rpx;
      border: 1rpx solid #e5e5e5;
      border-radius: 6rpx;
      background: #fafafa;
      font-size: 24rpx;
      color: #666;
      cursor: pointer;
      transition: all 0.2s ease;

      &.active {
        border-color: #fa2c19;
        background: #fff5f3;
        color: #fa2c19;
        font-weight: 500;
      }

      &:active {
        background: #f0f0f0;
      }
    }
  }
}

.custom-interval {
  display: flex;
  align-items: center;
  gap: 12rpx;

  .custom-label {
    font-size: 26rpx;
    color: #666;
    flex-shrink: 0;
  }

  .input-group {
    flex: 1;
  }
}

.reminder-disabled {
  padding: 12rpx 16rpx;
  background: #fafafa;
  border-radius: 8rpx;
  font-size: 26rpx;
  color: #999;
}

// 开关样式
.toggle-switch {
  position: relative;
  width: 52rpx;
  height: 32rpx;
  background: #e0e0e0;
  border-radius: 16rpx;
  cursor: pointer;
  transition: background 0.3s ease;

  .switch-checkbox {
    display: none;
  }

  .switch-slider {
    position: absolute;
    top: 2rpx;
    left: 2rpx;
    width: 28rpx;
    height: 28rpx;
    background: #ffffff;
    border-radius: 50%;
    transition: left 0.3s ease;
    box-shadow: 0 2rpx 4rpx rgba(0, 0, 0, 0.1);
  }

  &.active {
    background: #fa2c19;

    .switch-slider {
      left: 22rpx;
    }
  }
}

// 提交按钮区域
.submit-section {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 20rpx;
  background: #ffffff;
  box-shadow: 0 -4rpx 16rpx rgba(0, 0, 0, 0.08);
  z-index: 10;

  :deep(.nut-button) {
    height: 88rpx;
    font-size: 28rpx;
    font-weight: 500;
  }
}

// Popup 样式
:deep(.nut-popup) {
  .nut-date-picker {
    background: #ffffff;
  }
}
</style>
