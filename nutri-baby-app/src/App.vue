<script setup lang="ts">
import { onLaunch, onShow, onHide } from "@dcloudio/uni-app";

onLaunch(() => {
  console.log("App Launch");

  // 不在 App Launch 时进行页面跳转
  // 让各个页面自己处理权限检查和跳转逻辑
  // 避免与页面生命周期冲突导致的无限循环
});

onShow(() => {
  console.log("App Show");
});

onHide(() => {
  console.log("App Hide");
});
</script>
<style lang="scss">
/* 引入 nutui-uniapp 基础样式,避免导入有问题的 index.scss */
@import 'nutui-uniapp/styles/reset.css';
@import 'nutui-uniapp/styles/iconfont/iconfont.css';
/* 组件样式通过按需引入插件自动加载,无需手动导入 */
</style>
