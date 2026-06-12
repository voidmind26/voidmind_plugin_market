# Betterpowers 测试说明

本文说明如何测试 `betterpowers` 插件。当前 fork 不再维护依赖真实模型会话的行为测试；默认测试必须是零 token 成本、可本地重复运行的静态或脚本测试。

## 测试结构

```text
tests/
├── brainstorm-server/           # brainstorming 本地服务与协议测试
└── codex-plugin-sync/           # Codex 插件同步脚本测试
```

已删除的测试类型：

- 通过真实 Claude Code 会话验证技能触发的测试。
- 多轮对话、模型输出 grep、subagent 工作流行为测试。
- 任何默认会消耗 token 或依赖模型措辞稳定性的测试。

## 常用命令

### Brainstorm server

```bash
cd tests/brainstorm-server
npm test
node ws-protocol.test.js
./windows-lifecycle.test.sh
```

### Codex 同步脚本

```bash
cd tests/codex-plugin-sync
./test-sync-to-codex-plugin.sh
```

### Shell 语法检查

```bash
while IFS= read -r file; do
  bash -n "$file"
done < <(rg --files tests -g '*.sh')
```

## 测试策略

- 默认测试不调用模型，不运行 `claude -p`，不要求外部 agent harness。
- 技能行为通过人工验收、真实使用反馈和必要时的临时手动验证来评估。
- 如果未来确实需要模型行为验证，应作为手动实验脚本单独存放，并明确标注 token 成本，不纳入默认测试目录。
