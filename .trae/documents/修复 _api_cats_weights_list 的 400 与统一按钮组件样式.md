## 问题定位
- 权重列表 400：后端在 `/api/cats/weights/list` 为空 `catId` 时返回 400（server-go/main.go:1318 处理器）。现读取 `catId` 或 `id`（body），若未传或为空即报错。统计页调用点为 `src/pages/statistics/index.vue:629`，理论上传 `{ id, pageNum, pageSize }`；若 `activeCatId` 未就绪或传了 `undefined`，则会触发 400。
- UI按钮样式异常：部分页面在 WEAPP 环境使用原生 `<button class="nut-button">`，而不是 `<nut-button>` 组件（例如统计页头部与弹窗操作）。原生按钮不会套用 NutUI 的样式，造成“按钮不见了/样式异常”。

## 修复方案
- 保证 catId 传递：
  1) 在统计页 `loadWeights`、`loadRecent`、`loadSummary` 已增判空；进一步移除 `|| undefined`，只在有值时传递 `id`；保持无 `id` 就不调用。
  2) 确认 `activeCatId` 的设定流程：`loadCats()` 成功后设置首个猫 ID（src/pages/statistics/index.vue:603-605），并顺序执行 `await loadCats(); await loadStatistics(); await loadWeights();`（589-593）。再确保所有并发拉取前均检查 `activeCatId`（532-587）。
  3) 若仍需兜底：后端在 `/api/cats/weights/list` 空 `id` 时返回空列表而非 400（可选）。本次保持严格校验，前端负责传参。
- 统一按钮组件：
  1) 将统计页所有原生 `<button>` 改为 `<nut-button>`，避免样式丢失；包括头部“切换”、弹窗的“取消/保存”、订阅消息按钮等。
  2) 排行榜页已经统一使用 `<nut-button>`（检查并保持一致）。

## 具体修改点
- 前端：
  - `src/pages/statistics/index.vue`：移除 `id: activeCatId.value || undefined` 的 `undefined` 传递，保证只在存在时发送；所有原生 `<button>` 改为 `<nut-button>` 以统一样式。
- 后端：
  - 无需改动逻辑；保留严格校验以确保接口一致性。

## 验证
- 切换猫后，`/api/cats/weights/list` 返回 200，带分页 `items`；列表、摘要与月视图均有数据。
- 所有按钮使用 NutUI 样式正常显示（头部与弹窗）。

## 备注
- 构建警告为 Sass 废弃与样式顺序冲突，不影响运行；统一组件后可减少样式不一致问题。