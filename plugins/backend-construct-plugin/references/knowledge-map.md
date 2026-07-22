# 知识索引

## 用途

本文件定义分层标签到规范资产的读取路径。下表路径均相对于本文件所在的 `references/` 目录。所有实现先服从 `../knowledge/layering.md`，再按实际命中层渐进读取，不一次性加载全部资料。

## 读取顺序

1. 读取 `knowledge/`，确定职责边界和最小变更面。
2. 需要字段、签名、目录或禁忌项时，读取 `references/`。
3. 项目中没有可复用的相邻实现时，才读取 `examples/` 作为代码形态参考。

不要跳过项目代码和 `knowledge/`，直接从示例反推项目规则。

## 标签路由

| 标签 | knowledge | references | examples |
|---|---|---|---|
| `dto` | `../knowledge/dto.md` | `dto-conventions.md` | `../examples/dto-example.md` |
| `data` | `../knowledge/data.md` | `model-conventions.md` | `../examples/model-data-example.md` |
| `service` | `../knowledge/service.md` | `service-conventions.md` | `../examples/service-example.md` |
| `controller` | `../knowledge/controller-router.md` | `controller-router-conventions.md` | `../examples/controller-router-example.md` |
| `router` | `../knowledge/controller-router.md` | `controller-router-conventions.md` | `../examples/controller-router-example.md` |
| `task` | `../knowledge/task.md`、`../knowledge/logging.md` | `task-logging-conventions.md` | `../examples/task-registration-example.md`、`../examples/task-logging-example.md` |

涉及 model、持久化实体、表结构映射或缓存对象结构时，统一归入 `data`。

## 路由规则

- 只读取命中层对应的资料。
- 简单修改通常只需 `knowledge/` 和项目相邻实现。
- 需要精确生成代码时再补 `references/`；只有缺少项目样例时才用 `examples/`。
- 多层任务可组合多组资产，但不能因此扩展实现范围。
- 资料冲突时，遵循 `backend-dev` 中的规范优先级；示例始终不是强制模板。
