# Layering

- 先判断影响层，再确认标签，不默认扩成完整链路。
- 只为已确认标签对应的层创建任务；未命中标签必须显式标注“`不涉及`”。
- `controller` 只做参数绑定、调用 service、统一返回。
- `service` 只做业务编排、错误处理、日志记录。
- `data` 负责数据访问边界；若需求涉及 model、持久化实体或缓存对象结构，统一并入 `data` 层处理。
- 只有完整接口新增时，才允许接近 `dto+data+service+controller+router` 的全链路展开。
