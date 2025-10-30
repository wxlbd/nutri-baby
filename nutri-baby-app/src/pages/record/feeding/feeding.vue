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

        <!-- 记录时间选择 -->
        <view class="time-picker-section">
            <nut-cell-group>
                <nut-cell title="记录时间">
                    <template #link>
                        <view class="time-display" @click="showDatePicker">
                            <text>{{ formatRecordTime(recordDateTime) }}</text>
                            <nut-icon name="right"></nut-icon>
                        </view>
                    </template>
                </nut-cell>
            </nut-cell-group>
        </view>

        <!-- 日期选择器 -->
        <nut-date-picker
                v-model="recordDateTime"
                type="datetime"
                :min-date="minDateTime"
                :max-date="maxDateTime"
                @confirm="onDateTimeConfirm"
                @cancel="onDateTimeCancel"
                :visible="showDatetimePickerModal"
        ></nut-date-picker>

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
    </view>
</template>

<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref, watch} from 'vue'
import {onShow} from '@dcloudio/uni-app'
import {currentBaby, currentBabyId} from '@/store/baby'
import {getUserInfo} from '@/store/user'
import {padZero} from '@/utils/common'
import {StorageKeys, getStorage, setStorage, removeStorage} from '@/utils/storage'
import type {FeedingDetail} from '@/types'

// 直接调用 API 层
import * as feedingApi from '@/api/feeding'

// 临时喂养记录类型
interface TempBreastFeeding {
  babyId: string
  side: 'left' | 'right' | 'both'
  startTime: number // 开始时间戳(毫秒)
  feedingType: 'breast'
}

// 喂养类型
const feedingType = ref<'breast' | 'bottle' | 'food'>('breast')

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
const startTime = ref(0) // 开始时间戳 (毫秒)
const timerTrigger = ref(0) // 用于触发视图更新的虚拟响应式值
const tempRecordCheckDone = ref(false) // 防止重复检测临时记录
let timerInterval: number | null = null

// 日期时间选择器
const recordDateTime = ref<Date>(new Date()) // 记录时间,初始为当前时间
const showDatetimePickerModal = ref(false)
const minDateTime = ref<Date>(new Date(Date.now() - 30 * 24 * 60 * 60 * 1000)) // 最小: 30天前
const maxDateTime = ref<Date>(new Date()) // 最大: 当前时间

// 显示日期时间选择器
const showDatePicker = () => {
    showDatetimePickerModal.value = true
}

// 确认日期时间选择
const onDateTimeConfirm = (value: Date) => {
    recordDateTime.value = value
    showDatetimePickerModal.value = false
    console.log('[Feeding] 记录时间已更改为:', value)
}

// 取消日期时间选择
const onDateTimeCancel = () => {
    showDatetimePickerModal.value = false
}

// 格式化记录时间显示
const formatRecordTime = (date: Date): string => {
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    return `${year}-${month}-${day} ${hours}:${minutes}`
}

// 格式化时间显示 - 基于开始时间戳计算
const formattedTime = computed(() => {
    // 依赖 timerTrigger 以触发定期更新
    timerTrigger.value // 访问此值以建立依赖关系

    if (!timerRunning.value || startTime.value === 0) {
        return '00:00'
    }
    const elapsedSeconds = Math.floor((Date.now() - startTime.value) / 1000)
    const minutes = Math.floor(elapsedSeconds / 60)
    const seconds = elapsedSeconds % 60
    return `${padZero(minutes)}:${padZero(seconds)}`
})

// 保存临时记录到本地
const saveTempRecord = () => {
  const tempRecord: TempBreastFeeding = {
    babyId: currentBabyId.value,
    side: breastForm.value.side,
    startTime: startTime.value,
    feedingType: 'breast'
  }
  setStorage(StorageKeys.TEMP_BREAST_FEEDING, tempRecord)
  console.log('[Feeding] 临时记录已保存:', tempRecord)
}

// 清除临时记录
const clearTempRecord = () => {
  removeStorage(StorageKeys.TEMP_BREAST_FEEDING)
  tempRecordCheckDone.value = false // 重置标志，允许下次检测
  console.log('[Feeding] 临时记录已清除')
}

// 恢复临时记录
const restoreTempRecord = (tempRecord: TempBreastFeeding) => {
  breastForm.value.side = tempRecord.side
  startTime.value = tempRecord.startTime
  timerRunning.value = true

  // 启动定时器更新显示
  timerInterval = setInterval(() => {
    // 每秒改变 timerTrigger 以触发计算属性重新计算
    timerTrigger.value++
  }, 1000) as unknown as number

  console.log('[Feeding] 临时记录已恢复, 已过时长:', Math.floor((Date.now() - tempRecord.startTime) / 1000), '秒')
}

// 开始计时
const startTimer = () => {
    startTime.value = Date.now()
    timerRunning.value = true

    // 保存临时记录
    saveTempRecord()

    // 启动定时器以每秒更新视图
    timerInterval = setInterval(() => {
        // 每秒改变 timerTrigger 以触发计算属性重新计算
        timerTrigger.value++
    }, 1000) as unknown as number

    console.log('[Feeding] 开始计时')
}

// 停止计时
const stopTimer = () => {
    timerRunning.value = false
    if (timerInterval) {
        clearInterval(timerInterval)
        timerInterval = null
    }

    // 计算总时长(秒)
    const duration = Math.floor((Date.now() - startTime.value) / 1000)

    console.log('[Feeding] 停止计时,总时长:', duration, '秒')

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

    console.log('[Feeding] 喂养侧:', breastForm.value.side, '左侧:', breastForm.value.leftDuration, '右侧:', breastForm.value.rightDuration)

    // 清除临时记录
    clearTempRecord()
}

// 组件卸载时清除计时器
onUnmounted(() => {
    if (timerInterval) {
        clearInterval(timerInterval)
    }
})

// 页面加载时检测临时记录
onMounted(() => {
  checkTempRecord()
})

// 页面显示时也检测(从其他页面返回)
onShow(() => {
  // 每次页面显示时重置检测标志，允许再次检测
  tempRecordCheckDone.value = false
  checkTempRecord()

  // ❌ 注意: 不能在onShow中自动调用requestSubscribeMessage
  // 微信限制: requestSubscribeMessage只能通过用户主动点击(TAP)触发
  // 已改为在用户点击"保存记录"时申请
})

// 监听喂养侧变化,如果正在计时则更新临时记录
watch(() => breastForm.value.side, () => {
  if (timerRunning.value && startTime.value > 0) {
    saveTempRecord()
    console.log('[Feeding] 喂养侧已更改,临时记录已更新')
  }
})

// 检测并处理临时记录
const checkTempRecord = () => {
  // 如果已经在计时,不重复检测
  if (timerRunning.value) {
    return
  }

  // 如果已经检测过本次，不再重复检测（防止 onMounted 和 onShow 重复调用）
  if (tempRecordCheckDone.value) {
    return
  }

  const tempRecord = getStorage<TempBreastFeeding>(StorageKeys.TEMP_BREAST_FEEDING)

  if (!tempRecord) {
    tempRecordCheckDone.value = true // 标记已检测
    return
  }

  // 检查临时记录是否属于当前宝宝
  if (tempRecord.babyId !== currentBabyId.value) {
    console.log('[Feeding] 临时记录不属于当前宝宝,已忽略')
    tempRecordCheckDone.value = true // 标记已检测
    return
  }

  // 标记已检测（在显示弹窗前）
  tempRecordCheckDone.value = true

  // 计算已过时长
  const elapsedSeconds = Math.floor((Date.now() - tempRecord.startTime) / 1000)
  const minutes = Math.floor(elapsedSeconds / 60)
  const seconds = elapsedSeconds % 60

  console.log('[Feeding] 检测到临时记录,已过时长:', `${minutes}分${seconds}秒`)

  // 弹窗询问用户
  uni.showModal({
    title: '未完成的喂养记录',
    content: `检测到您之前有一次未完成的母乳喂养记录(${tempRecord.side === 'left' ? '左侧' : tempRecord.side === 'right' ? '右侧' : '两侧'}),已过 ${minutes} 分钟 ${seconds} 秒,是否继续?`,
    confirmText: '继续',
    cancelText: '重新开始',
    success: (res) => {
      if (res.confirm) {
        // 用户选择继续
        console.log('[Feeding] 用户选择继续临时记录')
        // 切换到母乳喂养标签
        feedingType.value = 'breast'
        // 恢复临时记录
        restoreTempRecord(tempRecord)
      } else {
        // 用户选择重新开始
        console.log('[Feeding] 用户选择重新开始,清除临时记录')
        clearTempRecord()
      }
    }
  })
}

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
        console.log('[Feeding] 验证母乳喂养,左侧:', breastForm.value.leftDuration, '右侧:', breastForm.value.rightDuration, '总时长:', totalDuration)
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
    // 如果还在计时中，先停止计时以获得准确的时长
    if (timerRunning.value && feedingType.value === 'breast') {
        console.log('[Feeding] 保存前检测到仍在计时,自动停止计时')
        stopTimer()
    }

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

        // 直接调用 API 层创建记录
        const requestData: feedingApi.CreateFeedingRecordRequest = {
            babyId: currentBabyId.value,
            feedingType: detail.type,
            feedingTime: recordDateTime.value.getTime(),
            detail: {}
        }

        // 根据类型填充 detail
        if (detail.type === 'breast') {
            requestData.duration = detail.duration
            requestData.detail = {
                breastSide: detail.side,
                leftTime: detail.leftDuration,
                rightTime: detail.rightDuration,
                duration: detail.duration
            }
        } else if (detail.type === 'bottle') {
            requestData.amount = detail.amount
            requestData.detail = {
                bottleType: detail.bottleType,
                unit: detail.unit,
                remaining: detail.remaining
            }
        } else {
            // food
            requestData.detail = {
                foodName: detail.foodName,
                note: detail.note
            }
        }

        await feedingApi.apiCreateFeedingRecord(requestData)
        console.log('[Feeding] 喂养记录保存成功')

        // 保存成功后清除临时记录 (如果是母乳喂养)
        if (feedingType.value === 'breast') {
          clearTempRecord()
        }

        uni.showToast({
            title: '记录成功',
            icon: 'success'
        })

        // 延迟返回上一页，让用户看到成功提示
        setTimeout(() => {
            uni.navigateBack()
        }, 1500)
    } catch (error: any) {
        console.error('[Feeding] 保存喂养记录失败:', error)
        uni.showToast({
            title: error.message || '记录失败',
            icon: 'none'
        })
    }
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

.time-picker-section {
    background: white;
    margin: 20rpx 0;
}

.time-display {
    display: flex;
    align-items: center;
    gap: 10rpx;
    color: #333;
    font-size: 28rpx;
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