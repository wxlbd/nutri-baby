# TabBar 图标说明

本目录包含宝宝喂养日志应用的底部导航栏图标,同时提供 SVG 和 PNG 两种格式。

## 图标列表

### 1. 首页 (Home)
- **未选中**: `home.png` / `home.svg` - 灰色描边房子图标
- **选中**: `home-active.png` / `home-active.svg` - 红色填充房子图标

**设计说明**: 采用经典的房子造型,象征"首页"/"主页"的概念。选中状态使用品牌红色(#fa2c19)填充,视觉冲击力强。

### 2. 时间轴 (Timeline)
- **未选中**: `timeline.png` / `timeline.svg` - 灰色时间线图标
- **选中**: `timeline-active.png` / `timeline-active.svg` - 红色时间线图标

**设计说明**: 垂直时间线设计,包含三个时间点和对应的描述线条,清晰表达时间流逝和事件记录的概念。

### 3. 统计 (Statistics)
- **未选中**: `statistics.png` / `statistics.svg` - 灰色柱状图和折线图组合
- **选中**: `statistics-active.png` / `statistics-active.svg` - 红色填充的柱状图和折线图

**设计说明**: 结合柱状图和趋势折线图,体现数据统计和分析功能。选中状态的数据点使用白色填充,增加层次感。

### 4. 我的 (User)
- **未选中**: `user.png` / `user.svg` - 灰色用户头像轮廓
- **选中**: `user-active.png` / `user-active.svg` - 红色填充用户头像

**设计说明**: 标准的用户头像轮廓,包括头部圆形和身体半圆形,简洁明了。

## 设计规范

### SVG 格式 (源文件)
- **画布尺寸**: 48x48px
- **未选中颜色**: #999999 (中灰色)
- **选中颜色**: #fa2c19 (品牌红)
- **描边宽度**: 2-2.5px
- **圆角**: 2px (适用于矩形元素)

### PNG 格式 (uni-app 使用)
- **图标尺寸**: 81x81px (3倍图,适配高清屏幕)
- **文件大小**: 844B ~ 2.1KB
- **背景**: 透明
- **抗锯齿**: 已启用,边缘平滑

## 使用方式

在 `pages.json` 中配置 (**使用 PNG 格式,uni-app 不支持 SVG**):

```json
{
  "tabBar": {
    "color": "#999999",
    "selectedColor": "#fa2c19",
    "list": [
      {
        "pagePath": "pages/index/index",
        "text": "首页",
        "iconPath": "static/tabbar/home.png",
        "selectedIconPath": "static/tabbar/home-active.png"
      }
      // ... 其他配置
    ]
  }
}
```

## 图标特点

✅ **双格式提供**: SVG 用于源文件编辑,PNG 用于实际显示
✅ **高清适配**: PNG 采用 81x81px (3倍图),支持 Retina 屏幕
✅ **双状态设计**: 每个图标都有选中和未选中两种状态
✅ **统一风格**: 所有图标采用相同的描边宽度和视觉风格
✅ **语义化命名**: 文件名清晰表达图标用途
✅ **品牌一致性**: 选中颜色与应用主题色保持一致
✅ **透明背景**: PNG 支持透明背景,适配各种主题
✅ **文件优化**: PNG 文件经过优化,体积小加载快

## 文件说明

- **SVG 文件**: 矢量格式源文件,可无损编辑和缩放
- **PNG 文件**: 从 SVG 转换的栅格图像,用于 uni-app TabBar

## 修改指南

如需修改图标:

1. **编辑 SVG 源文件**
   - 使用任何 SVG 编辑器或直接编辑代码
   - 保持 48x48px 的画布尺寸
   - 使用相同的描边宽度(2-2.5px)
   - 保持选中/未选中的颜色规范

2. **重新生成 PNG**
   ```bash
   cd src/static/tabbar
   rsvg-convert -w 81 -h 81 icon-name.svg -o icon-name.png
   ```

3. **批量转换** (如果修改了多个图标)
   ```bash
   for file in *.svg; do
     base="${file%.svg}"
     rsvg-convert -w 81 -h 81 "$file" -o "${base}.png"
   done
   ```

## 技术说明

- **转换工具**: rsvg-convert (librsvg)
- **输出尺寸**: 81×81px (uni-app 推荐的 3倍图尺寸)
- **图像质量**: 高质量抗锯齿渲染
- **兼容性**: 支持 uni-app 全平台(H5、小程序、App)

---

**创建日期**: 2025年10月20日
**设计工具**: 手工编写 SVG 代码
**转换工具**: rsvg-convert
**图标格式**: SVG(源文件) + PNG(生产环境)
**兼容性**: 支持 uni-app 全平台(H5、小程序、App)