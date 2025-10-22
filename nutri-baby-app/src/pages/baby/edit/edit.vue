<template>
  <view class="baby-edit-page">
    <view class="form-container">
      <nut-form ref="formRef" :model-value="formData">
        <!-- 头像 -->
        <nut-form-item label="宝宝头像">
          <view class="avatar-upload" @click="chooseAvatar">
            <image
              v-if="formData.avatarUrl"
              :src="formData.avatarUrl"
              mode="aspectFill"
              class="avatar"
            />
            <view v-else class="avatar-placeholder">
              <nut-icon name="photograph" size="40" />
              <text>点击上传</text>
            </view>
          </view>
        </nut-form-item>

        <!-- 姓名 -->
        <nut-form-item label="宝宝姓名" required>
          <nut-input
            v-model="formData.name"
            placeholder="请输入宝宝姓名"
            clearable
          />
        </nut-form-item>

        <!-- 昵称 -->
        <nut-form-item label="小名昵称">
          <nut-input
            v-model="formData.nickname"
            placeholder="请输入小名或昵称(可选)"
            clearable
          />
        </nut-form-item>

        <!-- 性别 -->
        <nut-form-item label="性别" required>
          <nut-radio-group v-model="formData.gender" direction="horizontal">
            <nut-radio label="male">男孩 👦</nut-radio>
            <nut-radio label="female">女孩 👧</nut-radio>
          </nut-radio-group>
        </nut-form-item>

        <!-- 出生日期 -->
        <nut-form-item label="出生日期" required>
          <view class="date-picker" @click="showDatePicker = true">
            <text v-if="formData.birthDate">{{ formData.birthDate }}</text>
            <text v-else class="placeholder">请选择出生日期</text>
            <nut-icon name="right" />
          </view>
        </nut-form-item>
      </nut-form>

      <!-- 提交按钮 -->
      <view class="submit-button">
        <nut-button
          type="primary"
          size="large"
          block
          @click="handleSubmit"
        >
          {{ isEdit ? '保存' : '添加宝宝' }}
        </nut-button>
      </view>
    </view>

    <!-- 日期选择器 -->
    <nut-date-picker
      v-model:visible="showDatePicker"
      v-model="selectedDate"
      type="date"
      title="选择出生日期"
      :min-date="minDate"
      :max-date="maxDate"
      @confirm="handleDateConfirm"
    />
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { addBaby, updateBaby, getBabyById } from '@/store/baby'
import { initializeVaccinePlansFromServer, generateRemindersForBaby } from '@/store/vaccine'
import { formatDate } from '@/utils/date'

// 表单数据
const formData = ref({
  name: '',
  nickname: '',
  gender: 'male' as 'male' | 'female',
  birthDate: '',
  avatarUrl: '',
})

// 是否为编辑模式
const isEdit = ref(false)
const editId = ref('')

// 日期选择器
const showDatePicker = ref(false)
const selectedDate = ref(new Date())
const minDate = new Date(2020, 0, 1)
const maxDate = new Date()

// 页面加载
onMounted(() => {
  // 获取页面参数
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1] as any
  const options = currentPage.options || {}

  if (options.id) {
    // 编辑模式
    isEdit.value = true
    editId.value = options.id

    const baby = getBabyById(options.id)
    if (baby) {
      formData.value = {
        name: baby.name,
        nickname: baby.nickname || '',
        gender: baby.gender,
        birthDate: baby.birthDate,
        avatarUrl: baby.avatarUrl || '',
      }

      // 设置选中的日期
      selectedDate.value = new Date(baby.birthDate)
    }
  }
})

// 选择头像
const chooseAvatar = () => {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: (res) => {
      formData.value.avatarUrl = res.tempFilePaths[0]
      // 这里可以上传到服务器
      // uploadFile(res.tempFilePaths[0])
    }
  })
}

// 日期确认
const handleDateConfirm = ({ selectedValue }: any) => {
  formData.value.birthDate = formatDate(new Date(selectedValue.join('-')).getTime(), 'YYYY-MM-DD')
  showDatePicker.value = false
}

// 表单验证
const validateForm = (): boolean => {
  if (!formData.value.name.trim()) {
    uni.showToast({
      title: '请输入宝宝姓名',
      icon: 'none'
    })
    return false
  }

  if (!formData.value.birthDate) {
    uni.showToast({
      title: '请选择出生日期',
      icon: 'none'
    })
    return false
  }

  return true
}

// 提交表单
const handleSubmit = async () => {
  if (!validateForm()) {
    return
  }

  try {
    if (isEdit.value) {
      // 更新
      await updateBaby(editId.value, {
        name: formData.value.name,
        nickname: formData.value.nickname,
        gender: formData.value.gender,
        birthDate: formData.value.birthDate,
        avatarUrl: formData.value.avatarUrl,
      })

      setTimeout(() => {
        uni.navigateBack()
      }, 1000)
    } else {
      // 添加（去家庭化架构 - 不需要传 familyId）
      const newBaby = await addBaby({
        name: formData.value.name,
        nickname: formData.value.nickname,
        gender: formData.value.gender,
        birthDate: formData.value.birthDate,
        avatarUrl: formData.value.avatarUrl,
      })

      // ✨ 为新宝宝自动生成疫苗计划和提醒
      console.log('[BabyEdit] 为新宝宝生成疫苗计划:', newBaby.name)

      // 1. 为该宝宝从服务器初始化疫苗计划
      await initializeVaccinePlansFromServer(newBaby.babyId)

      // 2. 为新宝宝生成疫苗提醒
      generateRemindersForBaby(newBaby.babyId, newBaby.birthDate)

      // 显示友好的提示
      uni.showModal({
        title: '✅ 宝宝添加成功',
        content: `已为 ${newBaby.name} 自动生成国家免疫规划疫苗计划和接种提醒，可在"疫苗管理"页面查看详情。`,
        showCancel: false,
        confirmText: '好的',
        success: () => {
          // 跳转到首页
          uni.reLaunch({
            url: '/pages/index/index'
          })
        }
      })
    }
  } catch (error) {
    console.error('保存失败:', error)
    // 错误已在 store 中处理,无需重复提示
  }
}
</script>

<style lang="scss" scoped>
.baby-edit-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 120rpx;
}

.form-container {
  background: white;
  padding: 40rpx 30rpx;
}

.avatar-upload {
  width: 160rpx;
  height: 160rpx;
  border-radius: 50%;
  overflow: hidden;
  border: 2rpx dashed #ddd;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.avatar {
  width: 100%;
  height: 100%;
}

.avatar-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12rpx;
  color: #999;
  font-size: 24rpx;
}

.date-picker {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 20rpx 0;
}

.placeholder {
  color: #999;
}

.submit-button {
  margin-top: 60rpx;
}
</style>