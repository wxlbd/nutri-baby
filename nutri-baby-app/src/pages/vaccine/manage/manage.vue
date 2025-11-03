<template>
  <view class="vaccine-plan-manage-page">
    <!-- 顶部操作栏 -->
    <view class="action-bar">
      <wd-button type="primary" size="small" @click="showAddDialog = true">
        <view class="button-content">
          <wd-icon name="plus" />
          <text>添加自定义计划</text>
        </view>
      </wd-button>
    </view>

    <!-- 疫苗计划列表 -->
    <view class="plan-list">
      <view
        v-for="plan in vaccinePlans"
        :key="plan.planId"
        class="plan-item"
      >
        <view class="plan-header">
          <view class="plan-name">
            <text class="required-badge" v-if="plan.isRequired">必打</text>
            <text class="custom-badge" v-if="isCustomPlan(plan)">自定义</text>
            {{ plan.vaccineName }}
          </view>
          <view class="plan-actions">
            <wd-button size="small" type="default" @click="handleEdit(plan)">编辑</wd-button>
            <wd-button v-if="isCustomPlan(plan)" size="small" type="danger" @click="handleDelete(plan)">删除</wd-button>
          </view>
        </view>

        <view class="plan-detail">
          <view class="detail-item">
            <text class="label">接种月龄:</text>
            <text class="value">{{ plan.ageInMonths }}个月</text>
          </view>
          <view class="detail-item">
            <text class="label">剂次:</text>
            <text class="value">第{{ plan.doseNumber }}针</text>
          </view>
          <view class="detail-item">
            <text class="label">提醒:</text>
            <text class="value">提前{{ plan.reminderDays }}天</text>
          </view>
        </view>

        <view v-if="plan.description" class="plan-desc">
          {{ plan.description }}
        </view>
      </view>
    </view>

    <!-- 添加/编辑对话框 -->
    <wd-popup
      v-model:visible="showAddDialog"
      position="bottom"
      :style="{ height: '80%' }"
      round
      closeable
    >
      <view class="dialog-content">
        <view class="dialog-title">{{ isEdit ? '编辑疫苗计划' : '添加疫苗计划' }}</view>

        <view class="form-section">
          <wd-form ref="formRef">
            <wd-form-item label="疫苗名称" required>
              <wd-input
                v-model="form.vaccineName"
                placeholder="请输入疫苗名称"
                clearable
              />
            </wd-form-item>

            <wd-form-item label="疫苗类型" required>
              <wd-input
                v-model="form.vaccineType"
                placeholder="例如: HepB, BCG, DTaP"
                clearable
              />
            </wd-form-item>

            <wd-form-item label="接种月龄" required>
              <wd-input
                v-model.number="form.ageInMonths"
                type="number"
                placeholder="请输入月龄"
                clearable
              />
            </wd-form-item>

            <wd-form-item label="剂次" required>
              <wd-input
                v-model.number="form.doseNumber"
                type="number"
                placeholder="请输入剂次"
                clearable
              />
            </wd-form-item>

            <wd-form-item label="提醒天数">
              <wd-input
                v-model.number="form.reminderDays"
                type="number"
                placeholder="提前几天提醒"
                clearable
              />
            </wd-form-item>

            <wd-form-item label="是否必打">
              <wd-switch v-model="form.isRequired" />
            </wd-form-item>

            <wd-form-item label="说明">
              <wd-textarea
                v-model="form.description"
                placeholder="疫苗说明(可选)"
                :max-length="200"
              />
            </wd-form-item>
          </wd-form>
        </view>

        <view class="dialog-footer">
          <wd-button
            type="default"
            size="large"
            block
            @click="showAddDialog = false"
          >
            取消
          </wd-button>
          <wd-button
            type="primary"
            size="large"
            block
            @click="handleSubmit"
          >
            {{ isEdit ? '保存' : '添加' }}
          </wd-button>
        </view>
      </view>
    </wd-popup>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { currentBaby, currentBabyId } from '@/store/baby'

// 直接调用 API 层
import * as vaccineApi from '@/api/vaccine'

// 对话框状态
const showAddDialog = ref(false)
const isEdit = ref(false)
const editPlanId = ref('')

// 疫苗计划列表(从 API 获取)
const vaccinePlans = ref<vaccineApi.VaccinePlanResponse[]>([])

// 表单数据
const form = ref({
  vaccineName: '',
  vaccineType: '',
  ageInMonths: 0,
  doseNumber: 1,
  reminderDays: 7,
  isRequired: true,
  description: ''
})

// 加载疫苗计划
const loadVaccinePlans = async () => {
  if (!currentBaby.value) return

  try {
    const data = await vaccineApi.apiFetchVaccinePlans(currentBaby.value.babyId)
    vaccinePlans.value = data.plans
  } catch (error) {
    console.error('加载疫苗计划失败:', error)
    uni.showToast({
      title: '加载数据失败',
      icon: 'none'
    })
  }
}

// 判断是否为自定义计划
const isCustomPlan = (plan: vaccineApi.VaccinePlanResponse): boolean => {
  // 简单判断:从模板生成的计划通常ID长度不同或有特定前缀
  // 这里可以根据实际情况调整判断逻辑
  return plan.vaccineName.includes('自定义') || !(plan.description || '').includes('个月接种')
}

// 编辑计划
const handleEdit = (plan: vaccineApi.VaccinePlanResponse) => {
  isEdit.value = true
  editPlanId.value = plan.planId
  form.value = {
    vaccineName: plan.vaccineName,
    vaccineType: plan.vaccineType,
    ageInMonths: plan.ageInMonths,
    doseNumber: plan.doseNumber,
    reminderDays: plan.reminderDays,
    isRequired: plan.isRequired,
    description: plan.description || ''
  }
  showAddDialog.value = true
}

// 删除计划
const handleDelete = (plan: vaccineApi.VaccinePlanResponse) => {
  uni.showModal({
    title: '确认删除',
    content: `确定要删除"${plan.vaccineName}"吗?`,
    success: async (res) => {
      if (res.confirm) {
        // ⚠️ API 层暂不支持删除疫苗计划
        uni.showToast({
          title: '暂不支持删除计划',
          icon: 'none'
        })
        // TODO: 等待后端API支持后实现
        // try {
        //   await vaccineApi.apiDeleteVaccinePlan(plan.planId)
        //   await loadVaccinePlans()
        //   uni.showToast({ title: '删除成功', icon: 'success' })
        // } catch (error: any) {
        //   uni.showToast({ title: error.message || '删除失败', icon: 'none' })
        // }
      }
    }
  })
}

// 提交表单
const handleSubmit = async () => {
  // 验证
  if (!form.value.vaccineName.trim()) {
    uni.showToast({
      title: '请输入疫苗名称',
      icon: 'none'
    })
    return
  }

  if (!form.value.vaccineType.trim()) {
    uni.showToast({
      title: '请输入疫苗类型',
      icon: 'none'
    })
    return
  }

  if (!currentBabyId.value) {
    uni.showToast({
      title: '请先选择宝宝',
      icon: 'none'
    })
    return
  }

  // ⚠️ API 层暂不支持创建/更新疫苗计划
  uni.showToast({
    title: isEdit.value ? '暂不支持编辑计划' : '暂不支持添加自定义计划',
    icon: 'none'
  })

  // TODO: 等待后端API支持后实现
  // if (isEdit.value) {
  //   try {
  //     await vaccineApi.apiUpdateVaccinePlan(editPlanId.value, { ... })
  //     await loadVaccinePlans()
  //     showAddDialog.value = false
  //     resetForm()
  //     uni.showToast({ title: '更新成功', icon: 'success' })
  //   } catch (error: any) {
  //     uni.showToast({ title: error.message || '更新失败', icon: 'none' })
  //   }
  // } else {
  //   try {
  //     await vaccineApi.apiCreateVaccinePlan({ ... })
  //     await loadVaccinePlans()
  //     showAddDialog.value = false
  //     resetForm()
  //     uni.showToast({ title: '创建成功', icon: 'success' })
  //   } catch (error: any) {
  //     uni.showToast({ title: error.message || '创建失败', icon: 'none' })
  //   }
  // }
}

// 重置表单
const resetForm = () => {
  isEdit.value = false
  editPlanId.value = ''
  form.value = {
    vaccineName: '',
    vaccineType: '',
    ageInMonths: 0,
    doseNumber: 1,
    reminderDays: 7,
    isRequired: true,
    description: ''
  }
}

// 页面加载
onMounted(async () => {
  if (!currentBaby.value) {
    uni.showToast({
      title: '请先选择宝宝',
      icon: 'none'
    })
    setTimeout(() => {
      uni.navigateBack()
    }, 1500)
    return
  }

  // 加载疫苗计划
  await loadVaccinePlans()
})
</script>

<style lang="scss" scoped>
.vaccine-plan-manage-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 20rpx;
  padding-bottom: 40rpx;
}

.action-bar {
  background: white;
  border-radius: 16rpx;
  padding: 20rpx;
  margin-bottom: 20rpx;
}

.button-content {
  display: flex;
  align-items: center;
  gap: 8rpx;
}

.plan-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.plan-item {
  background: white;
  border-radius: 16rpx;
  padding: 24rpx;
}

.plan-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}

.plan-name {
  font-size: 28rpx;
  font-weight: bold;
  color: #1a1a1a;
  display: flex;
  align-items: center;
  gap: 8rpx;
}

.required-badge, .custom-badge {
  display: inline-block;
  padding: 4rpx 8rpx;
  color: white;
  font-size: 20rpx;
  border-radius: 4rpx;
}

.required-badge {
  background: #fa2c19;
}

.custom-badge {
  background: #1890ff;
}

.plan-actions {
  display: flex;
  gap: 12rpx;
}

.plan-detail {
  display: flex;
  flex-wrap: wrap;
  gap: 20rpx;
  margin-bottom: 12rpx;
}

.detail-item {
  font-size: 24rpx;
  color: #666;

  .label {
    color: #999;
  }

  .value {
    color: #333;
    margin-left: 4rpx;
  }
}

.plan-desc {
  font-size: 24rpx;
  color: #666;
  line-height: 1.6;
}

.dialog-content {
  padding: 30rpx;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.dialog-title {
  font-size: 36rpx;
  font-weight: bold;
  text-align: center;
  margin-bottom: 30rpx;
}

.form-section {
  flex: 1;
  overflow-y: auto;
}

.dialog-footer {
  display: flex;
  gap: 20rpx;
  margin-top: 20rpx;
}
</style>
