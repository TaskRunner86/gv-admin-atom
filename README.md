# GV Dashboard 物联网设备监控平台

基于 **Vue 3 + Element Plus + Go + SQLite3** 的前后端分离物联网（IoT）管理平台，支持登录鉴权、设备运行看板（统计卡片 / 上报趋势 / 设备类型与状态分布 / 最新告警）与用户管理（增删改查）。

## 技术栈

| 层 | 技术 |
| --- | --- |
| 前端 | Vue 3、Element Plus、Vue Router、ECharts、Axios、Vite |
| 后端 | Go 1.22+（标准库 net/http 路由）、JWT-less Token 鉴权 |
| 数据库 | SQLite3（modernc.org/sqlite，纯 Go 驱动，无需 CGO） |

## 业务模型

平台围绕物联网设备的接入与运行监控组织数据：

| 表 | 说明 |
| --- | --- |
| `devices` | 设备台账：编号、名称、类型、部署点位、运行状态（在线 / 离线 / 告警 / 维护）、固件版本、最后上报时间 |
| `daily_metrics` | 每日运行指标：数据上报消息量、在线设备数 |
| `alerts` | 设备告警记录：告警编号、关联设备、告警指标、级别（严重 / 警告 / 提示）、触发值、处理状态（未处理 / 处理中 / 已恢复）、告警时间 |
| `users` / `tokens` | 平台用户与登录令牌 |

设备类型覆盖环境传感器、智能电表、边缘网关、高清摄像头、温控器、智能水表、空气监测站、门禁控制器八类。

## 目录结构

```
gv-admin-atom/
├── backend/                    # Go 后端
│   ├── main.go                 # 服务入口（路由 + 静态资源托管）
│   └── internal/
│       ├── database/           # SQLite 初始化、建表、种子数据
│       ├── handlers/           # 登录、看板、用户 CRUD 接口
│       └── middleware/         # CORS、Token 鉴权
├── frontend/                   # Vue 3 前端
│   ├── src/
│   │   ├── api/                # Axios 封装与接口定义
│   │   ├── router/             # 路由与登录守卫
│   │   ├── components/         # ECharts 封装组件
│   │   └── views/              # 登录页 / 布局 / 设备看板 / 用户管理
│   └── vite.config.js          # 开发代理 /api -> :8080
└── data/                       # SQLite 数据文件（首次启动自动生成）
```

## 快速开始

### 1. 启动后端（:8080）

```bash
cd backend
go build -o gv-dashboard .
./gv-dashboard
# 或开发模式：go run .
```

启动时会自动创建 SQLite 数据库并写入演示数据（用户、24 台设备台账、30 天上报指标、36 条告警记录）。

### 2. 构建前端

```bash
cd frontend
npm install
npm run build      # 产出 dist/，后端会自动托管
```

> 仅开发调试前端时：`npm run dev`（:5173，已配置代理转发 /api 到 :8080）。

### 3. 访问

浏览器打开 <http://localhost:8080>

- 默认账号：`admin` / `admin123`
- 页面：设备看板（接入设备 / 在线设备 / 今日上报 / 今日告警卡片、7 天上报趋势、设备类型占比、30 天上报量、设备状态分布、最新告警列表）、用户管理（搜索、分页、增删改，仅管理员可见）

## 主要 API

| 方法 | 路径 | 说明 | 鉴权 |
| --- | --- | --- | --- |
| POST | /api/login | 登录，返回 token | 否 |
| GET | /api/health | 健康检查 | 否 |
| GET | /api/user/profile | 当前用户信息 | 是 |
| GET | /api/dashboard/overview | 看板聚合数据（设备统计 / 上报趋势 / 类型与状态分布 / 最新告警） | 是 |
| GET | /api/users | 用户分页列表（page/pageSize/keyword） | 管理员 |
| POST | /api/users | 新增用户 | 管理员 |
| PUT | /api/users/{id} | 更新用户（密码留空不改） | 管理员 |
| DELETE | /api/users/{id} | 删除用户（不可删自己/admin） | 管理员 |

鉴权方式：`Authorization: Bearer <token>`，token 有效期 24 小时。

### 权限模型

| 角色 | 权限 |
| --- | --- |
| 管理员（admin） | 设备看板、个人信息、用户管理的全部操作（查看 / 新增 / 编辑 / 删除） |
| 普通用户（user） | 仅设备看板与个人信息；用户管理前后端双重拦截 |

用户管理的三层拦截：

1. **接口层**：`/api/users*` 全部由 `middleware.RequireAdmin` 校验角色，非管理员返回 `403 {"code":403,"message":"无权限，用户管理仅限管理员使用"}`。
2. **路由层**：`/users` 路由带 `meta.adminOnly`，普通用户访问会被守卫重定向到设备看板。
3. **界面层**：侧栏「用户管理」菜单项仅对管理员渲染。

`GET /api/dashboard/overview` 返回结构：

```json
{
  "stats": { "totalDevices": 24, "onlineDevices": 16, "todayMessages": 23000, "todayAlerts": 3 },
  "trend": [{ "day": "2026-09-23", "messages": 23000, "devices": 15 }],
  "monthTrend": [{ "day": "2026-09-23", "messages": 23000, "devices": 15 }],
  "typeDist": [{ "name": "环境传感器", "value": 5 }],
  "statusDist": [{ "name": "在线", "value": 16 }],
  "recentAlerts": [
    {
      "id": 1, "alertNo": "AL202609230001", "deviceNo": "IOT-ENV-1001",
      "deviceName": "1号厂房温湿度传感器", "location": "1号厂房A区",
      "metric": "湿度超限", "level": "提示", "value": "92.4 %RH",
      "status": "未处理", "createdAt": "2026-09-23 06:12:00"
    }
  ]
}
```

## 环境变量

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| PORT | 8080 | 服务监听端口 |
| DB_PATH | data/dashboard.db | SQLite 数据库文件路径 |
