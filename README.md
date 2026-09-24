# gv-admin-atom 平台

基于 **Vue 3 + Element Plus + Go + SQLite3** 的前后端分离管理平台，支持登录鉴权、数据看板与用户管理。

## 技术栈

| 层 | 技术 |
| --- | --- |
| 前端 | Vue 3、Element Plus、Vue Router、ECharts、Axios、Vite |
| 后端 | Go 1.22+（标准库 net/http 路由）、JWT-less Token 鉴权 |
| 数据库 | SQLite3（modernc.org/sqlite，纯 Go 驱动，无需 CGO） |


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
go build -o gv-admin-atom .
./gv-admin-atom
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


## 环境变量

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| PORT | 8080 | 服务监听端口 |
| DB_PATH | data/dashboard.db | SQLite 数据库文件路径 |
