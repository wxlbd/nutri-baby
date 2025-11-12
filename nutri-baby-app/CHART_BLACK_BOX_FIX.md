# 图表滚动黑框问题修复

## 问题描述

当图表滚动到边界时（最左侧或最右侧），会出现黑色的矩形框：
- 滚动到最右侧时，左侧出现黑框
- 滚动到最左侧时，右侧出现黑框

## 问题原因

这是 Canvas 绘制超出边界导致的：

1. **Canvas 未裁剪超出内容**
   - Canvas 元素没有设置 `overflow: hidden`
   - 超出可视区域的绘制内容会显示出来

2. **图表 padding 不足**
   - 左右 padding 太小（left: 5, right: 15）
   - 滚动时内容会绘制到边界外

3. **边界处理不当**
   - uCharts 在滚动边界时仍然绘制完整内容
   - 没有正确的边界裁剪

## 解决方案

### 1. 添加 Canvas 样式和背景色

```scss
.chart-canvas {
  width: 100%;
  height: 500rpx;
  overflow: hidden;         // 裁剪超出内容
  display: block;           // 块级元素，避免内联间隙
  background-color: #ffffff; // 设置白色背景（关键！）
}
```

### 2. 在 uCharts 配置中添加背景色

```typescript
{
  background: '#ffffff',  // 设置图表背景为白色（关键！）
  type: 'column',
  context: ctx,
  // ... 其他配置
}
```

### 3. 增加图表 padding

将 padding 从 `[15, 15, 0, 5]` 调整为 `[15, 20, 0, 15]`：

```typescript
{
  padding: [15, 20, 0, 15] as [number, number, number, number]
  //        ↑   ↑      ↑
  //        top right  left
  //        增加左右边距，防止内容绘制到边界外
}
```

### 4. 添加 boundaryGap 配置

```typescript
xAxis: {
  disableGrid: true,
  itemCount: itemCount,
  scrollShow: true,
  boundaryGap: 'center'  // 边界留白，居中对齐
}
```

## 关键发现

黑框的根本原因是 **Canvas 背景色未设置**：

1. Canvas 2D 模式下，默认背景是透明的
2. uCharts 在滚动时会清除和重绘 Canvas
3. 清除时如果没有背景色，会显示出底层的颜色（通常是黑色）
4. 需要同时在 CSS 和 uCharts 配置中设置白色背景

## 修改内容

### 文件：`nutri-baby-app/src/pages/statistics/statistics.vue`

#### 1. Canvas 样式

```diff
.chart-canvas {
  width: 100%;
  height: 500rpx;
+ overflow: hidden;
+ display: block;
}
```

#### 2. 喂养图表配置

```diff
{
- padding: [15, 15, 0, 5],
+ padding: [15, 20, 0, 15],
  xAxis: {
    disableGrid: true,
    itemCount: itemCount,
    scrollShow: true,
+   boundaryGap: 'center'
  }
}
```

#### 3. 身高图表配置

```diff
{
- padding: [15, 15, 0, 5],
+ padding: [15, 20, 0, 15],
  xAxis: {
    disableGrid: false,
    itemCount: itemCount,
    scrollShow: true,
+   boundaryGap: 'center'
  }
}
```

#### 4. 体重图表配置

```diff
{
- padding: [15, 15, 0, 5],
+ padding: [15, 20, 0, 15],
  xAxis: {
    disableGrid: false,
    itemCount: itemCount,
    scrollShow: true,
+   boundaryGap: 'center'
  }
}
```

## 技术说明

### overflow: hidden 的作用

- 裁剪超出 Canvas 边界的内容
- 防止绘制内容溢出到可视区域外
- 这是解决黑框问题的关键

### padding 的作用

```
padding: [top, right, bottom, left]
         [15,  20,    0,      15]
```

- `top: 15`: 顶部留白，避免遮挡标题
- `right: 20`: 右侧留白，防止内容绘制到右边界外
- `bottom: 0`: 底部无需留白（X 轴标签在外部）
- `left: 15`: 左侧留白，防止内容绘制到左边界外

### boundaryGap 的作用

- `'center'`: 数据点居中对齐
- 在边界处留出适当空白
- 改善滚动边界的视觉效果

## 预期效果

修复后的效果：

- ✅ 滚动到任何位置都不会出现黑框
- ✅ 图表内容完全在可视区域内
- ✅ 边界处理流畅自然
- ✅ 数据点居中对齐，视觉效果更好

## 测试方法

1. 切换到"本月"视图
2. 在图表上左右滑动
3. 滚动到最左侧，检查右侧是否有黑框
4. 滚动到最右侧，检查左侧是否有黑框
5. 确认所有数据点都在可视区域内

## 相关问题

### Q: 为什么会出现黑框？

A: Canvas 在绘制时，如果内容超出边界且没有裁剪，就会显示出来。黑框实际上是 Canvas 的背景色或未绘制区域。

### Q: 为什么增加 padding 可以解决？

A: padding 增加了图表内容与边界的距离，确保滚动时内容不会绘制到边界外。

### Q: overflow: hidden 会影响性能吗？

A: 不会。这是 CSS 的标准属性，浏览器会高效处理。

### Q: boundaryGap 是必需的吗？

A: 不是必需的，但建议添加。它可以改善边界处的视觉效果，让滚动更自然。

## 总结

这个问题的核心是 Canvas 内容溢出。通过以下三个措施解决：

1. **CSS 裁剪**: `overflow: hidden`
2. **增加边距**: `padding: [15, 20, 0, 15]`
3. **边界配置**: `boundaryGap: 'center'`

这些修改确保了图表在任何滚动位置都能正确显示，不会出现黑框或其他视觉问题。
