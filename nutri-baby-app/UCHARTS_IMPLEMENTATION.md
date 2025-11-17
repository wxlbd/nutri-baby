# uCharts 图表实现说明

## 概述

统计页面 (`statistics.vue`) 已使用 uCharts 库重新实现图表功能，替换了之前的 SVG 和 CSS 实现。

## 实现的图表

### 1. 喂养柱状图 (feedingChart)
- **类型**: 柱状图 (column)
- **数据**: 每日奶瓶奶量
- **颜色**: #7dd3a2 (主题绿色)
- **功能**: 显示本周/本月每日的奶瓶喂养量趋势

### 2. 身高折线图 (heightChart)
- **类型**: 折线图 (line)
- **数据**: 宝宝身高记录
- **颜色**: #7dd3a2 (主题绿色)
- **功能**: 显示宝宝身高成长曲线，使用平滑曲线

### 3. 体重折线图 (weightChart)
- **类型**: 折线图 (line)
- **数据**: 宝宝体重记录
- **颜色**: #52c41a (深绿色)
- **功能**: 显示宝宝体重成长曲线，使用平滑曲线

## 技术实现

### Canvas 初始化

```typescript
const getCanvasContext = (canvasId: string, callback: (ctx: any, width: number, height: number) => void) => {
  const query = uni.createSelectorQuery()
  query.select(`#${canvasId}`)
    .fields({ node: true, size: true })
    .exec((res) => {
      if (res && res[0]) {
        const canvas = res[0].node
        const ctx = canvas.getContext('2d')
        const dpr = uni.getSystemInfoSync().pixelRatio || 1
        
        canvas.width = res[0].width * dpr
        canvas.height = res[0].height * dpr
        ctx.scale(dpr, dpr)
        
        callback(ctx, res[0].width, res[0].height)
      }
    })
}
```

### 图表配置

每个图表都使用 uCharts 的标准配置，并根据数据量动态启用滚动：

```typescript
const dataLength = data.length
const itemCount = timeRange === 'week' ? 7 : 10 // 单屏显示数量
const enableScroll = dataLength > itemCount // 超出时启用滚动

new uCharts({
  type: 'column' | 'line',
  context: ctx,
  width: width,
  height: height,
  categories: [...],  // X 轴标签
  series: [{
    name: '数据名称',
    data: [...]       // Y 轴数据
  }],
  animation: true,
  color: ['#7dd3a2'],
  padding: [15, 15, 0, 5],
  enableScroll: enableScroll, // 动态启用滚动
  legend: { show: false },
  xAxis: { 
    disableGrid: true,
    itemCount: itemCount,    // 单屏数据点数量
    scrollShow: true         // 显示滚动条
  },
  yAxis: {
    gridType: 'dash',
    dashLength: 2
  },
  extra: {
    column: { 
      type: 'group', 
      width: enableScroll ? 15 : 20 // 滚动时柱子稍窄
    },
    line: { type: 'curve', width: 2 }
  }
})
```

### 滚动配置说明

- **喂养图表**: 本周显示 7 天，本月显示 10 天
- **成长图表**: 固定显示 6 个数据点
- **自动启用**: 当数据量超过 itemCount 时自动启用滚动
- **触摸交互**: 支持左右滑动查看更多数据

### 触摸交互

每个图表都支持触摸交互，需要正确调用 uCharts 的滚动方法：

```typescript
// touchstart 事件 - 开始滚动
const touchChart = (e: any) => {
  if (chartInstance) {
    chartInstance.scrollStart?.(e)   // 开始滚动
    chartInstance.showToolTip?.(e)   // 显示提示
  }
}

// touchmove 事件 - 滚动图表
const moveChart = (e: any) => {
  if (chartInstance) {
    chartInstance.scroll?.(e)        // 滚动图表（关键方法！）
    chartInstance.showToolTip?.(e)   // 更新提示位置
  }
}

// touchend 事件 - 结束滚动
const touchEndChart = (e: any) => {
  if (chartInstance) {
    chartInstance.scrollEnd?.(e)     // 结束滚动
  }
}
```

**重要说明**：
- `scrollStart`: 开始滚动，记录初始位置
- `scroll`: 执行滚动，这是让图表跟随手指移动的关键方法
- `scrollEnd`: 结束滚动，释放资源

Canvas 元素需要添加事件修饰符：

```vue
<canvas
  @touchstart.stop="touchChart"
  @touchmove.stop.prevent="moveChart"
  @touchend.stop="touchEndChart"
></canvas>
```

- `.stop`: 阻止事件冒泡
- `.prevent`: 阻止默认行为（防止页面滚动）

## 数据流

1. **加载数据**: `loadRecords()` 从 API 获取记录
2. **计算统计**: computed 属性计算统计数据
3. **绘制图表**: `drawCharts()` 初始化所有图表
4. **响应变化**: watch 监听时间范围变化，重新加载和绘制
5. **页面显示**: `onShow()` 钩子在页面显示时重新加载数据和绘制图表
6. **清理资源**: `onBeforeUnmount()` 在组件卸载时清理图表实例

## 注意事项

### 小程序兼容性

- 使用 `canvas-id` 和 `id` 双重标识
- 使用 `uni.createSelectorQuery()` 获取 Canvas 节点
- 正确处理设备像素比 (DPR)
- uCharts 构造函数需要两个参数：配置对象和回调函数
- 配置对象中需要包含 `$this: {}` 属性

### 性能优化

- 延迟 500ms 绘制，确保 DOM 完全渲染
- 只在有数据时才绘制图表
- 使用 `nextTick()` 确保响应式更新完成
- 重绘前清理旧图表实例，避免内存泄漏
- 页面显示时自动刷新数据，确保数据最新

### 样式

```scss
.chart-canvas {
  width: 100%;
  height: 500rpx;
}
```

## 依赖

- `@qiun/ucharts`: ^2.5.0-20230101
- uni-app 框架
- Vue 3 Composition API

## 测试建议

1. 测试不同时间范围（本周/本月）的数据显示
2. 测试空数据状态
3. 测试触摸交互功能
4. 在不同设备和平台上测试（H5、微信小程序等）
5. 测试数据更新时的图表重绘
6. 测试滚动功能（数据量超过 itemCount 时）

## 调试指南

### 检查图表是否正确创建

查看控制台日志：
```
[Statistics] 开始绘制喂养图表
[Statistics] 喂养数据: [...]
[Statistics] 创建喂养图表实例
[Statistics] 图表配置: { dataLength, itemCount, enableScroll, ... }
[Statistics] 喂养图表绘制完成
```

### 检查滚动是否启用

确认日志中的配置：
- `enableScroll: true` - 滚动已启用
- `itemCount: 10` - 单屏显示 10 个数据点
- `dataLength: 30` - 总共 30 个数据点

### 检查触摸事件

触摸图表时应该看到：
```
[Statistics] 喂养图表 touchstart { touches: [...] }
```

### 常见问题

1. **滑动不生效**
   - 检查 `enableScroll` 是否为 true
   - 检查数据量是否大于 `itemCount`
   - 检查触摸事件是否正确绑定
   - 确认调用了 `touchStart`、`touchMove`、`touchEnd` 方法

2. **图表不显示**
   - 检查 Canvas 节点是否正确获取
   - 检查数据是否正确加载
   - 查看是否有错误日志

3. **X 轴标签重叠**
   - 增加 `itemCount` 值
   - 或减少显示的数据点数量

## 生命周期管理

### 页面加载 (onMounted)
- 初始化页面数据
- 首次加载记录
- 首次绘制图表

### 页面显示 (onShow)
- 每次切换到统计页面时触发
- 重新加载最新数据
- 清理并重新绘制图表
- 确保显示最新的统计信息

### 组件卸载 (onBeforeUnmount)
- 清理所有图表实例
- 释放 Canvas 资源
- 防止内存泄漏

### 时间范围变化 (watch)
- 监听本周/本月切换
- 重新加载对应时间范围的数据
- 自动重绘图表

## 未来改进

- [ ] 添加图表加载状态指示器
- [ ] 支持图表缩放功能
- [ ] 添加更多图表类型（饼图、雷达图等）
- [ ] 优化大数据量时的性能
- [ ] 添加图表导出功能
- [ ] 添加下拉刷新功能
