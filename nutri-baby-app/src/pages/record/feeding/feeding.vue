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
import { padZero } from "@/utils/common";
import {
  StorageKeys,
  getStorage,
  setStorage,
  removeStorage,
} from "@/utils/storage";
import type { FeedingDetail } from "@/types";

// 直接调用 API 层
import * as feedingApi from "@/api/feeding";

// 临时喂养记录类型
interface TempBreastFeeding {
  babyId: string;
  side: "left" | "right" | "both";
  startTime: number; // 开始时间戳(毫秒)
  feedingType: "breast";
}

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

// 计时器相关
const timerRunning = ref(false);
const startTime = ref(0); // 开始时间戳 (毫秒)
const frozenTime = ref(0); // 停止后冻结的时间（秒）- 用于保留停止时的显示
const timerTrigger = ref(0); // 用于触发视图更新的虚拟响应式值
const tempRecordCheckDone = ref(false); // 防止重复检测临时记录
let timerInterval: number | null = null;

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

// 格式化时间显示 - 基于开始时间戳计算
const formattedTime = computed(() => {
  // 依赖 timerTrigger 以触发定期更新
  timerTrigger.value; // 访问此值以建立依赖关系

  // 如果停止了计时，使用冻结的时间（保持停止时的值）
  if (!timerRunning.value && frozenTime.value > 0) {
    const minutes = Math.floor(frozenTime.value / 60);
    const seconds = frozenTime.value % 60;
    return `${padZero(minutes)}:${padZero(seconds)}`;
  }

  // 如果没有开始计时，返回 00:00
  if (startTime.value === 0) {
    return "00:00";
  }

  // 计算正在运行的时间
  const elapsedSeconds = Math.floor((Date.now() - startTime.value) / 1000);
  const minutes = Math.floor(elapsedSeconds / 60);
  const seconds = elapsedSeconds % 60;
  return `${padZero(minutes)}:${padZero(seconds)}`;
});

// 保存临时记录到本地
const saveTempRecord = () => {
  const tempRecord: TempBreastFeeding = {
    babyId: currentBabyId.value,
    side: breastForm.value.side,
    startTime: startTime.value,
    feedingType: "breast",
  };
  setStorage(StorageKeys.TEMP_BREAST_FEEDING, tempRecord);
  console.log("[Feeding] 临时记录已保存:", tempRecord);
};

// 清除临时记录
const clearTempRecord = () => {
  removeStorage(StorageKeys.TEMP_BREAST_FEEDING);
  tempRecordCheckDone.value = false;
  console.log("[Feeding] 临时记录已清除");
};

// 恢复临时记录
const restoreTempRecord = (tempRecord: TempBreastFeeding) => {
  breastForm.value.side = tempRecord.side;
  startTime.value = tempRecord.startTime;
  timerRunning.value = true;

  // 启动定时器更新显示和临时记录
  timerInterval = setInterval(() => {
    // 每秒改变 timerTrigger 以触发计算属性重新计算
    timerTrigger.value++;
    // 持续更新临时记录
    saveTempRecord();
  }, 1000) as unknown as number;

  console.log(
    "[Feeding] 临时记录已恢复, 已过时长:",
    Math.floor((Date.now() - tempRecord.startTime) / 1000),
    "秒"
  );
};

// 开始计时
const startTimer = () => {
  startTime.value = Date.now();
  timerRunning.value = true;

  // 保存临时记录
  saveTempRecord();

  // 启动定时器以每秒更新视图和临时记录
  timerInterval = setInterval(() => {
    // 每秒改变 timerTrigger 以触发计算属性重新计算
    timerTrigger.value++;
    // 每秒更新一次临时记录，确保记录的时间戳总是最新的
    // （这样用户切页面再回来时，弹窗显示的"已过时长"是最新的）
    saveTempRecord();
  }, 1000) as unknown as number;

  console.log("[Feeding] 开始计时");
};

// 停止计时
const stopTimer = () => {
  if (timerInterval) {
    clearInterval(timerInterval);
    timerInterval = null;
  }

  // 计算总时长(秒)
  const duration = Math.floor((Date.now() - startTime.value) / 1000);

  // 保存冻结的时间，这样停止后显示不会改变
  frozenTime.value = duration;

  // 停止计时运行状态
  timerRunning.value = false;

  console.log("[Feeding] 停止计时,总时长:", duration, "秒");

  if (breastForm.value.side === "both") {
    // 两侧时平均分配
    breastForm.value.leftDuration = Math.floor(duration / 2);
    breastForm.value.rightDuration = duration - breastForm.value.leftDuration;
  } else {
    // 单侧时全部计入
    if (breastForm.value.side === "left") {
      breastForm.value.leftDuration = duration;
      breastForm.value.rightDuration = 0;
    } else {
      breastForm.value.leftDuration = 0;
      breastForm.value.rightDuration = duration;
    }
  }

  console.log(
    "[Feeding] 喂养侧:",
    breastForm.value.side,
    "左侧:",
    breastForm.value.leftDuration,
    "右侧:",
    breastForm.value.rightDuration
  );

  // ✨ 注意：保留 startTime 和 frozenTime 用于显示，在提交成功后清除
};

// 组件卸载时清除计时器
onUnmounted(() => {
  if (timerInterval) {
    clearInterval(timerInterval);
  }
});

// 页面加载时检测临时记录
onMounted(() => {
  loadReminderPreferences();
  checkTempRecord();
});

// 页面显示时不再重复检测临时记录（避免多次弹窗）
// 如果需要在返回时重新开始新的记录，用户可以手动刷新或清除临时记录
onShow(() => {
  // 仅加载提醒偏好，不再检测临时记录
  loadReminderPreferences();
});

// 监听喂养侧变化,如果正在计时则更新临时记录
watch(
  () => breastForm.value.side,
  () => {
    if (timerRunning.value && startTime.value > 0) {
      saveTempRecord();
      console.log("[Feeding] 喂养侧已更改,临时记录已更新");
    }
  }
);

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

// 检测并处理临时记录
const checkTempRecord = () => {
  // 如果已经在计时,不重复检测
  if (timerRunning.value) {
    return;
  }

  // 如果已经检测过本次，不再重复检测（防止 onMounted 和 onShow 重复调用）
  if (tempRecordCheckDone.value) {
    return;
  }

  const tempRecord = getStorage<TempBreastFeeding>(
    StorageKeys.TEMP_BREAST_FEEDING
  );

  if (!tempRecord) {
    tempRecordCheckDone.value = true; // 标记已检测
    return;
  }

  // 检查临时记录是否属于当前宝宝
  if (tempRecord.babyId !== currentBabyId.value) {
    console.log("[Feeding] 临时记录不属于当前宝宝,已忽略");
    tempRecordCheckDone.value = true; // 标记已检测
    return;
  }

  // 标记已检测（在显示弹窗前）
  tempRecordCheckDone.value = true;

  // 计算已过时长
  const elapsedSeconds = Math.floor((Date.now() - tempRecord.startTime) / 1000);
  const minutes = Math.floor(elapsedSeconds / 60);
  const seconds = elapsedSeconds % 60;

  console.log("[Feeding] 检测到临时记录,已过时长:", `${minutes}分${seconds}秒`);

  // 弹窗询问用户
  uni.showModal({
    title: "未完成的喂养记录",
    content: `检测到您之前有一次未完成的母乳喂养记录(${
      tempRecord.side === "left"
        ? "左侧"
        : tempRecord.side === "right"
        ? "右侧"
        : "两侧"
    }),已过 ${minutes} 分钟 ${seconds} 秒,是否继续?`,
    confirmText: "继续",
    cancelText: "重新开始",
    success: (res) => {
      if (res.confirm) {
        // 用户选择继续
        console.log("[Feeding] 用户选择继续临时记录");
        // 切换到母乳喂养标签
        feedingType.value = "breast";
        // 恢复临时记录
        restoreTempRecord(tempRecord);
      } else {
        // 用户选择重新开始
        console.log("[Feeding] 用户选择重新开始,清除临时记录");
        clearTempRecord();
      }
    },
  });
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
      detail: {},
    };

    // 根据类型填充 detail
    if (detail.type === "breast") {
      requestData.duration = detail.duration;
      requestData.detail = {
        breastSide: detail.side,
        leftTime: detail.leftDuration,
        rightTime: detail.rightDuration,
        duration: detail.duration,
      };
    } else if (detail.type === "bottle") {
      requestData.amount = detail.amount;
      requestData.detail = {
        bottleType: detail.bottleType,
        unit: detail.unit,
        remaining: detail.remaining,
      };
    } else {
      // food
      requestData.detail = {
        foodName: detail.foodName,
        note: detail.note,
      };
    }

    // 添加提醒间隔（如果启用了提醒）
    if (reminderEnabled.value) {
      requestData.reminderInterval = reminderInterval.value;
      console.log("[Feeding] 已设置提醒间隔:", reminderInterval.value, "分钟");
    }

    await feedingApi.apiCreateFeedingRecord(requestData);
    console.log("[Feeding] 喂养记录保存成功");

    // 保存成功后清除临时记录 (如果是母乳喂养)
    if (feedingType.value === "breast") {
      clearTempRecord();
    }

    uni.showToast({
      title: "记录成功",
      icon: "success",
    });

    // 提交成功后清除计时器显示和临时数据
    startTime.value = 0;
    frozenTime.value = 0;
    timerTrigger.value = 0;

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
