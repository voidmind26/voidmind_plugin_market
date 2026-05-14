# Parent Directory Git Discovery for Worktree Skills

**Date:** 2026-05-14
**Status:** Draft

## Problem

`using-git-worktrees` 和 `finishing-a-development-branch` 都默认“当前工作目录已经位于某个 git 仓库内”。它们直接从当前目录执行 `git rev-parse --git-dir`、`git rev-parse --git-common-dir`、`git rev-parse --show-toplevel`。

当用户打开的是一个聚合父目录，而真正的 git 仓库位于某个子目录时，这个前提不成立，结果是：

1. Step 0 的 worktree/normal-repo 检测直接失败。
2. 后续的原生 `EnterWorktree` 或 `git worktree add` 都没有明确目标仓库。
3. finishing skill 里的基线分支检测、cleanup、`git worktree remove` 也无法可靠定位主仓库。

现有测试主要覆盖“当前目录就是 git repo”的场景，没有覆盖“父目录打开、子目录才是 repo”的场景。

## Goals

1. 让 worktree 相关技能支持“当前目录不是 git repo，但其下存在目标子仓库”的工作方式。
2. 保持已有“当前目录已在 repo 内”行为不变。
3. 避免在多子仓库场景下拍脑袋选错仓库。
4. 为后续的 `EnterWorktree`、git fallback、cleanup 提供统一的目标仓库解析步骤。

## Non-Goals

- 不尝试自动支持“父目录下存在多个候选仓库但用户未指定目标”的全自动选择。
- 不修改 harness/tool 本身的 `EnterWorktree` 行为；只修改 betterpowers 技能指导文本与测试。
- 不扩展到非 git VCS。

## Options Considered

### Option A — 只要求用户先 `cd` 到子仓库

**优点**
- 改动最小。
- 不引入新的检测逻辑。

**缺点**
- 不能解决真实痛点；技能仍然在常见 monorepo/聚合目录工作流中失效。
- 把环境差异留给用户手动处理，和 skill 的“指导工作流”目标不符。

### Option B — 向下自动发现唯一子仓库并绑定后续步骤（推荐）

**优点**
- 对“父目录打开、下面只有一个 git 仓库”最顺滑。
- 保持技能主流程不变，只是在最前面增加仓库解析层。
- 可以明确规定：多个候选时不猜，必须让用户指定。

**缺点**
- 技能文本和测试都要补充。
- 需要在多个 skill 中复用同一套解析约定。

### Option C — 扫描到多个仓库时也自动选“最可能的一个”

**优点**
- 交互更少。

**缺点**
- 错选仓库的代价高。
- “最可能”标准很难稳定，容易变成不透明启发式。

## Decision

采用 **Option B**，但增加一条更高优先级规则：**始终先按当前任务所在的仓库识别目标仓库**。

技能先判断当前任务是否已经明确指向某个仓库；如果已明确，则直接以该仓库为目标，不再依据当前打开目录猜测。只有在任务上下文没有明确仓库时，才退回到当前目录/子目录发现流程；如果此时仍然无法唯一确定目标仓库，则应直接询问用户，而不是继续猜测。

具体分流为：

- **任务已明确指向仓库**：直接以该仓库为目标，后续所有 git 检测、`EnterWorktree` 意图、git fallback、cleanup 都以它为准。
- **任务未明确指向仓库，且当前目录本身在 repo 内**：沿用当前目录所在 repo。
- **任务未明确指向仓库，且当前目录不在 repo 内**：再向下扫描子目录中的 git 仓库。
  - **0 个候选**：当前上下文仍无法确定目标仓库，应直接询问用户目标仓库，而不是继续推进。
  - **1 个候选**：将该子仓库视为目标仓库并继续。
  - **多个候选**：当前上下文仍无法唯一确定目标仓库，应直接询问用户指定目标仓库路径；不要猜。

## Design

### 1. 在 `using-git-worktrees` 增加“目标仓库解析”前置步骤

在现有 Step 0 之前增加一个新的前置步骤，逻辑如下：

1. 先判断**当前任务是否已经明确指向某个仓库**。
2. 如果已明确，则直接绑定该仓库，后续步骤都在该仓库语义下执行。
3. 如果任务未明确仓库，再尝试判断当前目录是否已在 git 仓库内。
4. 如果当前目录在 repo 内，行为与今天一致。
5. 如果当前目录不在 repo 内，则在当前目录下一层或受限深度内查找 git 仓库候选。
6. 按候选数量分流：0/1/多。
7. 只要任务上下文 + 当前目录 + 子目录发现仍不足以唯一确定目标仓库，就立即询问用户。

建议在技能中使用平台无关的指导表达，而不是过度规定扫描命令细节；但必须明确这些行为约束：

- **任务仓库优先于当前目录**；
- 只有在任务上下文没有明确仓库时，才使用“当前目录所在 repo”；
- 只有在前两者都无法确定时，才向下发现子 repo；
- 只在“唯一候选”时自动继续；
- 多候选必须停下并让用户指定。
- 只要无法唯一确定目标仓库，就直接询问用户。

### 2. 后续所有 git 命令都必须以“解析出的目标仓库”为上下文

这里的“目标仓库”优先来自当前任务上下文，其次才是当前目录或向下发现得到的仓库。

`using-git-worktrees` 中后续依赖 repo 的步骤，需要从“默认当前目录”改成“默认目标仓库”。包括：

- `git rev-parse --git-dir`
- `git rev-parse --git-common-dir`
- `git branch --show-current`
- `git rev-parse --show-toplevel`
- `.worktrees/` / `worktrees/` 目录选择
- `.gitignore` 校验
- `git worktree add`

核心要求：即使用户会话起点在父目录，技能也应先把目标仓库语义讲清楚，再执行原来的 detect-and-defer 流程。

### 3. `finishing-a-development-branch` 复用同一解析约定

该 skill 的环境检测与 cleanup 同样需要在“目标仓库”上下文中进行，否则会在父目录场景下再次失败。

需要同步调整的点：

- Step 2 的环境检测
- 基线分支确定
- `MAIN_ROOT` 计算
- provenance-based cleanup
- `git worktree remove` 前的主仓库切换说明

### 4. 多仓库场景保持保守

如果当前任务已经明确指向某个仓库，则即使父目录下存在多个 git 仓库，也不应进入多候选歧义分支。

只有在**任务未明确仓库**时，才适用下面这条规则：如果当前父目录下发现多个 git 仓库：

- 不根据名字、最近修改时间、是否包含某文件来猜测；
- 直接要求用户指定目标仓库；
- 技能可提示“请在目标仓库根目录重新运行，或明确指定子仓库路径”。

这样可以避免在聚合目录里误对错误仓库创建/清理 worktree。

## Testing

至少补三类测试：

1. **唯一子仓库场景**
   - 当前目录不是 git repo。
   - 仅有一个子目录是 git repo。
   - 期望：技能识别目标仓库，并继续使用 native tool / git fallback 流程。

2. **无候选仓库场景**
   - 当前任务未明确仓库。
   - 当前目录不是 git repo。
   - 子目录下也没有 git 仓库。
   - 期望：技能直接询问用户目标仓库，而不是继续推进。

3. **多子仓库场景**
   - 当前目录不是 git repo。
   - 存在两个及以上子仓库。
   - 期望：技能不自动选择，转为要求用户指定。

4. **回归场景**
   - 当前目录本身就是 git repo。
   - 期望：保持现有行为，不受影响。

若补 finishing skill 测试，则应覆盖：

- 从父目录启动时仍能正确识别目标 repo；
- cleanup 逻辑基于目标 repo 而非父目录；
- 多子仓库场景下不执行危险 cleanup。

## Files Likely Affected

- `skills/using-git-worktrees/SKILL.md`
- `skills/finishing-a-development-branch/SKILL.md`
- `tests/claude-code/test-worktree-native-preference.sh`
- 可能新增一个更直接覆盖“父目录打开”场景的 skill 行为测试脚本

## Review Focus

请重点确认这五个边界：

1. **任务仓库优先于当前打开目录** 是否符合你的预期。
2. **唯一子仓库时自动继续** 是否仅应作为任务仓库缺失时的兜底。
3. **无法唯一确定仓库时直接询问用户** 是否符合你的预期。
4. **多个子仓库时必须停下让用户指定** 是否足够保守。
5. 这次修复是否只应覆盖 betterpowers 的 skill 文本/测试，而不去改 Claude Code 原生工具行为。
