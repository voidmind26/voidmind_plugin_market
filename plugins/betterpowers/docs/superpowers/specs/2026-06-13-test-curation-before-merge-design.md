# 合并前测试策展（Test Curation Before Merge）设计

> 日期：2026-06-13
> 范围：`plugins/betterpowers/skills/finishing-a-development-branch/`（主），`test-driven-development/`（轻量交叉引用）

## 1. 问题

betterpowers 的 TDD 红绿循环在开发中会写出一批 drop-down 的诊断/局部测试。它们在当时定位问题有用，但开发完成后多数不再有独立价值。现状是这些测试**默认全部留在仓库里**，随着迭代不断累积，导致基线回归套件越来越慢、维护成本越来越高。

TDD 技能本身已经反对"给每个函数补单测"，但它只说"诊断完回到大测试"，**没有规定这些诊断测试事后该被清理**，于是它们沉淀进了基线。

目标：让最终留存进基线的测试都是**高质量、有代表性**的；开发期产生的低价值测试在进入基线前被删除。

## 2. 已确认的决策

| 维度 | 决策 |
|------|------|
| 处置方式 | **直接删除**（不分层、不隔离） |
| 触发时机 | **合并前收口**——挂在 `finishing-a-development-branch`，merge/PR 之前 |
| 决定权 | **自动删明显冗余的；边界模糊的才列出来问用户** |

## 3. 范围边界

**策展只针对"本分支相对基线新增/修改的测试"**，通过 `git diff <base-branch>...<output-branch>` 在测试文件上的差异确定。**不触碰**基线里已有的历史测试。

理由：这个闸门的职责是"别让本分支把低价值测试倒进基线"，从源头阻止**未来**膨胀。

**已知限制（明确不在本次范围内）：** 本设计**不会缩小现有已经膨胀的基线套件**——它只防新增膨胀，不做存量清理。如果之后想一次性清理历史测试，需要另立一个独立的 `test-curation` 按需技能（用户当前已选择不做）。这意味着"基线回归慢"的存量痛点在本次改动后**不会立即缓解**，只是不再继续恶化。

## 4. 留存判定标准

对每个"本分支新增/修改"的测试，分三类处理：

### KEEP（代表性、高价值——不动）
- 证明真实工作流 / 主价值路径的 E2E / 集成测试
- 某行为唯一的覆盖来源，且该行为没有自然的 E2E 路径（TDD 的 exception path）
- 守护**独特失败模式**或**已修复 bug 回归**的小测试，且该失败模式没有被任何保留下来的更大测试完整覆盖（例如棘手边界、历史 bug 守卫）

### AUTO-DELETE（明显冗余——直接删，事后列入报告）
- 开发期为定位失败而写的 drop-down/诊断测试，其行为现已被某个保留下来的更大测试完整覆盖
- 仅测happy-path、与某 E2E 重复覆盖的琐碎单测
- 内部 helper 的测试，无独立回归价值，已被更高层测试包含

### ASK（边界模糊——列出来问用户，给默认建议后等待）
- 与某大测试重叠、但又触及大测试不一定覆盖到的边界的小测试
- 价值不明的慢测试
- 任何 agent 无法确信"删除后行为仍被覆盖"的情况

## 5. 安全不变式

1. **覆盖保全规则**：若删除某测试会导致某行为**没有任何保留测试覆盖**，则禁止自动删——它降级为 ASK。不确定是否仍被覆盖 → 一律按 ASK 处理。
2. **删除后复跑**：删完后重跑保留下来的完整套件，必须全绿。若出现新失败（说明被删测试是 load-bearing，例如承载共享 setup），**恢复该测试并重新归类**。
3. **独立提交**：策展删除作为一个单独、信息清晰的 commit（例如 `test: curate <feature> suite — remove N diagnostic tests subsumed by E2E`），便于审查和回滚。
4. **空操作短路**：若本分支没有新增/修改任何测试，闸门记录一行日志后直接跳过。

## 6. 在 `finishing-a-development-branch` 中的落点

在现有流程里插入一个新步骤 **Step 3b：Curate the branch's tests**，位于：
- **Step 1 Verify Tests（测试已全绿）之后** —— 只有套件本就是绿的才谈得上策展
- **Step 3 确定 base/output 分支之后** —— 这样 `TARGET_REPO` 与 `BASE_BRANCH` 都已解析，diff 才能算
- **Step 4（呈现 merge/PR 选项）之前** —— 保证 merge/PR 携带的是策展后的套件

新步骤流程：

```
Step 3b: Curate the branch's tests
  1. diff 出本分支相对 base 新增/修改的测试文件/用例
     若为空 → 记录"无新增测试，跳过策展" → 进入 Step 4
  2. 逐个按 §4 分类：KEEP / AUTO-DELETE / ASK
  3. 对 AUTO-DELETE：先用 §5.1 覆盖保全规则校验 → 删除
  4. 对 ASK：列出清单（每条带理由 + 推荐 keep/delete）→ 等用户拍板
  5. 重跑保留套件（§5.2）→ 必须全绿；若有新失败则恢复并重新归类
  6. 把删除作为单独 commit 提交（§5.3）
  7. 打印策展报告 → 进入 Step 4
```

同步更新该技能的：
- **Overview 的 Core principle**：`Verify tests → Curate branch tests → Resolve target repository → ...`
- **Red Flags / Common Mistakes**：新增"删除测试却没校验覆盖是否保全""删除后没复跑套件""把基线历史测试也纳入策展（越界）"等条目

### 策展报告格式

```
Test curation (vs <base-branch>):
  Kept (N):
    - <test> — <代表性理由>
  Deleted (M):
    - <test> — <为何冗余>
  Need your call (K):        # 仅当存在 ASK 项
    - <test> — <重叠/不确定点> — recommend keep/delete
Retained suite: <pass/fail>（runtime <before> → <after>，如可得）
```

## 7. 交叉引用（轻量伴随改动）

在 `test-driven-development/SKILL.md` 的 REFACTOR 小节补一句指向性说明：开发期的 drop-down 诊断测试不必纠结是否长期保留，分支收尾时 `finishing-a-development-branch` 的策展闸门会统一判定。这样实现者不会因"怕删"而默认全留。

> 这是一处一句话的交叉引用，不改 TDD 的判定逻辑本身。

## 8. 不做的事（YAGNI）

- 不做分层/隔离套件机制
- 不做存量基线的一次性清理（无独立 test-curation 技能）
- 不改 SDD / brainstorming / writing-plans 流程
- 不引入任何第三方依赖或外部工具

## 9. 验证方式

遵循本 fork "零 token 成本本地测试"约束，不新增调真实模型的行为测试。验证手段：
- shell 片段语法检查（若新步骤含 shell）：`bash -n`
- 人工走查：构造一个含 1 个 E2E + 2 个诊断单测的样例分支，确认报告分类正确、覆盖保全规则生效、删除后复跑通过
- 技能文本一致性：Overview / Red Flags / Quick Reference 与新步骤不矛盾
