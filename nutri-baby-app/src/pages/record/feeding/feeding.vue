<template>
    <view class="feeding-page">
        <!-- 喂养类型选择 -->
        <view class="type-selector">
            <nut-tabs v-model="feedingType" size="large">
                <nut-tab-pane title="母乳喂养" pane-key="breast">
                    <!-- 母乳喂养表单 -->
                    <view class="feeding-form">
                        <!-- 喂养侧选择 -->
                        <nut-cell-group>
                            <nut-cell title="喂养侧">
                                <nut-radio-group v-model="breastForm.side" direction="horizontal">
                                    <nut-radio label="left">左侧</nut-radio>
                                    <nut-radio label="right">右侧</nut-radio>
                                    <nut-radio label="both">两侧</nut-radio>
                                </nut-radio-group>
                            </nut-cell>
                        </nut-cell-group>

                        <!-- 计时器 -->
                        <view class="timer-section">
                            <view class="timer-display">
                                <text class="time">{{ formattedTime }}</text>
                                <text class="label">{{ timerRunning ? '进行中' : '未开始' }}</text>
                            </view>
                            <view class="timer-buttons">
                                <nut-button
                                        v-if="!timerRunning"
                                        type="primary"
                                        size="large"
                                        block
                                        @click="startTimer"
                                >
                                    开始计时
                                </nut-button>
                                <nut-button
                                        v-else
                                        type="success"
                                        size="large"
                                        block
                                        @click="stopTimer"
                                >
                                    停止计时
                                </nut-button>
                            </view>
                        </view>
                    </view>
                </nut-tab-pane>

                <nut-tab-pane title="奶瓶喂养" pane-key="bottle">
                    <!-- 奶瓶喂养表单 -->
                    <view class="feeding-form">
                        <nut-cell-group>
                            <nut-cell title="奶类型">
                                <nut-radio-group v-model="bottleForm.bottleType" direction="horizontal">
                                    <nut-radio label="formula">配方奶</nut-radio>
                                    <nut-radio label="breast-milk">母乳/冻奶</nut-radio>
                                </nut-radio-group>
                            </nut-cell>

                            <nut-cell title="单位">
                                <nut-radio-group v-model="bottleForm.unit" direction="horizontal">
                                    <nut-radio label="ml">毫升 (ml)</nut-radio>
                                    <nut-radio label="oz">盎司 (oz)</nut-radio>
                                </nut-radio-group>
                            </nut-cell>

                            <nut-cell title="喂养量">
                                <nut-input-number
                                        v-model="bottleForm.amount"
                                        :min="0"
                                        :max="500"
                                        :step="10"
                                />
                            </nut-cell>

                            <nut-cell title="剩余量(可选)">
                                <nut-input-number
                                        v-model="bottleForm.remaining"
                                        :min="0"
                                        :max="bottleForm.amount"
                                        :step="5"
                                />
                            </nut-cell>
                        </nut-cell-group>
                    </view>
                </nut-tab-pane>

                <nut-tab-pane title="辅食" pane-key="food">
                    <!-- 辅食表单 -->
                    <view class="feeding-form">
                        <nut-cell-group>
                            <nut-cell title="辅食名称">
                                <nut-input
                                        v-model="foodForm.foodName"
                                        placeholder="如:米粉、苹果泥等"
                                        clearable
                                />
                            </nut-cell>

                            <nut-cell title="备注(可选)">
                                <nut-textarea
                                        v-model="foodForm.note"
                                        placeholder="记录宝宝的接受程度、有无过敏反应等"
                                        :max-length="200"
                                        :rows="3"
                                />
                            </nut-cell>
                        </nut-cell-group>
                    </view>
                </nut-tab-pane>
            </nut-tabs>
        </view>

        <!-- 提交按钮 -->
        <view class="submit-section">
            <nut-button
                    type="primary"
                    size="large"
                    block
                    @click="handleSubmit"
            >
                保存记录
            </nut-button>
        </view>

        <!-- 订阅消息引导(母乳) -->
        <SubscribeGuide
          v-model="showBreastFeedingGuide"
          type="breast_feeding_reminder"
          :context-message="breastFeedingGuideContext"
          @confirm="handleSubscribeResult"
        />

        <!-- 订阅消息引导(奶瓶) -->
        <SubscribeGuide
          v-model="showBottleFeedingGuide"
          type="bottle_feeding_reminder"
          :context-message="bottleFeedingGuideContext"
          @confirm="handleSubscribeResult"
        />
    </view>
</template>

<script setup lang="ts">
import {computed, onUnmounted, ref} from 'vue'
import {currentBaby, currentBabyId} from '@/store/baby'
import {addFeedingRecord, feedingRecords} from '@/store/feeding'
import {getUserInfo} from '@/store/user'
import {padZero} from '@/utils/common'
import type {FeedingDetail} from '@/types'
import SubscribeGuide from '@/components/SubscribeGuide.vue'
import { getAuthStatus } from '@/store/subscribe'

// 喂养类型
const feedingType = ref<'breast' | 'bottle' | 'food'>('breast')

// 订阅消息引导状态
const showBreastFeedingGuide = ref(false)
const showBottleFeedingGuide = ref(false)

// 母乳喂养引导文案
const breastFeedingGuideContext = computed(() => {
  return '设置提醒,定时通知您该给宝宝喂奶了,不错过每次喂养时间'
})

// 奶瓶喂养引导文案
const bottleFeedingGuideContext = computed(() => {
  return '开启提醒,智能提示您宝宝的喂养时间,让喂养更有规律'
})

// 母乳喂养表单
const breastForm = ref({
    side: 'left' as 'left' | 'right' | 'both',
    leftDuration: 0,
    rightDuration: 0,
})

// 奶瓶喂养表单
const bottleForm = ref({
    bottleType: 'formula' as 'formula' | 'breast-milk',
    amount: 60,
    unit: 'ml' as 'ml' | 'oz',
    remaining: 0,
})

// 辅食表单
const foodForm = ref({
    foodName: '',
    note: '',
})

// 计时器相关
const timerRunning = ref(false)
const timerSeconds = ref(0)
let timerInterval: number | null = null

// 格式化时间显示
const formattedTime = computed(() => {
    const minutes = Math.floor(timerSeconds.value / 60)
    const seconds = timerSeconds.value % 60
    return `${padZero(minutes)}:${padZero(seconds)}`
})

// 开始计时
const startTimer = () => {
    timerRunning.value = true
    timerInterval = setInterval(() => {
        timerSeconds.value++
    }, 1000) as unknown as number
}

// 停止计时
const stopTimer = () => {
    timerRunning.value = false
    if (timerInterval) {
        clearInterval(timerInterval)
        timerInterval = null
    }

    // 使用秒数,不再转换为分钟
    const duration = timerSeconds.value
    if (breastForm.value.side === 'both') {
        // 两侧时平均分配
        breastForm.value.leftDuration = Math.floor(duration / 2)
        breastForm.value.rightDuration = duration - breastForm.value.leftDuration
    } else {
        // 单侧时全部计入
        if (breastForm.value.side === 'left') {
            breastForm.value.leftDuration = duration
            breastForm.value.rightDuration = 0
        } else {
            breastForm.value.leftDuration = 0
            breastForm.value.rightDuration = duration
        }
    }
}

// 组件卸载时清除计时器
onUnmounted(() => {
    if (timerInterval) {
        clearInterval(timerInterval)
    }
})

// 表单验证
const validateForm = (): boolean => {
    if (!currentBaby.value) {
        uni.showToast({
            title: '请先选择宝宝',
            icon: 'none'
        })
        return false
    }

    if (feedingType.value === 'breast') {
        const totalDuration = breastForm.value.leftDuration + breastForm.value.rightDuration
        if (totalDuration === 0) {
            uni.showToast({
                title: '请记录喂养时长',
                icon: 'none'
            })
            return false
        }
    } else if (feedingType.value === 'bottle') {
        if (bottleForm.value.amount <= 0) {
            uni.showToast({
                title: '请输入喂养量',
                icon: 'none'
            })
            return false
        }
    } else if (feedingType.value === 'food') {
        if (!foodForm.value.foodName.trim()) {
            uni.showToast({
                title: '请输入辅食名称',
                icon: 'none'
            })
            return false
        }
    }

    return true
}

// 提交记录
const handleSubmit = async () => {
    if (!validateForm()) {
        return
    }

    const user = getUserInfo()
    if (!user) {
        uni.showToast({
            title: '请先登录',
            icon: 'none'
        })
        return
    }

    let detail: FeedingDetail

    if (feedingType.value === 'breast') {
        const totalDuration = breastForm.value.leftDuration + breastForm.value.rightDuration
        detail = {
            type: 'breast',
            side: breastForm.value.side,
            duration: totalDuration, // 总时长(秒)
            leftDuration: breastForm.value.leftDuration, // 左侧时长(秒)
            rightDuration: breastForm.value.rightDuration, // 右侧时长(秒)
        }
    } else if (feedingType.value === 'bottle') {
        detail = {
            type: 'bottle',
            bottleType: bottleForm.value.bottleType,
            amount: bottleForm.value.amount,
            unit: bottleForm.value.unit,
            remaining: bottleForm.value.remaining || undefined,
        }
    } else {
        detail = {
            type: 'food',
            foodName: foodForm.value.foodName,
            note: foodForm.value.note || undefined,
        }
    }

    try {
        console.log('[Feeding] 开始保存喂养记录...')
        await addFeedingRecord({
            babyId: currentBabyId.value,
            time: Date.now(),
            detail: detail,
            createBy: user.openid,
            createByAvatar: user.avatarUrl,
            createByName: user.nickName
        })
        console.log('[Feeding] 喂养记录保存成功')

        // 每次保存成功后都直接询问授权(除非被Ban)
        const messageType = feedingType.value === 'breast' ? 'breast_feeding_reminder' : 'bottle_feeding_reminder'
        console.log('[Feeding] 检查订阅授权状态, messageType:', messageType)

        const authStatus = getAuthStatus(messageType)
        console.log('[Feeding] authStatus:', authStatus)

        // 只有被Ban时才不显示
        if (authStatus !== 'ban') {
          console.log('[Feeding] 准备显示订阅引导 (authStatus !== ban)')

          if (feedingType.value === 'breast') {
            console.log('[Feeding] 母乳喂养,1.5秒后显示引导')
            setTimeout(() => {
              console.log('[Feeding] 显示母乳喂养订阅引导')
              showBreastFeedingGuide.value = true
            }, 1500) // 延迟1.5秒显示,让用户看到保存成功的提示
            return // 等待用户处理引导,不立即返回上一页
          }

          if (feedingType.value === 'bottle') {
            console.log('[Feeding] 奶瓶喂养,1.5秒后显示引导')
            setTimeout(() => {
              console.log('[Feeding] 显示奶瓶喂养订阅引导')
              showBottleFeedingGuide.value = true
            }, 1500)
            return
          }
        } else {
          console.log('[Feeding] authStatus === ban, 不显示订阅引导')
        }

        // 被Ban或不需要引导,直接返回上一页
        console.log('[Feeding] 1秒后返回上一页')
        setTimeout(() => {
            uni.navigateBack()
        }, 1000)
    } catch (error) {
        // 错误已在 store 中处理
        console.error('[Feeding] 保存喂养记录失败:', error)
    }
}

// 处理订阅消息结果
const handleSubscribeResult = (result: 'accept' | 'reject') => {
  console.log('订阅结果:', result)
  // 返回上一页
  setTimeout(() => {
    uni.navigateBack()
  }, 500)
}
</script>

<style lang="scss" scoped>
.feeding-page {
    min-height: 100vh;
    background: #f5f5f5;
    padding-bottom: 120rpx;
}

.type-selector {
    background: white;
}

.feeding-form {
    padding: 20rpx;
}

.timer-section {
    background: white;
    border-radius: 16rpx;
    padding: 40rpx;
    margin: 20rpx 0;
    text-align: center;
}

.timer-display {
    margin-bottom: 40rpx;
}

.time {
    font-size: 80rpx;
    font-weight: bold;
    color: #fa2c19;
    display: block;
    margin-bottom: 16rpx;
}

.label {
    font-size: 28rpx;
    color: #808080;
}

.submit-section {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    padding: 20rpx;
    background: white;
    box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
}
</style>