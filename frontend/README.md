# Frontend

MVP V1 前端工程，基于 React + TypeScript + Vite。

## 技术栈

- React 18
- TypeScript
- Vite
- React Router

## 功能范围

- 工作台
- 模板列表 / 模板详情 / 流程发起
- 流程列表 / 流程详情 / 节点工作台
- 评论 / 附件 / 节点日志
- 交付物列表 / 详情 / 验收
- 人员管理 / 龙虾管理

## 环境要求

- Node.js 18+
- npm 9+
- 已启动的后端服务

默认后端地址：

- `http://localhost:8080`

可通过环境变量覆盖：

- `VITE_API_BASE_URL`

示例：

```bash
VITE_API_BASE_URL=http://127.0.0.1:8080 npm run dev
```

## 安装依赖

```bash
npm install
```

## 本地开发

```bash
npm run dev
```

默认启动后访问：

- `http://127.0.0.1:5173`

## 生产构建

```bash
npm run build
```

构建产物输出到：

- `dist/`

## 预览构建产物

```bash
npm run preview
```

## 联调说明

前端所有请求默认会从本地 actor 中带上以下请求头：

- `X-Person-ID`
- `X-Role-Type`

页面顶部提供 `ActorBar`，可手动切换身份。

另外，流程发起成功后，如果当前本地还没有 actor，前端会自动把本次发起人写入本地身份，保证进入流程详情后 `available_actions` 能正常返回。

建议联调测试账号：

- `1` / `leader`
- `2` / `middle_office`
- `3` / `operation`

## 目录结构

- `src/pages/`：页面
- `src/components/`：组件
- `src/api/`：接口调用
- `src/hooks/`：数据获取与动作 hooks
- `src/lib/`：基础工具
- `src/types/`：类型定义

## 当前状态

已完成一轮可执行验证：

- `npm install` 通过
- `npm run build` 通过
- `npm run dev` 可正常启动
