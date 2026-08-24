---
name: apifox-cli-setup
description: 安装、升级、登录和验证 Apifox CLI。当用户要求安装 Apifox CLI、检查版本、登录账号、排查 command not found 或未登录、配置默认 projectId，或其他 Apifox skill 发现 CLI 环境未就绪时使用。
---

# Apifox CLI Setup

按官方首次安装流程准备 Apifox CLI。开始前读取 `../../references/apifox-cli-rules.md`。

## 安装与版本

1. 执行 `node --version` 和 `npm --version`，确认 Node.js 与 npm 可用。
2. 执行 `command -v apifox` 和 `apifox --version`。
3. 仅在 CLI 不存在时安装：

```bash
npm i -g apifox-cli@latest --registry=https://registry.npmmirror.com/
```

4. 国内镜像安装失败时，改用官方 npm 源重试：

```bash
npm install -g apifox-cli@latest
```

5. 已安装时不要默认升级。只有 CLI 明确提示版本过低、缺少目标命令，或用户明确要求升级时，才执行升级。
6. 安装后执行 `apifox --version` 和 `apifox --help` 验证命令可用。

## 登录

1. 执行 `apifox whoami` 检查当前登录状态。
2. 已登录时不要重新索取 token。
3. 未登录时，请用户从 Apifox 的「账号设置 -> API 访问令牌」创建 token，并让用户在自己的终端执行：

```bash
apifox login --with-token <TOKEN>
```

4. 不要求用户把 token 粘贴到聊天中，也不由 agent 代为拼接含 token 的命令。不在命令回显、日志、聊天摘要或仓库文件中记录 token。CLI 自己将凭证保存在 `~/.apifox/config.toml`。
5. 登录后再次执行 `apifox whoami`。

## 项目配置

1. 执行 `apifox project list` 获取可访问项目。
2. 用户未指定项目时，先读取当前工作区的 `.apifox/settings.json`；存在 `projectId` 时优先使用。
3. 只有取得用户同意后，才创建或更新：

```json
{
  "projectId": 123456
}
```

4. 创建 `.apifox/` 时，确保 `.apifox/.gitignore` 至少包含 `*.private.*`。不要把 token 写进项目配置。
5. 已确认项目后，执行 `apifox project get <projectId>`；需要验证环境时执行 `apifox environment list --project <projectId>`。

## 完成条件

- `apifox --version` 成功。
- `apifox whoami` 返回已登录身份。
- 目标项目可由 `project get` 访问。
- 若用户要求保存默认项目，`.apifox/settings.json` 已写入并且不包含凭证。
