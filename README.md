# 公考刷题工具（Teex）📝

一个面向公务员、事业编、三支一扶、乡村振兴协理员等考试的在线刷题系统。采用 **Go 后端 + Vue 3 前端** 的前后端分离架构，支持解析模式刷题、试卷模式、错题重做、统计分析、历史记录回溯等功能。

## 功能特性

- ✅ **解析模式刷题** — 提交即反馈，逐题显示解析
- ✅ **题库管理** — 题目 CRUD + JSON 批量导入 + 难度/标签筛选
- ✅ **考试类型管理** — 考试与科目 CRUD，每个考试独立的题库
- ✅ **错题重做** — 智能识别"最后一答错误"的题目
- ✅ **统计分析** — 右侧面板实时显示当前考试的进度与正确率
- ✅ **历史记录** — 时间线式场次列表 + 详情弹窗 + 分页
- ✅ **键盘快捷键** — A/B/C/D 选选项，Enter 提交
- ✅ **数据导出/导入** — API 级别的数据备份能力
- ✅ **移动端适配** — 响应式布局
- 🚧 **收藏夹**（数据模型就绪，API 开发中）

### 支持的考试类型

| 考试类型 | 模块 |
|----------|------|
| 国家公务员 | 行测-言语理解、数量关系、判断推理、资料分析、申论 |
| 省级公务员 | 行测-言语理解、数量关系、判断推理、资料分析、申论、公共基础知识 |
| 事业编 | 公共基础知识、职业能力测验、综合应用能力 |
| 三支一扶 | 公共基础知识、职业能力测验 |
| 乡村振兴协理员 | 公共基础知识、农村工作知识、行测 |

## 技术栈

| 层 | 技术 | 版本 |
|----|------|------|
| 后端框架 | Go + Gin | Go 1.26.3 / Gin 1.12.0 |
| ORM | GORM | 1.31.1 |
| 数据库 | SQLite (WAL 模式) | mattn/go-sqlite3 |
| 前端框架 | Vue 3 (Composition API) | 3.5.34 |
| 构建工具 | Vite | 8.0.12 |
| 前端嵌入 | Go embed | 标准库 |

## 快速开始

### 推荐方式（使用脚本）

```bash
# 开发模式（一键启动前后端，热更新）
./run.sh dev

# 生产构建
./build.sh

# 启动生产版本
./run.sh start

# 停止所有服务
./run.sh stop
```

### 手动方式

```bash
# 1. 编译后端
go build -o server ./cmd/server/
./server
# 服务运行在 http://localhost:8080
# SQLite 数据库文件自动创建在 data/exam-quiz.db

# 2. 启动前端（开发模式，另一个终端）
cd web
npm install
npm run dev
# 前端运行在 http://localhost:5173
```

### 生产部署

```bash
./build.sh
./run.sh start
# 单文件部署，Go 二进制通过 go:embed 嵌入前端产物，无需 Nginx
```

> 💡 可通过环境变量 `PORT` 自定义端口，默认 `8080`。

## API 文档

### 健康检查
- `GET /api/health` — 健康检查

### 考试类型
- `GET /api/exams` — 获取所有考试类型（含模块列表）
- `POST /api/exams` — 创建考试类型（名称唯一校验）
- `PUT /api/exams/:id` — 更新考试类型
- `DELETE /api/exams/:id` — 删除考试类型（级联删除）
- `GET /api/exams/:id/modules` — 获取某考试类型下的模块列表（含题目数/未做数）

### 模块
- `POST /api/modules` — 创建模块（验证考试类型存在+名称唯一）
- `PUT /api/modules/:id` — 更新模块
- `DELETE /api/modules/:id` — 删除模块（级联删除）

### 题目管理
- `GET /api/questions?module_id=&exam_type_id=&type=&difficulty=&page=&size=` — 查询题目列表
- `GET /api/questions/:id` — 获取单个题目
- `POST /api/questions` — 创建题目（验证类型/难度/必填字段）
- `PUT /api/questions/:id` — 更新题目
- `DELETE /api/questions/:id` — 删除题目（级联删除答题记录+收藏）
- `POST /api/questions/import` — 批量导入题目（逐条校验，上限500条）

### 刷题

选题策略（`mode` 参数，与功能特性中的"解析模式"是不同概念）：
- `default`：未做优先，已做过的补随机
- `wrong`：仅错题（最后一次答题为错误的）
- `random`：纯随机

- `POST /api/quiz/start` — 开始刷题 `{module_id, count, mode, difficulty, tags}`
  - mode: `default`（未做优先）/ `wrong`（错题）/ `random`（随机）
  - count: 1-200，默认10
- `POST /api/quiz/answer` — 提交单题答案 `{question_id, user_input, duration}`
- `POST /api/quiz/submit-batch` — 批量提交答案 `{answers, session_id}`

### 统计
- `GET /api/stats` — 全局统计（正确率取每题最后一次）
- `GET /api/stats/module/:id` — 某模块统计

### 考试场次
- `GET /api/sessions?page=&size=` — 场次列表（分页）
- `GET /api/sessions/:id` — 单个场次详情
- `GET /api/sessions/:id/answers` — 某场次的答题记录

### 数据管理
- `DELETE /api/records` — 清空所有答题记录和考试场次
- `GET /api/export` — 导出全部数据为 JSON
- `POST /api/import` — 导入考试类型和模块

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
├── build.sh                       # 一键构建脚本
├── run.sh                         # 一键开发/启动/停止脚本
├── .gitignore
├── cmd/server/
│   └── main.go                    # 程序入口：初始化DB、种子、路由、静态文件、graceful shutdown
├── internal/
│   ├── model/
│   │   ├── model.go               # 数据模型（6个核心表 + 1个DTO）
│   │   └── model_test.go          # 模型测试
│   ├── database/db.go             # 数据库初始化（WAL模式 + 连接池 + 种子数据）
│   ├── handler/
│   │   ├── answer_handler.go      # 答题相关HTTP处理器
│   │   ├── exam_handler.go        # 考试类型/模块 CRUD + 健康检查 + 数据导出/导入
│   │   └── question_handler.go    # 题目 CRUD + 批量导入（含逐条校验）
│   ├── repository/
│   │   ├── answer_repo.go         # 答题记录CRUD + 统计查询
│   │   ├── exam_repo.go           # 考试类型/模块CRUD + 级联删除
│   │   ├── question_repo.go       # 题目CRUD + 筛选 + 收藏数据模型
│   │   └── session_repo.go        # 考试场次CRUD
│   ├── service/
│   │   ├── answer_service.go      # 答案比对 + 批量交卷 + 场次管理
│   │   ├── exam_service.go        # 考试类型业务 + 数据导出/导入
│   │   └── question_service.go    # 刷题模式分发 + 空题保护
│   └── router/router.go          # 路由注册（26个端点）+ CORS + 请求限制
├── data/seed/
│   ├── exams.json                 # 5种考试类型 + 19个模块
│   └── questions_sample.json      # 43道示例题目
├── web/
│   ├── src/
│   │   ├── api/index.js           # Axios封装 + 错误Toast通知
│   │   ├── views/                 # 7个页面组件
│   │   ├── stores/exam.js         # 全局状态管理
│   │   ├── components/Sidebar.vue # 侧边栏导航
│   │   └── style.css              # CSS变量系统 + 响应式
│   └── dist/                      # Vite构建产物（嵌入Go二进制，.gitignore已忽略）
├── webfs.go                       # Go embed指令
└── README.md
```

## 题目导入格式

```json
[
  {
    "module_id": 1,
    "type": "single",
    "content": "题目内容",
    "options": "[\"A. 选项A\", \"B. 选项B\"]",
    "answer": "A",
    "analysis": "解析说明",
    "difficulty": 2,
    "tags": "标签1,标签2",
    "source": "2024国考"
  }
]
```

## 设计决策

| 决策 | 选择 | 原因 |
|------|------|------|
| 数据库 | SQLite (WAL) | 零配置、单文件部署、WAL支持并发读 |
| 前端状态 | reactive() 闭包 | 项目规模小，无需额外状态管理库 |
| 前端嵌入 | Go embed | 单二进制部署，无需Nginx |
| 正确率算法 | 取每题最后一次 | 避免重复刷题导致正确率虚高 |
| 级联删除 | 手写事务 | GORM无外键级联约束，手动保证一致性 |
| 错题模式 | 最后一答错误 | 反映当前真实掌握情况 |

## License

MIT
