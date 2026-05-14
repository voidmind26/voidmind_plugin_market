# Apifox 测试用例断言真实样例

本文件沉淀一次真实验证通过的 Apifox 测试用例结构，用于指导插件后续生成测试用例时避免踩坑。

## 1. 测试用例 path 不要留空

当测试用例界面需要清楚展示请求路径时，测试用例应显式写入接口真实 `apiPath`。

示例：

```json
{
  "path": "/images/get_image_list"
}
```

如果 `path` 留空，虽然接口本身可能已经有 path，但测试用例界面中看不到这段 API 路径，容易误判为 URL 缺少接口部分。

## 2. 标准后置断言结构

Apifox UI 可正常识别的后置断言，应使用 `type: assertion`，而不是 `type: assert`。

### 单条断言样例

```json
{
  "type": "assertion",
  "data": {
    "name": "无报错",
    "subject": "responseJson",
    "comparison": "equal",
    "value": "0",
    "path": "$.errNo",
    "multipleValue": [],
    "extractSettings": {
      "expression": "$.errNo",
      "continueExtractorSettings": {
        "isContinueExtractValue": false,
        "JsonArrayValueIndexValue": ""
      }
    }
  },
  "defaultEnable": false,
  "enable": true
}
```

## 3. 已验证可接受的 comparison

当前真实验证可接受的 comparison 包括：

- `equal`
- `greaterThanOrEqual`
- `lessThanOrEqual`

## 4. 已验证可接受的 path 样例

- `$.errNo`
- `$.data.pn`
- `$.data.rn`
- `$.data.total`
- `$.data.list.length`

## 5. 当前接口的强断言正向测试样例

接口：`获取镜像列表`

推荐请求参数：

```json
{
  "query": [
    {"name": "pn", "value": "1"},
    {"name": "rn", "value": "1"}
  ]
}
```

推荐最小强断言：

1. `$.errNo == 0`
2. `$.data.pn == 1`
3. `$.data.rn == 1`
4. `$.data.total >= 1`
5. `$.data.list.length <= 1`

## 6. 踩坑结论

### 坑 1：`type: assert` 会在 UI 中显示为空白项

现象：
- MCP 更新成功
- `getTestCase` 可读回对象
- 但 Apifox UI 的后置操作面板显示为空白

原因：
- 使用了后端可存但前端不识别的非标准结构

修正：
- 统一改用 `type: assertion`

### 坑 2：测试用例 path 留空会误判为 URL 缺少 API 部分

现象：
- 环境模块地址其实已经配置正确
- 但测试用例界面中看不到 `/images/get_image_list`

修正：
- 为测试用例显式写入 `path: "/images/get_image_list"`
