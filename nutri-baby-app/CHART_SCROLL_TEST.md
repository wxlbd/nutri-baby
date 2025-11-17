# 图表滚动功能测试指南

## 测试前准备

1. 确保有足够的测试数据
   - 喂养记录：至少 15 条（本月视图需要 > 10 条）
   - 成长记录：至少 8 条（需要 > 6 条）

2. 打开浏览器控制台，查看调试日志

## 测试步骤

### 1. 测试喂养图表滚动

1. 切换到"统计"页面
2. 选择"本月"视图
3. 查看控制台日志，确认：
   ```
   [Statistics] 图表配置: {
     dataLength: 30,
     itemCount: 10,
     enableScroll: true
   }
   ```
4. 在喂养图表上左右滑动
5. 应该看到：
   - 图表内容随手指移动
   - X 轴标签滚动显示
   - 触摸日志输出

### 2. 测试成长图表滚动

1. 确保有 7 条以上的成长记录
2. 滚动到成长统计部分
3. 在身高/体重图表上左右滑动
4. 应该看到曲线随手指移动

### 3. 验证滚动配置

打开控制台，查找以下日志：

```javascript
// 喂养图表
{
  enableScroll: true,
  xAxis: {
    itemCount: 10,
    scrollShow: true
  }
}

// 成长图表
{
  enableScroll: true,
  xAxis: {
    itemCount: 6,
    scrollShow: true
  }
}
```

## 预期行为

### 正常滚动
- ✅ 手指左右滑动时，图表内容跟随移动
- ✅ 松开手指后，图表停留在当前位置
- ✅ 可以看到屏幕外的数据点
- ✅ 滚动流畅，无卡顿

### 不需要滚动
- ✅ 数据量 ≤ itemCount 时，不启用滚动
- ✅ 本周视图（7 天）通常不需要滚动
- ✅ 所有数据点都在屏幕内可见

## 故障排查

### 问题：滑动无反应

**检查项：**
1. 控制台是否有触摸事件日志？
   - 有 → 事件绑定正常，检查 uCharts 方法调用
   - 无 → 事件绑定有问题

2. `enableScroll` 是否为 true？
   - 是 → 检查触摸方法
   - 否 → 检查数据量是否足够

3. 图表实例是否创建成功？
   ```javascript
   console.log(feedingChartInstance) // 应该是对象，不是 null
   ```

**解决方案：**
```javascript
// 在触摸事件中添加日志
const touchFeeding = (e: any) => {
  console.log('触摸事件:', e)
  console.log('图表实例:', feedingChartInstance)
  console.log('scrollStart 方法:', feedingChartInstance?.scrollStart)
  console.log('scroll 方法:', feedingChartInstance?.scroll)
  
  if (feedingChartInstance?.scrollStart) {
    feedingChartInstance.scrollStart(e)
  }
}

const moveFeeding = (e: any) => {
  console.log('移动事件:', e)
  if (feedingChartInstance?.scroll) {
    feedingChartInstance.scroll(e)  // 关键：使用 scroll 而不是 touchMove
  }
}
```

**关键发现**：
- ❌ 错误：使用 `touchStart/touchMove/touchEnd`
- ✅ 正确：使用 `scrollStart/scroll/scrollEnd`
- `scroll` 方法是让图表跟随手指移动的关键！

### 问题：滚动卡顿

**可能原因：**
1. 数据量过大（> 100 个数据点）
2. 动画效果影响性能
3. 设备性能限制

**解决方案：**
1. 减少 `itemCount` 值
2. 关闭动画：`animation: false`
3. 优化数据处理逻辑

### 问题：X 轴标签仍然重叠

**解决方案：**
1. 减少 `itemCount`（如从 10 改为 7）
2. 调整字体大小：`fontSize: 10`
3. 旋转标签：
   ```javascript
   xAxis: {
     rotateLabel: true,
     rotateAngle: 45
   }
   ```

## 成功标准

- [x] 本月视图可以左右滑动查看所有 30 天数据
- [x] 成长图表可以滑动查看历史记录
- [x] 滑动流畅，无明显卡顿
- [x] X 轴标签清晰可读，不重叠
- [x] 触摸提示正常显示
- [x] 数据点可以准确点击

## 注意事项

1. **H5 和小程序行为可能不同**
   - H5 使用鼠标事件模拟触摸
   - 小程序使用原生触摸事件
   - 建议在真机上测试

2. **Canvas 2D 模式**
   - 确保 Canvas 设置了 `type="2d"`
   - 旧版小程序可能不支持

3. **事件修饰符**
   - `.stop` 阻止冒泡
   - `.prevent` 阻止默认行为
   - 两者配合使用效果最佳
