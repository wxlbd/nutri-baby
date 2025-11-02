<template>
    <!-- 自定义导航栏 - 与胶囊按钮对称对齐 -->
    <view class="navbar-wrapper" :style="{ paddingTop: statusBarHeight * 2 + 'rpx' }">
        <view
            class="navbar-content"
            :style="{
                height: menuButtonHeight * 2 + 18 + 'rpx'
            }"
        >
            <!-- 左侧宝宝信息 - 对齐胶囊位置 -->
            <view
                class="baby-info"
                @click="goToBabyList"
                :style="{
                    maxWidth: '360rpx',
                    height: menuButtonHeight * 2 + 'rpx',
                }"
            >
                <view v-if="currentBaby" class="baby-content">
                    <view class="baby-avatar">
                        <image
                            v-if="currentBaby.avatarUrl"
                            :src="currentBaby.avatarUrl"
                            mode="aspectFill"
                            class="avatar-img"
                        />
                        <image
                            v-else
                            src="/static/default.png"
                            mode="aspectFill"
                            class="avatar-img"
                        />
                    </view>
                    <view class="baby-text">
                        <text class="baby-name">{{ currentBaby.name }}</text>
                        <text class="baby-age">{{ babyAge }}</text>
                    </view>
                    <nut-icon
                        name="right"
                        size="12"
                        color="#999"
                        class="arrow-icon"
                    />
                </view>
                <view v-else class="add-baby-hint">
                    <text>添加宝宝</text>
                </view>
            </view>

            <!-- 中间标题 -->
            <view class="navbar-title">
                <text>今日概览</text>
            </view>

            <!-- 右侧占位符（与胶囊等宽） -->
            <view
                class="navbar-right"
                :style="{
                    width: menuButtonWidth * 2 + 'rpx',
                    height: menuButtonHeight * 2 + 'rpx',
                }"
            ></view>
        </view>
    </view>
    <view class="index-page" :style="{ paddingTop: navbarTotalHeight - 8 + 'rpx' }">
      

        <!-- 页面内容 -->
        <view class="page-content">
            <!-- 游客模式提示横幅 -->
            <view v-if="!isLoggedIn" class="guest-banner">
                <view class="banner-content">
                    <view class="banner-text">
                        <text class="banner-title">欢迎使用宝宝喂养日志</text>
                        <text class="banner-desc">登录后记录您的宝宝成长数据</text>
                    </view>
                    <nut-button size="small" type="primary" @click="goToLogin">
                        立即登录
                    </nut-button>
                </view>
            </view>

            <!-- 今日数据概览 -->
            <view class="today-stats">
                <view class="stats-title">今日数据</view>
                <view class="stats-grid">
                    <view class="stat-item stat-milk">
                        <image class="stat-bg" src="/static/stat-bg-milk.png" mode="aspectFill" />
                        <view class="stat-content">
                            <view class="stat-icon">🍼</view>
                            <view class="stat-value"
                                >{{ todayStats.totalMilk }}ml</view
                            >
                            <view class="stat-label">奶瓶奶量</view>
                        </view>
                    </view>
                    <view class="stat-item stat-breast">
                        <image class="stat-bg" src="/static/stat-bg-breast.png" mode="aspectFill" />
                        <view class="stat-content">
                            <view class="stat-icon">🤱</view>
                            <view class="stat-value"
                                >{{ todayStats.breastfeedingCount }}次</view
                            >
                            <view class="stat-label">母乳喂养</view>
                        </view>
                    </view>
                    <view class="stat-item stat-sleep">
                        <image class="stat-bg" src="/static/stat-bg-sleep.png" mode="aspectFill" />
                        <view class="stat-content">
                            <view class="stat-icon">💤</view>
                            <view class="stat-value">{{
                                formatDuration(todayStats.sleepDuration)
                            }}</view>
                            <view class="stat-label">睡眠时长</view>
                        </view>
                    </view>
                    <view class="stat-item stat-diaper">
                        <image class="stat-bg" src="/static/stat-bg-diaper.png" mode="aspectFill" />
                        <view class="stat-content">
                            <view class="stat-icon">🧷</view>
                            <view class="stat-value">{{
                                todayStats.diaperCount
                            }}次</view>
                            <view class="stat-label">换尿布</view>
                        </view>
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
            <view
                v-if="upcomingVaccines.length > 0"
                class="vaccine-reminder"
                @click="goToVaccine"
            >
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
                            <text class="vaccine-name"
                                >{{ vaccine.vaccineName }} (第{{
                                    vaccine.doseNumber
                                }}针)</text
                            >
                            <text class="vaccine-date">{{
                                formatVaccineDate(vaccine.scheduledDate)
                            }}</text>
                        </view>
                        <view class="vaccine-badge" :class="vaccine.status">
                            {{
                                vaccine.status === "due" ? "即将到期" : "已逾期"
                            }}
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
        </view>
    </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { onShow } from "@dcloudio/uni-app";
import { isLoggedIn, fetchUserInfo } from "@/store/user";
import { currentBaby, fetchBabyList } from "@/store/baby";
import {
    formatRelativeTime,
    formatDuration,
    formatDate,
    getTodayStart,
    getTodayEnd,
} from "@/utils/date";
import {
    getFeedingGuidelineByAge,
    calculateAgeInMonths,
} from "@/utils/feeding";
import { calculateAge } from "@/utils/date";

// 直接调用 API 层
import * as feedingApi from "@/api/feeding";
import * as diaperApi from "@/api/diaper";
import * as sleepApi from "@/api/sleep";
import * as vaccineApi from "@/api/vaccine";

// 喂养订阅消息管理
import {
    shouldShowFeedingSubscribeRequest,
    requestAllFeedingSubscribeMessages,
} from "@/utils/feeding-subscribe";

// ============ 导航栏相关 ============

// 导航栏相关
const statusBarHeight = ref(0); // 状态栏高度（px）
const menuButtonWidth = ref(0); // 胶囊按钮宽度（px）
const menuButtonHeight = ref(0); // 胶囊按钮高度（px）
const menuButtonTop = ref(0); // 胶囊按钮顶部距离（px）

// 宝宝年龄
const babyAge = computed(() => {
    if (!currentBaby.value) return "";
    return calculateAge(currentBaby.value.birthDate);
});

// 导航栏总高度
const navbarTotalHeight = computed(() => {
    // 总高度计算与 join.vue 保持一致
    // = 状态栏高度 (px×2→rpx) + 胶囊顶部距离 (px×2→rpx) + 导航栏内容高度
    return (
        Math.round(statusBarHeight.value * 2) +
        Math.round(menuButtonTop.value * 2) 
    );
});

// 跳转到宝宝列表
const goToBabyList = () => {
    uni.navigateTo({
        url: "/pages/baby/list/list",
    });
};

// ============ 响应式数据 ============

// 今日喂养记录
const todayFeedingRecords = ref<feedingApi.FeedingRecordResponse[]>([]);

// 今日换尿布记录
const todayDiaperRecords = ref<diaperApi.DiaperRecordResponse[]>([]);

// 今日睡眠记录
const todaySleepRecords = ref<sleepApi.SleepRecordResponse[]>([]);

// 疫苗提醒
const vaccineReminders = ref<vaccineApi.VaccineReminderResponse[]>([]);

// ============ 计算属性 ============

// 今日数据统计
const todayStats = computed(() => {
    if (!currentBaby.value) {
        return {
            totalMilk: 0,
            breastfeedingCount: 0,
            sleepDuration: 0,
            diaperCount: 0,
        };
    }

    // 计算奶瓶奶量 (仅统计奶瓶喂养,母乳无法测量毫升数)
    const totalMilk = todayFeedingRecords.value
        .filter((r) => r.feedingType === "bottle")
        .reduce((sum, r) => sum + (r.amount || 0), 0);

    // 计算母乳喂养次数
    const breastfeedingCount = todayFeedingRecords.value.filter(
        (r) => r.feedingType === "breast",
    ).length;

    // 计算睡眠总时长 (秒)
    const sleepDuration = todaySleepRecords.value.reduce(
        (sum, r) => sum + (r.duration || 0),
        0,
    );

    // 换尿布次数
    const diaperCount = todayDiaperRecords.value.length;

    return {
        totalMilk: Math.round(totalMilk),
        breastfeedingCount,
        sleepDuration,
        diaperCount,
    };
});

// 距上次喂奶时间
const lastFeedingTime = computed(() => {
    
    if (!currentBaby.value || todayFeedingRecords.value.length === 0)
        return "-";

    // 按时间倒序排列,取第一条
    const sortedRecords = [...todayFeedingRecords.value].sort(
        (a, b) => b.feedingTime - a.feedingTime,
    );
    const lastRecord = sortedRecords[0];

    return formatRelativeTime(lastRecord.feedingTime);
});

// 下次喂奶建议 - 基于医学指南
const nextFeedingTime = computed(() => {
    if (!currentBaby.value || todayFeedingRecords.value.length === 0) return "";

    // 获取最后一次喂奶记录
    const sortedRecords = [...todayFeedingRecords.value].sort(
        (a, b) => b.feedingTime - a.feedingTime,
    );
    const lastRecord = sortedRecords[0];

    // 计算宝宝精确月龄
    const ageInMonths = calculateAgeInMonths(currentBaby.value.birthDate);

    // 根据月龄和医学指南获取推荐喂奶间隔
    const guideline = getFeedingGuidelineByAge(ageInMonths);

    // 使用推荐间隔的中位数（分钟）
    const intervalMinutes = Math.round(
        ((guideline.intervalMinHours + guideline.intervalMaxHours) / 2) * 60,
    );

    const nextTime = lastRecord.feedingTime + intervalMinutes * 60 * 1000;
    const timeDiff = nextTime - Date.now();

    // 喂养类型提示
    const feedingTypeHint =
        guideline.feedingType === "demand"
            ? "（按需喂养，请观察宝宝信号）"
            : "";

    if (timeDiff <= 0) {
        return `建议现在喂奶 ${feedingTypeHint}`.trim();
    }

    const hours = Math.floor(timeDiff / (60 * 60 * 1000));
    const minutes = Math.floor((timeDiff % (60 * 60 * 1000)) / (60 * 1000));

    // 显示推荐间隔范围
    const intervalRange = `${Math.floor(guideline.intervalMinHours)}-${Math.ceil(guideline.intervalMaxHours)}小时`;

    if (hours > 0) {
        return `建议 ${hours}小时${minutes}分钟后喂奶（推荐间隔：${intervalRange}）`;
    } else {
        return `建议 ${minutes}分钟后喂奶（推荐间隔：${intervalRange}）`;
    }
});

// 即将到期的疫苗(最多显示2个)
const upcomingVaccines = computed(() => {
    // 仅显示 due 和 overdue 状态的提醒
    return vaccineReminders.value
        .filter((r) => r.status === "due" || r.status === "overdue")
        .slice(0, 2);
});

// 格式化疫苗日期
const formatVaccineDate = (timestamp: number): string => {
    return formatDate(timestamp, "MM-DD");
};

// 页面加载 (仅在首次挂载时执行)
onMounted(() => {
    console.log("[Index] onMounted");
    // 初始化导航栏
    initializeNavbar();
});

// 初始化导航栏
const initializeNavbar = () => {
    // 获取系统信息
    const systemInfo = uni.getSystemInfoSync();
    statusBarHeight.value = systemInfo.statusBarHeight || 0;

    // 获取胶囊按钮信息（仅微信小程序）
    // #ifdef MP-WEIXIN
    try {
        const menuButton = uni.getMenuButtonBoundingClientRect();
        if (menuButton) {
            // 胶囊按钮的宽度和高度（保持 px，与导航栏样式中使用 rpx 统一处理）
            menuButtonWidth.value = menuButton.width; // px
            menuButtonHeight.value = menuButton.height; // px
            menuButtonTop.value = menuButton.top; // px（状态栏下的距离）

            console.log("[Index] 胶囊对齐:", {
                statusBarHeight: statusBarHeight.value,
                menuButtonTop: menuButtonTop.value,
                menuButtonWidth: menuButton.width,
                menuButtonHeight: menuButton.height,
                menuButtonBottom: menuButton.top + menuButton.height,
                navbarTotalHeight: navbarTotalHeight.value,
            });
        }
    } catch (e) {
        console.warn("[Index] 获取胶囊信息失败，使用默认高度", e);
        // 使用默认值
        menuButtonWidth.value = 88; // 默认宽度
        menuButtonHeight.value = 32; // 默认高度
    }
    // #endif
};

// 页面显示 (每次页面显示时执行,包括 switchTab)
onShow(async () => {
    console.log("[Index] onShow - 开始检查登录和宝宝信息");

    // 检查登录和宝宝信息
    await checkLoginAndBaby();
});

// 计算页面内容的 padding-top
// 已改为计算属性 pageContentPaddingTop，无需手动计算

// 检查登录和宝宝信息
const checkLoginAndBaby = async () => {
    console.log("[Index] checkLoginAndBaby - 登录状态:", isLoggedIn.value);

    // 1. 检查登录状态
    if (!isLoggedIn.value) {
        console.log("[Index] 未登录，显示游客模式");
        // ✅ 未登录时不强制跳转，显示游客模式提示
        // 游客模式：用户可以浏览首页，但无法查看真实数据
        return;
    }

    try {
        // 2. 获取用户信息
        await fetchUserInfo();

        // 3. 获取宝宝列表
        const babies = await fetchBabyList();

        console.log("[Index] 宝宝列表:", babies);
        console.log("[Index] 当前宝宝:", currentBaby.value);

        // 4. 检查是否有宝宝 - 使用 babies 数组判断而不是 currentBaby
        if (!babies || babies.length === 0) {
            // 没有宝宝,跳转到添加宝宝页面
            console.log("[Index] 没有宝宝,提示添加");
            uni.showModal({
                title: "提示",
                content: "请先添加宝宝信息",
                showCancel: false,
                success: () => {
                    uni.navigateTo({
                        url: "/pages/baby/edit/edit",
                    });
                },
            });
            return;
        }

        // 5. 有宝宝,加载今日数据
        if (currentBaby.value) {
            await loadTodayData();
        }
    } catch (error) {
        console.error("[Index] 获取用户/宝宝信息失败:", error);
        uni.showToast({
            title: "加载数据失败",
            icon: "none",
        });
    }
};

// 加载今日数据
const loadTodayData = async () => {
    if (!currentBaby.value) return;

    const babyId = currentBaby.value.babyId;
    const todayStart = getTodayStart();
    const todayEnd = getTodayEnd();

    try {
        // 并行加载所有数据
        const [feedingData, diaperData, sleepData, vaccineData] =
            await Promise.all([
                // 获取今日喂养记录
                feedingApi.apiFetchFeedingRecords({
                    babyId,
                    startTime: todayStart,
                    endTime: todayEnd,
                    pageSize: 100,
                }),
                // 获取今日换尿布记录
                diaperApi.apiFetchDiaperRecords({
                    babyId,
                    startTime: todayStart,
                    endTime: todayEnd,
                    pageSize: 100,
                }),
                // 获取今日睡眠记录
                sleepApi.apiFetchSleepRecords({
                    babyId,
                    startTime: todayStart,
                    endTime: todayEnd,
                    pageSize: 100,
                }),
                // 获取疫苗提醒
                vaccineApi.apiFetchVaccineReminders({
                    babyId,
                    status: ["upcoming", "due", "overdue"],
                }),
            ]);

        // 更新响应式数据
        todayFeedingRecords.value = feedingData.records;
        todayDiaperRecords.value = diaperData.records;
        todaySleepRecords.value = sleepData.records;
        vaccineReminders.value = vaccineData.reminders;

        console.log("[Index] 今日数据加载完成", {
            feeding: feedingData.records.length,
            diaper: diaperData.records.length,
            sleep: sleepData.records.length,
            vaccine: vaccineData.reminders.length,
        });
    } catch (error) {
        console.error("[Index] 加载今日数据失败:", error);
        // 不显示错误提示,静默失败
    }
};

// 跳转到登录
const goToLogin = () => {
    uni.navigateTo({
        url: "/pages/user/login",
    });
};

// 喂养记录（需要检查登录状态）
const handleFeeding = async () => {
    if (!isLoggedIn.value) {
        uni.showModal({
            title: "提示",
            content: "该功能需要登录，是否前往登录？",
            success: (res) => {
                if (res.confirm) {
                    goToLogin();
                }
            },
        });
        return;
    }

    if (!currentBaby.value) {
        uni.showToast({
            title: "请先添加宝宝",
            icon: "none",
        });
        return;
    }
    // ✨ 在跳转前申请喂养订阅消息权限
    try {
        console.log("[Index] 检查是否需要申请喂养订阅消息");

        const { shouldShow, bannedCount } = shouldShowFeedingSubscribeRequest();

        if (shouldShow) {
            console.log("[Index] 显示喂养订阅申请, 已Ban数:", bannedCount);
            // 申请喂养订阅消息
            await requestAllFeedingSubscribeMessages();
        } else {
            console.log("[Index] 不需要显示订阅申请");
        }
    } catch (error: any) {
        console.error("[Index] 申请订阅消息失败:", error);
        // 静默失败，不影响主功能
    }

    // 申请完成后跳转到喂养记录页面
    uni.navigateTo({
        url: "/pages/record/feeding/feeding",
    });
};

// 换尿布记录
const handleDiaper = () => {
    if (!isLoggedIn.value) {
        uni.showModal({
            title: "提示",
            content: "该功能需要登录，是否前往登录？",
            success: (res) => {
                if (res.confirm) {
                    goToLogin();
                }
            },
        });
        return;
    }

    if (!currentBaby.value) {
        uni.showToast({
            title: "请先添加宝宝",
            icon: "none",
        });
        return;
    }
    uni.navigateTo({
        url: "/pages/record/diaper/diaper",
    });
};

// 睡眠记录
const handleSleep = () => {
    if (!isLoggedIn.value) {
        uni.showModal({
            title: "提示",
            content: "该功能需要登录，是否前往登录？",
            success: (res) => {
                if (res.confirm) {
                    goToLogin();
                }
            },
        });
        return;
    }

    if (!currentBaby.value) {
        uni.showToast({
            title: "请先添加宝宝",
            icon: "none",
        });
        return;
    }
    uni.navigateTo({
        url: "/pages/record/sleep/sleep",
    });
};

// 成长记录
const handleGrowth = () => {
    if (!isLoggedIn.value) {
        uni.showModal({
            title: "提示",
            content: "该功能需要登录，是否前往登录？",
            success: (res) => {
                if (res.confirm) {
                    goToLogin();
                }
            },
        });
        return;
    }

    if (!currentBaby.value) {
        uni.showToast({
            title: "请先添加宝宝",
            icon: "none",
        });
        return;
    }
    uni.navigateTo({
        url: "/pages/record/growth/growth",
    });
};

// 跳转到疫苗提醒
const goToVaccine = () => {
    if (!isLoggedIn.value) {
        uni.showModal({
            title: "提示",
            content: "该功能需要登录，是否前往登录？",
            success: (res) => {
                if (res.confirm) {
                    goToLogin();
                }
            },
        });
        return;
    }

    if (!currentBaby.value) {
        uni.showToast({
            title: "请先添加宝宝",
            icon: "none",
        });
        return;
    }
    uni.navigateTo({
        url: "/pages/vaccine/vaccine",
    });
};
</script>

<style lang="scss" scoped>
// ===== 设计系统变量 =====
$spacing: 20rpx; // 统一间距

// ===== 导航栏样式 =====
.navbar-wrapper {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    background: #ffffff;
    z-index: 9999;
}

.navbar-content {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 20rpx; // 左右边距
    // 高度由内联样式动态设置
}

// 左侧宝宝信息 - 对齐胶囊位置
.baby-info {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    flex-shrink: 0;
    min-width: 200rpx;
    // 宽高由内联样式动态设置
}

.baby-content {
    display: flex;
    align-items: center;
    gap: 8rpx;
    padding: 6rpx 16rpx 6rpx 6rpx;
    background: #f5f7fa;
    border-radius: 40rpx;
    height: 100%;
    max-width: 100%;
}

.baby-avatar {
    width: 52rpx;
    height: 52rpx;
    border-radius: 50%;
    overflow: hidden;
    flex-shrink: 0;
}

.avatar-img {
    width: 100%;
    height: 100%;
}

.baby-text {
    display: flex;
    flex-direction: column;
    gap: 2rpx;
    flex: 1;
    min-width: 0;
    max-width: 200rpx;
}

.baby-name {
    font-size: 26rpx;
    font-weight: 500;
    color: #333;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    line-height: 1.3;
}

.baby-age {
    font-size: 22rpx;
    color: #999;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    line-height: 1.2;
}

.arrow-icon {
    flex-shrink: 0;
    margin-left: 2rpx;
}

.add-baby-hint {
    padding: 16rpx 32rpx;
    background: #f5f7fa;
    border-radius: 40rpx;
    font-size: 24rpx;
    color: #999;
}

// 中间标题 - 居中显示
.navbar-title {
    position: absolute;
    left: 50%;
    transform: translateX(-50%);
    font-size: 34rpx; // 标准导航栏标题大小 (17px = 34rpx)
    font-weight: 600;
    color: #000;
    pointer-events: none;
}

// 右侧占位符（与胶囊等宽）
.navbar-right {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    flex-shrink: 0;
    // 宽高由内联样式动态设置
}

.index-page {
    // padding-top 由内联样式动态设置
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

// 游客模式提示横幅
.guest-banner {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 16rpx;
    padding: 30rpx;
    margin-bottom: $spacing;
    box-shadow: 0 4rpx 12rpx rgba(102, 126, 234, 0.2);
}

.banner-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 20rpx;
}

.banner-text {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 8rpx;
    color: white;
}

.banner-title {
    font-size: 32rpx;
    font-weight: bold;
}

.banner-desc {
    font-size: 24rpx;
    opacity: 0.9;
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
    position: relative;
    text-align: center;
    padding: 20rpx;
    border-radius: 12rpx;
    overflow: hidden;
    background-color: #f5f5f5;
}

.stat-bg {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    z-index: 0;
    object-fit: cover;
    object-position: center;
}

.stat-content {
    position: relative;
    z-index: 1;
}

// 奶瓶奶量背景
.stat-milk {
}

// 母乳喂养背景
.stat-breast {
}

// 睡眠时长背景
.stat-sleep {
}

// 换尿布背景
.stat-diaper {
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
</style>
