**目标**

* 将 `setBaseURL('http://localhost:8081')` 与 `setAuthConfig({ publicPaths: [/^\/api\/auth\//] })` 的默认初始化直接放到 `src/utils/request.ts`，减少跨文件配置。

* 保留外部覆盖能力，避免硬编码不可调整。

**实现方案**

* 在 `src/utils/request.ts` 顶层增加“惰性默认初始化”：

  * 若 `BASE_URL` 为空，则 `setBaseURL(process.env.API_BASE_URL || 'http://localhost:8081')`

  * 无 `AUTH_CONFIG.publicPaths` 时，默认 `[/^\/api\/auth\//]`

* 保留导出的 `setBaseURL` 与 `setAuthConfig`，外部仍可随时覆盖默认值。

* 删除 `src/app.ts` 中的初始化调用，避免重复设置；其它逻辑不变。

**注意点**

* `process.env.API_BASE_URL`（H5/WeApp）若未注入则回退到 `'http://localhost:8081'`。

* 不引入循环依赖；仅在 request.ts 内完成默认初始化。

**验证**

* 构建 WeApp 与 H5，确认默认配置生效；公开接口匹配 `/api/auth/` 不加鉴权，其余自动带 token。

* 401 提示与跳转行为保持一致。

**你的确认**

* 确认后我将更新 `request.ts` 添加默认初始化，并移除 `app.ts` 的两处显式设置；随后执行小程序打包验证。

