## 问题与原因
- 补卡与创建记录调用 `/api/records/create` 未看到记录：前端 `addRecordForDate` 未传 `catId`，后端在插入时未校验 `catId` 与用户归属，且忽略数据库错误，导致创建的记录可能为空 `catId` 或插入失败而不报错。
- 统计页 `/api/records/list` 无数据：统计页查询按猫过滤；如果记录创建时未写入 `catId`，列表自然为空。
- 摘要 `/api/statistics/summary`：已支持 `id` 别名，若记录缺少 `catId`，摘要也可能不匹配。

## 方案
- 前端：在 `store/simple.ts` 的 `addRecordForDate` 将 `catId` 写入请求 body。
- 后端：强化 `/api/records/create` 校验与错误处理：
  - 要求 `catId` 非空；校验该 `catId` 属于当前用户，否则 403。
  - 插入执行不再忽略错误；如失败返回 500，避免“创建成功但无数据”的假象。

## 修改点
- 前端：`src/store/simple.ts` 中 `addRecordForDate` 的 `payload` 增加 `catId: params.catId`。
- 后端：`server-go/main.go` 中 `/api/records/create`：
  - 判断 `catId==""` 则 400。
  - 查询 `cats` 表验证所有权；不匹配则 403。
  - `db.Exec` 与后续查询增加错误处理。

## 验证
- 在统计页补卡，确认新记录在 `records` 表存在并带有该猫的 `catId`。
- `/api/records/list?id=<catId>` 拉取可见；摘要按猫显示正确。

## 备注
- 保持现有接口的入参结构（POST+JSON body），不影响已修改的列表接口。