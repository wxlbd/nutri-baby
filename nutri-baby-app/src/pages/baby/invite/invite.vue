<template>
  <view>
    <wd-navbar
      title="邀请协作者"
      left-arrow
      safeAreaInsetTop
      @click-left="handleBack"
    >
      <template #capsule>
    <wd-navbar-capsule @back="handleBack" @back-home="handleBackHome" />
  </template>
    </wd-navbar>
    <view class="invite-page">
      <!-- 顶部信息卡片 -->
      <view class="header-card">
        <view class="baby-info">
          <text class="baby-icon">👶</text>
          <view class="baby-detail">
            <text class="baby-name">{{ babyName }}</text>
            <text class="baby-desc">邀请家人共同记录成长</text>
          </view>
        </view>
      </view>

      <!-- 设置表单 -->
      <wd-cell-group border>
        <wd-cell title="协作者角色">
          <wd-radio-group v-model="selectedRole">
            <wd-radio value="editor">编辑者</wd-radio>
            <wd-radio value="viewer">查看者</wd-radio>
          </wd-radio-group>
        </wd-cell>

        <wd-cell title="访问权限">
          <wd-radio-group v-model="accessType">
            <wd-radio value="permanent">永久</wd-radio>
            <wd-radio value="temporary">临时</wd-radio>
          </wd-radio-group>
        </wd-cell>

        <wd-cell
          v-if="accessType === 'temporary'"
          title="过期时间"
          :value="validityText"
          is-link
          @click="showDatetimePickerModal = true"
        />
      </wd-cell-group>

      <!-- 角色说明 -->
      <view class="role-tips">
        <text class="tip-icon">ℹ️</text>
        <text class="tip-text" v-if="selectedRole === 'editor'">
          编辑者可以记录和编辑所有数据
        </text>
        <text class="tip-text" v-else> 查看者只能查看数据，不能编辑 </text>
      </view>

      <!-- 生成按钮 -->
      <view class="button-wrapper">
        <wd-button
          type="primary"
          size="large"
          block
          @click="handleGenerateQRCode"
          :loading="generating"
        >
          {{ generating ? "生成中..." : "生成邀请二维码" }}
        </wd-button>
      </view>

      <!-- 二维码展示 -->
      <view v-if="qrcodeUrl" class="qrcode-section">
        <wd-card>
          <view class="qrcode-wrapper">
            <image :src="qrcodeUrl" class="qrcode-image" mode="aspectFit" />
          </view>

          <view class="qrcode-footer">
            <text class="footer-text">长按识别二维码或保存到相册</text>
            <wd-button type="success" size="small" @click="saveQRCode">
              保存到相册
            </wd-button>
          </view>
        </wd-card>
      </view>

      <!-- 日期时间选择器 -->
      <wd-popup v-model="showDatetimePickerModal" position="bottom">
        <wd-datetime-picker
          v-model="expiresDateValue"
          type="datetime"
          title="选择过期时间"
          :min-date="minDate"
          :max-date="maxDate"
          @confirm="onDateTimeConfirm"
          @cancel="showDatetimePickerModal = false"
        />
      </wd-popup>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { onLoad } from "@dcloudio/uni-app";
import { inviteCollaborator } from "@/store/collaborator";
import { formatDate } from "@/utils";

// 页面参数
const babyId = ref("");
const babyName = ref("");

// 表单数据
const selectedRole = ref<"editor" | "viewer">("editor");
const accessType = ref<"permanent" | "temporary">("permanent");
const expiresDateValue = ref<number>(Date.now() + 7 * 24 * 60 * 60 * 1000); // 默认7天后
const showDatetimePickerModal = ref(false);

// 二维码相关
const qrcodeUrl = ref("");
const generating = ref(false);

// 日期选择器范围
const minDate = Date.now();
const maxDate = Date.now() + 365 * 24 * 60 * 60 * 1000;

// 有效期文本
const validityText = computed(() => {
  if (accessType.value === "permanent") {
    return "永久有效";
  }
  return formatDate(expiresDateValue.value, "YYYY-MM-DD HH:mm");
});

// 页面加载
onLoad((options) => {
  if (options?.babyId) {
    babyId.value = options.babyId;
  }
  if (options?.babyName) {
    babyName.value = decodeURIComponent(options.babyName);
  }
});

// 日期时间选择确认
function onDateTimeConfirm({ value }: { value: number }) {
  expiresDateValue.value = value;
  showDatetimePickerModal.value = false;
}

// 生成二维码
async function handleGenerateQRCode() {
  if (!babyId.value) {
    uni.showToast({
      title: "宝宝ID不能为空",
      icon: "none",
    });
    return;
  }

  generating.value = true;

  try {
    // 计算过期时间戳
    const expiresAt =
      accessType.value === "temporary" ? expiresDateValue.value : undefined;

    // 调用API生成邀请（二维码方式）
    const invitationData = await inviteCollaborator(
      babyId.value,
      "qrcode",
      selectedRole.value,
      accessType.value,
      expiresAt
    );

    const { qrcodeParams } = invitationData;

    if (!qrcodeParams || !qrcodeParams.qrcodeUrl) {
      uni.showToast({
        title: "二维码生成失败",
        icon: "none",
      });
      return;
    }

    // 显示二维码
    qrcodeUrl.value = qrcodeParams.qrcodeUrl;

    uni.showToast({
      title: "二维码生成成功",
      icon: "success",
    });
  } catch (error: any) {
    console.error("Generate QR code error:", error);
    uni.showToast({
      title: error.message || "生成失败",
      icon: "none",
    });
  } finally {
    generating.value = false;
  }
}

// 保存二维码
function saveQRCode() {
  if (!qrcodeUrl.value) {
    uni.showToast({
      title: "二维码未生成",
      icon: "none",
    });
    return;
  }

  // 下载二维码图片
  uni.downloadFile({
    url: qrcodeUrl.value,
    success: (res) => {
      if (res.statusCode === 200) {
        uni.saveImageToPhotosAlbum({
          filePath: res.tempFilePath,
          success: () => {
            uni.showToast({
              title: "保存成功",
              icon: "success",
            });
          },
          fail: () => {
            uni.showToast({
              title: "保存失败,请授予相册权限",
              icon: "none",
            });
          },
        });
      }
    },
    fail: (err) => {
      console.error("Download QR code error:", err);
      uni.showToast({
        title: "下载失败",
        icon: "none",
      });
    },
  });
}
function handleBackHome() {
  uni.switchTab({
    url: "/pages/index/index",
  });
}
function handleBack() {
  uni.navigateBack();
}
</script>

<style lang="scss" scoped>
.invite-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 20rpx;
}

.header-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16rpx;
  padding: 32rpx;
  margin-bottom: 20rpx;

  .baby-info {
    display: flex;
    align-items: center;
    gap: 24rpx;

    .baby-icon {
      font-size: 64rpx;
      line-height: 1;
    }

    .baby-detail {
      flex: 1;
      display: flex;
      flex-direction: column;
      gap: 8rpx;

      .baby-name {
        font-size: 36rpx;
        font-weight: bold;
        color: white;
      }

      .baby-desc {
        font-size: 26rpx;
        color: rgba(255, 255, 255, 0.85);
      }
    }
  }
}

// 角色提示
.role-tips {
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 20rpx 24rpx;
  margin-top: 20rpx;
  background: #fff8e1;
  border-radius: 12rpx;
  border-left: 6rpx solid #ffc107;

  .tip-icon {
    font-size: 32rpx;
  }

  .tip-text {
    flex: 1;
    font-size: 26rpx;
    color: #f57c00;
    line-height: 1.5;
  }
}

// 按钮包装器
.button-wrapper {
  margin-top: 40rpx;
  margin-bottom: 40rpx;
}

// 二维码区域
.qrcode-section {
  margin-top: 20rpx;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20rpx);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.qrcode-wrapper {
  display: flex;
  justify-content: center;
  padding: 40rpx 20rpx;

  .qrcode-image {
    width: 480rpx;
    height: 480rpx;
    border-radius: 12rpx;
    background: white;
  }
}

.qrcode-footer {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20rpx;
  padding: 20rpx 0;
  border-top: 1px solid #f0f0f0;

  .footer-text {
    font-size: 26rpx;
    color: #999;
  }
}
</style>
