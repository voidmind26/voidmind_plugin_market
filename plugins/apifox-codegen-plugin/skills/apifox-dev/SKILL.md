---
name: apifox-dev
description: This skill should be used when the user asks to "生成接口文档", "生成场景化测试", "同时生成接口和测试", "从代码生成 Apifox 接口", "同步 Apifox 接口", "更新 Apifox 接口文档", "补 Apifox 测试用例", or "先生成接口再补测试" for a Go Web project.
version: 0.1.0
allowed-tools:
  - Read
  - Bash
  - AskUserQuestion
  - Skill
---

# Apifox Dev

作为 `apifox-codegen-plugin` 的统一入口，先规范输入，再将任务继续路由到下游 skill。不要停留在识别层；完成确认后直接调用目标 skill。

## Goals

- 识别用户要生成接口文档、场景化测试，还是两者连续执行。
- 在执行前补齐最小必要输入，尤其是路由前缀。
- 对“全部接口”请求做全量扫描提醒，并显式确认范围。
- 在新增接口场景下，确保接口最终落在目标模块，且测试环境中已存在该模块地址。
- 在生成测试场景下，确保正向测试默认是强断言行为测试，而不是只验证 `HTTP 200`。
- 对明显非 Go Web 风格项目直接拒绝接管。
- 输入确认后继续调用下游 skill，不把路由责任留给外层。

## Required Checks

按以下顺序执行：

1. 判断项目是否明显属于 Go Web 风格。
   - 若明显不是：直接说明首版仅接管 Go Web 风格项目，当前任务不继续路由。
   - 若无法明显排除：继续。

2. 识别用户意图，归一为以下三类之一：
   - `interfaces_only`
   - `tests_only`
   - `interfaces_then_tests`

3. 收集范围输入。
   - 默认要求用户提供路由前缀，如 `/user`、`/order`。
   - 若用户未提供路由前缀，先追问，再继续。
   - 若用户明确说“全部接口”，必须说明这是全量扫描，并要求显式确认扫描范围；未确认前不得继续。

4. 形成最终执行载荷。
   - 包含 `intent`、`route_scope`、必要的用户补充说明。
   - 若是“全部接口”，在载荷中明确标注 `route_scope=all_confirmed`。
   - 若是路由前缀，使用用户原始前缀值，不擅自改写。

5. 直接调用下游 skill。
   - 不只返回判断结果。
   - 不把“接下来该调哪个 skill”留给外层决定。

## Input Normalization

优先将用户输入整理为以下内部结构，再继续路由：

```json
{
  "intent": "interfaces_only | tests_only | interfaces_then_tests",
  "route_scope": "/user | /order | all_confirmed",
  "user_request": "原始需求的简短归纳"
}
```

使用规则：

- 用户说“生成接口文档”“同步接口到 Apifox”“从代码生成接口” → `interfaces_only`
- 用户说“生成场景化测试”“补测试用例”“只做测试” → `tests_only`
- 用户说“接口和测试都生成”“先出接口再出测试” → `interfaces_then_tests`
- 用户给出 `/user`、`/order` 这类前缀时，直接作为 `route_scope`
- 用户说“全部接口”时，仅在完成显式确认后写入 `all_confirmed`

## Confirmation Rules

### 缺少路由前缀

缺少范围时，先追问：

- 要处理哪个路由前缀？例如 `/user`、`/order`
- 如果确实要处理全部接口，请明确确认“全部接口”范围

### 用户要求“全部接口”

先明确提示：

- 这是全量扫描，不是局部生成
- 可能覆盖范围较大
- 需要用户显式确认后才继续

未拿到明确确认前，不调用下游 skill。

### 非 Go Web 风格

若仓库结构、路由方式或任务描述明显不是 Go Web 风格，直接说明：

- 首版统一入口仅接管 Go Web 风格项目
- 当前任务不继续路由到下游 skill

## Routing

按以下映射继续执行：

- 仅生成接口 -> `generate-interfaces-from-code`
- 仅生成测试 -> `generate-scenario-tests`
- 两者都要 -> 先 `generate-interfaces-from-code`，再 `generate-scenario-tests`

执行“两者都要”时，先完成接口生成，再基于同一份已确认范围继续调用测试 skill，不要再次把路由决策抛回外层。

## Execution Notes

- 只在输入已经确认后再调用下游 skill。
- 将已确认的 `intent`、`route_scope` 和必要上下文一并传给下游 skill。
- 不擅自把局部前缀扩成全量扫描。
- 不因用户语义模糊而猜测范围；宁可先问清楚。
- 若用户目标只是其中一项，不额外触发另一项。
- 若任务被拒绝接管，清楚说明原因后停止，不做伪路由。
- 若后续会新增 Apifox 接口，下游必须把 reference 中的模块归属与测试环境模块地址校验提升为主流程强检查，不允许默认落到默认模块。
- 若后续会生成 Apifox 测试，下游必须把 reference 中的强断言要求提升为主流程强检查，不允许只产出“返回 200/请求成功”类弱测试。
- 若后续会生成 Apifox 测试，下游还必须显式落实 `apiPath` 与标准 `type: assertion` 结构要求，不能只把这些规则留在 reference 里。

## Response Style

保持短促、明确、可执行：

- 先说识别出的意图
- 再说当前缺什么输入或确认了什么范围
- 一旦确认完成，立即继续调用下游 skill
