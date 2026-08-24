# Apifox 接口写入规则

## 1. Path 拆分

扫描和写入阶段维护四个概念：

| 字段 | 含义 | 用途 |
|---|---|---|
| `routePrefix` | 用户指定的扫描边界 | 只筛选扫描范围 |
| `fullRoute` | 代码中含部署前缀的完整路由 | 定位 router/handler |
| `routeMountPath` | 移除部署前缀后的业务路由 | 表达真实业务挂载路径 |
| `apifoxPath` | 最终写入 Apifox 的 path | `endpoint` payload 的唯一 path 来源 |

默认 `apifoxPath = routeMountPath`。只能移除部署前缀，不能移除 `/flowconfig`、`/order` 等业务模块前缀。

## 2. 模块、目录与环境

新增接口前必须确认：

1. 目标项目和分支。
2. 接口所属业务模块。
3. 该模块下目标接口目录的 `folderId`。
4. 测试环境存在该模块的 base URL。

使用当前 CLI 返回结构取证：

```bash
apifox project get <projectId> --with endpointFolders
apifox folder list --project <projectId> --branch <branch> --type endpoint
apifox environment list --project <projectId> --branch <branch>
apifox environment get <environmentId> --project <projectId> --branch <branch>
```

不要把模块 ID 和目录 ID 混为一谈。无法确认目标目录或测试环境模块地址时停止创建，不默认落到根目录。

## 3. 查重

先按 path 搜索候选：

```bash
apifox endpoint list --project <projectId> --branch <branch> \
  --path-contains <apifoxPath> --page-size 500
```

再从结构化 JSON 中按 method 和完整 path 精确匹配。不能只凭 path 包含结果判断接口已存在。

## 4. 写入顺序

使用“最小创建 -> 回读 -> 完整结构分步更新”：

1. 获取 `endpoint-create` 和 `endpoint-update` schema。
2. 创建最小骨架：`name`、`type=http`、`method`、`path`、`folderId`。
3. schema 校验成功后执行 create。
4. 立即 `endpoint get` 回读。
5. 基于完整回读结构依次补充 description、请求、响应、example。
6. 每一步都验证 `endpoint-update` schema、执行 update，再次回读。

修改数组或嵌套对象时不得提交局部 JSON Patch；必须保留完整结构中的未知字段和已有数组项。

## 5. JSON 结构

- method 使用当前 `endpoint-create` schema 接受的小写枚举。
- query/path/header/cookie 参数按 schema 放入 `parameters` 对应数组。
- JSON 请求体使用 `requestBody.type=application/json` 与 `requestBody.jsonSchema`。
- 响应放入 `responses`，示例放入 `responseExamples`。
- 复杂对象先用结构化 API 构造，再序列化为 JSON 文件；不要通过字符串拼接生成嵌套 payload。
- 临时 payload 放入 `mktemp -d`，完成后不留在业务仓库。

## 6. 失败缩圈

- `cli-schema validate` 失败时不得远端写入。
- 服务端 422 时停在当前维度，不继续叠加字段。
- request/response 失败时先保留顶层，再逐层恢复字段、数组元素与 example。
- 创建骨架后补全失败时报告 endpoint ID 和缺失维度；未经确认不要删除骨架。

## 7. 完成标准

- 项目、分支、模块、目录和测试环境均已确认。
- path 精确等于 `apifoxPath`。
- 每次写入前均通过对应 schema 校验。
- 创建或更新后均已回读。
- 最终回读的请求、响应和示例与代码事实一致。
