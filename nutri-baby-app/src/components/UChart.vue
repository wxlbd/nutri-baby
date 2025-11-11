<template>
  <view class="ucharts-wrapper" :class="{ 'is-loading': isLoading }">
    <!-- 加载中指示器 -->
    <view v-if="isLoading" class="ucharts-loading">
      <view class="loading-spinner"></view>
      <text class="loading-text">{{ loadingText }}</text>
    </view>

    <!-- Canvas 容器 -->
    <canvas
      :canvas-id="canvasId"
      :type="canvasType"
      class="ucharts-canvas"
      :style="{ width: width, height: height }"
      @click="handleCanvasClick"
    />

    <!-- 错误提示 -->
    <view v-if="error" class="ucharts-error">
      <text class="error-text">{{ error }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import type { PropType } from 'vue'
import uCharts from '@qiun/ucharts'
import type { ChartType, ChartData, ChartOptions } from '@/composables/useUChart'

interface Props {
  /**
   * Canvas ID，必须唯一
   */
  canvasId: string

  /**
   * 图表类型
   */
  chartType: ChartType

  /**
   * 图表数据
   */
  chartData: ChartData

  /**
   * 图表配置选项
   */
  chartOptions?: ChartOptions

  /**
   * Canvas 宽度，默认 100%
   */
  width?: string

  /**
   * Canvas 高度，默认 400rpx
   */
  height?: string

  /**
   * Canvas 类型，微信小程序推荐使用 '2d'
   */
  canvasType?: '2d' | 'webgl'

  /**
   * 加载中文本
   */
  loadingText?: string

  /**
   * 是否正在加载
   */
  isLoading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  width: '100%',
  height: '400rpx',
  canvasType: '2d',
  loadingText: '加载中...',
  isLoading: false
})

const emit = defineEmits<{
  /**
   * Canvas 完成绘制事件
   */
  complete: [detail: any]

  /**
   * Canvas 点击事件
   */
  click: [event: any]

  /**
   * 错误事件
   */
  error: [error: string]
}>()

// 图表实例
const chartInstance = ref<any>(null)

// 错误信息
const error = ref<string>('')

/**
 * 处理 Canvas 完成绘制
 */
const handleChartComplete = (event: any) => {
  try {
    chartInstance.value = event.detail?.instance
    error.value = ''
    emit('complete', event.detail)
  } catch (err: any) {
    error.value = err?.message || '图表绘制失败'
    emit('error', error.value)
  }
}

/**
 * 处理 Canvas 点击
 */
const handleCanvasClick = (event: any) => {
  emit('click', event)
}

/**
 * 重新绘制图表
 */
const redraw = async () => {
  await nextTick()
  if (chartInstance.value) {
    // 触发重新绘制
    chartInstance.value.setOption?.({
      ...props.chartOptions
    })
  }
}

/**
 * 监听数据变化，自动更新图表
 */
watch(
  () => props.chartData,
  async (newData) => {
    if (chartInstance.value && newData.series?.length > 0) {
      await nextTick()
      try {
        // 更新数据会自动触发重绘
        chartInstance.value.setOption?.({
          categories: newData.categories || [],
          series: newData.series
        })
      } catch (err: any) {
        error.value = err?.message || '更新数据失败'
        emit('error', error.value)
      }
    }
  },
  { deep: true }
)

/**
 * 监听配置变化
 */
watch(
  () => props.chartOptions,
  async (newOptions) => {
    if (chartInstance.value && newOptions) {
      await nextTick()
      try {
        chartInstance.value.setOption?.(newOptions)
      } catch (err: any) {
        error.value = err?.message || '更新配置失败'
        emit('error', error.value)
      }
    }
  },
  { deep: true }
)

// 暴露方法给父组件
defineExpose({
  redraw,
  getChartInstance: () => chartInstance.value
})

onMounted(async () => {
  await nextTick()
  // uCharts 会自动处理 Canvas 初始化和绘制
})
</script>

<style scoped lang="scss">
.ucharts-wrapper {
  position: relative;
  width: 100%;
  background: white;
  border-radius: 12rpx;
  overflow: hidden;

  &.is-loading {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 400rpx;
  }
}

.ucharts-canvas {
  display: block;
  width: 100%;
}

.ucharts-loading {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.9);
  z-index: 10;
}

.loading-spinner {
  width: 40rpx;
  height: 40rpx;
  border: 4rpx solid #f0f0f0;
  border-top-color: #7dd3a2;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.loading-text {
  margin-top: 20rpx;
  font-size: 24rpx;
  color: #999;
}

.ucharts-error {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 240, 245, 0.95);
  z-index: 10;
  border: 2rpx solid #ffebee;
}

.error-text {
  font-size: 24rpx;
  color: #d32f2f;
  text-align: center;
  padding: 20rpx;
}
</style>
