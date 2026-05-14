# apifox-codegen-plugin

用于从 Go Web 代码生成 Apifox 接口文档与场景化测试用例的统一插件。

## 核心能力

- 根据 Go Web 路由与实现生成 Apifox 接口文档。
- 基于接口与业务场景生成场景化测试用例。
- 在新增接口时校验目标模块归属与测试环境模块地址，避免接口误落默认模块。
- 在生成测试时默认产出强断言正向测试，而不是只验证 `HTTP 200`。

## 输入规范

- 默认传路由前缀，按前缀范围生成对应接口内容。
- 显式说明“全部接口”时，按全量接口范围生成。

示例：

- `生成 /user 前缀的 Apifox 接口文档`
- `先生成 /order 的接口，再补场景化测试`
- `我要生成全部接口，但先帮我确认范围`

## 使用方式

- 安装插件后，通过统一入口触发接口生成或测试生成。
- 默认优先按路由前缀工作；只有明确确认“全部接口”时才全量扫描。
- 需要先确保当前仓库属于 Go Web 风格，且可从 router -> controller -> service -> dto 链路取证。
- 若会新增 Apifox 接口，插件会继续检查：
  - 接口是否应落在目标模块
  - 测试环境中是否已为该模块配置专属地址
- 若会生成 Apifox 测试，插件默认要求：
  - 正向测试不能只验证 `HTTP 200`
  - 测试用例建议显式写入 `apiPath`
  - 后置断言优先使用 Apifox UI 可识别的 `type: assertion` 结构

## MCP 前置条件

- 插件通过 Apifox 官方 HTTP MCP 接入相关能力。
- 依赖环境变量 `APIFOX_ACCESS_TOKEN`，并通过 `Authorization: Bearer ${APIFOX_ACCESS_TOKEN}` 完成鉴权。
- 同时依赖请求头 `X-Apifox-Api-Version: 2025-09-01`。
- 使用前需要先在 Claude Code 运行环境中配置好 `APIFOX_ACCESS_TOKEN`，不要把真实 token 写入仓库。

## 环境变量配置示例

临时设置当前终端会话：

```bash
export APIFOX_ACCESS_TOKEN="your_apifox_token"
```

如果你是从当前终端启动 Claude Code，可以先在终端中执行上面的命令，再启动 `claude` 或 `zcode`。

如果你想直接在当前 Claude Code 会话里执行，也可以输入：

```bash
! export APIFOX_ACCESS_TOKEN="your_apifox_token"
```

请不要把真实 token 写入仓库文件、`.mcp.json` 或 README 示例中。

## 参考资料

- `references/apifox-write-rules.md`：接口写入顺序、模块归属与环境地址约束、422 缩圈规则。
- `references/scenario-test-rules.md`：场景化测试最低覆盖、强断言要求、测试用例 path 约束。
- `references/apifox-testcase-assertion-example.md`：真实可用的 `type: assertion` 后置断言结构与测试用例 path 样例。
