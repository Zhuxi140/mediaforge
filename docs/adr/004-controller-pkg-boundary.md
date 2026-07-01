# ADR-004: Controller → pkg 分层架构

**状态**: ✅ 已采纳  
**日期**: 2026-06-29  

## 背景

Go 后端需要组织代码，明确职责边界，避免 Wails 绑定代码与核心业务逻辑耦合。

## 可选方案

1. **全在一个包** — 所有 Go 代码放在同一包中
2. **Controller + Service 分层** — controller 处理 Wails 绑定，service/pkg 处理业务
3. **Clean Architecture** — 严格的四层架构（entity/usecase/repo/controller）

## 决策

选择方案 **2（Controller + pkg 分层）**。

## 结构

```
controller/          # 薄层：Wails 绑定，持有 context.Context
  renamer_app.go     #   代理 pkg/renamer 的调用
  media_app.go       #   代理 pkg/media 的调用
pkg/
  renamer/           # 核心逻辑：与 Wails 无依赖
    renamer.go       #   文件重命名算法
  media/             # 核心逻辑：与 Wails 仅依赖 runtime.EventsEmit
    engine.go        #   FFmpeg 嵌入引擎
    transcode.go     #   转码流程 + 进度监控
    subtitle.go      #   字幕扫描/提取/转换
    types.go         #   共享类型 + 格式映射
```

## 理由

- `controller/` 层仅做参数转发和 context 管理，薄而清晰
- `pkg/` 层无 Wails 框架依赖，可独立单元测试（尽管当前未写测试）
- `pkg/media/` 因需要向前端推送进度，依赖 `runtime.EventsEmit`，但可通过接口抽象
- 新增功能时：新建 controller 方法 + 扩展对应 pkg 即可

## 后果

- `controller/` 层方法签名由 Wails 绑定生成约束，修改需同步更新前端调用
- `pkg/` 层不应直接调用 `runtime` 包（media 包当前是特例）
- 后期可将 `runtime.EventsEmit` 抽象为接口，使 media 包可测试
