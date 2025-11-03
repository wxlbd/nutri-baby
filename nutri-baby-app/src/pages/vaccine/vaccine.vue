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
                        {{ completionStats.completed }} /
                        {{ completionStats.total }} ({{
                            completionStats.percentage
                        }}%)
                    </text>
                </view>
            </view>
        </view>

        <!-- 即将到期提醒 -->
        <view
            v-if="upcomingReminders && upcomingReminders.length > 0"
            class="reminders-section"
        >
            <view class="section-title"
                >⏰ 近期待接种 ({{ upcomingReminders.length }})</view
            >

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
                            {{ reminder.vaccineName }} (第{{
                                reminder.doseNumber
                            }}针)
                        </view>
                        <view class="vaccine-date">
                            预定时间:
                            {{
                                formatDate(reminder.scheduledDate, "YYYY-MM-DD")
                            }}
                        </view>
                        <view class="vaccine-status">
                            <text
                                v-if="reminder.status === 'due'"
                                class="status-badge due"
                                >即将到期</text
                            >
                            <text
                                v-if="reminder.status === 'overdue'"
                                class="status-badge overdue"
                                >已逾期</text
                            >
                        </view>
                    </view>
                    <view class="reminder-action">
                        <wd-button size="small" type="primary"
                            >记录接种</wd-button
                        >
                    </view>
                </view>
            </view>
        </view>

        <!-- 疫苗计划列表 -->
        <view class="plan-section">
            <view class="section-header">
                <text class="section-title">📋 疫苗计划</text>
                <wd-button size="small" @click="goToManage">
                    管理计划
                </wd-button>
            </view>

            <wd-tabs v-model="activeTab">
                <wd-tab title="全部" pane-key="all" />
                <wd-tab title="已完成" pane-key="completed" />
                <wd-tab title="未完成" pane-key="pending" />
            </wd-tabs>

            <view class="plan-list">
                <view
                    v-for="plan in filteredPlans"
                    :key="plan.id"
                    class="plan-item"
                    :class="{ completed: isPlanCompleted(plan.id) }"
                >
                    <view class="plan-header">
                        <view class="plan-name">
                            <text class="required-badge" v-if="plan.isRequired"
                                >必打</text
                            >
                            {{ plan.vaccineName }}
                        </view>
                        <text class="plan-age">{{ plan.ageInMonths }}个月</text>
                    </view>

                    <view class="plan-detail">
                        <text class="plan-dose">第{{ plan.doseNumber }}针</text>
                        <text v-if="plan.description" class="plan-desc">{{
                            plan.description
                        }}</text>
                    </view>

                    <view v-if="isPlanCompleted(plan.id)" class="plan-record">
                        <text class="completed-icon">✓</text>
                        <text class="completed-text">已接种</text>
                        <text class="completed-date">
                            {{ getRecordDate(plan.id) }}
                        </text>
                    </view>

                    <view v-else class="plan-action">
                        <wd-button
                            size="small"
                            type="primary"
                            @click="handleRecordByPlan(plan)"
                        >
                            记录接种
                        </wd-button>
                    </view>
                </view>
            </view>
        </view>

        <!-- 接种记录对话框 -->
        <wd-popup
            v-model:visible="showRecordDialog"
            position="bottom"
            :style="{ height: '75%' }"
            round
            closeable
        >
            <view class="dialog-container">
                <view class="dialog-header">
                    <view class="dialog-title">记录疫苗接种</view>
                </view>

                <scroll-view class="dialog-body" scroll-y>
                    <view class="form-section">
                        <view class="form-item">
                            <view class="form-label">疫苗名称</view>
                            <wd-input
                                v-model="recordForm.vaccineName"
                                placeholder="疫苗名称"
                                readonly
                            />
                        </view>

                        <view class="form-item">
                            <view class="form-label">接种日期</view>
                            <wd-input
                                :model="
                                    formatDate(
                                        recordForm.vaccineDate,
                                        'YYYY-MM-DD',
                                    )
                                "
                                readonly
                                @click="showDatePicker = true"
                            />
                        </view>

                        <view class="form-item">
                            <view class="form-label"
                                >接种医院 <text class="required">*</text></view
                            >
                            <wd-input
                                v-model="recordForm.hospital"
                                placeholder="请输入医院名称"
                                clearable
                            />
                        </view>

                        <view class="form-item">
                            <view class="form-label">疫苗批号</view>
                            <wd-input
                                v-model="recordForm.batchNumber"
                                placeholder="请输入疫苗批号(可选)"
                                clearable
                            />
                        </view>

                        <view class="form-item">
                            <view class="form-label">不良反应</view>
                            <wd-textarea
                                v-model="recordForm.reaction"
                                placeholder="如有不良反应请记录(可选)"
                                :max-length="200"
                                :rows="1"
                                :autosize="{ minHeight: 60, maxHeight: 120 }"
                            />
                        </view>

                        <view class="form-item">
                            <view class="form-label">备注</view>
                            <wd-textarea
                                v-model="recordForm.note"
                                placeholder="其他备注信息(可选)"
                                :max-length="200"
                                :rows="1"
                                :autosize="{ minHeight: 60, maxHeight: 120 }"
                            />
                        </view>
                    </view>
                </scroll-view>

                <view class="dialog-footer">
                    <wd-button
                        type="primary"
                        size="large"
                        @click="handleSaveRecord"
                    >
                        保存
                    </wd-button>
                    <wd-button
                        type="default"
                        size="large"
                        @click="showRecordDialog = false"
                    >
                        取消
                    </wd-button>
                </view>
            </view>
        </wd-popup>

        <!-- 日期选择器 -->
        <wd-datetime-picker
            v-model:visible="showDatePicker"
            v-model="selectedDate"
            type="date"
            title="选择接种日期"
            @confirm="handleDateConfirm"
        />

        <!-- 订阅消息引导 -->
        <SubscribeGuide
            v-model="showVaccineGuide"
            type="vaccine_reminder"
            :context-message="vaccineGuideContext"
            @confirm="handleSubscribeResult"
        />
    </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { currentBaby, currentBabyId } from "@/store/baby";
import { userInfo } from "@/store/user";
import { formatDate } from "@/utils/date";
import type { VaccinePlan, VaccineReminder } from "@/types";
import SubscribeGuide from "@/components/SubscribeGuide.vue";
import { shouldShowGuide } from "@/store/subscribe";

// 直接调用 API 层
import * as vaccineApi from "@/api/vaccine";

// Tab状态
const activeTab = ref("all");

// 对话框状态
const showRecordDialog = ref(false);
const showDatePicker = ref(false);
const selectedDate = ref(new Date());

// 订阅消息引导状态
const showVaccineGuide = ref(false);

// 疫苗数据(从 API 获取)
const vaccinePlans = ref<vaccineApi.VaccinePlanResponse[]>([]);
const vaccineRecords = ref<vaccineApi.VaccineRecordResponse[]>([]);
const vaccineReminders = ref<vaccineApi.VaccineReminderResponse[]>([]);

// 后端返回的统计数据
const vaccineStats = ref({
    completed: 0,
    percentage: 0,
    total: 0,
});

// 表单数据
const recordForm = ref({
    planId: "",
    vaccineType: "",
    vaccineName: "",
    doseNumber: 1,
    vaccineDate: Date.now(),
    hospital: "",
    batchNumber: "",
    reaction: "",
    note: "",
});

// 加载疫苗数据
const loadVaccineData = async () => {
    if (!currentBaby.value) return;

    const babyId = currentBaby.value.babyId;
    console.log("加载疫苗数据", babyId);
    try {
        const [plansData, recordsData, remindersData] = await Promise.all([
            vaccineApi.apiFetchVaccinePlans(babyId),
            vaccineApi.apiFetchVaccineRecords({ babyId, pageSize: 200 }),
            vaccineApi.apiFetchVaccineReminders({ babyId }),
        ]);

        vaccinePlans.value = plansData.plans;
        vaccineRecords.value = recordsData.records;
        vaccineReminders.value = remindersData.reminders;

        // 保存后端返回的统计数据
        vaccineStats.value = {
            completed: plansData.completed || 0,
            percentage: plansData.percentage || 0,
            total: plansData.total || vaccinePlans.value.length,
        };
    } catch (error) {
        console.error("加载疫苗数据失败:", error);
        uni.showToast({
            title: "加载数据失败",
            icon: "none",
        });
    }
};

// 完成度统计 - 直接使用后端返回的数据
const completionStats = computed(() => {
    if (!currentBaby.value || !vaccineStats.value) {
        return { total: 0, completed: 0, percentage: 0 };
    }

    // 直接使用后端返回的统计数据（更准确、更高效）
    return {
        total: vaccineStats.value.total || 0,
        completed: vaccineStats.value.completed || 0,
        percentage: vaccineStats.value.percentage || 0,
    };
});

// 即将到期的提醒
const upcomingReminders = computed(() => {
    if (!currentBaby.value) return [];

    // 筛选出 upcoming, due, overdue 状态的提醒
    return vaccineReminders.value
        .filter((r) => ["upcoming", "due", "overdue"].includes(r.status))
        .sort((a, b) => a.scheduledDate - b.scheduledDate)
        .slice(0, 3); // 只显示前3个
});

// 疫苗引导的场景化文案
const vaccineGuideContext = computed(() => {
    const reminders = upcomingReminders.value;
    if (reminders && reminders.length > 0) {
        const nextReminder = reminders[0];
        const daysLeft = Math.ceil(
            (nextReminder.scheduledDate - Date.now()) / (1000 * 60 * 60 * 24),
        );
        return `宝宝下次需要接种「${nextReminder.vaccineName}第${nextReminder.doseNumber}针」,距离接种日期还有 ${daysLeft}天`;
    }
    return "下次接种前我们会提前3天提醒您哦~";
});

// 过滤后的计划列表
const filteredPlans = computed(() => {
    let plans = vaccinePlans.value || [];

    if (activeTab.value === "completed") {
        plans = plans.filter((plan) => isPlanCompleted(plan.planId));
    } else if (activeTab.value === "pending") {
        plans = plans.filter((plan) => !isPlanCompleted(plan.planId));
    }

    return plans.sort((a, b) => a.ageInMonths - b.ageInMonths);
});

// 判断计划是否已完成
const isPlanCompleted = (planId: string): boolean => {
    if (!currentBabyId.value || !vaccineRecords.value) return false;
    return vaccineRecords.value.some(
        (record) =>
            record.babyId === currentBabyId.value && record.planId === planId,
    );
};

// 获取接种记录日期
const getRecordDate = (planId: string): string => {
    if (!currentBabyId.value || !vaccineRecords.value) return "";
    const record = vaccineRecords.value.find(
        (r) => r.babyId === currentBabyId.value && r.planId === planId,
    );
    return record ? formatDate(record.vaccineDate, "YYYY-MM-DD") : "";
};

// 根据计划 ID 查找计划
const getVaccinePlanById = (
    planId: string,
): vaccineApi.VaccinePlanResponse | undefined => {
    return vaccinePlans.value.find((p) => p.planId === planId);
};

// 处理记录接种(通过提醒)
const handleRecordVaccine = (reminder: VaccineReminder) => {
    const plan = getVaccinePlanById(reminder.planId);
    if (!plan) return;

    recordForm.value = {
        planId: plan.planId,
        vaccineType: plan.vaccineType,
        vaccineName: plan.vaccineName,
        doseNumber: plan.doseNumber,
        vaccineDate: Date.now(),
        hospital: "",
        batchNumber: "",
        reaction: "",
        note: "",
    };

    showRecordDialog.value = true;
};

// 处理记录接种(通过计划)
const handleRecordByPlan = (plan: vaccineApi.VaccinePlanResponse) => {
    recordForm.value = {
        planId: plan.planId,
        vaccineType: plan.vaccineType,
        vaccineName: plan.vaccineName,
        doseNumber: plan.doseNumber,
        vaccineDate: Date.now(),
        hospital: "",
        batchNumber: "",
        reaction: "",
        note: "",
    };

    showRecordDialog.value = true;
};

// 日期选择确认
const handleDateConfirm = ({ selectedValue }: any) => {
    const date = new Date(selectedValue.join("-"));
    recordForm.value.vaccineDate = date.getTime();
    showDatePicker.value = false;
};

// 保存接种记录
const handleSaveRecord = async () => {
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

    // 保存前记录当前记录数
    const recordCountBefore = vaccineRecords.value.length;

    try {
        await vaccineApi.apiCreateVaccineRecord({
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
        });

        uni.showToast({
            title: "记录成功",
            icon: "success",
        });

        showRecordDialog.value = false;

        // 重新加载数据
        await loadVaccineData();

        // 检查是否是首次添加疫苗记录
        const isFirstRecord = recordCountBefore === 0;

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

.dialog-container {
    height: 100%;
    display: flex;
    flex-direction: column;
    background: #fff;
}

.dialog-header {
    flex-shrink: 0;
    padding: 20rpx 30rpx;
    border-bottom: 1rpx solid #f0f0f0;
}

.dialog-title {
    font-size: 32rpx;
    font-weight: bold;
    text-align: center;
    color: #333;
}

.dialog-body {
    flex: 1;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
}

.form-section {
    padding: 20rpx 30rpx 20rpx 30rpx;
}

.form-item {
    margin-bottom: 20rpx;
}

.form-label {
    font-size: 26rpx;
    font-weight: bold;
    margin-bottom: 8rpx;
    color: #333;
}

.form-label .required {
    color: #fa2c19;
    margin-left: 4rpx;
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

.dialog-footer .nut-button {
    width: 100%;
}
</style>
