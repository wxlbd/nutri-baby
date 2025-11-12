# 宝宝协作者与邀请系统 - 完整分析报告

## 一、核心数据模型关系图

### 1. Entity 层结构（去家庭化架构）

```
┌──────────────────────────────────────────────────────────────┐
│                      Backend Entities                        │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────┐         ┌──────────────────┐                 │
│  │  User    │         │     Baby         │                 │
│  ├──────────┤         ├──────────────────┤                 │
│  │ ID       │◄────────│ ID               │                 │
│  │ OpenID   │1    *   │ Name             │                 │
│  │ NickName │         │ BirthDate        │                 │
│  │ AvatarURL│         │ Gender           │                 │
│  └────▲─────┘         │ AvatarURL        │                 │
│       │               │ UserID (创建者)   │                 │
│       │               │ FamilyGroup      │                 │
│       │               │ CreatedAt        │                 │
│       │               │ UpdatedAt        │                 │
│       │               └────────┬─────────┘                 │
│       │                        │                            │
│       │        ┌───────────────┴───────────────┐            │
│       │        │                               │            │
│       │    1   │  BabyCollaborator             │ 1          │
│       │  ◄─────┤  (宝宝与用户的关系)            │────►      │
│       │        │                               │            │
│       │        │ ID                            │            │
│       │        │ BabyID (FK -> Baby.ID)        │            │
│       │        │ UserID (FK -> User.ID)        │            │
│       │        │ Role (admin/editor/viewer)    │            │
│       │        │ AccessType (permanent/temp)   │            │
│       │        │ ExpiresAt (null 永久)         │            │
│       │        │ CreatedAt                     │            │
│       │        │ UpdatedAt                     │            │
│       │        └───────────────────────────────┘            │
│       │                                                     │
│       │        ┌─────────────────────────────┐              │
│       │        │  BabyInvitation             │              │
│       │        │  (邀请流程管理)               │              │
│       └────────┤                             │              │
│                │ ID                          │              │
│                │ BabyID (FK -> Baby.ID)      │              │
│                │ UserID (FK -> User.ID)      │              │
│                │ Token (长token,分享用)      │              │
│                │ ShortCode (6位,二维码用)    │              │
│                │ InviteType (share/qrcode)   │              │
│                │ Role (角色)                 │              │
│                │ AccessType (访问类型)       │              │
│                │ ExpiresAt (权限过期时间)    │              │
│                │ CreatedAt                   │              │
│                └─────────────────────────────┘              │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### 2. 表结构详解

#### Baby 表（宝宝档案）
```go
type Baby struct {
    ID          int64    // 主键 (雪花ID)
    Name        string   // 宝宝名字
    Nickname    string   // 昵称
    BirthDate   string   // 出生日期 (YYYY-MM-DD)
    Gender      string   // 性别 (male/female)
    AvatarURL   string   // 头像URL
    Height      float64  // 身高 (cm)
    Weight      float64  // 体重 (kg)
    UserID      int64    // 创建者用户ID (FK)
    FamilyGroup string   // 家庭分组名称 (可选)
    CreatedAt   int64    // 创建时间(毫秒)
    UpdatedAt   int64    // 更新时间(毫秒)
    DeletedAt   soft_delete // 软删除时间
    
    // 关联
    Collaborators []*BabyCollaborator // 协作者列表
}
```

#### BabyCollaborator 表（协作者）
```go
type BabyCollaborator struct {
    ID         int64    // 主键 (雪花ID)
    BabyID     int64    // 宝宝ID (FK, 复合唯一索引)
    UserID     int64    // 用户ID (FK, 复合唯一索引)
    Role       string   // 角色: "admin", "editor", "viewer"
    AccessType string   // 访问类型: "permanent" 或 "temporary"
    ExpiresAt  *int64   // 临时权限过期时间(毫秒) - null表示永久
    CreatedAt  int64    // 创建时间(毫秒)
    UpdatedAt  int64    // 更新时间(毫秒)
    DeletedAt  soft_delete // 软删除时间
    
    // 关联
    User *User // 用户详情
    Baby *Baby // 宝宝详情
}

// 约束条件:
// - (BabyID, UserID) 唯一复合索引: idx_baby_user
// - 权限检查时会调用 IsExpired() 检查是否过期
```

#### BabyInvitation 表（邀请记录）
```go
type BabyInvitation struct {
    ID         int64    // 主键 (雪花ID)
    BabyID     int64    // 宝宝ID (FK)
    UserID     int64    // 邀请人用户ID (FK)
    Token      string   // 临时token (64位十六进制, 唯一索引)
    ShortCode  string   // 6位短码 (唯一索引, 用于二维码)
    InviteType string   // "share" 分享 或 "qrcode" 二维码
    Role       string   // 邀请的角色
    AccessType string   // 访问类型: "permanent" 或 "temporary"
    ExpiresAt  *int64   // 权限过期时间(毫秒) - null表示永久
    CreatedAt  int64    // 创建时间(毫秒)
    DeletedAt  soft_delete // 软删除时间
}

// 特点:
// - 同一用户对同一宝宝只有一个有效邀请
// - Token用于微信分享传递邀请信息
// - ShortCode用于二维码scene参数(受32字符限制)
```

---

## 二、协作者角色权限体系

### 1. 角色定义

| 角色 | admin | editor | viewer |
|------|-------|--------|--------|
| 管理宝宝信息 | ✅ | ❌ | ❌ |
| 邀请/移除协作者 | ✅ | ❌ | ❌ |
| 记录和编辑数据 | ✅ | ✅ | ❌ |
| 查看数据 | ✅ | ✅ | ✅ |
| 删除宝宝 | ✅ | ❌ | ❌ |
| 修改其他协作者角色 | ✅ | ❌ | ❌ |

### 2. 访问类型

- **permanent（永久）**: ExpiresAt = NULL, 协作关系永不过期
- **temporary（临时）**: ExpiresAt = 具体时间戳, 到期后权限自动失效

### 3. 权限检查流程

```go
// 在 BabyCollaborator 中定义的权限检查方法
IsExpired()  // 检查临时权限是否过期
IsAdmin()    // 检查是否为admin角色
CanEdit()    // 检查是否可编辑(admin 或 editor)

// 在 BabyCollaboratorRepository 中的检查方法
CheckPermission()    // 检查是否有权限访问,考虑过期时间
IsCollaborator()     // 检查是否为协作者
IsAdmin()            // 检查是否为管理员
CanEdit()            // 检查是否可编辑
```

---

## 三、邀请流程详解

### 工作流程图

```
┌─────────────────────────────────────────────────────────────────┐
│                        邀请流程                                  │
└─────────────────────────────────────────────────────────────────┘

1️⃣  邀请人(admin/editor)生成邀请
    ↓
    POST /babies/{babyId}/collaborators/invite
    │
    ├─ 检查权限: CanEdit(babyId, inviterID) ✓
    │
    ├─ 查询现有邀请: FindByBabyAndInviter()
    │   │
    │   ├─ 如果存在未使用邀请 → 直接返回现有邀请信息
    │   │
    │   └─ 如果不存在 → 生成新邀请
    │       │
    │       ├─ 生成 Token (64位十六进制)
    │       ├─ 生成 ShortCode (6位)
    │       ├─ 存储 BabyInvitation 记录
    │       └─ 调用微信生成二维码
    │
    ├─ 响应邀请DTO:
    │   ├─ babyId, babyName
    │   ├─ inviterName
    │   ├─ role, accessType, expiresAt
    │   ├─ shortCode
    │   └─ qrcodeParams { scene, page, qrcodeUrl }
    │
    └─ 返回客户端

2️⃣  被邀请人接收邀请
    ↓
    a) 微信分享路径:
       Page route: pages/baby/join/join?babyId=xxx&token=xxx
    
    b) 二维码路径:
       QR Code scene: c=ABC123
       解析后: pages/baby/join/join?scene=c=ABC123

    ↓
    GET /invitations/code/{shortCode}
    │
    ├─ 查询 BabyInvitation 记录
    ├─ 获取宝宝信息
    ├─ 获取邀请人信息
    │
    └─ 返回 InvitationDetailDTO:
       ├─ babyId, babyName, babyAvatar
       ├─ inviterName
       ├─ role, accessType, expiresAt
       └─ token (用于确认)

3️⃣  被邀请人确认加入
    ↓
    POST /babies/join
    │
    ├─ 请求: { babyId, token }
    │
    ├─ 查询邀请记录: FindByToken(token)
    │   └─ 校验: invitation.BabyID == babyId
    │
    ├─ 获取被邀请人的 UserID (通过openid)
    │
    ├─ 检查是否已是协作者
    │   └─ FindByBabyAndUser(babyId, userID)
    │
    ├─ 创建 BabyCollaborator 记录
    │   ├─ babyId = invitation.BabyID
    │   ├─ userId = invitedUser.ID
    │   ├─ role = invitation.Role
    │   ├─ accessType = invitation.AccessType
    │   └─ expiresAt = invitation.ExpiresAt
    │
    ├─ 成功: 返回宝宝详情 BabyDTO
    │
    └─ 该邀请后续可复用,不自动删除
```

### 邀请码的两种模式

#### 模式A: Token (用于微信分享)
```
特点:
- 长度: 64位十六进制字符串
- 传递方式: URL参数
- URL示例: pages/baby/join/join?babyId=123&token=abc123...
- 优点: 完整信息,安全性高
- 缺点: URL较长
```

#### 模式B: ShortCode (用于二维码)
```
特点:
- 长度: 6位字母数字
- 传递方式: 二维码场景值
- 二维码scene: c=ABC123
- 页面route: pages/baby/join/join?scene=c=ABC123
- 优点: 短码易于扫描和输入
- 缺点: 需要服务器查询

流程:
1. 前端获取shortCode和二维码图片
2. 用户扫描二维码或手动输入shortCode
3. 前端调用: GET /invitations/code/{shortCode}
4. 后端返回详细邀请信息
5. 前端显示邀请详情,用户确认
6. 用户点击确认,调用: POST /babies/join { babyId, token }
```

---

## 四、后端实现架构

### 1. Repository 层接口

#### BabyCollaboratorRepository
```go
interface BabyCollaboratorRepository {
    Create(ctx, collaborator) error                    // 创建协作者
    FindByBabyID(ctx, babyID) ([]*BabyCollaborator)    // 获取宝宝的所有协作者
    FindByUserID(ctx, userID) ([]*BabyCollaborator)    // 获取用户的所有协作宝宝
    FindByBabyAndUser(ctx, babyID, userID) (*BabyCollaborator)
    CheckPermission(ctx, babyID, userID) (*BabyCollaborator) // 考虑过期时间
    Update(ctx, collaborator) error                    // 更新协作者
    Delete(ctx, babyID, userID) error                  // 移除协作者
    BatchCreate(ctx, collaborators) error              // 批量创建(复制时用)
    CleanExpired(ctx) error                            // 清理过期的临时权限
    IsCollaborator(ctx, babyID, userID) (bool)         // 检查是否为协作者
    IsAdmin(ctx, babyID, userID) (bool)                // 检查是否为管理员
    CanEdit(ctx, babyID, userID) (bool)                // 检查是否有编辑权限
}
```

#### BabyInvitationRepository
```go
interface BabyInvitationRepository {
    Create(ctx, invitation) error
    FindByToken(ctx, token) (*BabyInvitation)
    FindByShortCode(ctx, shortCode) (*BabyInvitation)
    FindByBabyID(ctx, babyID) ([]*BabyInvitation)
    FindByBabyAndInviter(ctx, babyID, inviterID) (*BabyInvitation) // 关键:查询唯一邀请
    Delete(ctx, invitationID) error
    CleanExpired(ctx) error
}
```

### 2. Service 层业务逻辑

```go
type BabyService struct {
    babyRepo               BabyRepository
    collaboratorRepo       BabyCollaboratorRepository
    invitationRepo         BabyInvitationRepository
    userRepo               UserRepository
    vaccineScheduleService *VaccineScheduleService
    wechatService          *WechatService
    logger                 *zap.Logger
}

// 核心方法:
CreateBaby()              // 创建宝宝+自动成为admin
GetUserBabies()           // 获取用户可访问的宝宝列表
GetCollaborators()        // 获取宝宝的协作者列表
InviteCollaborator()      // 邀请协作者 (生成邀请码和二维码)
GetInvitationByShortCode() // 通过短码获取邀请详情
JoinBaby()                // 加入协作 (确认邀请)
RemoveCollaborator()      // 移除协作者
UpdateCollaboratorRole()  // 更新协作者角色
checkPermission()         // 权限检查
copyCollaborators()       // 复制协作者列表到新宝宝
```

### 3. DTO 数据传输层

#### 请求DTO
```go
// 邀请请求
type InviteCollaboratorRequest struct {
    InviteType string  // "share" 或 "qrcode"
    Role       string  // "admin", "editor", "viewer"
    AccessType string  // "permanent" 或 "temporary"
    ExpiresAt  *int64  // 仅当 accessType=temporary 时
}

// 加入请求
type JoinBabyRequest struct {
    BabyID string  // 宝宝ID
    Token  string  // 邀请token
}
```

#### 响应DTO
```go
// 协作者DTO (用于列表展示)
type CollaboratorDTO struct {
    OpenID     string  // 用户openid
    NickName   string  // 用户昵称
    AvatarURL  string  // 用户头像
    Role       string  // 角色
    AccessType string  // 访问类型
    ExpiresAt  *int64  // 过期时间(临时权限)
    JoinTime   int64   // 加入时间
}

// 邀请DTO (用于生成邀请)
type BabyInvitationDTO struct {
    BabyID       string        // 宝宝ID
    Name         string        // 宝宝名字
    InviterName  string        // 邀请人昵称
    Role         string        // 角色
    ShortCode    string        // 6位短码
    QRCodeParams *QRCodeParams // 二维码参数
    ExpiresAt    *int64        // 权限过期时间
}

// 邀请详情DTO (通过短码查询返回)
type InvitationDetailDTO struct {
    BabyID      string  // 宝宝ID
    BabyName    string  // 宝宝名字
    BabyAvatar  string  // 宝宝头像
    InviterName string  // 邀请人昵称
    Role        string  // 角色
    AccessType  string  // 访问类型
    ExpiresAt   *int64  // 过期时间
    Token       string  // token(用于加入)
}
```

### 4. HTTP Handler 处理器

```
API 端点:

1. GET /babies                              // 获取用户的宝宝列表
2. POST /babies                             // 创建宝宝
3. GET /babies/{babyId}                    // 获取宝宝详情
4. PUT /babies/{babyId}                    // 更新宝宝信息
5. DELETE /babies/{babyId}                 // 删除宝宝

协作者相关:
6. GET /babies/{babyId}/collaborators       // 获取协作者列表
7. POST /babies/{babyId}/collaborators/invite // 邀请协作者
8. DELETE /babies/{babyId}/collaborators/{openid} // 移除协作者
9. PUT /babies/{babyId}/collaborators/{openid}/role // 更新角色

邀请相关:
10. GET /babies/{babyId}/qrcode?shortCode=xxx  // 生成二维码(可复用)
11. POST /babies/join                           // 加入宝宝协作
12. GET /invitations/code/{shortCode}           // 获取邀请详情
```

---

## 五、前端实现

### 1. TypeScript 类型定义

```typescript
// src/types/index.ts

interface BabyCollaborator {
    openid: string                    // 用户openid
    nickName: string                  // 昵称
    avatarUrl: string                 // 头像
    role: 'admin' | 'editor' | 'viewer'
    accessType: 'permanent' | 'temporary'
    expiresAt?: number                // 过期时间(毫秒)
    joinTime: number                  // 加入时间(毫秒)
}

type CollaboratorRole = 'admin' | 'editor' | 'viewer'
type AccessType = 'permanent' | 'temporary'
```

### 2. API 模块 (src/api/baby.ts)

关键API:
- `apiFetchBabyList()` - 获取宝宝列表
- `apiFetchBabyDetail(babyId)` - 获取宝宝详情
- `apiGenerateQRCode(babyId, shortCode)` - 生成二维码
- `apiGetInvitationByCode(shortCode)` - 通过短码获取邀请详情

### 3. Store 模块 (src/store/collaborator.ts)

核心函数:
```typescript
// 获取协作者列表
async fetchCollaborators(babyId: string): Promise<BabyCollaborator[]>

// 邀请协作者(生成邀请码和二维码)
async inviteCollaborator(
    babyId: string,
    inviteType: 'share' | 'qrcode',
    role: CollaboratorRole,
    accessType: AccessType,
    expiresAt?: number
): Promise<邀请DTO>

// 通过邀请加入宝宝协作
async joinBabyCollaboration(
    babyId: string,
    token: string
): Promise<{ babyId, name, role }>

// 移除协作者(仅admin)
async removeCollaborator(babyId: string, openid: string): Promise<boolean>

// 更新协作者角色(仅admin)
async updateCollaboratorRole(
    babyId: string,
    openid: string,
    role: CollaboratorRole
): Promise<boolean>

// 本地缓存管理
getCollaborators(babyId: string)  // 获取本地缓存
clearCollaborators(babyId?)       // 清除缓存
```

### 4. 前端页面

#### 宝宝列表页面 (pages/baby/list/list.vue)
- 展示用户可访问的宝宝列表
- 每个宝宝卡片有: 头像、名字、性别、年龄
- 按钮: 邀请协作、设为默认、编辑、删除
- 点击卡片切换宝宝

#### 邀请页面 (pages/baby/invite/invite.vue)
- 设置邀请参数: 角色(editor/viewer)、访问类型(永久/临时)、过期时间
- 生成邀请二维码
- 展示二维码,支持长按保存或扫描

#### 加入页面 (pages/baby/join/join.vue)
- 通过shortCode查询邀请详情
- 展示: 宝宝信息、邀请人、角色权限说明
- 确认或取消加入

---

## 六、关键流程代码

### 1. 创建宝宝流程

```go
func (s *BabyService) CreateBaby(ctx, openID, req) {
    // 1. 验证日期格式
    // 2. 获取用户ID (通过openid)
    // 3. 创建Baby实体
    // 4. 存储到数据库
    // 5. 创建者自动成为admin (BabyCollaborator)
    // 6. [可选] 复制协作者列表 (CopyCollaboratorsFrom)
    // 7. 初始化疫苗计划
}
```

### 2. 邀请流程

```go
func (s *BabyService) InviteCollaborator(ctx, babyID, openID, req) {
    // 1. 转换ID: babyID string → int64
    // 2. 获取邀请人用户ID
    // 3. 检查权限: CanEdit(babyID, userID)
    // 4. 获取宝宝信息
    // 5. 检查现有邀请: FindByBabyAndInviter()
    //    - 如果存在: 重新生成二维码,返回现有邀请
    //    - 如果不存在: 创建新邀请
    // 6. 生成Token (64位)
    // 7. 生成ShortCode (6位)
    // 8. 创建BabyInvitation记录
    // 9. 调用微信生成二维码
    // 10. 返回邀请DTO
}
```

### 3. 加入流程

```go
func (s *BabyService) JoinBaby(ctx, openID, req) {
    // 1. 转换ID: req.BabyID string → int64
    // 2. 查询邀请记录: FindByToken(token)
    // 3. 校验: invitation.BabyID == req.BabyID
    // 4. 获取被邀请人用户ID
    // 5. 检查是否已是协作者: FindByBabyAndUser()
    // 6. 创建BabyCollaborator记录
    //    - babyId = invitation.BabyID
    //    - userId = invitedUser.ID
    //    - role = invitation.Role
    //    - accessType = invitation.AccessType
    //    - expiresAt = invitation.ExpiresAt
    // 7. 返回宝宝详情
}
```

### 4. 权限检查

```go
// 在 BabyService.checkPermission() 中:
func (s *BabyService) checkPermission(ctx, babyIDInt64, openID) error {
    // 1. 获取用户ID (通过openid)
    // 2. 检查是否为协作者: IsCollaborator(babyID, userID)
    // 3. 如果不是: 返回 PermissionDenied 错误
}

// 在 BabyCollaboratorRepository.CheckPermission() 中:
func (r *babyCollaboratorRepositoryImpl) CheckPermission(ctx, babyID, userID) {
    // 1. 查询BabyCollaborator记录: FindByBabyAndUser()
    // 2. 检查是否过期: IsExpired()
    // 3. 如果过期: 返回nil (视为无权限)
    // 4. 返回协作者信息或nil
}
```

---

## 七、数据流向示例

### 例1: 邀请协作者完整流程

```
前端 (vue)                    后端 (Go)                      数据库

用户点击邀请按钮
    ↓
进入 invite.vue 页面
设置: role=editor, accessType=permanent
    ↓
POST /babies/{id}/collaborators/invite
    │                           ↓
    │                    BabyHandler.InviteCollaborator()
    │                           ↓
    │                    BabyService.InviteCollaborator()
    │                           ├─ checkPermission()
    │                           ├─ invitationRepo.FindByBabyAndInviter()
    │                           ├─ generateInvitationToken()  ("abc123...")
    │                           ├─ generateUniqueShortCode()  ("AB12CD")
    │                           ├─ invitationRepo.Create()
    │                           │       ↓
    │                           │   INSERT INTO baby_invitations
    │                           │   (id, baby_id, user_id, token, short_code, 
    │                           │    invite_type, role, access_type, expires_at)
    │                           │
    │                           ├─ wechatService.GenerateQRCode()
    │                           │       ↓
    │                           │   微信API: 生成小程序码
    │                           │
    │                           └─ 返回 BabyInvitationDTO
    │
    ←── { qrcodeUrl, shortCode, ... }
    │
展示二维码
长按保存或扫描
    ↓
被邀请人扫描二维码

(用户手机中微信识别二维码)
    ↓
页面跳转: pages/baby/join/join?scene=c=AB12CD
    ↓
GET /invitations/code/AB12CD
    │                           ↓
    │                    BabyHandler.GetInvitationByShortCode()
    │                           ↓
    │                    BabyService.GetInvitationByShortCode()
    │                           ├─ invitationRepo.FindByShortCode("AB12CD")
    │                           │       ↓
    │                           │   SELECT * FROM baby_invitations
    │                           │   WHERE short_code = 'AB12CD'
    │                           │
    │                           ├─ babyRepo.FindByID()
    │                           └─ userRepo.FindByID()  (邀请人)
    │
    ←── { babyId, babyName, inviterName, role, ... }
    │
展示邀请详情
用户点击"确认加入"
    ↓
POST /babies/join
body: { babyId: "123", token: "abc123..." }
    │                           ↓
    │                    BabyHandler.JoinBaby()
    │                           ↓
    │                    BabyService.JoinBaby()
    │                           ├─ invitationRepo.FindByToken("abc123...")
    │                           ├─ collaboratorRepo.FindByBabyAndUser()
    │                           ├─ collaboratorRepo.Create()
    │                           │       ↓
    │                           │   INSERT INTO baby_collaborators
    │                           │   (id, baby_id, user_id, role, 
    │                           │    access_type, expires_at)
    │                           │   VALUES (123, 456, 789, 'editor', 
    │                           │           'permanent', NULL)
    │                           │
    │                           └─ GetBabyDetail()
    │
    ←── { babyId, name, ... }
    │
加入成功,显示宝宝详情
用户可以开始记录数据
```

### 例2: 删除协作者的权限检查

```
前端调用:
DELETE /babies/{babyId}/collaborators/{targetOpenid}

后端流程:
BabyService.RemoveCollaborator()
    ├─ 获取操作人用户ID
    ├─ 检查操作人权限: IsAdmin(babyId, operatorUserID)
    │   └─ collaboratorRepo.CheckPermission()
    │       └─ 检查是否过期
    ├─ 获取目标用户ID
    ├─ 检查不能删除创建者: baby.UserID == targetUserID
    └─ collaboratorRepo.Delete(babyId, targetUserID)
```

---

## 八、当前前端宝宝列表展示方式

### list.vue 关键代码分析

```vue
<template>
  <!-- 宝宝卡片 (循环展示) -->
  <view v-for="baby in babyList" :key="baby.babyId" class="baby-card">
    
    <!-- 卡片内容 -->
    <view class="card-header" @click="handleSelectBaby(baby.babyId)">
      <!-- 头像 -->
      <image :src="baby.avatarUrl" />
      
      <!-- 基本信息 -->
      <view class="baby-info">
        <text>{{ baby.name }}</text>
        <text>{{ baby.gender === "male" ? "男宝" : "女宝" }}</text>
        <text>{{ calculateAge(baby.birthDate) }}</text>
      </view>
      
      <!-- 选中标记 -->
      <wd-icon v-if="baby.babyId === currentBabyId" name="check-circle-fill" />
    </view>
    
    <!-- 操作按钮 -->
    <view class="card-actions">
      <!-- 邀请协作按钮 -->
      <wd-button @click="handleInvite(baby.babyId, baby.name)">
        邀请协作
      </wd-button>
      
      <!-- 设为默认按钮 -->
      <wd-button @click="handleSetDefault(baby.babyId)">
        设为默认
      </wd-button>
      
      <!-- 编辑按钮 -->
      <wd-button @click="handleEdit(baby.babyId)">
        编辑
      </wd-button>
      
      <!-- 删除按钮 -->
      <wd-button @click="handleDelete(baby.babyId)">
        删除
      </wd-button>
    </view>
  </view>
</template>

<script setup>
// 关键数据:
const babyList = ref<BabyProfile[]>([])  // 宝宝列表
const currentBabyId = ref<string>()      // 当前选中的宝宝ID
const userInfo = ref<UserInfo>()         // 用户信息(包含defaultBabyId)

// 页面加载时获取列表:
onLoad(() => {
  loadBabyList()
})

// 获取宝宝列表 (只显示当前用户可访问的宝宝)
async function loadBabyList() {
  try {
    babyList.value = await apiFetchBabyList()
    // 返回的列表已经是用户有权限访问的宝宝
    // (后端通过 BabyRepository.FindByUserID(userID) 获取)
  } catch (error) {
    // 处理错误
  }
}

// 当前前端没有显示协作者列表
// 协作者管理页面目前还未实现
</script>
```

### 当前协作者展示的缺失点

基于代码审查,前端目前**没有展示协作者列表**的页面。缺失:

1. ❌ 显示宝宝的协作者列表
2. ❌ 显示协作者的角色和权限
3. ❌ 显示协作者的加入时间
4. ❌ 显示临时权限的过期时间
5. ❌ 协作者管理界面(移除、更新角色)

---

## 九、总结

### 去家庭化架构的优势

1. **灵活性**: 一个用户可以协作管理多个宝宝,一个宝宝可以有多个协作者
2. **细粒度控制**: 每个宝宝的协作者独立管理,支持不同角色权限
3. **临时权限**: 支持设置权限过期时间,适用于临时协作者
4. **邀请机制**: 通过微信分享或二维码邀请,无需提前注册
5. **权限继承**: 创建宝宝时支持复制现有协作者列表

### 核心概念对比

| 旧架构 | 新架构 |
|--------|--------|
| FamilyInfo | ❌ 已移除 |
| FamilyMember | BabyCollaborator |
| Invitation | BabyInvitation |
| 一个家庭 → 多个宝宝 | 一个宝宝 → 多个协作者 |

### 关键特性

1. **协作模式**: admin (创建者) ↔ editor/viewer (协作者)
2. **邀请方式**: 微信分享 (Token) + 二维码 (ShortCode)
3. **权限检查**: 考虑角色、访问类型、过期时间
4. **数据隔离**: 每个宝宝的数据独立,协作者基于权限访问

