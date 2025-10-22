# API 集成改造指南

本文档说明如何将剩余的 Store 模块集成 API 接口。

## 📋 改造进度

### ✅ 已完成模块

1. **user.ts** - 用户认证
   - 微信登录 API
   - Token 刷新 API
   - 用户信息获取 API

2. **family.ts** - 家庭管理
   - 完整的家庭 CRUD API
   - 邀请码生成和加入 API
   - 成员管理 API

3. **baby.ts** - 宝宝档案
   - 完整的宝宝 CRUD API
   - 字段映射 (babyId ↔ id, babyName ↔ name)

4. **feeding.ts** - 喂养记录
   - 创建记录 API (POST /feeding-records)
   - 查询记录列表 API (GET /feeding-records)
   - 本地删除功能 (待API完善)

### 🔄 待完成模块

以下模块结构相似,可按照 `feeding.ts` 的模式进行改造:

#### 5. sleep.ts - 睡眠记录

**对应 API**:
- `POST /sleep-records` - 创建睡眠记录
- `GET /sleep-records` - 获取睡眠记录列表

**改造要点**:
```typescript
export async function addSleepRecord(
  record: Omit<SleepRecord, 'id' | 'createTime'>
): Promise<SleepRecord> {
  const response = await post<any>('/sleep-records', {
    babyId: record.babyId,
    startTime: record.startTime,
    endTime: record.endTime,
    duration: record.duration,
    quality: record.quality,
    note: record.note,
  })
  // ... 处理响应
}

export async function fetchSleepRecords(params: {
  babyId: string
  startTime?: number
  endTime?: number
  page?: number
  pageSize?: number
}): Promise<SleepRecord[]> {
  const response = await get<PagedResponse>('/sleep-records', params)
  // ... 处理响应
}
```

#### 6. diaper.ts - 换尿布记录

**对应 API**:
- `POST /diaper-records` - 创建换尿布记录
- `GET /diaper-records` - 获取换尿布记录列表

**字段映射**:
- `diaperType` → API 字段
- `changeTime` → API 字段

**改造要点**:
```typescript
export async function addDiaperRecord(
  record: Omit<DiaperRecord, 'id' | 'createTime'>
): Promise<DiaperRecord> {
  const response = await post<any>('/diaper-records', {
    babyId: record.babyId,
    diaperType: record.type, // wet/dirty/both
    note: record.note,
    changeTime: record.time,
  })
  // ... 处理响应
}
```

#### 7. growth.ts - 成长记录

**对应 API**:
- `POST /growth-records` - 创建成长记录
- `GET /growth-records` - 获取成长记录列表

**字段映射**:
- `headCircumference` → `headCircum` (API)

**改造要点**:
```typescript
export async function addGrowthRecord(
  record: Omit<GrowthRecord, 'id' | 'createTime'>
): Promise<GrowthRecord> {
  const response = await post<any>('/growth-records', {
    babyId: record.babyId,
    height: record.height, // cm
    weight: record.weight, // g
    headCircum: record.headCircumference, // cm
    note: record.note,
    recordTime: record.time,
  })
  // ... 处理响应
}
```

#### 8. vaccine.ts - 疫苗管理

**对应 API**:
- `GET /babies/{babyId}/vaccine-plans` - 获取疫苗计划
- `POST /babies/{babyId}/vaccine-records` - 创建疫苗接种记录
- `GET /babies/{babyId}/vaccine-reminders` - 获取疫苗提醒列表
- `GET /babies/{babyId}/vaccine-statistics` - 获取疫苗接种统计

**改造要点**:
```typescript
export async function fetchVaccinePlans(babyId: string) {
  const response = await get<any>(`/babies/${babyId}/vaccine-plans`)
  // ... 处理响应
}

export async function addVaccineRecord(
  babyId: string,
  record: VaccineRecord
): Promise<VaccineRecord> {
  const response = await post<any>(
    `/babies/${babyId}/vaccine-records`,
    {
      planId: record.planId,
      vaccineType: record.vaccineType,
      vaccineName: record.vaccineName,
      doseNumber: record.doseNumber,
      vaccineDate: record.vaccineDate,
      hospital: record.hospital,
      batchNumber: record.batchNumber,
      doctor: record.doctor,
      reaction: record.reaction,
      note: record.note,
    }
  )
  // ... 处理响应
}
```

## 🎯 统一改造模式

### 1. 引入 API 方法

```typescript
import { get, post, put, del } from '@/utils/request'
```

### 2. 创建 API 调用函数

```typescript
/**
 * 从服务器获取记录列表
 *
 * API: GET /xxx-records
 */
export async function fetchXxxRecords(params: {
  babyId: string
  startTime?: number
  endTime?: number
  page?: number
  pageSize?: number
}): Promise<XxxRecord[]> {
  try {
    const response = await get<{
      records: any[]
      total: number
      page: number
      pageSize: number
    }>('/xxx-records', params)

    if (response.code === 0 && response.data) {
      const records = response.data.records as XxxRecord[]

      // 更新本地缓存
      xxxRecords.value = records
      setStorage(StorageKeys.XXX_RECORDS, records)

      return records
    } else {
      throw new Error(response.message || '获取记录失败')
    }
  } catch (error: any) {
    console.error('fetch xxx records error:', error)
    throw error
  }
}

/**
 * 添加记录
 *
 * API: POST /xxx-records
 */
export async function addXxxRecord(
  record: Omit<XxxRecord, 'id' | 'createTime'>
): Promise<XxxRecord> {
  try {
    const response = await post<any>('/xxx-records', {
      babyId: record.babyId,
      // ... 其他字段映射
    })

    if (response.code === 0 && response.data) {
      const newRecord: XxxRecord = {
        ...record,
        id: response.data.recordId,
        createTime: response.data.createTime,
      }

      // 添加到本地列表
      xxxRecords.value.unshift(newRecord)
      setStorage(StorageKeys.XXX_RECORDS, xxxRecords.value)

      uni.showToast({
        title: '记录成功',
        icon: 'success',
      })

      return newRecord
    } else {
      throw new Error(response.message || '添加记录失败')
    }
  } catch (error: any) {
    console.error('add xxx record error:', error)
    uni.showToast({
      title: error.message || '记录失败',
      icon: 'none',
    })
    throw error
  }
}
```

### 3. 保留本地查询方法

```typescript
/**
 * 本地查询方法 (不调用 API)
 */
export function getXxxRecords(): XxxRecord[] {
  return xxxRecords.value
}

export function getXxxRecordsByBabyId(babyId: string): XxxRecord[] {
  return xxxRecords.value.filter((record) => record.babyId === babyId)
}

export function getTodayXxxRecords(babyId: string): XxxRecord[] {
  const todayStart = getTodayStart()
  const todayEnd = getTodayEnd()

  return xxxRecords.value.filter(
    (record) =>
      record.babyId === babyId &&
      record.time >= todayStart &&
      record.time <= todayEnd
  )
}
```

### 4. 待实现的 API 标注 TODO

```typescript
/**
 * 删除记录 (本地实现,待 API 完善)
 * TODO: 集成 DELETE /xxx-records/{recordId} API
 */
export function deleteXxxRecord(id: string): boolean {
  const index = xxxRecords.value.findIndex((record) => record.id === id)
  if (index === -1) {
    return false
  }

  xxxRecords.value.splice(index, 1)
  setStorage(StorageKeys.XXX_RECORDS, xxxRecords.value)
  return true
}
```

## 📝 字段映射规则

### API → 本地类型

| API 字段 | 本地字段 | 说明 |
|---------|---------|------|
| `recordId` | `id` | 记录 ID |
| `babyId` | `babyId` | 宝宝 ID |
| `feedingTime` | `time` | 喂养时间 |
| `changeTime` | `time` | 换尿布时间 |
| `recordTime` | `time` | 记录时间 |
| `startTime` | `startTime` | 开始时间 |
| `endTime` | `endTime` | 结束时间 |
| `createTime` | `createTime` | 创建时间 |
| `updateTime` | `updateTime` | 更新时间 |

### 特殊字段映射

- 宝宝档案: `babyName` ↔ `name`
- 成长记录: `headCircum` ↔ `headCircumference`

## ⚠️ 注意事项

1. **时间戳格式**: 所有时间戳使用毫秒级 Unix 时间戳
2. **分页查询**: 默认 `page=1`, `pageSize=20`
3. **错误处理**: 统一使用 try-catch + uni.showToast
4. **本地缓存**: 每次 API 调用成功后更新本地缓存
5. **Token 认证**: request.ts 已自动处理 Bearer Token
6. **离线支持**: 保留本地查询方法,支持离线访问缓存数据

## 🔧 环境变量配置

确保已配置 API 基础地址:

```bash
# .env.development
VITE_API_BASE_URL=https://api.nutribaby.com/v1

# .env.production
VITE_API_BASE_URL=https://api.nutribaby.com/v1
```

## 🚀 后续优化建议

1. **实现完整的 CRUD API** (更新、删除)
2. **添加离线队列机制** (网络恢复后自动同步)
3. **实现 WebSocket 实时推送** (多端数据同步)
4. **优化分页加载** (下拉刷新、上拉加载更多)
5. **添加数据缓存策略** (减少不必要的 API 调用)

## 📚 参考资料

- API 文档: `nutri-baby-app/API.md`
- 已完成模块:
  - `src/store/user.ts`
  - `src/store/family.ts`
  - `src/store/baby.ts`
  - `src/store/feeding.ts`
- 请求工具: `src/utils/request.ts`
- 类型定义: `src/types/index.ts`
