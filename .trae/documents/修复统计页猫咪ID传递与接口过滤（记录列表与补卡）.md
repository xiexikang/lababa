## 问题诊断
- 记录列表无数据：统计页在挂载时可能在 `activeCatId` 未就绪时发起请求；已顺序化加载与判空保护，但仍需确保全链路均携带猫咪ID。
- 统计摘要未按猫过滤：`/api/statistics/summary` 仅在传入 `catId/id` 时才执行过滤；前端需保证传参为 `id`。
- 补卡页面未传猫咪ID：`pages/statistics/detail/index.vue` 的补卡提交未携带 `catId`，服务端会按用户ID创建但无法归属到具体猫。

## 修改方案
- 导航携带ID：在统计页点击日历跳转详情时，URL增加 `id=${activeCatId}`，确保补卡页拿到猫咪ID。
- 补卡传ID：在 `detail/index.vue` 解析路由的 `id`，`confirmFill` 时调用 `store.addRecordForDate({ date, catId: id, ... })`。
- 日详情过滤：`dayRecords` 同时按日期与 `catId` 过滤，避免混入其他猫。
- 接口参数一致：
  - 统计页所有接口统一传 `{ id }`（已覆盖列表/摘要/月视图/体重/提醒设置/模板）。
  - 保留已做的判空保护与顺序加载，确保 `activeCatId` 存在后再发请求。
- 后端核对：当前 `/api/records/list`、`/api/statistics/summary`、`/api/statistics/month-days`、`/api/cats/settings/get`、`/api/cats/weights/list` 均已兼容 `id` 别名；无需再改动SQL。

## 实施变更
- `src/pages/statistics/index.vue`
  - 在 `onDayClick` 的跳转加入 `&id=${activeCatId.value}`。
- `src/pages/statistics/detail/index.vue`
  - 读取 `router?.params?.id` 为 `catIdParam`。
  - `confirmFill` 时传 `catId: catIdParam` 给 `store.addRecordForDate(...)`。
  - `dayRecords` 过滤条件增加 `String(r.catId) === catIdParam`。

## 验证
- 进入统计页后，列表、摘要均有数据；打开某日详情补卡成功，新增记录归属到当前猫；返回统计页后列表与周统计更新。

## 备注
- 如需强制后端在摘要缺少 `id` 时返回错误，可再加校验；当前保持兼容以避免旧页面报错。