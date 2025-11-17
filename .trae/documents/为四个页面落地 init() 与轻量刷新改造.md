**总体策略**

* 每个页面新增 `init()`，仅并发请求本页所需接口，不刷新整站；在 `onMounted` 与 `onShow` 调用。

* 显式交互前统一用 `ensureAuth(redirect?)` 校验；请求层 401/403 已统一跳转登录并回跳。

**统计页（pages/statistics/index.vue）**

* 替换 `store.init()` 为 `init()`：并发拉取个人周概览（含 `score/scoreDetails`）、月日历（`days/dayStatusMap`）、最近分页与摘要。

* 增加 `loading/lastLoadedAt` 防抖；在评分条旁展示 `scoreDetails` 明细。

* 生命周期：`onMounted(() => init())` + `onShow(() => init())`。

**首页（pages/index/index.vue）**

* 新增 `init()`：预取邀请 `inviteId`（需鉴权）、解析并接受参数中的 `inviteId`、刷新首页轻量数据。

* 分享按钮点击前 `ensureAuth()`；生命周期：`onMounted/onShow` 均调用 `init()`。

**排行页（pages/ranking/index.vue）**

* 新增 `init(period)`：

  * 公共榜单：`GET /api/ranking/list?period=...`（公开）

  * 个人区块：`GET /api/ranking/me?period=...`、`GET /api/user/me`

* 切换 period 仅调用 `init(period)`，不刷新整页；`onMounted/onShow` 调用。

* 邀请与接受：按钮前 `ensureAuth()`，打开首页携带 `inviteId` 自动 `accept`。

**个人中心（pages/profile/index.vue）**

* 新增 `init()`：拉取个人摘要与展示；未授权卡片/按钮触发授权。

* 生命周期：`onMounted/onShow` 调用。

**请求层与公共接口声明**

* 增加 `ensureAuth(redirect?)`（导出），用于显式交互前校验。

* 将 `/api/ranking/list` 声明为公开接口（`requireAuth:false` 或加入 `publicPaths`）。

**验证**

* Tab 切换仅触发当前页 `init()`；未登录进入登录页，登录成功按 `redirect` 回到原页并在 `onShow` 自动 `init()`。

* 排行页除榜单外均为个人数据；统计页评分明细显示；H5/小程序一致。

