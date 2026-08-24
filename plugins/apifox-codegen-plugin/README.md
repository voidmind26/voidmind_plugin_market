# apifox-codegen-plugin

用于通过官方 Apifox CLI 从 Go Web 代码生成接口文档与场景化测试用例。插件直接调用本机 `apifox` 命令，不注册 Apifox MCP 服务。

## 核心能力

- 根据 Go Web 路由与实现生成 Apifox 接口文档。
- 基于接口与业务场景生成场景化测试用例。
- 安装、登录并验证官方 Apifox CLI。
- 在新增接口时校验目标模块归属与测试环境模块地址，避免接口误落默认模块。
- 所有 JSON 写入先按 CLI 当前 schema 校验，写入后立即回读。
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
- CLI 缺失、未登录或项目未配置时，先使用 `apifox-cli-setup` 完成环境准备。
- 默认优先按路由前缀工作；只有明确确认“全部接口”时才全量扫描。
- 需要先确保当前仓库属于 Go Web 风格，且可从 router -> controller -> service -> dto 链路取证。
- 若会新增 Apifox 接口，插件会继续检查：
  - 接口是否应落在目标模块
  - 测试环境中是否已为该模块配置专属地址
- 若会生成 Apifox 测试，插件默认要求：
  - 正向测试不能只验证 `HTTP 200`
  - 测试用例建议显式写入 endpoint 的真实 `apifoxPath`
  - 后置断言优先使用 Apifox UI 可识别的 `type: assertion` 结构

## CLI 前置条件

- 本机已安装 Node.js 与 npm。
- 按[官方安装说明](https://apifox.com/apifox-cli-installation-guide.md)安装最新 CLI：

```bash
npm i -g apifox-cli@latest --registry=https://registry.npmmirror.com/
```

国内镜像不可用时改用官方 npm 源：

```bash
npm install -g apifox-cli@latest
```

安装后验证：

```bash
apifox --version
apifox --help
```

## 登录与项目配置

先检查登录状态：

```bash
apifox whoami
```

未登录时，在自己的终端中使用 Apifox API 访问令牌登录，不要把 token 粘贴到聊天、日志或仓库文件：

```bash
apifox login --with-token <TOKEN>
```

执行 `apifox project list` 获取可访问项目。可选地在业务仓库的 `.apifox/settings.json` 保存默认项目：

```json
{
  "projectId": 123456
}
```

`.apifox/settings.json` 只能保存非敏感项目配置；凭证由 CLI 保存在用户配置目录。

## CLI 写入约束

所有远端写入遵循：

1. 用 `--help` 和 `apifox cli-schema get` 确认当前命令与 JSON 结构。
2. 用 `list/get` 确认项目、分支、目录、环境和现有资源。
3. 在临时目录生成 payload，并执行 `apifox cli-schema validate`。
4. 校验通过后执行 `create/update`，随后用 `get` 回读。
5. 接口测试创建或更新后至少执行一次 `apifox test-case run`。

## 参考资料

- `references/apifox-cli-rules.md`：CLI 预检、命令发现、schema 校验、权限与回读规则。
- `references/apifox-write-rules.md`：接口写入顺序、模块归属与环境地址约束、422 缩圈规则。
- `references/scenario-test-rules.md`：场景化测试最低覆盖、强断言要求、测试用例 path 约束。
- `references/apifox-testcase-assertion-example.md`：真实可用的 `type: assertion` 后置断言结构与测试用例 path 样例。
