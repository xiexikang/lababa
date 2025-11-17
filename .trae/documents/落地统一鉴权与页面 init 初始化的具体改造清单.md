**总体目标**

* 完成统一登录跳转与回跳；为各页面增加 `init()` 初始化方法，Tab 切换只调用该页接口不刷新整站；排行页除榜单外按个人接口查询；统计页展示评分明细。

**统一鉴权**

* 在现有请求层（已实现 401/403 跳转）增加 `ensureAuth(redirect?)` 方法（放在 request.ts），用于显式交互前校验：有 token 返回 true；无 token `navigateTo('/pages/login/index?redirect=...')` 并返回 false。

* 登录页 `/pages/login/index`：保持现流程 `Taro.login + Taro.getUserProfile → /api/auth/weapp`，写入 `auth-token`/`user-info`，触发 `user-updated`，按 `redirect` 回跳。

**页面 init() 与生命周期**

* `pages/index/index.vue`：新增 `async init()`（预取邀请 id、接受邀请、刷新首页轻量数据），在 `onMounted/onShow` 调用；受保护操作前 `ensureAuth()`。

* `pages/statistics/index.vue`：将 `loadStatistics()` 重命名为 `init()`；`onMounted/onShow` 调用；评分处渲染 `week.scoreDetails`（接口返回）。

* `pages/ranking/index.vue`：新增 `async init(period)`：

  * 公共榜：`GET /api/ranking/list?period=...`（public）

  * 我的排名：`GET /api/ranking/me?period=...`（个人）

  * 我的头像昵称：`GET /api/user/me` 或本地 `user-info`

  * 在 `onMounted/onShow` 调用；period 切换仅调用 `init(period)` 不刷新整页。

* `pages/profile/index.vue`：新增 `async init()` 拉取个人摘要数据与展示；`onMounted/onShow` 调用。

**好友邀请**

* 生成邀请：在排行/首页的按钮点击：`ensureAuth()` 后调用 `POST /api/friends/invite`，将 `inviteId` 注入分享参数。

* 接受邀请：首页读取 `inviteId` 参数后 `POST /api/friends/accept`；失败 toast 提示；接口缺失时降级为提示“邀请功能暂不可用”。

**统计页评分明细**

* 接口 `GET /api/statistics/week` 返回 `scoreDetails: Array<{label:string,value:number}>`；在评分条下方渲染明细列表；色条阈值不变。

**验证**

* 未登录访问或交互 → 统一跳登录页 → 登录成功自动回到原页面并在 `onShow` 执行 `init()`。

* Tab 切换只调用该页 `init()` 更新数据；排行页除榜单外均为个人数据；统计页显示评分明细；H5/小程序一致。

**若你确认**

* 我将为各页面补充 `init()` 并接入接口，增加 `ensureAuth()`，完善排行页个人区块与评分明细，实现不刷新页面的 Tab 切换行为。

