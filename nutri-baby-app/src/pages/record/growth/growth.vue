<template>
    <view class="growth-container">
        <!-- 页面内容 -->
        <view class="growth-content">
            <!-- 最新数据卡片 -->
            <view v-if="latestRecord" class="latest-card">
                <view class="card-title">最新记录</view>
                <view class="data-grid">
                    <view v-if="latestRecord.height" class="data-item">
                        <view class="data-icon">📏</view>
                        <view class="data-value">{{
                            latestRecord.height
                        }}</view>
                        <view class="data-label">身高(cm)</view>
                    </view>
                    <view v-if="latestRecord.weight" class="data-item">
                        <view class="data-icon">⚖️</view>
                        <view class="data-value">{{
                            latestRecord.weight
                        }}</view>
                        <view class="data-label">体重(kg)</view>
                    </view>
                    <view
                        v-if="latestRecord.headCircumference"
                        class="data-item"
                    >
                        <view class="data-icon">📐</view>
                        <view class="data-value">{{
                            latestRecord.headCircumference
                        }}</view>
                        <view class="data-label">头围(cm)</view>
                    </view>
                </view>
                <view class="record-time">
                    记录于
                    {{
                        formatDate(latestRecord.measureTime, "YYYY-MM-DD HH:mm")
                    }}
                </view>
            </view>

            <!-- 添加记录按钮 -->
            <view class="add-section">
                <wd-button
                    type="primary"
                    size="large"
                    block
                    @click="showAddDialog = true"
                >
                    + 添加成长记录
                </wd-button>
            </view>

            <!-- 历史记录列表 -->
            <view class="records-section">
                <view class="section-title">历史记录</view>

                <view v-if="recordList.length === 0" class="empty-state">
                    <wd-status-tip description="暂无成长记录" />
                </view>

                <view v-else class="record-list">
                    <view
                        v-for="record in recordList"
                        :key="record.recordId"
                        class="record-item"
                    >
                        <view class="record-header">
                            <view class="record-date">
                                {{
                                    formatDate(record.measureTime, "YYYY-MM-DD")
                                }}
                            </view>
                            <wd-button
                                size="small"
                                type="default"
                                @click="handleDelete(record.recordId)"
                            >
                                删除
                            </wd-button>
                        </view>

                        <view class="record-data">
                            <view v-if="record.height" class="data-row">
                                <text class="data-label">身高:</text>
                                <text class="data-value"
                                    >{{ record.height }} cm</text
                                >
                            </view>
                            <view v-if="record.weight" class="data-row">
                                <text class="data-label">体重:</text>
                                <text class="data-value"
                                    >{{ record.weight }} kg</text
                                >
                            </view>
                            <view
                                v-if="record.headCircumference"
                                class="data-row"
                            >
                                <text class="data-label">头围:</text>
                                <text class="data-value"
                                    >{{ record.headCircumference }} cm</text
                                >
                            </view>
                            <view v-if="record.note" class="data-row">
                                <text class="data-label">备注:</text>
                                <text class="data-value">{{
                                    record.note
                                }}</text>
                            </view>
                        </view>
                    </view>
                </view>
            </view>
        </view>

        <!-- 添加记录对话框 -->
        <wd-popup
            v-model:visible="showAddDialog"
            position="bottom"
            round
            closeable
        >
            <view class="dialog-content">
                <view class="dialog-title">添加成长记录</view>

                <view class="form-section">
                    <!-- 身高 -->
                    <view class="form-item">
                        <view class="form-label">
                            <text class="icon">📏</text>
                            <text>身高 (cm)</text>
                        </view>
                        <wd-input
                            v-model="formData.height"
                            type="digit"
                            placeholder="请输入身高"
                            clearable
                        />
                    </view>

                    <!-- 体重 -->
                    <view class="form-item">
                        <view class="form-label">
                            <text class="icon">⚖️</text>
                            <text>体重 (kg)</text>
                        </view>
                        <wd-input
                            v-model="formData.weight"
                            type="digit"
                            placeholder="请输入体重"
                            clearable
                        />
                    </view>

                    <!-- 头围 -->
                    <view class="form-item">
                        <view class="form-label">
                            <text class="icon">📐</text>
                            <text>头围 (cm)</text>
                        </view>
                        <wd-input
                            v-model="formData.headCircumference"
                            type="digit"
                            placeholder="请输入头围"
                            clearable
                        />
                    </view>

                    <!-- 记录时间 -->
                    <view class="form-item">
                        <view class="form-label">
                            <text class="icon">📅</text>
                            <text>记录时间</text>
                        </view>
                        <wd-input
                            :model="
                                formatDate(formData.time, 'YYYY-MM-DD HH:mm')
                            "
                            readonly
                            @click="showDatePicker = true"
                        />
                    </view>

                    <!-- 备注 -->
                    <view class="form-item">
                        <view class="form-label">
                            <text class="icon">📝</text>
                            <text>备注</text>
                        </view>
                        <wd-textarea
                            v-model="formData.note"
                            placeholder="可选,记录特殊情况"
                            :max-length="200"
                            :rows="3"
                            role="2"
                        />
                    </view>
                </view>

                <view class="dialog-footer">
                    <wd-button
                        type="default"
                        size="large"
                        block
                        @click="showAddDialog = false"
                    >
                        取消
                    </wd-button>
                    <wd-button
                        type="primary"
                        size="large"
                        block
                        @click="handleSubmit"
                    >
                        保存
                    </wd-button>
                </view>
            </view>
        </wd-popup>

        <!-- 日期选择器 -->
        <wd-datetime-picker
            v-model:visible="showDatePicker"
            v-model="selectedDate"
            type="datetime"
            title="选择记录时间"
            @confirm="handleDateConfirm"
        />
    </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { currentBaby } from "@/store/baby";
import { formatDate } from "@/utils/date";

// 直接调用 API 层
import * as growthApi from "@/api/growth";

// 对话框显示状态
const showAddDialog = ref(false);
const showDatePicker = ref(false);
const selectedDate = ref(new Date());

// 表单数据
const formData = ref({
    height: "",
    weight: "",
    headCircumference: "",
    time: Date.now(),
    note: "",
});

// 成长记录列表(从 API 获取)
const records = ref<growthApi.GrowthRecordResponse[]>([]);

// 最新记录
const latestRecord = computed(() => {
    return records.value.length > 0 ? records.value[0] : null;
});

// 历史记录列表
const recordList = computed(() => {
    return records.value;
});

// 加载成长记录
const loadRecords = async () => {
    if (!currentBaby.value) return;

    try {
        const data = await growthApi.apiFetchGrowthRecords({
            babyId: currentBaby.value.babyId,
            pageSize: 100,
        });
        console.log("成长记录 API 响应:", data); // 调试日志
        records.value = data.records;
    } catch (error) {
        console.error("加载成长记录失败:", error);
    }
};

// 页面加载
onMounted(() => {
    loadRecords();
});

// 日期选择确认
const handleDateConfirm = ({ selectedValue }: any) => {
    const date = new Date(selectedValue.join(" "));
    formData.value.time = date.getTime();
    showDatePicker.value = false;
};

// 提交表单
const handleSubmit = async () => {
    if (!currentBaby.value) {
        uni.showToast({
            title: "请先选择宝宝",
            icon: "none",
        });
        return;
    }

    // 验证至少填写一项
    if (
        !formData.value.height &&
        !formData.value.weight &&
        !formData.value.headCircumference
    ) {
        uni.showToast({
            title: "请至少填写一项数据",
            icon: "none",
        });
        return;
    }

    // 验证数据范围
    const height = parseFloat(formData.value.height);
    const weight = parseFloat(formData.value.weight);
    const headCircumference = parseFloat(formData.value.headCircumference);

    if (
        formData.value.height &&
        (isNaN(height) || height <= 0 || height > 200)
    ) {
        uni.showToast({
            title: "身高数据不合理",
            icon: "none",
        });
        return;
    }

    if (
        formData.value.weight &&
        (isNaN(weight) || weight <= 0 || weight > 100)
    ) {
        uni.showToast({
            title: "体重数据不合理",
            icon: "none",
        });
        return;
    }

    if (
        formData.value.headCircumference &&
        (isNaN(headCircumference) ||
            headCircumference <= 0 ||
            headCircumference > 100)
    ) {
        uni.showToast({
            title: "头围数据不合理",
            icon: "none",
        });
        return;
    }

    // 添加记录
    try {
        await growthApi.apiCreateGrowthRecord({
            babyId: currentBaby.value.babyId,
            measureTime: formData.value.time,
            height: formData.value.height ? height : undefined,
            weight: formData.value.weight ? weight : undefined,
            headCircumference: formData.value.headCircumference
                ? headCircumference
                : undefined,
            note: formData.value.note || undefined,
        });

        uni.showToast({
            title: "添加成功",
            icon: "success",
        });

        // 重新加载记录
        await loadRecords();

        // 重置表单
        formData.value = {
            height: "",
            weight: "",
            headCircumference: "",
            time: Date.now(),
            note: "",
        };

        showAddDialog.value = false;
    } catch (error: any) {
        uni.showToast({
            title: error.message || "添加失败",
            icon: "none",
        });
    }
};

// 删除记录
const handleDelete = async (id: string) => {
    uni.showModal({
        title: "确认删除",
        content: "确定要删除这条成长记录吗?",
        success: async (res) => {
            if (res.confirm) {
                try {
                    await growthApi.apiDeleteGrowthRecord(id);
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
.growth-container {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
}

.growth-content {
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    background: #f5f5f5;
    padding: 20rpx;
}

.latest-card {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 16rpx;
    padding: 30rpx;
    margin-bottom: 20rpx;
    color: white;
    flex-shrink: 0;
}

.card-title {
    font-size: 32rpx;
    font-weight: bold;
    margin-bottom: 20rpx;
}

.data-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 20rpx;
    margin-bottom: 20rpx;
}

.data-item {
    text-align: center;
}

.data-icon {
    font-size: 40rpx;
    margin-bottom: 8rpx;
}

.data-value {
    font-size: 36rpx;
    font-weight: bold;
    margin-bottom: 4rpx;
}

.data-label {
    font-size: 24rpx;
    opacity: 0.9;
}

.record-time {
    font-size: 24rpx;
    opacity: 0.8;
    text-align: center;
}

.add-section {
    margin-bottom: 20rpx;
    flex-shrink: 0;
}

.records-section {
    background: white;
    border-radius: 16rpx;
    padding: 30rpx;
    flex: 1;
    overflow-y: auto;
}

.section-title {
    font-size: 32rpx;
    font-weight: bold;
    margin-bottom: 20rpx;
}

.empty-state {
    padding: 80rpx 0;
}

.record-list {
    display: flex;
    flex-direction: column;
    gap: 20rpx;
}

.record-item {
    background: #f5f5f5;
    border-radius: 12rpx;
    padding: 24rpx;
}

.record-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16rpx;
}

.record-date {
    font-size: 28rpx;
    font-weight: bold;
    color: #1a1a1a;
}

.record-data {
    display: flex;
    flex-direction: column;
    gap: 12rpx;
}

.data-row {
    display: flex;
    justify-content: space-between;
    font-size: 26rpx;

    .data-label {
        color: #666;
    }

    .data-value {
        color: #1a1a1a;
        font-weight: 500;
    }
}

.dialog-content {
    padding: 24rpx;
    display: flex;
    flex-direction: column;
    max-height: 100%;
}

.dialog-title {
    font-size: 36rpx;
    font-weight: bold;
    text-align: center;
    margin-bottom: 24rpx;
}

.form-section {
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    min-height: 0;
}

.form-item {
    margin-bottom: 0;
    flex: 0 1 auto;
}

.form-label {
    display: flex;
    align-items: center;
    gap: 8rpx;
    font-size: 26rpx;
    font-weight: bold;
    margin-bottom: 8rpx;

    .icon {
        font-size: 28rpx;
    }
}

.dialog-footer {
    display: flex;
    flex-direction: column;
    gap: 16rpx;
    margin-top: 16rpx;
}
</style>
