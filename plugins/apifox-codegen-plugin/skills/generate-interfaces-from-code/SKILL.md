---
name: generate-interfaces-from-code
description: 从 Go Web 代码扫描路由并通过 Apifox CLI 创建或更新接口文档。当用户要求从代码生成 Apifox 接口、按路由前缀同步接口、整理现有路由或全量扫描接口时使用。
---

# Generate Interfaces From Code

从 router -> controller -> service -> dto 提取接口事实，并通过 `apifox endpoint` 写入已确认的项目、分支和模块目录。

## 开始前

读取以下规范：

- `../../references/apifox-cli-rules.md`：CLI 预检、schema 校验、权限和回读规则。
- `../../references/route-scan-rules.md`：扫描范围和代码链路取证。
- `../../references/apifox-write-rules.md`：path、模块、环境和写入顺序。

执行 `apifox --version`、`apifox whoami` 和 `apifox endpoint --help`。CLI 未就绪时使用 `apifox-cli-setup`，不要回退到 MCP。

## 扫描

1. 确认具体路由前缀；全量扫描必须经过用户确认。
2. 从 router 提取 HTTP method、`fullRoute` 和 handler。
3. 继续读取 controller、service 和 dto，提取接口名称、业务说明、请求与响应结构。
4. 为每个接口形成：

```json
{
  "method": "POST",
  "fullRoute": "/zybuosmis/flowconfig/create",
  "routeMountPath": "/flowconfig/create",
  "apifoxPath": "/flowconfig/create",
  "interfaceName": "创建流量配置",
  "request": {},
  "response": {},
  "descriptionDraft": "根据代码事实整理的说明"
}
```

`fullRoute` 只用于取证。写入的 `path` 必须等于 `apifoxPath`，默认与 `routeMountPath` 相同；只能移除部署前缀，不能移除业务模块前缀。

## CLI 上下文

1. 确认 `projectId` 和目标分支。
2. 执行：

```bash
apifox project get <projectId> --branch <branch> --with endpointFolders
apifox folder list --project <projectId> --branch <branch> --type endpoint
apifox environment list --project <projectId> --branch <branch>
```

3. 从项目与目录返回结构确认目标模块及其接口目录 `folderId`；不要把模块 ID 和目录 ID 混为一谈，也不要默认使用 `folderId=0`。
4. 对测试环境执行 `environment get <environmentId> --project <projectId> --branch <branch>`，确认其模块地址覆盖目标模块。缺少模块地址时停止写接口，先报告阻断。
5. 用 `endpoint list --project <projectId> --branch <branch> --path-contains <apifoxPath> --page-size 500` 查找候选，再按 method 与完整 path 精确比较，判断创建还是更新。

## 创建接口

在临时目录中执行：

1. 获取 schema：

```bash
apifox cli-schema get endpoint-create
apifox cli-schema get endpoint-update
```

2. 生成最小创建 JSON，至少包含 `name`、小写 `method`、`path`、`type=http` 和已确认的 `folderId`。
3. 校验并创建：

```bash
apifox cli-schema validate endpoint-create --file <create.json>
apifox endpoint create --project <projectId> --branch <branch> --file <create.json>
```

4. 从结构化 JSON 获取 `endpointId`，立即执行 `endpoint get` 回读。
5. 基于回读的完整结构依次补充：
   - description
   - parameters/requestBody
   - responses
   - responseExamples
6. 每一步都先在完整结构上合并本维度，执行 `endpoint-update` schema 校验，再 `endpoint update --file`，随后再次 `get` 回读。

## 更新接口

1. 先执行 `endpoint get <endpointId>` 获取完整资源。
2. 仅修改由本次代码扫描确认的字段，保留未知字段、数组项和已有配置。
3. 使用结构化 JSON 序列化生成文件，不用字符串替换拼 payload。
4. 执行：

```bash
apifox cli-schema validate endpoint-update --file <update.json>
apifox endpoint update <endpointId> --project <projectId> --branch <branch> --file <update.json>
apifox endpoint get <endpointId> --project <projectId> --branch <branch>
```

## 失败恢复

- 本地 validate 失败：按 schema 修正，不执行远端写入。
- 服务端返回 422：保持最小创建成功结果，每次只补一个维度；schema 复杂时从顶层对象逐层恢复。
- 写入权限受限：让用户选择开启直接编辑权限或使用 AI 分支。
- 已创建骨架但后续补全失败：明确报告该 endpoint ID 和残缺范围；未经确认不要删除。
- CLI 返回成功但客户端不可见：核对 project、branch、folderId 和回读结果。

## 输出

报告扫描范围、接口数量、method + `apifoxPath` + 名称、项目/分支、目标模块与目录、测试环境模块地址检查、创建/更新的 endpoint ID、回读结果和未完成项。
