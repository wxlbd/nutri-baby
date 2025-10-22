<template>
  <view class="vaccine-plan-manage-page">
    <!-- 顶部操作栏 -->
    <view class="action-bar">
      <nut-button type="primary" size="small" @click="showAddDialog = true">
        <view class="button-content">
          <nut-icon name="plus" />
          <text>添加自定义计划</text>
        </view>
      </nut-button>
    </view>

    <!-- 疫苗计划列表 -->
    <view class="plan-list">
      <view
        v-for="plan in vaccinePlans"
        :key="plan.id"
        class="plan-item"
      >
        <view class="plan-header">
          <view class="plan-name">
            <text class="required-badge" v-if="plan.isRequired">必打</text>
            <text class="custom-badge" v-if="isCustomPlan(plan)">自定义</text>
            {{ plan.vaccineName }}
          </view>
          <view class="plan-actions">
            <nut-button size="small" type="default" @click="handleEdit(plan)">编辑</nut-button>
            <nut-button v-if="isCustomPlan(plan)" size="small" type="danger" @click="handleDelete(plan)">删除</nut-button>
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
    <nut-popup
      v-model:visible="showAddDialog"
      position="bottom"
      :style="{ height: '80%' }"
      round
      closeable
    >
      <view class="dialog-content">
        <view class="dialog-title">{{ isEdit ? '编辑疫苗计划' : '添加疫苗计划' }}</view>

        <view class="form-section">
          <nut-form ref="formRef">
            <nut-form-item label="疫苗名称" required>
              <nut-input
                v-model="form.vaccineName"
                placeholder="请输入疫苗名称"
                clearable
              />
            </nut-form-item>

            <nut-form-item label="疫苗类型" required>
              <nut-input
                v-model="form.vaccineType"
                placeholder="例如: HepB, BCG, DTaP"
                clearable
              />
            </nut-form-item>

            <nut-form-item label="接种月龄" required>
              <nut-input
                v-model.number="form.ageInMonths"
                type="number"
                placeholder="请输入月龄"
                clearable
              />
            </nut-form-item>

            <nut-form-item label="剂次" required>
              <nut-input
                v-model.number="form.doseNumber"
                type="number"
                placeholder="请输入剂次"
                clearable
              />
            </nut-form-item>

            <nut-form-item label="提醒天数">
              <nut-input
                v-model.number="form.reminderDays"
                type="number"
                placeholder="提前几天提醒"
                clearable
              />
            </nut-form-item>

            <nut-form-item label="是否必打">
              <nut-switch v-model="form.isRequired" />
            </nut-form-item>

            <nut-form-item label="说明">
              <nut-textarea
                v-model="form.description"
                placeholder="疫苗说明(可选)"
                :max-length="200"
              />
            </nut-form-item>
          </nut-form>
        </view>

        <view class="dialog-footer">
          <nut-button
            type="default"
            size="large"
            block
            @click="showAddDialog = false"
          >
            取消
          </nut-button>
          <nut-button
            type="primary"
            size="large"
            block
            @click="handleSubmit"
          >
            {{ isEdit ? '保存' : '添加' }}
          </nut-button>
        </view>
      </view>
    </nut-popup>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { currentBaby, currentBabyId } from '@/store/baby'
import {
  vaccinePlans,
  createVaccinePlan,
  updateVaccinePlan,
  deleteVaccinePlan
} from '@/store/vaccine'
import type { VaccinePlan } from '@/types'

// 对话框状态
const showAddDialog = ref(false)
const isEdit = ref(false)
const editPlanId = ref('')

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

// 判断是否为自定义计划
const isCustomPlan = (plan: VaccinePlan): boolean => {
  // 简单判断:从模板生成的计划通常ID长度不同或有特定前缀
  // 这里可以根据实际情况调整判断逻辑
  return plan.vaccineName.includes('自定义') || !plan.description.includes('个月接种')
}

// 编辑计划
const handleEdit = (plan: VaccinePlan) => {
  isEdit.value = true
  editPlanId.value = plan.id
  form.value = {
    vaccineName: plan.vaccineName,
    vaccineType: plan.vaccineType,
    ageInMonths: plan.ageInMonths,
    doseNumber: plan.doseNumber,
    reminderDays: plan.reminderDays,
    isRequired: plan.isRequired,
    description: plan.description
  }
  showAddDialog.value = true
}

// 删除计划
const handleDelete = (plan: VaccinePlan) => {
  uni.showModal({
    title: '确认删除',
    content: `确定要删除"${plan.vaccineName}"吗?`,
    success: async (res) => {
      if (res.confirm) {
        const success = await deleteVaccinePlan(plan.id)
        if (success) {
          // 刷新列表
        }
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

  if (isEdit.value) {
    // 更新
    const success = await updateVaccinePlan(editPlanId.value, {
      vaccineName: form.value.vaccineName,
      description: form.value.description,
      ageInMonths: form.value.ageInMonths,
      doseNumber: form.value.doseNumber,
      isRequired: form.value.isRequired,
      reminderDays: form.value.reminderDays
    })

    if (success) {
      showAddDialog.value = false
      resetForm()
    }
  } else {
    // 创建
    const result = await createVaccinePlan(currentBabyId.value, {
      vaccineType: form.value.vaccineType,
      vaccineName: form.value.vaccineName,
      description: form.value.description,
      ageInMonths: form.value.ageInMonths,
      doseNumber: form.value.doseNumber,
      isRequired: form.value.isRequired,
      reminderDays: form.value.reminderDays
    })

    if (result) {
      showAddDialog.value = false
      resetForm()
    }
  }
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
onMounted(() => {
  if (!currentBaby.value) {
    uni.showToast({
      title: '请先选择宝宝',
      icon: 'none'
    })
    setTimeout(() => {
      uni.navigateBack()
    }, 1500)
  }
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
