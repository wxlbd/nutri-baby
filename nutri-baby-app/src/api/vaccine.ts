/**
 * 疫苗管理 API 接口
 * 职责: 纯 API 调用,无状态,无副作用
 */
import { get, post, put, del } from "@/utils/request";

// ============ 类型定义 ============

/**
 * 疫苗类型枚举
 */
export type VaccineType =
  | "BCG"
  | "HepB"
  | "OPV"
  | "DTaP"
  | "MMR"
  | "JE"
  | "MenA"
  | "MenAC"
  | "HepA"
  | "Varicella"
  | "InfluenzaA"
  | "Rotavirus"
  | "Pneumococcal"
  | "HIB"
  | "EV71"
  | "Other";

/**
 * 提醒状态
 */
export type VaccineReminderStatus =
  | "upcoming"
  | "due"
  | "overdue"
  | "completed";

/**
 * API 响应: 疫苗计划详情
 */
export interface VaccinePlanResponse {
  planId: string;
  vaccineName: string;
  vaccineType: VaccineType;
  ageInMonths: number;
  doseNumber: number;
  isRequired: boolean;
  reminderDays: number;
  description?: string;
  isCustom?: boolean; // 是否为自定义计划
  templateId?: string; // 模板ID
}

/**
 * API 响应: 疫苗接种记录详情
 */
export interface VaccineRecordResponse {
  recordId: string;
  babyId: string;
  planId: string;
  vaccineName: string;
  vaccineType: VaccineType;
  doseNumber: number;
  vaccineDate: number;
  hospital?: string;
  batchNumber?: string;
  doctor?: string;
  reaction?: string;
  note?: string;
  createBy: string;
  createTime: number;
}

/**
 * API 响应: 疫苗提醒详情
 */
export interface VaccineReminderResponse {
  reminderId: string;
  babyId: string;
  planId: string;
  vaccineName: string;
  doseNumber: number;
  scheduledDate: number;
  status: VaccineReminderStatus;
  isSent: boolean;
  sentTime?: number;
  completedTime?: number;
}

/**
 * API 响应: 疫苗计划列表
 */
export interface VaccinePlansListResponse {
  plans: VaccinePlanResponse[];
  total: number;
  completed?: number; // 已完成接种的计划数
  percentage?: number; // 完成百分比
}

/**
 * API 响应: 疫苗接种记录列表
 */
export interface VaccineRecordsListResponse {
  records: VaccineRecordResponse[];
  total: number;
  page: number;
  pageSize: number;
}

/**
 * API 响应: 疫苗提醒列表
 */
export interface VaccineRemindersListResponse {
  reminders: VaccineReminderResponse[];
  total: number;
}

/**
 * API 请求: 创建疫苗接种记录
 */
export interface CreateVaccineRecordRequest {
  babyId: string;
  planId: string;
  vaccineDate: number;
  hospital?: string;
  batchNumber?: string;
  doctor?: string;
  reaction?: string;
  note?: string;
}

/**
 * API 响应: 疫苗接种统计
 */
export interface VaccineStatsResponse {
  totalPlans: number;
  completedCount: number;
  upcomingCount: number;
  overdueCount: number;
  completionRate: number;
}

// ============ API 函数 ============

/**
 * 获取疫苗计划列表
 *
 * @param babyId 宝宝ID
 * @returns Promise<VaccinePlansListResponse>
 */
export async function apiFetchVaccinePlans(
  babyId: string,
): Promise<VaccinePlansListResponse> {
  const response = await get<VaccinePlansListResponse>(
    `/babies/${babyId}/vaccine-plans`,
  );
  return response.data || { plans: [], total: 0, completed: 0, percentage: 0 };
}

/**
 * 获取宝宝的疫苗接种记录列表
 *
 * ⚠️ 注意: 后端暂未实现该接口,需要添加 GET /v1/babies/:babyId/vaccine-records 路由
 *
 * @param params 查询参数
 * @returns Promise<VaccineRecordsListResponse>
 */
export async function apiFetchVaccineRecords(params: {
  babyId: string;
  page?: number;
  pageSize?: number;
}): Promise<VaccineRecordsListResponse> {
  // TODO: 后端需要实现 GET /v1/babies/:babyId/vaccine-records 接口
  const { babyId, ...queryParams } = params;
  const response = await get<VaccineRecordsListResponse>(
    `/babies/${babyId}/vaccine-records`,
    queryParams,
  );
  return response.data || { records: [], total: 0, page: 1, pageSize: 10 };
}

/**
 * 获取宝宝的疫苗提醒列表
 *
 * @param params 查询参数
 * @returns Promise<VaccineRemindersListResponse>
 */
export async function apiFetchVaccineReminders(params: {
  babyId: string;
  status?: VaccineReminderStatus[];
}): Promise<VaccineRemindersListResponse> {
  const { babyId, ...queryParams } = params;
  const response = await get<VaccineRemindersListResponse>(
    `/babies/${babyId}/vaccine-reminders`,
    queryParams,
  );
  return response.data || { reminders: [], total: 0 };
}

/**
 * 获取疫苗接种统计
 *
 * @param babyId 宝宝ID
 * @returns Promise<VaccineStatsResponse>
 */
export async function apiFetchVaccineStats(
  babyId: string,
): Promise<VaccineStatsResponse> {
  const response = await get<VaccineStatsResponse>(
    `/babies/${babyId}/vaccine-statistics`,
  );
  return (
    response.data || {
      totalPlans: 0,
      completedCount: 0,
      upcomingCount: 0,
      overdueCount: 0,
      completionRate: 0,
    }
  );
}

/**
 * 创建疫苗接种记录
 *
 * @param data 创建请求数据
 * @returns Promise<VaccineRecordResponse>
 */
export async function apiCreateVaccineRecord(
  data: CreateVaccineRecordRequest,
): Promise<VaccineRecordResponse> {
  const { babyId, ...recordData } = data;
  const response = await post<VaccineRecordResponse>(
    `/babies/${babyId}/vaccine-records`,
    recordData,
  );
  if (!response.data) {
    throw new Error(response.message || "创建疫苗接种记录失败");
  }
  return response.data;
}

/**
 * 更新疫苗接种记录
 *
 * ⚠️ 注意: 后端暂未实现该接口,需要添加 PUT /v1/vaccine-records/:recordId 路由
 *
 * @param recordId 记录ID
 * @param data 更新数据
 * @returns Promise<void>
 */
export async function apiUpdateVaccineRecord(
  recordId: string,
  data: Partial<CreateVaccineRecordRequest>,
): Promise<void> {
  // TODO: 后端需要实现 PUT /v1/vaccine-records/:recordId 接口
  await put(`/vaccine-records/${recordId}`, data);
}

/**
 * 删除疫苗接种记录
 *
 * ⚠️ 注意: 后端暂未实现该接口,需要添加 DELETE /v1/vaccine-records/:recordId 路由
 *
 * @param recordId 记录ID
 * @returns Promise<void>
 */
export async function apiDeleteVaccineRecord(recordId: string): Promise<void> {
  // TODO: 后端需要实现 DELETE /v1/vaccine-records/:recordId 接口
  await del(`/vaccine-records/${recordId}`);
}

/**
 * 标记疫苗提醒已发送
 *
 * ⚠️ 注意: 后端暂未实现该接口,需要添加 POST /v1/vaccine-reminders/:reminderId/mark-sent 路由
 *
 * @param reminderId 提醒ID
 * @returns Promise<void>
 */
export async function apiMarkReminderSent(reminderId: string): Promise<void> {
  // TODO: 后端需要实现 POST /v1/vaccine-reminders/:reminderId/mark-sent 接口
  await post(`/vaccine-reminders/${reminderId}/mark-sent`, {});
}
