import { defineConfig } from "vite";
import uni from "@dcloudio/vite-plugin-uni";
import UniComponents from "@uni-helper/vite-plugin-uni-components";
import { NutResolver } from "nutui-uniapp";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    // 确保放在 UniApp() 之前
    UniComponents({
      resolvers: [NutResolver()]
    }),
    uni(),
  ],
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `@import "nutui-uniapp/styles/variables.scss";`,
      }
    }
  }
});