# FlyAI Web UI

FlyAI 的管理端使用 Next.js、React、Tailwind CSS 和 shadcn/ui。生产环境采用静态导出，生成结果位于 `dist/`，由 Go 后端启动时直接读取，因此运行 FlyAI 时不需要单独启动 Node.js。

## 本地开发

```powershell
npm ci
npm run dev
```

开发服务器适合调试页面外观。需要联调真实 API 时，请执行生产构建并从仓库根目录重新运行 Go 服务，使前后端保持同源。

## 生产构建

```powershell
npm run lint
npm run typecheck
npm run build
```

修改前端后需要重新执行 `npm run build` 并提交更新后的 `dist/`。前端不再嵌入可执行文件，因此无需仅为页面变化重新执行 `go build`。
