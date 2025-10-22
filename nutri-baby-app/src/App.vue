<script setup lang="ts">
import { onLaunch, onShow, onHide } from "@dcloudio/uni-app";
import { checkLoginStatus } from '@/store/user'

onLaunch(async () => {
  console.log("App Launch");

  // 启动时检查用户状态并重定向
  await checkUserStatusAndRedirect()
});

onShow(() => {
  console.log("App Show");
});

onHide(() => {
  console.log("App Hide");
});

/**
 * 检查用户状态并重定向到正确的页面
 *
 * 流程:
 * 1. 未登录 -> 登录页
 * 2. 已登录 -> 首页
 */
async function checkUserStatusAndRedirect() {
  try {
    // 检查登录状态(同步检查本地 token)
    const isLoggedIn = checkLoginStatus()

    if (!isLoggedIn) {
      // 未登录,跳转到登录页
      console.log('[App] 未登录,跳转到登录页')
      uni.reLaunch({
        url: '/pages/user/login'
      })
      return
    }

    // 已登录,直接跳转到首页
    console.log('[App] 已登录,跳转到首页')
    uni.reLaunch({
      url: '/pages/index/index'
    })

  } catch (error) {
    // 检查失败,跳转到登录页
    console.error('[App] 检查用户状态失败:', error)
    uni.reLaunch({
      url: '/pages/user/login'
    })
  }
}
</script>
<style lang="scss">
/* 引入 nutui-uniapp 基础样式,避免导入有问题的 index.scss */
@import 'nutui-uniapp/styles/reset.css';
@import 'nutui-uniapp/styles/iconfont/iconfont.css';
/* 组件样式通过按需引入插件自动加载,无需手动导入 */
</style>
