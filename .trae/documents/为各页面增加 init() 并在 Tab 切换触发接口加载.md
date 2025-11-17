**目标**

* 切换底部 Tab 时，不刷新整个应用，仅调用“当前页面所需接口”。

* 为每个页面增加 `init()` 初始化方法，统一在生命周期中调用。

**统一方法与约定**

* 新增通用方法 `ensureAuth(redirect?)`：显式交互前校验登录，无 token 时跳登录页携带回跳参数。

* 每页提供 `async init()`：只并发加载本页接口，不修改全局状态；必要时 `ensureAuth()`。

* 生命周期：`onMounted(() => init())` + `onShow(() => init())`（Taro Tab 切换回该页时触发）。

* 防抖：用 `loading`/`lastLoadedAt` 防重复请求；短时间内不重复触发。

**页面落地清单**

* 首页 `/pages/index/index`：

  * `init()`：预取邀请 `inviteId`（需鉴权）、解析并接受路由参数中的 `inviteId`、刷新首页轻量数据（如头部信息）。

  * 分享/邀请点击前 `ensureAuth()`。

* 统计页 `/pages/statistics/index`：

  * `init()`：并发拉取个人周概览（含 `score/scoreDetails`）、月日历（`days/dayStatusMap`）、最近记录分页与摘要。

  * 在评分条旁展示 `scoreDetails`。

* 排行页 `/pages/ranking/index`：

  * `init(period)`：公共榜单 `GET /api/ranking/list?period=...`（公开）+ 个人数据 `GET /api/ranking/me?period=...`、`GET /api/user/me`。

  * Period 切换仅 `init(period)`，不刷新整页。

  * 邀请/接受：按钮前 `ensureAuth()`；打开首页携带 `inviteId` 自动接受。

* 个人中心 `/pages/profile/index`：

  * `init()`：拉取个人摘要与展示；未授权时卡片/按钮执行授权。

**代码模式示例**

```ts
// 每页
const loading = ref(false)
const lastLoadedAt = ref(0)
const init = async () => {
  if (loading.value) return
  if (Date.now() - lastLoadedAt.value < 1500) return
  loading.value = true
  try {
    // Promise.all 并发拉取本页所需接口
  } finally {
    loading.value = false
    lastLoadedAt.value = Date.now()
  }
}
import { onMounted } from 'vue'
import { useDidShow as onShow } from '@tarojs/taro'
onMounted(() => init())
onShow(() => init())
```

**鉴权与回跳**

* 请求层 401/403 统一跳登录页 `/pages/login/index?redirect=<当前路由>`。

* 登录成功触发 `user-updated` 并按 `redirect` 回跳；`onShow` 自动再次调用 `init()`。

**验证**

* 切换底部 Tab 只触发当前页 `init()`，不刷新其他页。

* 未登录触发鉴权的页面或操作进入登录页；登录后回到原页并自动加载。

* H5 与小程序端行为一致。

