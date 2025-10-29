/**
 * 宝宝数据状态管理
 * 职责: 状态管理 + 本地计算,API 调用委托给 api 层
 *
 * ⚠️ 向后兼容: 所有导出函数的签名保持不变,页面组件无需修改
 */
import { ref, computed } from "vue";
import type { BabyProfile } from "@/types";
import { StorageKeys, getStorage, setStorage } from "@/utils/storage";
import * as babyApi from "@/api/baby";
import { getUserInfo } from "./user";

// ============ 状态定义 ============

// 宝宝列表 - 延迟初始化
const babyList = ref<BabyProfile[]>([]);

// 当前选中的宝宝 ID - 延迟初始化
const currentBabyId = ref<string>("");

// 初始化标记
let isInitialized = false;

// 延迟初始化 - 仅在首次访问时从存储读取
function initializeIfNeeded() {
  if (!isInitialized) {
    babyList.value = getStorage<BabyProfile[]>(StorageKeys.BABY_LIST) || [];
    currentBabyId.value = getStorage<string>(StorageKeys.CURRENT_BABY_ID) || "";
    isInitialized = true;
  }
}

// 当前宝宝信息
const currentBaby = computed(() => {
  initializeIfNeeded(); // 确保数据已加载
  console.log("babyList.value", babyList.value);
  return (
    babyList.value.find((baby) => baby.babyId === currentBabyId.value) || null
  );
});

// ============ 本地查询函数 ============

/**
 * 获取宝宝列表(本地)
 *
 * ⚠️ 向后兼容: 函数签名保持不变
 */
export function getBabyList() {
  initializeIfNeeded();
  return babyList.value;
}

/**
 * 获取当前宝宝(本地)
 *
 * ⚠️ 向后兼容: 函数签名保持不变
 */
export function getCurrentBaby() {
  return currentBaby.value;
}

// ============ API 调用函数(委托给 api 层) ============

/**
 * 获取用户可访问的宝宝列表 (去家庭化架构)
 *
 * API: GET /babies
 *
 * ⚠️ 向后兼容: 函数签名保持不变
 */
export async function fetchBabyList(): Promise<BabyProfile[]> {
  initializeIfNeeded();
  try {
    const apiResponse = await babyApi.apiFetchBabyList();

    // 将 API 响应的字段映射到本地类型
    const babies: BabyProfile[] = apiResponse.map((baby) => ({
      babyId: baby.babyId,
      name: baby.name,
      nickname: baby.nickname,
      gender: baby.gender,
      birthDate: baby.birthDate,
      avatarUrl: baby.avatarUrl,
      creatorId: baby.creatorId,
      familyGroup: baby.familyGroup,
      createTime: baby.createTime,
      updateTime: baby.updateTime,
    }));

    babyList.value = babies;
    setStorage(StorageKeys.BABY_LIST, babies);

    // 设置当前宝宝的逻辑优化：
    // 1. 如果用户设置了默认宝宝且该宝宝在列表中,使用默认宝宝
    // 2. 如果没有默认宝宝或默认宝宝不在列表中,选中第一个
    // 3. 每次调用 fetchBabyList 都会重新设置 currentBabyId（用户手动切换除外）
    const userInfo = getUserInfo();
    const defaultBabyId = userInfo?.defaultBabyId;

    if (babies.length > 0) {
      if (defaultBabyId && babies.some((b) => b.babyId === defaultBabyId)) {
        setCurrentBaby(defaultBabyId);
      } else {
        setCurrentBaby(babies[0].babyId);
      }
    }

    return babies;
  } catch (error: any) {
    console.error("fetch baby list error:", error);
    throw error;
  }
}

/**
 * 获取宝宝详情 (去家庭化架构)
 *
 * API: GET /babies/{babyId}
 *
 * ⚠️ 向后兼容: 函数签名保持不变
 */
export async function fetchBabyDetail(babyId: string): Promise<BabyProfile> {
  try {
    const apiResponse = await babyApi.apiFetchBabyDetail(babyId);

    // 映射字段
    const baby: BabyProfile = {
      babyId: apiResponse.babyId,
      name: apiResponse.babyName,
      nickname: apiResponse.nickname,
      gender: apiResponse.gender,
      birthDate: apiResponse.birthDate,
      avatarUrl: apiResponse.avatarUrl,
      creatorId: apiResponse.creatorId,
      familyGroup: apiResponse.familyGroup,
      createTime: apiResponse.createTime,
      updateTime: apiResponse.updateTime,
    };

    // 更新本地列表
    const index = babyList.value.findIndex((b) => b.babyId === baby.babyId);
    if (index !== -1) {
      babyList.value[index] = baby;
    } else {
      babyList.value.push(baby);
    }
    setStorage(StorageKeys.BABY_LIST, babyList.value);

    return baby;
  } catch (error: any) {
    console.error("fetch baby detail error:", error);
    throw error;
  }
}

// ============ 增删改操作已移除 ============
// 页面组件应直接调用 API 层:
// - import * as babyApi from '@/api/baby'
// - await babyApi.apiCreateBaby(data)
// - await babyApi.apiUpdateBaby(id, data)
// - await babyApi.apiDeleteBaby(id)
//
// 调用成功后,需要手动刷新宝宝列表:
// - await fetchBabyList()

// ============ 本地操作函数 ============

/**
 * 设置当前宝宝
 *
 * ⚠️ 向后兼容: 函数签名保持不变
 */
export function setCurrentBaby(id: string) {
  currentBabyId.value = id;
  setStorage(StorageKeys.CURRENT_BABY_ID, id);
}

/**
 * 根据 ID 获取宝宝信息(本地)
 *
 * ⚠️ 向后兼容: 函数签名保持不变
 */
export function getBabyById(id: string): BabyProfile | null {
  return babyList.value.find((baby) => baby.babyId === id) || null;
}

/**
 * 清除宝宝数据 (用于登出)
 *
 * ⚠️ 向后兼容: 函数签名保持不变
 */
export function clearBabyData() {
  babyList.value = [];
  currentBabyId.value = "";
  setStorage(StorageKeys.BABY_LIST, []);
  setStorage(StorageKeys.CURRENT_BABY_ID, "");
}

// ============ 导出 ============

export { babyList, currentBabyId, currentBaby };
