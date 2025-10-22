# 前端去家庭化改造总结文档

## 📋 改造概览

本次改造完全移除了前端的"家庭"概念，转向基于宝宝的协作模式，与后端去家庭化架构保持一致。

## ✅ 已完成的工作

### 1. TypeScript类型定义 (`src/types/index.ts`)

**已完成的更新：**
- ✅ 保留 `BabyProfile` 并更新字段 (`babyId`, `creatorId`, `familyGroup`)
- ✅ 新增 `BabyCollaborator` 类型 (替代 `FamilyMember`)
- ✅ 新增 `CollaboratorRole` 类型 (`admin` | `editor` | `viewer`)
- ✅ 新增 `AccessType` 类型 (`permanent` | `temporary`)
- ✅ 标记家庭相关类型为 `@deprecated`

### 2. 状态管理Store重构

#### ✅ 新增 `src/store/collaborator.ts`
- 宝宝协作者管理
- 邀请协作者功能 (微信分享/二维码)
- 加入宝宝协作
- 协作者权限管理 (添加/移除/更新角色)

#### ✅ 更新 `src/store/user.ts`
- 登录响应新增 `isNewUser` 标志
- 移除家庭相关检查逻辑
- 新增 `getIsNewUser()` 和 `setIsNewUser()` 方法

#### ✅ 更新 `src/store/baby.ts`
- `fetchBabyList()`: 改为 `GET /babies` (无需 familyId 参数)
- `addBaby()`: 支持 `familyGroup` 和 `copyCollaboratorsFrom` 参数
- 字段映射: `baby.id` → `baby.babyId`
- 移除所有 familyId 相关逻辑

#### ✅ 更新 `src/store/index.ts`
- 移除 `export * from './family'`
- 新增 `export * from './collaborator'`

#### ⚠️ 废弃 `src/store/family.ts`
- 保留文件但不再导出
- 待后续完全移除

### 3. 本地存储策略更新 (`src/utils/storage.ts`)

**已完成的更新：**
- ✅ 移除 `FAMILY_LIST`, `CURRENT_FAMILY_ID`, `FAMILY_MEMBERS`, `INVITATIONS`
- ✅ 新增 `COLLABORATORS_PREFIX` (按宝宝ID存储协作者)
- ✅ 标记废弃键为 `@deprecated`
- ✅ 新增 `clearDeprecatedFamilyData()` 函数用于清理旧数据

### 4. 用户登录流程重构

#### ✅ `src/App.vue`
- 启动时检查用户的 `babies` 数组
- 无宝宝 → 跳转欢迎页
- 有宝宝 → 跳转首页
- 移除家庭检查逻辑

#### ✅ `src/pages/user/login.vue`
- 登录成功后检查 `isNewUser` 标志
- 新用户 → 欢迎页
- 老用户无宝宝 → 欢迎页
- 老用户有宝宝 → 首页

#### ✅ `src/pages/welcome/welcome.vue`
- 引导新用户创建宝宝或加入协作
- 实现 `joinBabyCollaboration()` 调用

## 🔄 待完成的工作

### 1. 家庭管理页面重构

**当前状态：**
- 文件存在: `src/pages/family/family.vue`
- 需要重构为: 协作者管理页面

**待实现功能：**
- [ ] 列出当前宝宝的协作者
- [ ] 生成微信分享邀请
- [ ] 生成二维码邀请
- [ ] 移除协作者 (仅 admin 权限)
- [ ] 更新协作者角色 (仅 admin 权限)
- [ ] 查看协作者的权限信息

**建议方案：**
- 将 `family.vue` 重命名为 `collaborators.vue`
- 或创建新的 `src/pages/collaborators/` 目录

### 2. pages.json 配置更新

**需要更新的路由：**
```json
{
  "pages": [
    // 移除或重命名
    {
      "path": "pages/family/family",
      "style": {
        "navigationBarTitleText": "协作者管理" // 更新标题
      }
    },
    // 确保欢迎页已注册
    {
      "path": "pages/welcome/welcome",
      "style": {
        "navigationBarTitleText": "欢迎"
      }
    }
  ]
}
```

### 3. 其他页面中的家庭相关引用

**需要检查和更新的文件：**
- [ ] `src/pages/index/index.vue` - 首页仪表盘
- [ ] `src/pages/baby/list/list.vue` - 宝宝列表页
- [ ] `src/pages/baby/edit/edit.vue` - 编辑宝宝页
- [ ] `src/pages/user/user.vue` - 用户中心页
- [ ] `src/pages/record/*/*.vue` - 各类记录页面

**需要更新的逻辑：**
- 移除任何对 `family store` 的引用
- 移除 `currentFamilyId` 的使用
- 移除家庭成员选择逻辑
- 更新为基于宝宝的协作模式

### 4. API接口调用更新

**需要更新的接口调用：**
- [ ] 所有记录类接口 (喂养/睡眠/换尿布/成长)
  - 参数从 `familyId` 改为 `babyId`
  - 响应字段更新

**示例：**
```typescript
// 旧: POST /feeding-records?familyId=xxx
// 新: POST /feeding-records?babyId=xxx

// 旧: GET /babies?familyId=xxx
// 新: GET /babies (返回用户有权限的所有宝宝)
```

### 5. 数据迁移和清理

**需要执行的操作：**
- [ ] 在用户首次启动新版本时，调用 `clearDeprecatedFamilyData()`
- [ ] 提示用户数据已迁移到新架构
- [ ] 可选：提供数据导出功能以防万一

## 📝 API对接清单

### 后端已实现的API (需前端对接)

#### 认证相关
- `POST /v1/auth/wechat-login` ✅ 已对接
- `POST /v1/auth/refresh-token` ✅ 已对接
- `GET /v1/auth/user-info` ✅ 已对接

#### 宝宝管理
- `POST /v1/babies` ✅ 已对接 (需测试 `copyCollaboratorsFrom`)
- `GET /v1/babies` ✅ 已对接
- `GET /v1/babies/:babyId` ✅ 已对接
- `PUT /v1/babies/:babyId` ✅ 已对接
- `DELETE /v1/babies/:babyId` ✅ 已对接

#### 协作者管理
- `GET /v1/babies/:babyId/collaborators` ✅ Store已实现，待页面使用
- `POST /v1/babies/:babyId/collaborators/invite` ✅ Store已实现，待页面使用
- `POST /v1/babies/join` ✅ 已在欢迎页实现
- `DELETE /v1/babies/:babyId/collaborators/:openid` ✅ Store已实现，待页面使用
- `PUT /v1/babies/:babyId/collaborators/:openid/role` ✅ Store已实现，待页面使用

#### 记录管理
- `POST /v1/feeding-records` ⚠️ 需更新参数
- `GET /v1/feeding-records?babyId=xxx` ⚠️ 需更新参数
- `POST /v1/sleep-records` ⚠️ 需更新参数
- `GET /v1/sleep-records?babyId=xxx` ⚠️ 需更新参数
- `POST /v1/diaper-records` ⚠️ 需更新参数
- `GET /v1/diaper-records?babyId=xxx` ⚠️ 需更新参数
- `POST /v1/growth-records` ⚠️ 需更新参数
- `GET /v1/growth-records?babyId=xxx` ⚠️ 需更新参数

#### 疫苗管理
- `GET /v1/babies/:babyId/vaccine-plans` ⚠️ 需对接
- `POST /v1/babies/:babyId/vaccine-records` ⚠️ 需对接
- `GET /v1/babies/:babyId/vaccine-reminders` ⚠️ 需对接
- `GET /v1/babies/:babyId/vaccine-statistics` ⚠️ 需对接

## ⚠️ 重要注意事项

### 1. 兼容性处理
- 旧版本用户升级时，需要清理家庭相关缓存
- 建议在 App.vue 中检测版本号，执行一次性迁移

### 2. 权限控制
- admin: 完全控制 (删除宝宝、管理协作者)
- editor: 可添加/编辑记录
- viewer: 仅查看

### 3. 微信小程序限制
- 微信分享需要在用户操作事件中触发
- 二维码生成可能需要后端支持

### 4. 数据同步
- 协作者列表需要实时更新
- 考虑使用 WebSocket 推送协作者变更

## 🎯 下一步建议

1. **优先级 P0 (阻塞上线):**
   - 重构家庭管理页面为协作者管理页面
   - 更新所有记录页面的API调用
   - 更新 pages.json 配置

2. **优先级 P1 (影响体验):**
   - 实现协作者邀请功能 (微信分享/二维码)
   - 实现协作者权限管理界面
   - 数据迁移和清理逻辑

3. **优先级 P2 (优化增强):**
   - 添加协作者变更通知
   - 优化邀请流程的用户体验
   - 添加权限说明和引导

## 📊 测试清单

### 功能测试
- [ ] 新用户首次登录流程
- [ ] 老用户登录后的宝宝列表加载
- [ ] 创建宝宝功能
- [ ] 加入协作功能
- [ ] 协作者管理 (添加/移除/更新角色)
- [ ] 各类记录的创建/查询

### 数据测试
- [ ] 旧版本用户升级后数据正常
- [ ] 家庭相关缓存已清理
- [ ] 宝宝列表正确显示
- [ ] 协作者信息正确显示

### 权限测试
- [ ] admin 可以删除宝宝
- [ ] editor 可以添加/编辑记录
- [ ] viewer 仅可查看
- [ ] 非协作者无法访问

---

**文档版本:** v1.0
**更新时间:** 2025-10-22
**作者:** Claude Code AI Assistant
