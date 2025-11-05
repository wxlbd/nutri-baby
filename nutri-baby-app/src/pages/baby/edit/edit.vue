<template>
  <view>
    <wd-message-box />
    <wd-toast />

    <wd-form ref="formRef" :model="formData" :rules="formRules">
      <!-- 基础信息 -->
      <wd-cell-group custom-class="group" title="基础信息" border>
        <!-- 宝宝头像 -->
        <wd-cell title="宝宝头像" title-width="200rpx">
          <view style="text-align: left">
            <view style="margin-bottom: 24rpx">
              <!-- 用户上传的头像 -->
              <image
                v-if="formData.avatarUrl"
                :src="formData.avatarUrl"
                mode="aspectFill"
                style="width: 160rpx; height: 160rpx; border-radius: 50%; object-fit: cover"
              />
              <!-- 默认头像 -->
              <image
                v-else
                src="@/static/default.png"
                mode="aspectFill"
                style="width: 160rpx; height: 160rpx; border-radius: 50%; object-fit: cover"
              />
            </view>
            <wd-button size="small" @click="chooseAvatar">
              <wd-icon name="photograph" size="16" />
              {{ formData.avatarUrl ? '更换头像' : '选择头像' }}
            </wd-button>
          </view>
        </wd-cell>

        <!-- 宝宝姓名 -->
        <wd-input
          label="宝宝姓名"
          label-width="200rpx"
          :maxlength="20"
          show-word-limit
          prop="name"
          required
          clearable
          v-model="formData.name"
          placeholder="请输入宝宝姓名"
        />

        <!-- 小名昵称 -->
        <wd-input
          label="小名昵称"
          label-width="200rpx"
          :maxlength="20"
          show-word-limit
          clearable
          v-model="formData.nickname"
          placeholder="请输入小名或昵称（可选）"
        />

        <!-- 性别 -->
        <wd-cell title="性别" title-width="200rpx" prop="gender" center>
          <view style="text-align: left">
            <wd-radio-group v-model="formData.gender" inline>
              <wd-radio value="male">
                <text>👦 男孩</text>
              </wd-radio>
              <wd-radio value="female">
                <text>👧 女孩</text>
            </wd-radio>
            </wd-radio-group>
          </view>
        </wd-cell>

        <!-- 出生日期 -->
        <wd-datetime-picker
          label="出生日期"
          label-width="200rpx"
          placeholder="请选择出生日期"
          prop="birthDate"
          type="date"
          @confirm="handleDateConfirm"
        />
      </wd-cell-group>
    </wd-form>

    <!-- 底部按钮 -->
    <view style="padding: 24rpx;display: flex;flex-direction: column; justify-content: center;gap: 10rpx;" >
      <wd-button type="primary" size="large" @click="handleSubmit" block :loading="isSubmitting">
        {{ isEdit ? '保存更改' : '添加宝宝' }}
      </wd-button>
      <wd-button v-if="isEdit" plain size="large" @click="handleCancel" block style="margin-top: 24rpx">
        取消
      </wd-button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { formatDate } from '@/utils/date'

// 直接调用 API 层
import * as babyApi from '@/api/baby'
import * as vaccineApi from '@/api/vaccine'

// 表单数据
const formData = ref({
  name: '',
  nickname: '',
  gender: 'male' as 'male' | 'female',
  birthDate: '',
  avatarUrl: '',
})

// 表单验证规则
const formRules = ref({
  name: [
    { required: true, message: '请输入宝宝姓名', errorType: 'message' },
    { validator: (val: string) => val.trim().length > 0, message: '宝宝姓名不能为空', errorType: 'message' },
  ],
  nickname: [],
  gender: [
    { required: true, message: '请选择宝宝性别', errorType: 'message' },
  ],
  birthDate: [
    { required: true, message: '请选择出生日期', errorType: 'message' },
  ],
})

// 是否为编辑模式
const isEdit = ref(false)
const editId = ref('')

// 日期选择器
const selectedDate = ref()

// 提交状态
const isSubmitting = ref(false)

// 表单 ref
const formRef = ref<any>(null)

// 页面加载
onMounted(async () => {
  // 获取页面参数
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1] as any
  const options = currentPage.options || {}

  if (options.id) {
    // 编辑模式
    isEdit.value = true
    editId.value = options.id

    try {
      const baby = await babyApi.apiFetchBabyDetail(options.id)
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
    } catch (error) {
      console.error('加载宝宝信息失败:', error)
      uni.showToast({
        title: '加载数据失败',
        icon: 'none'
      })
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
      formData.value.avatarUrl = res.tempFilePaths[0] || ''
      // 这里可以上传到服务器
      // uploadFile(res.tempFilePaths[0])
    }
  })
}

// 日期确认
const handleDateConfirm = (val: any) => {
  const { value } = val
  console.log('selectedValue:', value)
  formData.value.birthDate = formatDate(value, 'YYYY-MM-DD')
}

// 提交表单
const handleSubmit = async () => {
  try {
    // 验证表单
    const valid = await formRef.value?.validate()
    if (!valid) {
      return
    }

    isSubmitting.value = true

    if (isEdit.value) {
      // 更新
      await babyApi.apiUpdateBaby(editId.value, {
        name: formData.value.name,
        nickname: formData.value.nickname,
        gender: formData.value.gender,
        birthDate: formData.value.birthDate,
        avatarUrl: formData.value.avatarUrl,
      })

      uni.showToast({
        title: '更新成功',
        icon: 'success'
      })

      setTimeout(() => {
        uni.navigateBack()
      }, 1000)
    } else {
      // 添加
      const newBaby = await babyApi.apiCreateBaby({
        name: formData.value.name,
        nickname: formData.value.nickname,
        gender: formData.value.gender,
        birthDate: formData.value.birthDate,
        avatarUrl: formData.value.avatarUrl,
      })

      // ✨ 为新宝宝自动获取疫苗计划
      console.log('[BabyEdit] 为新宝宝获取疫苗计划:', newBaby.name)

      try {
        // 从服务器获取该宝宝的疫苗计划
        await vaccineApi.apiFetchVaccinePlans(newBaby.babyId)

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
      } catch (vaccineError) {
        console.error('获取疫苗计划失败:', vaccineError)
        // 即使疫苗计划获取失败,宝宝添加仍然成功
        uni.showToast({
          title: '宝宝添加成功',
          icon: 'success'
        })
        setTimeout(() => {
          uni.reLaunch({
            url: '/pages/index/index'
          })
        }, 1500)
      }
    }
  } catch (error: any) {
    console.error('保存失败:', error)
    uni.showToast({
      title: error.message || '保存失败',
      icon: 'none'
    })
  } finally {
    isSubmitting.value = false
  }
}

// 取消编辑
const handleCancel = () => {
  uni.navigateBack()
}
</script>

<style scoped></style>