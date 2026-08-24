---
name: apifox-dev
description: 统一处理 Go Web 项目的 Apifox 接口文档与测试用例生成。当用户要求生成、同步或更新 Apifox 接口，生成场景化测试，或先生成接口再补测试时使用；通过 Apifox CLI 完成项目查询和写入。
---

# Apifox Dev

作为插件统一入口，确认 CLI、项目、分支和扫描范围后，直接路由到接口生成或测试生成 skill。

## 开始前

1. 读取 `../../references/apifox-cli-rules.md`。
2. 执行 `apifox --version`、`apifox --help` 和 `apifox whoami`。
3. CLI 缺失或未登录时，先使用 `apifox-cli-setup`，不要回退到 MCP 或 HTTP API。
4. 用户未指定项目时，读取 `.apifox/settings.json`；没有默认 `projectId` 时执行 `apifox project list` 并让用户确认。
5. 涉及写入时确认目标分支。未传 `--branch` 会写主分支，不要静默采用。

## 输入确认

按顺序执行：

1. 判断项目是否属于 Go Web 风格；明显没有 HTTP router/handler 时停止。
2. 将意图归一为：
   - `interfaces_only`
   - `tests_only`
   - `interfaces_then_tests`
3. 确认路由范围：
   - 默认要求具体前缀，如 `/user`、`/order`。
   - 用户明确要求全部接口时，说明扫描面较大并取得显式确认。
4. 形成内部执行载荷：

```json
{
  "intent": "interfaces_only | tests_only | interfaces_then_tests",
  "route_scope": "/user | /order | all_confirmed",
  "project_id": 123456,
  "branch": "用户确认的分支",
  "user_request": "原始需求简述"
}
```

## 路由

- 仅生成接口：使用 `generate-interfaces-from-code`。
- 仅生成测试：使用 `generate-scenario-tests`。
- 两者都要：先完成 `generate-interfaces-from-code`，再把已写入接口 ID、真实 path、项目和分支传给 `generate-scenario-tests`。

不要只返回路由建议，也不要把局部前缀扩成全量扫描。

## 强制边界

- 新接口必须写入已确认的模块/目录，不能落到根目录或默认模块。
- 写入前必须确认测试环境存在该模块地址。
- 所有 JSON 写入必须先通过 `apifox cli-schema validate`。
- 创建或更新后必须 `get` 回读。
- 正向测试不能只有 HTTP 200；至少验证成功语义和一个业务行为或结构结果。
- 删除、覆盖导入、AI 分支创建以及 merge/MR 不在本入口中自动执行。
