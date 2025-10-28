<script setup lang="ts">
import { onLaunch, onShow, onHide } from "@dcloudio/uni-app";
import { initialize } from "@/store/user";

onLaunch(() => {
  console.log("App Launch");

  // 应用启动时立即初始化用户状态,从本地存储恢复登录信息
  // 这样可以避免代码重新编译时丢失登录状态
  initialize();

  console.log("[App] User state initialized from storage");
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
