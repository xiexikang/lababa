## 可能根因
- 你的 MySQL 中 `records` 表可能是旧版本，缺少我们现在插入使用的列（如 `catId`、`shape`、`amount` 等），导致插入时报错并返回 `500`。
- 目前的建表逻辑只执行 `CREATE TABLE IF NOT EXISTS`，不会对已存在表做列升级，因此无法自动修复。

## 方案
- 启动时检测 `records` 表的列集合，缺啥补啥：查询 `INFORMATION_SCHEMA.COLUMNS`，对缺失列执行 `ALTER TABLE records ADD COLUMN ...`。
- 覆盖列：`userId, catId, startTime, endTime, duration, color, status, shape, amount, note, isCompleted, createdAt`。
- 保留现有索引，若缺少必要索引也补上（如 `idx_records_user_end`、`idx_records_user_cat_end`）。
- 保持向后兼容，不影响现有数据。

## 执行
1. 新增启动迁移函数 `ensureRecordsSchema(db)` 并在 `openMySQL()` 创建表后调用。
2. 重建并重启 `server-go`。
3. 复测 `/api/records/create` 与 `/api/records/list`。

## 验证
- 成功创建返回 `code=0`，`data.record.id` 非空；列表与数据库均可见新记录。
- 如仍失败，查看启动或插入日志，定位具体数据库错误。