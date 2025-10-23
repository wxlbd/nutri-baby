<template>
  <view class="baby-list-page">
    <!-- 头部 -->

    <!-- 宝宝列表 -->
    <view class="baby-list">
      <view
        v-for="baby in babyList"
        :key="baby.babyId"
        class="baby-card"
        :class="{ active: baby.babyId === currentBabyId, 'is-default': baby.babyId === userInfo?.defaultBabyId }"
      >
        <!-- 默认标签 -->
        <view v-if="baby.babyId === userInfo?.defaultBabyId" class="default-badge">
          <nut-icon name="star-fill" size="12" color="#ff9800" />
          <text>默认</text>
        </view>

        <!-- 卡片头部 - 点击切换宝宝 -->
        <view class="card-header" @click="handleSelectBaby(baby.babyId)">
          <!-- 头像 -->
          <view class="baby-avatar">
            <image
              v-if="baby.avatarUrl"
              :src="baby.avatarUrl"
              mode="aspectFill"
            />
            <view v-else class="avatar-placeholder">
              {{ baby.name.charAt(0) }}
            </view>
          </view>

          <!-- 信息 -->
          <view class="baby-info">
            <view class="name-row">
              <text class="baby-name">{{ baby.name }}</text>
              <text v-if="baby.nickname" class="nickname">{{ baby.nickname }}</text>
            </view>
            <view class="baby-meta">
              <text class="gender">{{ baby.gender === 'male' ? '👦 男宝' : '👧 女宝' }}</text>
              <text class="divider">|</text>
              <text class="age">{{ calculateAge(baby.birthDate) }}</text>
            </view>
          </view>

          <!-- 选中标记 -->
          <view v-if="baby.babyId === currentBabyId" class="check-icon">
            <nut-icon name="check-circle-fill" size="24" color="#fa2c19" />
          </view>
        </view>

        <!-- 分割线 -->
        <view class="divider-line" />

        <!-- 操作按钮区域 -->
        <view class="card-actions" @click.stop>
          <!-- 第一行按钮 -->
          <view class="action-row">
            <nut-button
              v-if="baby.babyId !== userInfo?.defaultBabyId"
              size="small"
              plain
              type="warning"
              @click="handleSetDefault(baby.babyId, baby.name)"
            >
              <nut-icon name="star" size="14" />
              设为默认
            </nut-button>
            <nut-button
              size="small"
              plain
              type="primary"
              @click="handleInvite(baby.babyId, baby.name)"
            >
              <nut-icon name="share" size="14" />
              邀请协作
            </nut-button>
          </view>

          <!-- 第二行按钮 -->
          <view class="action-row">
            <nut-button
              size="small"
              plain
              type="info"
              @click="handleEdit(baby.babyId)"
            >
              <nut-icon name="edit" size="14" />
              编辑
            </nut-button>
            <nut-button
              size="small"
              plain
              type="danger"
              @click="handleDelete(baby.babyId)"
            >
              <nut-icon name="del" size="14" />
              删除
            </nut-button>
          </view>
        </view>
      </view>

      <!-- 空状态 -->
      <nut-empty
        v-if="babyList.length === 0"
        description="还没有添加宝宝"
        image="empty"
      >
        <template #description>
          <text class="empty-text">还没有添加宝宝哦~</text>
        </template>
      </nut-empty>
    </view>

    <!-- 添加按钮 -->
    <view class="add-button">
      <nut-button
        type="primary"
        size="large"
        block
        @click="handleAdd"
      >
        <nut-icon name="plus" size="18" />
        添加宝宝
      </nut-button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { babyList, currentBabyId, setCurrentBaby, deleteBaby } from '@/store/baby'
import { userInfo, setDefaultBaby } from '@/store/user'
import { calculateAge } from '@/utils/date'

// 页面加载时初始化
onMounted(() => {
  // 如果只有一个宝宝且没有选中任何宝宝，默认选中这个宝宝
  if (babyList.value.length === 1 && !currentBabyId.value) {
    setCurrentBaby(babyList.value[0].babyId)
    console.log('[BabyList] 自动选中唯一的宝宝:', babyList.value[0].name)
  }
})

// 选择宝宝
const handleSelectBaby = (id: string) => {
  setCurrentBaby(id)
  console.log('[BabyList] 切换宝宝:', id)
  uni.showToast({
    title: '已切换',
    icon: 'success',
    duration: 1000
  })

  // 延迟返回首页
  setTimeout(() => {
    uni.navigateBack()
  }, 1000)
}

// 设置为默认宝宝
const handleSetDefault = async (id: string, name: string) => {
  try {
    await setDefaultBaby(id)
    console.log('[BabyList] 设置默认宝宝:', name)
  } catch (error) {
    console.error('[BabyList] 设置默认宝宝失败:', error)
  }
}

// 添加宝宝
const handleAdd = () => {
  uni.navigateTo({
    url: '/pages/baby/edit/edit'
  })
}

// 邀请协作者
const handleInvite = (id: string, name: string) => {
  uni.navigateTo({
    url: `/pages/baby/invite/invite?babyId=${id}&babyName=${encodeURIComponent(name)}`
  })
}

// 编辑宝宝
const handleEdit = (id: string) => {
  uni.navigateTo({
    url: `/pages/baby/edit/edit?id=${id}`
  })
}

// 删除宝宝
const handleDelete = (id: string) => {
  uni.showModal({
    title: '确认删除',
    content: '删除后无法恢复,确定要删除这个宝宝吗?',
    success: (res) => {
      if (res.confirm) {
        const success = deleteBaby(id)
        if (success) {
          uni.showToast({
            title: '删除成功',
            icon: 'success'
          })
        }
      }
    }
  })
}
</script>

<style lang="scss" scoped>
.baby-list-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #f8f9fa 0%, #e9ecef 100%);
  padding-bottom: 140rpx;
}

.header {
  background: white;
  padding: 40rpx 30rpx;
  text-align: center;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.title {
  font-size: 36rpx;
  font-weight: bold;
  color: #1a1a1a;
}

.baby-list {
  padding: 24rpx;
}

/* 卡片样式 */
.baby-card {
  background: white;
  border-radius: 20rpx;
  margin-bottom: 24rpx;
  overflow: hidden;
  box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
  position: relative;

  &.active {
    box-shadow: 0 4rpx 20rpx rgba(250, 44, 25, 0.25);
    border: 2px solid #fa2c19;
  }

  &.is-default {
    background: linear-gradient(135deg, #fff8e1 0%, #ffffff 20%);
  }
}

/* 默认标签 */
.default-badge {
  position: absolute;
  top: 16rpx;
  right: 16rpx;
  background: linear-gradient(135deg, #ffd54f 0%, #ffb300 100%);
  color: white;
  font-size: 22rpx;
  padding: 8rpx 16rpx;
  border-radius: 20rpx;
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 6rpx;
  font-weight: bold;
  box-shadow: 0 2rpx 8rpx rgba(255, 152, 0, 0.3);
  z-index: 10;

  text {
    line-height: 1;
  }

  .nut-icon {
    line-height: 1;
  }
}

/* 卡片头部 */
.card-header {
  padding: 30rpx;
  display: flex;
  align-items: center;
  cursor: pointer;
  transition: background 0.2s;

  &:active {
    background: rgba(0, 0, 0, 0.02);
  }
}

.baby-avatar {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);

  image {
    width: 100%;
    height: 100%;
  }

  .avatar-placeholder {
    width: 100%;
    height: 100%;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 48rpx;
    font-weight: bold;
    color: white;
  }
}

.baby-info {
  flex: 1;
  margin-left: 24rpx;
  overflow: hidden;
}

.name-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 12rpx;
  flex-wrap: wrap;
}

.baby-name {
  font-size: 34rpx;
  font-weight: bold;
  color: #1a1a1a;
  line-height: 1.2;
}

.nickname {
  font-size: 26rpx;
  color: #999;
  background: #f5f5f5;
  padding: 4rpx 12rpx;
  border-radius: 12rpx;
  font-weight: normal;
}

.baby-meta {
  font-size: 26rpx;
  color: #666;
  display: flex;
  align-items: center;
  gap: 12rpx;

  .divider {
    color: #ddd;
  }

  .gender {
    font-weight: 500;
  }

  .age {
    color: #999;
  }
}

.check-icon {
  margin-left: 16rpx;
  flex-shrink: 0;
  animation: scaleIn 0.3s ease;
}

@keyframes scaleIn {
  from {
    transform: scale(0);
  }
  to {
    transform: scale(1);
  }
}

/* 分割线 */
.divider-line {
  height: 1rpx;
  background: linear-gradient(90deg, transparent 0%, #e0e0e0 50%, transparent 100%);
  margin: 0 30rpx;
}

/* 操作按钮区域 */
.card-actions {
  padding: 20rpx 30rpx 30rpx;
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.action-row {
  display: flex;
  gap: 16rpx;
  justify-content: space-between;

  :deep(.nut-button) {
    flex: 1;
    height: 64rpx;
    font-size: 26rpx;
    border-radius: 12rpx;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
    gap: 8rpx;
    transition: all 0.2s;

    &:active {
      transform: scale(0.96);
    }

    // 确保图标和文字垂直居中对齐
    .nut-icon {
      line-height: 1;
      vertical-align: middle;
    }
  }
}

/* 空状态 */
.empty-text {
  color: #999;
  font-size: 28rpx;
}

/* 添加按钮 */
.add-button {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 24rpx;
  background: linear-gradient(180deg, transparent 0%, white 20%);
  backdrop-filter: blur(10rpx);

  :deep(.nut-button) {
    height: 88rpx;
    font-size: 32rpx;
    border-radius: 16rpx;
    box-shadow: 0 4rpx 16rpx rgba(250, 44, 25, 0.3);
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
    gap: 12rpx;

    &:active {
      transform: scale(0.98);
    }

    // 图标文字对齐
    .nut-icon {
      line-height: 1;
    }
  }
}
</style>