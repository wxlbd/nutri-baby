# 宝宝喂养日志 - 云端同步 API 接口文档

## 接口说明

### 基础信息

- **Base URL**: `https://api.nutribaby.com/v1`
- **数据格式**: JSON
- **认证方式**: Bearer Token (JWT)
- **字符编码**: UTF-8

### 通用响应格式

```json
{
  "code": 0,           // 0: 成功, 其他: 错误码
  "message": "success", // 响应消息
  "data": {},          // 响应数据
  "timestamp": 1234567890
}
```

### 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 未授权 |
| 1003 | 资源不存在 |
| 1004 | 数据冲突 |
| 1005 | 权限不足 |
| 2001 | 服务器内部错误 |

---

## 1. 用户认证

### 1.1 微信登录

**接口**: `POST /auth/wechat-login`

**请求参数**:
```json
{
  "code": "微信登录code",
  "nickName": "昵称(可选)",
  "avatarUrl": "头像URL(可选)"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "JWT Token",
    "userInfo": {
      "openid": "用户openid",
      "nickName": "昵称",
      "avatarUrl": "头像URL",
      "createTime": 1234567890,
      "lastLoginTime": 1234567890
    },
    "families": [
      {
        "familyId": "家庭ID",
        "familyName": "家庭名称",
        "description": "家庭描述(可选)",
        "role": "admin", // admin: 管理员, member: 成员
        "joinTime": 1234567890
      }
    ]
  },
  "timestamp": 1234567890
}
```

### 1.2 刷新 Token

**接口**: `POST /auth/refresh-token`

**Headers**: `Authorization: Bearer {token}`

**说明**: 从JWT Token中自动获取用户openid,无需请求参数

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "新的JWT Token",
    "expiresIn": 7200
  }
}
```

### 1.3 获取用户信息

**接口**: `GET /auth/user-info`

**Headers**: `Authorization: Bearer {token}`

**说明**: 从JWT Token中自动获取用户openid,无需请求参数

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "openid": "用户openid",
    "nickName": "昵称",
    "avatarUrl": "头像URL",
    "defaultBabyId": "默认宝宝ID (可选,如未设置则为空字符串)",
    "createTime": 1234567890,
    "lastLoginTime": 1234567890
  }
}
```

### 1.4 设置默认宝宝

**接口**: `PUT /auth/default-baby`

**Headers**: `Authorization: Bearer {token}`

**说明**: 设置用户的默认宝宝,每次进入应用时自动选中该宝宝

**请求参数**:
```json
{
  "babyId": "宝宝ID"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

**错误响应**:
- `1001`: babyId参数错误或为空
- `1003`: 指定的宝宝不存在
- `1005`: 当前用户无权访问该宝宝

---

## 2. 家庭管理

> **重要说明 - 单家庭模式**:
> - 每个用户只能属于一个家庭
> - 首次登录时必须选择"创建家庭"或"加入家庭"
> - 已有家庭的用户不能再创建或加入其他家庭
> - 如需加入其他家庭,必须先退出当前家庭

### 2.1 获取我的家庭

**接口**: `GET /families/my`

**Headers**: `Authorization: Bearer {token}`

**说明**: 获取当前用户所属的家庭信息(单个)

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "familyId": "家庭ID",
    "familyName": "我们的家",
    "description": "家庭描述(可选)",
    "role": "admin",
    "joinTime": 1234567890,
    "members": [...],
    "babyIds": ["baby1", "baby2"]
  }
}
```

**用户无家庭时**:
```json
{
  "code": 404,
  "message": "用户尚未创建或加入家庭",
  "data": null
}
```

### 2.2 创建家庭

**接口**: `POST /families`

**Headers**: `Authorization: Bearer {token}`

**请求参数**:
```json
{
  "familyName": "我们的家",
  "description": "可选描述"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "familyId": "家庭ID",
    "familyName": "我们的家",
    "description": "家庭描述(可选)",
    "role": "admin",
    "joinTime": 1234567890,
    "members": [
      {
        "openid": "当前用户openid",
        "nickName": "昵称",
        "avatarUrl": "头像",
        "role": "admin",
        "joinTime": 1234567890
      }
    ],
    "babyIds": []
  }
}
```

**错误情况**:
```json
{
  "code": 400,
  "message": "您已经有家庭了,不能再创建新家庭",
  "data": null
}
```

### 2.3 获取家庭详情

**接口**: `GET /families/{familyId}`

**Headers**: `Authorization: Bearer {token}`

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "familyId": "家庭ID",
    "familyName": "我们的家",
    "createBy": "创建者openid",
    "createTime": 1234567890,
    "members": [
      {
        "memberId": "成员ID",
        "familyId": "家庭ID",
        "userId": "用户ID",
        "userInfo": {
          "openid": "成员openid",
          "nickName": "昵称",
          "avatarUrl": "头像URL",
          "createTime": 1234567890,
          "lastLoginTime": 1234567890
        },
        "role": "admin",
        "joinTime": 1234567890
      }
    ]
  }
}
```

### 2.4 更新家庭

**接口**: `PUT /families/{familyId}`

**Headers**: `Authorization: Bearer {token}`

**权限**: 仅管理员可操作

**请求参数**:
```json
{
  "familyName": "新的家庭名称",
  "description": "新的描述(可选)"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 2.5 解散家庭

**接口**: `DELETE /families/{familyId}`

**Headers**: `Authorization: Bearer {token}`

**权限**: 仅管理员可操作

**说明**: 解散后所有成员都会被移除,家庭数据将被删除

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 2.6 创建邀请

**接口**: `POST /families/invitations`

**Headers**: `Authorization: Bearer {token}`

**请求参数**:
```json
{
  "familyId": "家庭ID",
  "role": "member",  // admin 或 member
  "expireDays": 7    // 可选,默认7天
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "invitationCode": "ABCD1234",
    "familyId": "家庭ID",
    "familyName": "我们的家",
    "role": "member",
    "createBy": "创建者openid",
    "createTime": 1234567890,
    "expireTime": 1234657890
  }
}
```

### 2.7 加入家庭

**接口**: `POST /families/join`

**Headers**: `Authorization: Bearer {token}`

**请求参数**:
```json
{
  "invitationCode": "ABCD1234"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "familyId": "家庭ID",
    "familyName": "我们的家",
    "description": "家庭描述(可选)",
    "role": "member",
    "joinTime": 1234567890,
    "members": [...],
    "babyIds": [...]
  }
}
```

**错误情况**:
```json
{
  "code": 400,
  "message": "您已经有家庭了,请先退出当前家庭",
  "data": null
}
```

### 2.8 移除家庭成员

**接口**: `DELETE /families/{familyId}/members/{memberId}`

**Headers**: `Authorization: Bearer {token}`

**权限**: 仅管理员可操作

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 2.9 退出家庭

**接口**: `POST /families/{familyId}/leave`

**Headers**: `Authorization: Bearer {token}`

**说明**:
- 如果是管理员且家庭中有其他成员,需先转让管理员权限
- 退出后可以创建新家庭或加入其他家庭

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

**错误情况(管理员需先转让)**:
```json
{
  "code": 400,
  "message": "管理员需要先转让权限或解散家庭",
  "data": null
}
```

### 2.10 转让管理员权限

**接口**: `POST /families/{familyId}/transfer`

**Headers**: `Authorization: Bearer {token}`

**权限**: 仅管理员可操作

**请求参数**:
```json
{
  "newAdminId": "新管理员的openid"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## 3. 宝宝档案

### 3.1 创建宝宝档案

**接口**: `POST /babies`

**Headers**: `Authorization: Bearer {token}`

**请求参数**:
```json
{
  "familyId": "家庭ID",
  "babyName": "宝宝姓名",
  "gender": "male",  // male, female
  "birthDate": "2024-01-01",  // YYYY-MM-DD
  "avatarUrl": "头像URL(可选)"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "babyId": "宝宝ID",
    "familyId": "家庭ID",
    "babyName": "宝宝姓名",
    "gender": "male",
    "birthDate": "2024-01-01",
    "avatarUrl": "头像URL",
    "height": 0,
    "weight": 0,
    "createBy": "创建者openid",
    "createTime": 1234567890,
    "updateTime": 1234567890
  }
}
```

### 3.2 获取宝宝列表

**接口**: `GET /families/{familyId}/babies`

**Headers**: `Authorization: Bearer {token}`

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "babyId": "宝宝ID",
      "familyId": "家庭ID",
      "babyName": "宝宝姓名",
      "gender": "male",
      "birthDate": "2024-01-01",
      "avatarUrl": "头像URL",
      "height": 50,
      "weight": 3500,
      "createBy": "创建者openid",
      "createTime": 1234567890,
      "updateTime": 1234567890
    }
  ]
}
```

### 3.3 获取宝宝详情

**接口**: `GET /babies/{babyId}`

**Headers**: `Authorization: Bearer {token}`

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "babyId": "宝宝ID",
    "familyId": "家庭ID",
    "babyName": "宝宝姓名",
    "gender": "male",
    "birthDate": "2024-01-01",
    "avatarUrl": "头像URL",
    "height": 50,
    "weight": 3500,
    "createBy": "创建者openid",
    "createTime": 1234567890,
    "updateTime": 1234567890
  }
}
```

### 3.4 更新宝宝档案

**接口**: `PUT /babies/{babyId}`

**Headers**: `Authorization: Bearer {token}`

**请求参数**: (所有字段可选)
```json
{
  "babyName": "宝宝姓名",
  "gender": "male",
  "birthDate": "2024-01-01",
  "avatarUrl": "头像URL",
  "height": 50,    // cm
  "weight": 3500   // g
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 3.5 删除宝宝档案

**接口**: `DELETE /babies/{babyId}`

**Headers**: `Authorization: Bearer {token}`

**权限**: 仅管理员可操作

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## 4. 喂养记录

### 4.1 创建喂养记录

**接口**: `POST /feeding-records`

**Headers**: `Authorization: Bearer {token}`

**请求参数**:
```json
{
  "babyId": "宝宝ID",
  "feedingType": "breast",  // breast: 母乳, formula: 配方奶, mixed: 混合喂养
  "amount": 120,            // ml (奶瓶喂养时使用)
  "duration": 15,           // 分钟 (母乳喂养时使用)
  "detail": {
    "breastSide": "left",   // left, right, both (母乳喂养时使用)
    "leftTime": 10,         // 左侧时长(分钟)
    "rightTime": 5,         // 右侧时长(分钟)
    "formulaType": "惠氏"   // 奶粉类型
  },
  "note": "备注(可选)",
  "feedingTime": 1234567890 // 毫秒时间戳
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "recordId": "记录ID",
    "babyId": "宝宝ID",
    "feedingType": "breast",
    "amount": 120,
    "duration": 15,
    "detail": {
      "breastSide": "left",
      "leftTime": 10,
      "rightTime": 5,
      "formulaType": "惠氏"
    },
    "note": "备注",
    "feedingTime": 1234567890,
    "createBy": "创建者openid",
    "createTime": 1234567890
  }
}
```

### 4.2 获取喂养记录列表

**接口**: `GET /feeding-records`

**Headers**: `Authorization: Bearer {token}`

**Query参数**:
- `babyId`: 宝宝ID (必填)
- `startTime`: 开始时间戳(毫秒,可选)
- `endTime`: 结束时间戳(毫秒,可选)
- `page`: 页码,默认1
- `pageSize`: 每页数量,默认20

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "records": [
      {
        "recordId": "记录ID",
        "babyId": "宝宝ID",
        "feedingType": "breast",
        "amount": 120,
        "duration": 15,
        "detail": {},
        "note": "备注",
        "feedingTime": 1234567890,
        "createBy": "创建者openid",
        "createTime": 1234567890
      }
    ],
    "total": 100,
    "page": 1,
    "pageSize": 20
  }
}
```

---

## 5. 睡眠记录

### 5.1 创建睡眠记录

**接口**: `POST /sleep-records`

**Headers**: `Authorization: Bearer {token}`

**请求参数**:
```json
{
  "babyId": "宝宝ID",
  "startTime": 1234567890,  // 毫秒时间戳
  "endTime": 1234577890,    // 可选,进行中时为0或null
  "duration": 120,          // 时长(分钟),可选
  "quality": "good",        // good, fair, poor (可选)
  "note": "备注(可选)"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "recordId": "记录ID",
    "babyId": "宝宝ID",
    "startTime": 1234567890,
    "endTime": 1234577890,
    "duration": 120,
    "quality": "good",
    "note": "备注",
    "createBy": "创建者openid",
    "createTime": 1234567890
  }
}
```

### 5.2 获取睡眠记录列表

**接口**: `GET /sleep-records`

**Headers**: `Authorization: Bearer {token}`

**Query参数**: 同喂养记录

**响应**: 同喂养记录格式

---

## 6. 换尿布记录

### 6.1 创建换尿布记录

**接口**: `POST /diaper-records`

**Headers**: `Authorization: Bearer {token}`

**请求参数**:
```json
{
  "babyId": "宝宝ID",
  "diaperType": "pee",      // pee: 小便, poop: 大便, both: 两者
  "note": "备注(可选)",
  "changeTime": 1234567890  // 毫秒时间戳
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "recordId": "记录ID",
    "babyId": "宝宝ID",
    "diaperType": "pee",
    "note": "备注",
    "changeTime": 1234567890,
    "createBy": "创建者openid",
    "createTime": 1234567890
  }
}
```

### 6.2 获取换尿布记录列表

**接口**: `GET /diaper-records`

**Headers**: `Authorization: Bearer {token}`

**Query参数**: 同喂养记录

**响应**: 同喂养记录格式

---

## 7. 成长记录

### 7.1 创建成长记录

**接口**: `POST /growth-records`

**Headers**: `Authorization: Bearer {token}`

**请求参数**:
```json
{
  "babyId": "宝宝ID",
  "height": 50,              // 身高(cm),必填
  "weight": 3500,            // 体重(g),必填
  "headCircum": 35,          // 头围(cm),可选
  "note": "备注(可选)",
  "recordTime": 1234567890   // 毫秒时间戳
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "recordId": "记录ID",
    "babyId": "宝宝ID",
    "height": 50,
    "weight": 3500,
    "headCircum": 35,
    "note": "备注",
    "recordTime": 1234567890,
    "createBy": "创建者openid",
    "createTime": 1234567890
  }
}
```

### 7.2 获取成长记录列表

**接口**: `GET /growth-records`

**Headers**: `Authorization: Bearer {token}`

**Query参数**: 同喂养记录

**响应**: 同喂养记录格式

---

## 8. 疫苗接种管理

### 8.1 获取疫苗计划

**接口**: `GET /babies/{babyId}/vaccine-plans`

**Headers**: `Authorization: Bearer {token}`

**说明**: 获取宝宝的疫苗接种计划,根据出生日期自动计算预定日期和状态

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "plans": [
      {
        "planId": "计划ID",
        "vaccineType": "HepB",
        "vaccineName": "乙肝疫苗",
        "description": "出生24小时内接种",
        "ageInMonths": 0,
        "doseNumber": 1,
        "isRequired": true,
        "reminderDays": 3,
        "scheduledDate": 1234567890,  // 根据宝宝出生日期计算的预定日期
        "status": "pending"  // pending: 未接种, completed: 已接种, overdue: 已逾期
      }
    ],
    "total": 20,
    "completed": 5,
    "percentage": 25
  }
}
```

### 8.2 创建疫苗接种记录

**接口**: `POST /babies/{babyId}/vaccine-records`

**Headers**: `Authorization: Bearer {token}`

**请求参数**:
```json
{
  "planId": "计划ID",
  "vaccineType": "HepB",
  "vaccineName": "乙肝疫苗",
  "doseNumber": 1,
  "vaccineDate": 1234567890,    // 毫秒时间戳
  "hospital": "市妇幼保健院",
  "batchNumber": "202401001",   // 疫苗批号(可选)
  "doctor": "张医生",           // 接种医生(可选)
  "reaction": "无不良反应",     // 不良反应(可选)
  "note": "备注信息"            // 备注(可选)
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "recordId": "记录ID",
    "babyId": "宝宝ID",
    "planId": "计划ID",
    "vaccineType": "HepB",
    "vaccineName": "乙肝疫苗",
    "doseNumber": 1,
    "vaccineDate": 1234567890,
    "hospital": "市妇幼保健院",
    "batchNumber": "202401001",
    "doctor": "张医生",
    "reaction": "无不良反应",
    "note": "备注信息",
    "createBy": "创建者openid",
    "createTime": 1234567890
  }
}
```

### 8.3 获取疫苗提醒列表

**接口**: `GET /babies/{babyId}/vaccine-reminders`

**Headers**: `Authorization: Bearer {token}`

**Query参数**:
- `status`: 提醒状态筛选(可选) - upcoming/due/overdue
- `limit`: 返回数量限制(可选),默认10

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "reminders": [
      {
        "reminderId": "提醒ID",
        "babyId": "宝宝ID",
        "babyName": "宝宝姓名",
        "planId": "计划ID",
        "vaccineName": "乙肝疫苗",
        "doseNumber": 2,
        "scheduledDate": 1234567890,
        "status": "due",  // upcoming: 即将到期(7天以上), due: 7天内, overdue: 已逾期
        "daysUntilDue": 5,  // 距离预定日期的天数(负数表示逾期)
        "reminderSent": false,
        "createTime": 1234567890
      }
    ],
    "total": 15,
    "upcoming": 10,
    "due": 3,
    "overdue": 2
  }
}
```

### 8.4 获取疫苗接种统计

**接口**: `GET /babies/{babyId}/vaccine-statistics`

**Headers**: `Authorization: Bearer {token}`

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 20,           // 总计划数
    "completed": 5,        // 已完成数
    "pending": 13,         // 未完成数
    "overdue": 2,          // 已逾期数
    "percentage": 25,      // 完成百分比
    "nextVaccine": {
      "vaccineName": "乙肝疫苗",
      "doseNumber": 2,
      "scheduledDate": 1234567890,
      "daysUntilDue": 5
    },
    "recentRecords": [
      {
        "recordId": "记录ID",
        "babyId": "宝宝ID",
        "planId": "计划ID",
        "vaccineType": "BCG",
        "vaccineName": "卡介苗",
        "doseNumber": 1,
        "vaccineDate": 1234567890,
        "hospital": "市妇幼保健院",
        "batchNumber": "202401001",
        "doctor": "张医生",
        "reaction": "无不良反应",
        "note": "备注",
        "createBy": "创建者openid",
        "createTime": 1234567890
      }
    ]
  }
}
```

---

## 9. 数据同步

### 9.1 WebSocket同步

**接口**: `GET /sync`

**Headers**: `Authorization: Bearer {token}`

**说明**: WebSocket同步功能待实现,当前版本暂不支持

**响应**:
```json
{
  "code": 5001,
  "message": "WebSocket同步功能待实现",
  "data": null
}
```

**注意**:
- 数据同步功能计划通过WebSocket实现
- 目前建议使用各记录接口的GET方法定期拉取数据
- 未来版本将支持实时推送和批量同步

---

## 10. 统计分析

**说明**: 统计分析功能待实现,可通过记录列表接口获取数据后在客户端进行统计

---

## 11. 文件上传

**说明**: 文件上传功能待实现,当前版本暂不支持图片上传

---

## 数据库设计说明

### 关键表结构

1. **users** - 用户表
2. **families** - 家庭表
3. **family_members** - 家庭成员关系表
4. **babies** - 宝宝档案表
5. **feeding_records** - 喂养记录表
6. **sleep_records** - 睡眠记录表
7. **diaper_records** - 换尿布记录表
8. **growth_records** - 成长记录表
9. **vaccine_plans** - 疫苗计划表
10. **vaccine_records** - 疫苗接种记录表
11. **vaccine_reminders** - 疫苗提醒表
12. **invitations** - 邀请码表

### 关键字段映射

**宝宝表 (babies)**
- `baby_name` → `babyName` (JSON字段名)
- `birth_date` → `birthDate` (YYYY-MM-DD格式)
- 体重单位: 数据库存储为克(g)
- 身高单位: 数据库存储为厘米(cm)

**喂养记录表 (feeding_records)**
- `feeding_type` → `feedingType`: breast/formula/mixed
- `feeding_time` → `feedingTime`: 毫秒时间戳
- `detail` → JSON字段,包含breastSide/leftTime/rightTime/formulaType

**睡眠记录表 (sleep_records)**
- `start_time` → `startTime`: 毫秒时间戳
- `end_time` → `endTime`: 毫秒时间戳
- `quality` → `quality`: good/fair/poor

**换尿布记录表 (diaper_records)**
- `diaper_type` → `diaperType`: pee/poop/both
- `change_time` → `changeTime`: 毫秒时间戳

**成长记录表 (growth_records)**
- `head_circum` → `headCircum`: 头围,厘米(cm)
- `record_time` → `recordTime`: 毫秒时间戳
- 体重单位: 克(g)
- 身高单位: 厘米(cm)

**疫苗记录表 (vaccine_records)**
- `vaccine_date` → `vaccineDate`: 毫秒时间戳
- `batch_number` → `batchNumber`: 可选字段
- `create_by` → `createBy`: openid

### 时间戳说明

- **所有时间戳均使用毫秒级Unix时间戳**
- **日期字段(birthDate)使用 YYYY-MM-DD 格式字符串**
- 数据库存储时间戳为BIGINT类型

### 疫苗相关表结构详细说明

#### vaccine_plans (疫苗计划表)
系统预设的疫苗接种计划,适用于所有宝宝

#### vaccine_records (疫苗接种记录表)
用户的实际接种记录

#### vaccine_reminders (疫苗提醒表)
基于宝宝出生日期和计划生成的提醒

### 索引设计

**feeding_records 表索引**:
- PRIMARY KEY: `record_id`
- INDEX: `idx_baby_id` (baby_id)
- INDEX: `idx_feeding_time` (feeding_time)
- INDEX: `idx_create_time` (create_time)

**sleep_records 表索引**:
- PRIMARY KEY: `record_id`
- INDEX: `idx_baby_id` (baby_id)
- INDEX: `idx_start_time` (start_time)

**diaper_records 表索引**:
- PRIMARY KEY: `record_id`
- INDEX: `idx_baby_id` (baby_id)
- INDEX: `idx_change_time` (change_time)

**growth_records 表索引**:
- PRIMARY KEY: `record_id`
- INDEX: `idx_baby_id` (baby_id)
- INDEX: `idx_record_time` (record_time)

**vaccine_records 表索引**:
- PRIMARY KEY: `record_id`
- INDEX: `idx_baby_id` (baby_id)
- INDEX: `idx_plan_id` (plan_id)
- INDEX: `idx_vaccine_date` (vaccine_date)

**vaccine_reminders 表索引**:
- PRIMARY KEY: `reminder_id`
- INDEX: `idx_baby_id` (baby_id)
- INDEX: `idx_status` (status)
- INDEX: `idx_scheduled_date` (scheduled_date)
- UNIQUE INDEX: `idx_baby_plan` (baby_id, plan_id)

---

## WebSocket 实时推送

### 说明

WebSocket实时推送功能待实现,未来版本将支持以下特性:

### 连接地址 (计划)

`wss://api.nutribaby.com/ws?token={token}&familyId={familyId}`

### 消息类型 (计划)

- `record_created`: 新记录创建(喂养/睡眠/换尿布/成长/疫苗)
- `record_updated`: 记录更新
- `record_deleted`: 记录删除
- `vaccine_reminder`: 疫苗提醒通知
- `member_joined`: 成员加入
- `member_left`: 成员离开
- `sync_required`: 需要同步

**当前版本建议**: 使用轮询方式定期调用各记录列表接口获取最新数据

---

## 安全说明

1. **HTTPS**: 所有接口必须使用HTTPS
2. **Token认证**: 使用JWT Bearer Token认证
3. **Token过期**: Token有效期2小时,使用refresh-token接口刷新
4. **权限验证**: 每个接口都需验证家庭成员权限
5. **频率限制**: 建议单用户每分钟最多60次请求
6. **数据加密**: 敏感数据加密存储

---

## 版本说明

- **当前版本**: v1.0
- **Base URL**: `/v1`
- **更新日期**: 2025-01-20
- **维护者**: 开发团队

### 已实现功能

**v1.0 (2025-01-20)**
- ✅ 用户认证系统(微信登录、Token刷新、用户信息获取)
- ✅ 家庭协作功能(创建、列表、详情、更新、删除、邀请、加入、成员管理)
- ✅ 宝宝档案管理(CRUD操作)
- ✅ 喂养记录管理(创建、列表查询)
- ✅ 睡眠记录管理(创建、列表查询)
- ✅ 换尿布记录管理(创建、列表查询)
- ✅ 成长记录管理(创建、列表查询)
- ✅ 疫苗接种管理(计划查询、记录创建、提醒列表、统计)

### 待实现功能

- ⏳ WebSocket实时推送
- ⏳ 数据批量同步
- ⏳ 统计分析接口
- ⏳ 文件上传功能
- ⏳ 记录更新和删除接口

---

## 注意事项

1. **时间戳**: 所有时间戳使用毫秒级Unix时间戳
2. **日期格式**: birthDate使用 YYYY-MM-DD 格式
3. **分页**: 从1开始,默认每页20条
4. **删除操作**: 使用软删除,实际不删除数据
5. **Query参数**: babyId为必填参数(记录查询接口)
6. **单位约定**:
   - 奶量: 毫升(ml)
   - 体重: 克(g)
   - 身高/头围: 厘米(cm)
   - 时长: 分钟(min)
7. **字段命名**: 使用驼峰命名法(camelCase)
8. **认证方式**: 所有需要认证的接口使用 `Authorization: Bearer {token}` Header

