<template>
    <view class="sleep-page">
        <!-- 当前状态 -->
        <view class="status-card">
            <view v-if="ongoingRecord" class="sleeping">
                <view class="status-icon">💤</view>
                <view class="status-text">宝宝正在睡觉</view>
                <view class="sleep-duration">
                    <text class="duration">{{ formattedTime }}</text>
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

        <!-- 快速补记睡眠 -->
        <view v-if="!ongoingRecord" class="quick-record-section">
            <view class="section-title">快速补记睡眠</view>
            <nut-button
                type="info"
                size="large"
                block
                @click="showQuickRecordModal = true"
            >
                <view class="button-content">
                    <text class="icon">⏰</text>
                    <text>补记历史睡眠</text>
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
                        <text class="duration-text">{{
                            formatDuration(lastRecord.duration)
                        }}</text>
                    </template>
                </nut-cell>
            </nut-cell-group>
        </view>
    </view>

    <!-- 快速补记睡眠对话框 -->
    <nut-dialog
        v-model:visible="showQuickRecordModal"
        title="补记睡眠"
        @confirm="handleQuickSleepConfirm"
        @cancel="showQuickRecordModal = false"
    >
        <view class="quick-record-form">
            <!-- 睡眠类型 -->
            <view class="form-item">
                <view class="form-label">睡眠类型</view>
                <nut-radio-group v-model="quickRecord.type" direction="horizontal">
                    <nut-radio label="nap">小睡</nut-radio>
                    <nut-radio label="night">夜间长睡</nut-radio>
                </nut-radio-group>
            </view>

            <!-- 开始时间 -->
            <view class="form-item">
                <view class="form-label">开始时间</view>
                <view class="time-input" @click="showStartTimePicker = true">
                    {{ formatQuickTime(quickRecord.startTime) }}
                </view>
            </view>

            <!-- 结束时间 -->
            <view class="form-item">
                <view class="form-label">结束时间</view>
                <view class="time-input" @click="showEndTimePicker = true">
                    {{ formatQuickTime(quickRecord.endTime) }}
                </view>
            </view>
        </view>
    </nut-dialog>

    <!-- 开始时间选择器 -->
    <nut-date-picker
        v-model="quickRecord.startTime"
        type="datetime"
        :min-date="minDateTime"
        :max-date="quickRecord.endTime"
        @confirm="onStartTimeConfirm"
        @cancel="showStartTimePicker = false"
        :visible="showStartTimePicker"
    ></nut-date-picker>

    <!-- 结束时间选择器 -->
    <nut-date-picker
        v-model="quickRecord.endTime"
        type="datetime"
        :min-date="quickRecord.startTime"
        :max-date="maxDateTime"
        @confirm="onEndTimeConfirm"
        @cancel="showEndTimePicker = false"
        :visible="showEndTimePicker"
    ></nut-date-picker>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from "vue";
import { onShow } from "@dcloudio/uni-app";
import { currentBabyId, currentBaby } from "@/store/baby";
import { getUserInfo } from "@/store/user";
import { formatDate, formatDuration } from "@/utils/date";
import { padZero } from "@/utils/common";
import {
    StorageKeys,
    getStorage,
    setStorage,
    removeStorage,
} from "@/utils/storage";
import type { SleepRecord } from "@/types";

// 直接调用 API 层
import * as sleepApi from "@/api/sleep";

// 临时睡眠记录类型
interface TempSleepRecording {
    babyId: string;
    type: "nap" | "night";
    startTime: number; // 开始时间戳(毫秒)
}

// 睡眠类型
const sleepType = ref<"nap" | "night">("nap");

// 进行中的睡眠记录
const ongoingRecord = ref<SleepRecord | null>(null);

// 最后一次睡眠记录
const lastRecord = ref<SleepRecord | null>(null);

// 快速补记相关
const showQuickRecordModal = ref(false);
const showStartTimePicker = ref(false);
const showEndTimePicker = ref(false);
const minDateTime = ref<Date>(new Date(Date.now() - 30 * 24 * 60 * 60 * 1000)); // 30天前
const maxDateTime = ref<Date>(new Date());

const quickRecord = ref({
    type: 'nap' as 'nap' | 'night',
    startTime: new Date(Date.now() - 2 * 60 * 60 * 1000), // 默认2小时前
    endTime: new Date()
});

// 快速补记时间格式化
const formatQuickTime = (date: Date): string => {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    return `${year}-${month}-${day} ${hours}:${minutes}`;
};

// 开始时间确认
const onStartTimeConfirm = (value: Date) => {
    quickRecord.value.startTime = value;
    showStartTimePicker.value = false;
};

// 结束时间确认
const onEndTimeConfirm = (value: Date) => {
    quickRecord.value.endTime = value;
    showEndTimePicker.value = false;
};

// 处理快速补记睡眠
const handleQuickSleepConfirm = async () => {
    const user = getUserInfo();
    if (!user) {
        uni.showToast({
            title: "请先登录",
            icon: "none",
        });
        return;
    }

    if (!currentBaby.value) {
        uni.showToast({
            title: "请先选择宝宝",
            icon: "none",
        });
        return;
    }

    // 验证时间
    if (quickRecord.value.startTime >= quickRecord.value.endTime) {
        uni.showToast({
            title: "开始时间必须早于结束时间",
            icon: "none",
        });
        return;
    }

    try {
        const elapsedSeconds = Math.floor((quickRecord.value.endTime.getTime() - quickRecord.value.startTime.getTime()) / 1000);

        await sleepApi.apiCreateSleepRecord({
            babyId: currentBabyId.value,
            sleepType: quickRecord.value.type,
            startTime: quickRecord.value.startTime.getTime(),
            endTime: quickRecord.value.endTime.getTime(),
            duration: elapsedSeconds,
        });

        uni.showToast({
            title: "保存成功",
            icon: "success",
        });

        showQuickRecordModal.value = false;

        // 重置表单
        quickRecord.value = {
            type: 'nap',
            startTime: new Date(Date.now() - 2 * 60 * 60 * 1000),
            endTime: new Date()
        };

        setTimeout(() => {
            uni.navigateBack();
        }, 1000);
    } catch (error: any) {
        console.error("[Sleep] 保存快速补记睡眠失败:", error);
        uni.showToast({
            title: error.message || "保存失败",
            icon: "none",
        });
    }
};

// 定时器相关
const timerRunning = ref(false);
const startTime = ref(0); // 开始时间戳 (毫秒)
const timerTrigger = ref(0); // 用于触发视图更新的虚拟响应式值
const tempRecordCheckDone = ref(false); // 防止重复检测临时记录
let timerInterval: number | null = null;

// 格式化时间显示 - 基于开始时间戳计算
const formattedTime = computed(() => {
    // 依赖 timerTrigger 以触发定期更新
    timerTrigger.value; // 访问此值以建立依赖关系

    if (!timerRunning.value || startTime.value === 0) {
        return "00:00:00";
    }

    const elapsedSeconds = Math.floor((Date.now() - startTime.value) / 1000);
    const hours = Math.floor(elapsedSeconds / 3600);
    const minutes = Math.floor((elapsedSeconds % 3600) / 60);
    const seconds = elapsedSeconds % 60;

    return `${padZero(hours)}:${padZero(minutes)}:${padZero(seconds)}`;
});

// 保存临时睡眠记录到本地
const saveTempRecord = () => {
    const tempRecord: TempSleepRecording = {
        babyId: currentBabyId.value,
        type: sleepType.value,
        startTime: startTime.value,
    };
    setStorage(StorageKeys.TEMP_SLEEP_RECORDING, tempRecord);
    console.log("[Sleep] 临时记录已保存:", tempRecord);
};

// 清除临时睡眠记录
const clearTempRecord = () => {
    removeStorage(StorageKeys.TEMP_SLEEP_RECORDING);
    tempRecordCheckDone.value = false; // 重置标志，允许下次检测
    console.log("[Sleep] 临时记录已清除");
};

// 恢复临时睡眠记录
const restoreTempRecord = (tempRecord: TempSleepRecording) => {
    const user = getUserInfo();
    if (!user) {
        return;
    }

    sleepType.value = tempRecord.type;
    startTime.value = tempRecord.startTime;
    timerRunning.value = true;

    // 创建本地睡眠记录对象以显示计时器
    ongoingRecord.value = {
        id: `temp_${Date.now()}`, // 临时ID
        babyId: tempRecord.babyId,
        startTime: tempRecord.startTime,
        type: tempRecord.type,
        createBy: user.openid,
        createByName: user.nickName,
        createByAvatar: user.avatarUrl,
        createTime: Date.now(),
    };

    // 启动定时器更新显示
    timerInterval = setInterval(() => {
        // 每秒改变 timerTrigger 以触发计算属性重新计算
        timerTrigger.value++;
    }, 1000) as unknown as number;

    console.log(
        "[Sleep] 临时记录已恢复, 已过时长:",
        Math.floor((Date.now() - tempRecord.startTime) / 1000),
        "秒",
    );
};

// 开始睡觉
const startSleep = async () => {
    const user = getUserInfo();
    if (!user) {
        uni.showToast({
            title: "请先登录",
            icon: "none",
        });
        return;
    }

    if (!currentBaby.value) {
        uni.showToast({
            title: "请先选择宝宝",
            icon: "none",
        });
        return;
    }

    try {
        // 使用本地时间戳开始计时
        startTime.value = Date.now();
        timerRunning.value = true;

        // 创建本地睡眠记录对象以显示计时器
        ongoingRecord.value = {
            id: `temp_${Date.now()}`, // 临时ID
            babyId: currentBabyId.value,
            startTime: startTime.value,
            type: sleepType.value,
            createBy: user.openid,
            createByName: user.nickName,
            createByAvatar: user.avatarUrl,
            createTime: Date.now(),
        };

        // 保存临时记录到本地存储
        saveTempRecord();

        // 启动定时器以每秒更新视图
        timerInterval = setInterval(() => {
            // 每秒改变 timerTrigger 以触发计算属性重新计算
            timerTrigger.value++;
        }, 1000) as unknown as number;

        uni.showToast({
            title: "开始记录睡眠",
            icon: "success",
        });

        console.log("[Sleep] 开始计时");
    } catch (error: any) {
        uni.showToast({
            title: error.message || "开始失败",
            icon: "none",
        });
    }
};

// 结束睡觉
const endSleep = async () => {
    if (!timerRunning.value || startTime.value === 0) {
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

    try {
        // 停止计时器
        if (timerInterval) {
            clearInterval(timerInterval);
            timerInterval = null;
        }

        timerRunning.value = false;

        // 计算总时长(秒)
        const elapsedSeconds = Math.floor(
            (Date.now() - startTime.value) / 1000,
        );

        console.log("[Sleep] 停止计时,总时长:", elapsedSeconds, "秒");

        // 调用 API 创建睡眠记录
        await sleepApi.apiCreateSleepRecord({
            babyId: currentBabyId.value,
            sleepType: sleepType.value,
            startTime: startTime.value,
            endTime: Date.now(),
            duration: elapsedSeconds, // 添加时长字段
        });

        console.log("[Sleep] 睡眠记录保存成功");

        // 清除临时记录和进行中的记录
        clearTempRecord();
        ongoingRecord.value = null;

        uni.showToast({
            title: "保存成功",
            icon: "success",
        });

        setTimeout(() => {
            uni.navigateBack();
        }, 1000);
    } catch (error: any) {
        console.error("[Sleep] 保存睡眠记录失败:", error);

        // 如果保存失败,恢复计时器
        timerRunning.value = true;
        timerInterval = setInterval(() => {
            timerTrigger.value++;
        }, 1000) as unknown as number;

        uni.showToast({
            title: error.message || "保存失败",
            icon: "none",
        });
    }
};

// 页面卸载时清除计时器
onUnmounted(() => {
    if (timerInterval) {
        clearInterval(timerInterval);
    }
});

// 页面加载
onMounted(() => {
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

    checkTempRecord();
});

// 页面显示时也检测(从其他页面返回)
onShow(() => {
    // 每次页面显示时重置检测标志，允许再次检测
    tempRecordCheckDone.value = false;
    checkTempRecord();
});

// 监听睡眠类型变化,如果正在计时则更新临时记录
watch(
    () => sleepType.value,
    () => {
        if (timerRunning.value && startTime.value > 0) {
            saveTempRecord();
            console.log("[Sleep] 睡眠类型已更改,临时记录已更新");
        }
    },
);

// 检测并处理临时睡眠记录
const checkTempRecord = () => {
    // 如果已经在计时,不重复检测
    if (timerRunning.value) {
        return;
    }

    // 如果已经检测过本次，不再重复检测（防止 onMounted 和 onShow 重复调用）
    if (tempRecordCheckDone.value) {
        return;
    }

    const tempRecord = getStorage<TempSleepRecording>(
        StorageKeys.TEMP_SLEEP_RECORDING,
    );

    if (!tempRecord) {
        tempRecordCheckDone.value = true; // 标记已检测
        return;
    }

    // 检查临时记录是否属于当前宝宝
    if (tempRecord.babyId !== currentBabyId.value) {
        console.log("[Sleep] 临时记录不属于当前宝宝,已忽略");
        tempRecordCheckDone.value = true; // 标记已检测
        return;
    }

    // 标记已检测（在显示弹窗前）
    tempRecordCheckDone.value = true;

    // 计算已过时长
    const elapsedSeconds = Math.floor(
        (Date.now() - tempRecord.startTime) / 1000,
    );
    const hours = Math.floor(elapsedSeconds / 3600);
    const minutes = Math.floor((elapsedSeconds % 3600) / 60);
    const seconds = elapsedSeconds % 60;

    console.log(
        "[Sleep] 检测到临时记录,已过时长:",
        `${hours}小时${minutes}分${seconds}秒`,
    );

    // 弹窗询问用户
    uni.showModal({
        title: "未完成的睡眠记录",
        content: `检测到您之前有一次未完成的${tempRecord.type === "nap" ? "小睡" : "夜间长睡"}记录,已过 ${hours} 小时 ${minutes} 分钟 ${seconds} 秒,是否继续?`,
        confirmText: "继续",
        cancelText: "重新开始",
        success: (res) => {
            if (res.confirm) {
                // 用户选择继续
                console.log("[Sleep] 用户选择继续临时记录");
                // 恢复临时记录
                restoreTempRecord(tempRecord);
            } else {
                // 用户选择重新开始
                console.log("[Sleep] 用户选择重新开始,清除临时记录");
                clearTempRecord();
            }
        },
    });
};

// 格式化记录时间
const formatRecordTime = (record: SleepRecord) => {
    return formatDate(record.startTime, "MM-DD HH:mm");
};
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
    font-size: 72rpx;
    font-weight: bold;
    font-family: "Courier New", monospace;
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

.quick-record-section {
    background: white;
    border-radius: 16rpx;
    padding: 30rpx;
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

.quick-record-form {
    padding: 20rpx 0;
}

.form-item {
    margin-bottom: 30rpx;
}

.form-label {
    font-size: 28rpx;
    font-weight: bold;
    margin-bottom: 12rpx;
    display: block;
}

.time-input {
    padding: 16rpx;
    border: 1rpx solid #eee;
    border-radius: 8rpx;
    text-align: center;
    font-size: 28rpx;
    color: #333;
}
</style>
