## 目标
- 前端对所有以 `/list` 结尾的接口改为使用 `POST`，参数通过 JSON body 传递，不再拼接到 URL。
- 后端对应列表接口优先从 `POST` 的 JSON body 读取参数，若未提供则回退到 URL Query；继续允许 `GET` 以兼容旧路径。

## 受影响调用点
- `'/api/cats/list'`：`src/pages/index/index.vue:377`，`src/pages/ranking/index.vue:270`，`src/pages/statistics/index.vue:601`，`src/pages/profile/index.vue:278`
- `'/api/ranking/list'`：`src/pages/ranking/index.vue:193,229`
- `'/api/records/list'`：`src/pages/ranking/index.vue:308`，`src/pages/statistics/index.vue:331,546`，`src/store/simple.ts:114`
- `'/api/cats/weights/list'`：`src/pages/statistics/index.vue:629`

## 前端改动
- 修改 `src/utils/request.ts` 的 `get(...)`：
  - 检测 `url` 是否以 `/list` 结尾；命中时改为 `POST` 并将 `params` 作为 `data` 发送；不再将 `params` 拼接到 URL。
  - 其它 `get` 保持现有 `GET+query`。

## 后端改动
- 增加通用读取函数：读取 JSON body 到 `map[string]any`，以及 `getBodyString/getBodyInt/getBodyInt64`。
- 在以下处理器优先读取 body 参数：
  - `/api/records/list`：`start,end,catId,pageNum,pageSize`
  - `/api/ranking/list`：`period,pageNum,pageSize`
  - `/api/cats/list`：`q,pageNum,pageSize`
  - `/api/cats/weights/list`：`catId`
  - `/api/index/list`：`start,end,catId,pageNum,pageSize`（当前前端未用，保持一致）
- 保留 `GET` 兼容；CORS 无需改动。

## 验证
- 在首页、排行榜、统计、个人中心中操作列表，确认请求方法为 `POST` 且 URL 无查询串；分页与筛选正常。

## 回滚
- 如需回退，仅还原前端 `get(...)` 的条件分支；后端保留双读取策略。

请确认后我将开始实施并完成修改与验证。