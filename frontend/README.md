# 星诺图床 — 前端

基于 Vue 3 + TypeScript + Tailwind CSS 构建的图床管理界面。

## 技术栈

- **框架**: Vue 3 (Composition API + `<script setup>`)
- **语言**: TypeScript
- **构建工具**: Vite
- **样式**: Tailwind CSS
- **组件库**: shadcn-vue + Radix Vue primitives
- **状态管理**: Pinia
- **路由**: Vue Router 4
- **HTTP 客户端**: Axios
- **表单验证**: VeeValidate + Zod
- **图标**: Lucide Vue Next
- **测试**: Vitest + Vue Test Utils
- **代码质量**: ESLint + Prettier + Oxlint

## 目录结构

```
src/
├── api/              # API 接口封装 (按模块组织)
├── assets/           # 静态资源 (CSS, 字体)
├── components/       # Vue 组件
│   └── ui/          # shadcn-vue 基础组件
├── layouts/          # 布局组件
├── lib/             # 工具函数
├── router/          # 路由配置
├── stores/          # Pinia 状态管理
├── types/           # TypeScript 类型定义
├── utils/           # 通用工具
└── views/           # 页面视图
    └── admin/       # 管理后台页面
```

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 类型检查
npm run type-check

# 运行测试
npm run test:unit

# 代码检查
npm run lint

# 格式化
npm run format
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `VITE_API_BASE_URL` | 后端 API 地址 | `http://localhost:8000/api/v1` |
| `VITE_APP_TITLE` | 应用标题 | `星诺图床` |

复制 `.env.example` 为 `.env.local` 进行本地配置。

## 构建

```bash
npm run build
```

构建产物输出到 `dist/` 目录，可通过 Nginx 或任意静态服务器托管。

## Docker

```bash
docker build -t sino-imgbed-frontend .
docker run -p 80:80 sino-imgbed-frontend
```
