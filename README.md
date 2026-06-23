# 公考刷题工具（Teex）📝

一个面向公务员、事业编、三支一扶、乡村振兴协理员等考试的在线刷题系统。采用 **Go 后端 + Vue 3 前端** 的前后端分离架构，支持多用户、解析模式刷题、试卷模式、错题重做、统计分析、历史记录回溯等功能。

## 功能特性

- ✅ **多用户系统** — 注册/登录、JWT 认证、用户数据隔离、管理员权限
- ✅ **解析模式刷题** — 提交即反馈，逐题显示解析
- ✅ **题库管理** — 题目 CRUD + JSON 批量导入/删除 + 难度/标签筛选
- ✅ **考试类型管理** — 考试与科目 CRUD，每个考试独立的题库
- ✅ **错题重做** — 智能识别"最后一答错误"的题目
- ✅ **统计分析** — 右侧面板实时显示当前考试的进度与正确率（按用户隔离，30s 缓存）
- ✅ **历史记录** — 时间线式场次列表 + 详情弹窗 + 分页
- ✅ **键盘快捷键** — A/B/C/D 选选项，Enter 提交
- ✅ **设置中心** — 账户资料管理、修改昵称/密码、个人刷题概览、刷题偏好配置
- ✅ **数据导出/导入** — API 级别的数据备份能力
- ✅ **移动端适配** — 响应式布局
- ✅ **管理员权限控制** — 路由守卫 + API 中间件双重保护，管理页面仅管理员可访问
- ✅ **用户管理** — 管理员可创建/编辑/删除用户，分配角色，控制注册开关
- ✅ **单文件部署** — 前端通过 Go embed 嵌入二进制，一个文件即完整服务

### 支持的考试类型

| 考试类型 | 模块 |
|----------|------|
| 国家公务员 | 行测-言语理解、数量关系、判断推理、资料分析、申论 |
| 省级公务员 | 行测-言语理解、数量关系、判断推理、资料分析、申论、公共基础知识 |
| 事业编 | 公共基础知识、职业能力测验、综合应用能力 |
| 三支一扶 | 公共基础知识、职业能力测验 |
| 乡村振兴协理员 | 公共基础知识、农村工作知识、行测 |

## 技术栈

### 后端

| 组件 | 技术 | 版本 |
|------|------|------|
| 语言 | Go | 1.26.3 |
| HTTP 框架 | Gin | 1.12.0 |
| ORM | GORM | 1.31.1 |
| 数据库 | SQLite (WAL 模式) | glebarez/sqlite（纯 Go，无 CGO） |
| 认证 | JWT (HS256) | golang-jwt/v5 5.3.1 |
| 密码哈希 | bcrypt | golang.org/x/crypto 0.53.0 |
| 跨域 | CORS | gin-contrib/cors 1.7.7 |
| 缓存 | sync.Map + 30s TTL | 标准库（按用户隔离） |

### 前端

| 组件 | 技术 | 版本 |
|------|------|------|
| 框架 | Vue 3 (Composition API) | 3.5.34 |
| 路由 | Vue Router | 4.6.4 |
| HTTP 客户端 | Axios | 1.18.0 |
| 构建工具 | Vite | 8.0.12 |
| 状态管理 | reactive() 闭包 | 无需 Vuex/Pinia |

## 快速开始

### 环境要求

- Go >= 1.26
- Node.js >= 18（仅开发/构建前端时需要）

### 开发模式

```bash
# 终端 1：启动后端
go build -o server ./cmd/server/
./server
# 服务运行在 http://localhost:8080

# 终端 2：启动前端（热更新）
cd web
npm install
npm run dev
# 前端运行在 http://localhost:5173（自动代理 /api 到 :8080）
```

### 生产构建

```bash
# Linux / macOS
./build.sh

# Windows
build.bat

# 构建完成后，单文件启动
./server
# 前端已通过 go:embed 嵌入，无需 Nginx
```

> `build.sh` 会自动检测 Go 和 Node.js 环境（支持 NVM），安装依赖，构建前端，最后编译 Go 二进制。

### 默认账户

首次启动自动创建管理员账户：`admin` / `admin123`

### 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | `8080` | 服务端口（端口被占用时自动递增，最多尝试 20 次） |
| `DATA_DIR` | `.` | SQLite 数据库文件存放目录 |
| `JWT_SECRET` | 随机 32 字节 | JWT 签名密钥（未设置则每次重启随机生成） |
| `ADMIN_PASSWORD` | `admin123` | 默认管理员密码 |
| `CORS_ORIGINS` | （空=允许所有） | 允许的跨域来源，逗号分隔 |
| `GORM_LOG` | `false` | 设为 `true` 启用 GORM SQL 日志 |

## API 文档

### 公开接口（无需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/health` | 健康检查 |
| `POST` | `/api/auth/register` | 用户注册 `{username, password, nickname?}`（需管理员开启注册） |
| `POST` | `/api/auth/login` | 用户登录 `{username, password}` → 返回 `{token, user}` |
| `GET` | `/api/settings/registration` | 查询注册功能是否开启 |

> 以下接口均需在请求头中携带 `Authorization: Bearer <token>`

### 用户信息

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/profile` | 获取当前用户信息 |
| `PUT` | `/api/profile` | 修改昵称 `{nickname}` |
| `PUT` | `/api/profile/password` | 修改密码 `{old_password, new_password}` |

### 考试类型

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/exams` | 获取所有考试类型 |
| `POST` | `/api/exams` | 创建考试类型（名称唯一校验） |
| `PUT` | `/api/exams/:id` | 更新考试类型 |
| `DELETE` | `/api/exams/:id` | 删除考试类型（级联删除模块、题目、答题记录） |
| `GET` | `/api/exams/:id/modules` | 获取模块列表（含题目数 / 当前用户未做数） |
| `GET` | `/api/exams/:id/stats` | 获取某考试类型下各模块的统计数据 |

### 模块

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/modules` | 创建模块（验证考试类型存在 + 名称唯一） |
| `PUT` | `/api/modules/:id` | 更新模块 |
| `DELETE` | `/api/modules/:id` | 删除模块（级联删除） |

### 题目管理

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/questions` | 查询题目列表（支持 `module_id`, `exam_type_id`, `type`, `difficulty`, `page`, `size` 筛选） |
| `GET` | `/api/questions/:id` | 获取单个题目 |
| `POST` | `/api/questions` | 创建题目（验证类型/难度/必填字段） |
| `PUT` | `/api/questions/:id` | 更新题目 |
| `DELETE` | `/api/questions/:id` | 删除题目（级联删除答题记录） |
| `POST` | `/api/questions/import` | 批量导入题目（逐条校验，上限 500 条） |
| `DELETE` | `/api/questions/batch` | 批量删除题目 `{ids: [1, 2, 3]}`（上限 500 条） |

### 刷题

选题策略（`mode` 参数）：
- `default` — 未做优先，已做过的补随机
- `wrong` — 仅错题（最后一次答题为错误的）
- `random` — 纯随机

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/quiz/start` | 开始刷题 `{module_id, count, mode, difficulty?, tags?}`（count: 1-200，默认 10） |
| `POST` | `/api/quiz/answer` | 提交单题答案 `{question_id, user_input, duration, session_id?}` |
| `POST` | `/api/quiz/submit-batch` | 批量提交答案（试卷模式） `{answers, session_id?}`（上限 500） |

### 统计（按当前用户隔离）

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/stats` | 全局统计（正确率取每题最后一次答题） |
| `GET` | `/api/stats/module/:id` | 某模块统计 |

### 考试场次（按当前用户隔离）

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/sessions` | 场次列表（支持 `page`, `size` 分页） |
| `GET` | `/api/sessions/:id` | 单个场次详情 |
| `GET` | `/api/sessions/:id/answers` | 某场次的答题记录（分页） |

### 数据管理

| 方法 | 路径 | 说明 |
|------|------|------|
| `DELETE` | `/api/records` | 清空当前用户的答题记录和考试场次 |
| `GET` | `/api/export` | 导出全部考试类型、模块和题目为 JSON |
| `POST` | `/api/import` | 导入考试类型、模块和题目 |

### 管理员接口（需 admin 角色）

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/admin/users` | 获取用户列表 |
| `POST` | `/api/admin/users` | 创建用户 `{username, password, nickname?, role?}` |
| `PUT` | `/api/admin/users/:id` | 更新用户（昵称、密码、角色） |
| `DELETE` | `/api/admin/users/:id` | 删除用户（不可删除自身） |
| `PUT` | `/api/admin/settings/registration` | 开启/关闭公开注册 `{enabled: true/false}` |

### 统一响应格式

```json
// 成功
{ "data": ... }
{ "data": ..., "total": 100 }
{ "message": "删除成功" }

// 错误
{ "error": "操作失败，请稍后重试" }
```

## 项目结构

```
teex/
├── cmd/server/
│   └── main.go                       # 程序入口：初始化 DB、种子数据、路由、静态文件、graceful shutdown
├── internal/
│   ├── cache/
│   │   └── cache.go                  # sync.Map 内存缓存（30s TTL，按用户隔离）
│   ├── database/
│   │   ├── db.go                     # 数据库初始化（WAL 模式 + 连接池 + 自动建表）
│   │   ├── seeds.go                  # 种子数据加载 + 默认管理员账户创建
│   │   └── seeddata/
│   │       ├── exams.json            # 5 种考试类型 + 19 个模块
│   │       └── questions_sample.json # 43 道示例题目
│   ├── handler/
│   │   ├── answer_handler.go         # 刷题开始/提交/统计/场次 HTTP 处理器
│   │   ├── exam_handler.go           # 考试类型/模块 CRUD + 健康检查 + 数据导出/导入
│   │   ├── question_handler.go       # 题目 CRUD + 批量导入/删除（含逐条校验）
│   │   └── user_handler.go           # 认证/用户信息/管理员用户管理
│   ├── middleware/
│   │   └── auth.go                   # JWT 认证中间件 + 管理员权限中间件（含 DB 实时校验）
│   ├── model/
│   │   ├── model.go                  # 核心数据模型（ExamType, Module, Question, UserAnswer, ExamSession, SystemConfig）
│   │   ├── user.go                   # 用户模型（密码 json:"-" 不序列化）
│   │   └── model_test.go             # 模型校验测试
│   ├── repository/
│   │   ├── answer_repo.go            # 答题记录 CRUD + 按用户统计查询 + 级联计数
│   │   ├── config_repo.go            # SystemConfig 键值存储
│   │   ├── exam_repo.go              # 考试类型/模块 CRUD + 级联删除 + 聚合统计
│   │   ├── question_repo.go          # 题目 CRUD + 筛选随机选题 + 错题查询
│   │   ├── question_repo_test.go     # splitTags + QuizFilter 测试
│   │   ├── session_repo.go           # 考试场次 CRUD + 分页答题记录
│   │   └── user_repo.go             # 用户 CRUD + 修改昵称/密码
│   ├── response/
│   │   └── response.go               # 统一 JSON 响应工具（OK, Created, List, Error）
│   ├── router/
│   │   └── router.go                 # 路由注册 + CORS + 请求体大小限制（JSON 2MB / 多部分 8MB）
│   ├── service/
│   │   ├── answer_service.go         # 答案比对 + 批量交卷 + 场次自动结束 + 缓存失效
│   │   ├── answer_service_test.go    # 答案比对单元测试
│   │   ├── exam_service.go           # 考试类型业务 + 数据导出/导入 + 统计缓存层
│   │   ├── question_service.go       # 刷题模式分发（3 种选题策略）+ 空题保护
│   │   └── user_service.go           # 注册/登录/个人资料/管理员用户管理
│   ├── util/
│   │   ├── jwt.go                    # JWT 生成/解析（HS256, 7 天过期）
│   │   └── password.go               # bcrypt 密码哈希/验证
│   └── validator/
│       └── validator.go              # 题目字段校验 + ID/查询参数解析
├── web/
│   ├── index.html                    # 入口 HTML（zh-CN, Inter 字体）
│   ├── package.json
│   ├── vite.config.js                # Vite 配置（Vue 插件 + API 代理）
│   └── src/
│       ├── main.js                   # Vue 应用引导
│       ├── App.vue                   # 根布局：导航栏 + 侧边栏 + 统计面板
│       ├── style.css                 # CSS 变量系统 + 响应式样式
│       ├── api/
│       │   └── index.js              # Axios 封装 + token 注入 + 401 拦截 + 全部 API 函数
│       ├── components/
│       │   ├── Sidebar.vue           # 侧边栏导航 + 用户信息 + 退出登录
│       │   └── StatsPanel.vue        # 右侧统计面板（进度、正确率）
│       ├── router/
│       │   └── index.js              # Vue Router（认证/管理员路由守卫）
│       ├── stores/
│       │   ├── auth.js               # 认证状态管理（token + user 持久化到 localStorage）
│       │   └── exam.js               # 考试选择 + 刷题偏好持久化
│       ├── utils/
│       │   └── quiz.js               # 刷题工具函数
│       └── views/
│           ├── HomeView.vue          # 首页仪表盘
│           ├── LoginView.vue         # 登录/注册表单
│           ├── ExamView.vue          # 模块列表页
│           ├── QuizView.vue          # 主刷题界面（键盘快捷键 A/B/C/D, Enter）
│           ├── HistoryView.vue       # 场次历史时间线 + 详情弹窗
│           ├── SettingsView.vue      # 个人设置（资料/密码/偏好/概览）
│           ├── ExamManageView.vue    # 管理员：考试类型 + 模块 CRUD
│           ├── QuestionManageView.vue# 管理员：题目 CRUD + 批量导入
│           └── UserManageView.vue    # 管理员：用户管理
├── build.sh                          # Linux/macOS 生产构建脚本
├── build.bat                         # Windows 生产构建脚本
├── webfs.go                          # go:embed 指令（嵌入 web/dist）
├── go.mod
├── go.sum
└── README.md
```

## 题目导入格式

```json
[
  {
    "module_id": 1,
    "type": "single",
    "content": "题目内容",
    "options": "[\"A. 选项A\", \"B. 选项B\", \"C. 选项C\", \"D. 选项D\"]",
    "answer": "A",
    "analysis": "解析说明",
    "difficulty": 2,
    "tags": "标签1,标签2",
    "source": "2024国考"
  }
]
```

**字段说明：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `module_id` | int | ✅ | 所属模块 ID |
| `type` | string | ✅ | 题型：`single`（单选）/ `multiple`（多选）/ `judge`（判断）/ `fill`（填空） |
| `content` | string | ✅ | 题目内容 |
| `options` | string | 选择题必填 | 选项 JSON 数组字符串 |
| `answer` | string | ✅ | 正确答案（多选用逗号分隔，如 `"A,C"`） |
| `analysis` | string | ❌ | 解析说明 |
| `difficulty` | int | ❌ | 难度等级 1-5，默认 1 |
| `tags` | string | ❌ | 标签，逗号分隔 |
| `source` | string | ❌ | 来源（如 "2024国考"） |

## 设计决策

| 决策 | 选择 | 原因 |
|------|------|------|
| 数据库 | SQLite (WAL) | 零配置、单文件部署、WAL 支持并发读 |
| 认证方案 | JWT (HS256, 7天) | 无状态、前后端分离友好、纯 Go 实现 |
| 密码存储 | bcrypt | 业界标准、自带盐值、抗暴力破解 |
| 前端状态 | reactive() 闭包 | 项目规模小，无需额外状态管理库 |
| 前端嵌入 | Go embed | 单二进制部署，无需 Nginx |
| 正确率算法 | 取每题最后一次 | 避免重复刷题导致正确率虚高 |
| 级联删除 | 手写事务 | GORM 无外键级联约束，手动保证一致性 |
| 错题模式 | JOIN 子查询 | 用 INNER JOIN 替代嵌套子查询，提升性能 |
| 随机选题 | ORDER BY RANDOM() | SQLite 下简洁可靠，满足万级数据量 |
| 统计缓存 | sync.Map + 30s TTL + 按用户隔离 | 减少重复查询，答题后自动失效 |
| 注册控制 | 默认关闭 | 安全优先，管理员按需开启 |

## 更新日志

### 2025-06

- 🔧 **刷题错误处理增强** — API 失败时显示具体错误信息（区分 401/服务端/网络错误），支持重试
- 🔧 **题目内容容错** — content/options 为 null 时显示 fallback 文案，不再白屏
- 🔧 **选项解析健壮性** — `parseOptions` 增加类型校验和日志，非法 JSON 不再静默吞掉
- 🔧 **quizMode 响应式** — 设置页切换模式后刷题页实时生效
- 🔧 **路由参数监听** — 切换模块自动重新加载题目，修复组件复用导致的旧题残留
- 🔧 **答题 session_id** — 提交答案时携带 session_id，修复答题记录未关联场次的问题
- 🧹 **清理临时文件** — 移除根目录的年度题目数据文件和无关配置

## License

MIT
