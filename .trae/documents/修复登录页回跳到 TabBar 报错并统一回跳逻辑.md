**问题**
- 登录页报错："redirectTo:fail can not redirectTo a tabbar page"。原因是回跳时使用了 `redirectTo` 路由到 TabBar 页面（TabBar 页面必须使用 `switchTab`，且不能携带 query）。

**修复方案**
- 在 `/pages/login/index.vue` 中统一回跳逻辑：
  - 解析 `redirect` 参数，提取纯路径 `path` 与 `query`；
  - 若 `path` 属于 TabBar（例如 `/pages/index/index`、`/pages/statistics/index`、`/pages/profile/index`、`/pages/ranking/index`）：
    - 使用 `Taro.switchTab({ url: path })`，忽略 query（TabBar 不支持 query）；
    - 如需传参，使用 `Taro.setStorageSync('redirect-query', query)` 或事件总线在目标页 `onShow` 读取。
  - 否则使用 `Taro.redirectTo({ url: path + (query ? '?' + queryString : '') })`。
  - 若 `redirect` 无效或为空，默认 `switchTab({ url: '/pages/index/index' })`。

**实现要点**
- 新增工具函数 `parseRoute(redirect)`：返回 `{ path, queryObj, queryString }`；路径通过 `split('?')[0]`，query 通过 `URLSearchParams` 或手动解析。
- 维护 `tabPages` 集合用于判断（已存在，改用 `path` 判断，不含 query）。
- 在 `ensureAuth(redirect?)` 中保持当前路由拼接逻辑，登录完成后按上述规则回跳。

**验证**
- 从未登录状态点击：
  - TabBar 入口（首页、统计、个人中心、排行）→ 登录成功应 `switchTab` 到目标页，无报错；
  - 非 TabBar 页面（详情页等）→ 登录成功应 `redirectTo` 到目标页；
- 若目标页需要参数：
  - TabBar 页通过存储或事件在 `onShow` 读取；非 TabBar 页使用 `redirectTo` 携带 query。

**可选优化**
- 将 `redirect` 的 query 统一存储为 `last-redirect-query`，目标页在 `onShow` 读取并清理，避免多次使用。

确认后我将更新登录页的回跳逻辑并进行打包验证，确保不再出现该报错。