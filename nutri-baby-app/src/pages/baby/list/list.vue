<template>
  <view class="baby-list-page">
    <!-- 头部 -->
    <view class="header">
      <text class="title">选择宝宝</text>
    </view>

    <!-- 宝宝列表 -->
    <view class="baby-list">
      <view
        v-for="baby in babyList"
        :key="baby.babyId"
        class="baby-item"
        :class="{ active: baby.babyId === currentBabyId }"
        @click="handleSelectBaby(baby.babyId)"
      >
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
          <view class="baby-name">
            {{ baby.name }}
            <text v-if="baby.nickname" class="nickname">({{ baby.nickname }})</text>
          </view>
          <view class="baby-meta">
            <text class="gender">{{ baby.gender === 'male' ? '👦' : '👧' }}</text>
            <text class="age">{{ calculateAge(baby.birthDate) }}</text>
          </view>
        </view>

        <!-- 选中标记 -->
        <view v-if="baby.babyId === currentBabyId" class="check-icon">
          <nut-icon name="checked" size="20" color="#fa2c19" />
        </view>

        <!-- 操作按钮 -->
        <view class="baby-actions" @click.stop>
          <nut-button
            size="small"
            type="primary"
            @click="handleInvite(baby.babyId, baby.name)"
          >
            邀请
          </nut-button>
          <nut-button
            size="small"
            type="default"
            @click="handleEdit(baby.babyId)"
          >
            编辑
          </nut-button>
          <nut-button
            size="small"
            type="default"
            @click="handleDelete(baby.babyId)"
          >
            删除
          </nut-button>
        </view>
      </view>

      <!-- 空状态 -->
      <nut-empty
        v-if="babyList.length === 0"
        description="还没有添加宝宝"
        image="empty"
      />
    </view>

    <!-- 添加按钮 -->
    <view class="add-button">
      <nut-button
        type="primary"
        size="large"
        block
        @click="handleAdd"
      >
        <nut-icon name="plus" />
        添加宝宝
      </nut-button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { babyList, currentBabyId, setCurrentBaby, deleteBaby } from '@/store/baby'
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
  background: #f5f5f5;
  padding-bottom: 120rpx;
}

.header {
  background: white;
  padding: 40rpx 30rpx;
  text-align: center;
}

.title {
  font-size: 36rpx;
  font-weight: bold;
  color: #1a1a1a;
}

.baby-list {
  padding: 20rpx;
}

.baby-item {
  background: white;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
  display: flex;
  align-items: center;
  position: relative;
  transition: all 0.3s;

  &.active {
    border: 2px solid #fa2c19;
  }
}

.baby-avatar {
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  margin-right: 24rpx;
  overflow: hidden;
  flex-shrink: 0;

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
    font-size: 40rpx;
    font-weight: bold;
    color: white;
  }
}

.baby-info {
  flex: 1;
}

.baby-name {
  font-size: 32rpx;
  font-weight: bold;
  color: #1a1a1a;
  margin-bottom: 12rpx;
}

.nickname {
  font-size: 28rpx;
  color: #666;
  font-weight: normal;
}

.baby-meta {
  font-size: 28rpx;
  color: #808080;
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.check-icon {
  margin-left: 20rpx;
}

.baby-actions {
  display: flex;
  gap: 12rpx;
  margin-left: 20rpx;
}

.add-button {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 20rpx;
  background: white;
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
}
</style>