# Knowledge Map

## Purpose

本文件只负责做知识索引与消费顺序说明，不替代 `knowledge/layering.md` 的原则性约束。

所有标签路由默认先服从：

- `knowledge/layering.md`

## Consumption Order

当 skill 或 agent 命中某个标签时，按下面顺序取资料：

1. 先读 `knowledge/`：拿高层边界与最小变更面规则
2. 再读 `references/`：补细规则、签名模式、禁忌项、目录约定
3. 最后读 `examples/`：在需要具体代码形态时再取最小模板

不要跳过 `knowledge/` 直接从 `examples/` 反推规则。

## Tag Routing

### dto

- knowledge:
  - `knowledge/dto.md`
- references:
  - `references/dto-conventions.md`
- examples:
  - `examples/dto-example.md`

### data

- knowledge:
  - `knowledge/data.md`
- references:
  - `references/model-conventions.md`
- examples:
  - `examples/model-data-example.md`

说明：若需求涉及 model、持久化实体、表结构映射或缓存对象结构，统一归入 `data` 标签处理。

### service

- knowledge:
  - `knowledge/service.md`
- references:
  - `references/service-conventions.md`
- examples:
  - `examples/service-example.md`

### controller

- knowledge:
  - `knowledge/controller-router.md`
- references:
  - `references/controller-router-conventions.md`
- examples:
  - `examples/controller-router-example.md`

### router

- knowledge:
  - `knowledge/controller-router.md`
- references:
  - `references/controller-router-conventions.md`
- examples:
  - `examples/controller-router-example.md`

### task

- knowledge:
  - `knowledge/task.md`
  - `knowledge/logging.md`
- references:
  - `references/task-logging-conventions.md`
- examples:
  - `examples/task-registration-example.md`
  - `examples/task-logging-example.md`

## Routing Rules

- 只为命中标签加入对应路径，不把全部知识资产一股脑传给下游。
- 简单任务优先保留最小集合：通常是 `knowledge/`，必要时再补 `references/` 或 `examples/`。
- 复杂任务可以同时带多组标签资产，但仍按命中层裁剪，不默认扩成全链路。
- 若知识来源之间冲突，以 `knowledge/layering.md` 和命中标签对应的 `knowledge/*.md` 为先，再参考 `references/` 和 `examples/`。
