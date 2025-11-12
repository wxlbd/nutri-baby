# 协作者系统快速参考

## 核心概念速查表

### 1. 三个核心实体关系

```
Baby (宝宝)
├── ID (int64) - 主键
├── UserID (int64) - 创建者用户ID
└── Collaborators []*BabyCollaborator

BabyCollaborator (协作者)
├── BabyID + UserID - 复合唯一键
├── Role: "admin" | "editor" | "viewer"
├── AccessType: "permanent" | "temporary"
└── ExpiresAt: *int64 (NULL=永久)

BabyInvitation (邀请)
├── Token (64位) - 微信分享用
├── ShortCode (6位) - 二维码用
├── Role 和 AccessType - 邀请权限信息
└── ExpiresAt - 权限过期时间
```

### 2. 角色权限对照

| 权限 | admin | editor | viewer |
|------|:-----:|:------:|:------:|
| 查看数据 | ✅ | ✅ | ✅ |
| 编辑/记录 | ✅ | ✅ | ❌ |
| 管理宝宝 | ✅ | ❌ | ❌ |
| 邀请协作 | ✅ | ❌ | ❌ |
| 移除协作 | ✅ | ❌ | ❌ |

### 3. 访问类型

- **permanent**: ExpiresAt = NULL, 永不过期
- **temporary**: ExpiresAt = 时间戳, 到期失效

---

## API 端点速查

### 宝宝管理
```
POST /babies                        创建宝宝
GET  /babies                        获取宝宝列表 (用户可访问的)
GET  /babies/{babyId}               获取宝宝详情
PUT  /babies/{babyId}               更新宝宝信息
DELETE /babies/{babyId}             删除宝宝 (需admin)
```

### 协作者管理
```
GET  /babies/{babyId}/collaborators              获取协作者列表
POST /babies/{babyId}/collaborators/invite       生成邀请码
DELETE /babies/{babyId}/collaborators/{openid}   移除协作者 (需admin)
PUT  /babies/{babyId}/collaborators/{openid}/role 更新角色 (需admin)
```

### 邀请处理
```
GET  /invitations/code/{shortCode}   通过shortCode查询邀请详情
POST /babies/join                     确认加入协作
GET  /babies/{babyId}/qrcode?shortCode=xxx  生成二维码
```

---

## 前端 Store 函数速查

### collaborator.ts

```typescript
// 获取协作者列表
fetchCollaborators(babyId: string)

// 邀请协作者
inviteCollaborator(
  babyId: string,
  inviteType: 'share' | 'qrcode',
  role: 'admin' | 'editor' | 'viewer',
  accessType: 'permanent' | 'temporary',
  expiresAt?: number
)

// 加入协作
joinBabyCollaboration(babyId: string, token: string)

// 移除协作者
removeCollaborator(babyId: string, openid: string)

// 更新角色
updateCollaboratorRole(babyId: string, openid: string, role: string)

// 缓存管理
getCollaborators(babyId: string)
clearCollaborators(babyId?: string)
```

---

## 邀请码生成逻辑

### Token 模式（微信分享）
```
用途: 微信分享邀请
长度: 64位十六进制字符串
形式: "abc123def456..."
传递: URL参数 ?babyId=123&token=abc123...
检查: 直接查询 invitationRepo.FindByToken(token)
```

### ShortCode 模式（二维码）
```
用途: 二维码scene参数
长度: 6位字母数字
形式: "AB12CD"
传递: 二维码scene="c=AB12CD"
流程:
  1. 扫码 → pages/baby/join/join?scene=c=AB12CD
  2. GET /invitations/code/AB12CD → 获取详情
  3. POST /babies/join { babyId, token } → 加入
```

---

## 邀请唯一性规则

**关键**: 同一用户对同一宝宝只有一个有效邀请

```go
// 查询现有邀请
invitation, err := invitationRepo.FindByBabyAndInviter(babyID, userID)

if invitation != nil {
    // 邀请已存在: 复用邀请信息
    // - 重新生成二维码并返回
    // - 不创建新的 BabyInvitation 记录
} else {
    // 邀请不存在: 创建新邀请
    // - 生成 Token (64位)
    // - 生成 ShortCode (6位)
    // - 创建 BabyInvitation 记录
    // - 调用微信生成二维码
}
```

---

## 权限检查流程

### 在何处检查

```
1. 生成邀请: CanEdit(babyId, userID) ✓
2. 移除协作: IsAdmin(babyId, userID) ✓
3. 更新角色: IsAdmin(babyId, userID) ✓
4. 删除宝宝: IsAdmin(babyId, userID) ✓
5. 查看任何数据: IsCollaborator(babyId, userID) ✓
```

### 过期检查

```go
// 在 BabyCollaborator 中
IsExpired() bool {
    if accessType != "temporary" || expiresAt == nil {
        return false
    }
    return time.Now().UnixMilli() > *expiresAt
}

// 在 Repository 中
CheckPermission(babyID, userID) {
    collab := FindByBabyAndUser(babyID, userID)
    if collab != nil && collab.IsExpired() {
        return nil  // 视为无权限
    }
    return collab
}
```

---

## 常见操作流程

### 1. 创建宝宝 → 自动成为 admin

```
POST /babies
  ↓
BabyService.CreateBaby()
  ├─ 创建 Baby 实体
  ├─ 插入 babies 表
  ├─ 创建 BabyCollaborator { role="admin", accessType="permanent" }
  └─ [可选] 复制其他宝宝的协作者列表
```

### 2. 邀请协作者流程

```
POST /babies/{babyId}/collaborators/invite
  ↓
1. 检查权限: CanEdit(babyId, operatorID)
2. 查询现有邀请: FindByBabyAndInviter(babyId, operatorID)
   - 存在: 重用并返回
   - 不存在: 创建新邀请
3. 生成 Token 和 ShortCode
4. 插入 baby_invitations 表
5. 生成二维码
6. 返回邀请DTO
```

### 3. 通过邀请加入

```
GET /invitations/code/{shortCode}
  ├─ 查询邀请记录
  ├─ 获取宝宝和邀请人信息
  └─ 返回邀请详情

POST /babies/join
  ├─ 验证 Token
  ├─ 检查是否已是协作者
  ├─ 创建 BabyCollaborator
  │  └─ 从邀请记录复制角色、访问类型、过期时间
  └─ 返回宝宝详情
```

### 4. 移除协作者

```
DELETE /babies/{babyId}/collaborators/{targetOpenid}
  ├─ 检查操作人权限: IsAdmin(babyId, operatorID)
  ├─ 校验: 不能删除创建者 (baby.UserID != targetUserID)
  ├─ 软删除 BabyCollaborator
  └─ 返回成功
```

---

## 数据库索引关键点

```
baby_collaborators:
  - (baby_id, user_id) UNIQUE  // 保证一个用户对一个宝宝只有一条记录

baby_invitations:
  - token UNIQUE              // 保证 token 唯一
  - short_code UNIQUE         // 保证短码唯一
  - (baby_id, user_id) INDEX  // 查询邀请人对该宝宝的邀请
```

---

## 前端页面对应关系

```
页面名称           文件路径                      功能
─────────────────────────────────────────────────────────
宝宝列表          pages/baby/list/list.vue       显示可访问宝宝
                                                 包含邀请按钮

邀请生成          pages/baby/invite/invite.vue   设置邀请参数
                                                 生成二维码

邀请加入          pages/baby/join/join.vue       展示邀请详情
                                                 确认加入

协作者管理        ❌ 未实现                      需要新建
```

---

## 常见问题排查

### Q: 为什么同一用户邀请同一宝宝多次不生成新邀请？
A: 邀请唯一性设计。调用 FindByBabyAndInviter() 查询现有邀请，存在则复用。

### Q: Token 和 ShortCode 的区别？
A: Token(64位)用于微信分享，ShortCode(6位)用于二维码(受32字符限制)。

### Q: 如何检查临时权限是否过期？
A: 调用 CheckPermission() 时自动检查 IsExpired()，过期返回 nil。

### Q: 删除协作者后能否撤销？
A: 是软删除，可通过直接数据库操作恢复。建议改为硬删除或添加撤销功能。

### Q: 一个用户能否同时是多个宝宝的协作者？
A: 可以。BabyCollaborator 表中 (BabyID, UserID) 复合唯一。

---

## 后端主要文件位置

```
实体层:
  internal/domain/entity/
    ├── baby.go                      // Baby + BabyCollaborator
    └── baby_invitation.go           // BabyInvitation

仓储层:
  internal/domain/repository/
    ├── user_repository.go           // BabyRepository 接口
    ├── baby_collaborator_repository.go
    └── baby_invitation_repository.go

实现层:
  internal/infrastructure/persistence/
    ├── baby_repository_impl.go
    ├── baby_collaborator_repository_impl.go
    └── baby_invitation_repository_impl.go

服务层:
  internal/application/service/
    └── baby_service.go              // 核心业务逻辑

DTO层:
  internal/application/dto/
    └── baby_dto.go                  // 所有请求/响应 DTO

处理器层:
  internal/interface/http/handler/
    └── baby_handler.go              // HTTP 端点实现
```

---

## 前端主要文件位置

```
类型定义:
  src/types/index.ts                // BabyCollaborator 类型

API 模块:
  src/api/baby.ts                   // API 调用函数

状态管理:
  src/store/collaborator.ts         // 协作者相关函数

页面:
  src/pages/baby/
    ├── list/list.vue               // 宝宝列表 ✅ 有邀请按钮
    ├── invite/invite.vue           // 邀请生成 ✅
    └── join/join.vue               // 邀请加入 ✅

缺失:
  协作者管理页面 (需要新建)
    └── collaborators/manage.vue    // 显示、编辑、删除协作者
```

