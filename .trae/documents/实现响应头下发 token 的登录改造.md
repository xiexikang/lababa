## 改造目标
- 后端登录接口 `/api/auth/weapp` 不在响应体返回 token，改为通过响应头下发（`Authorization: Bearer <token>`，附带备选 `X-Token`）。
- 前端在小程序与 H5 两端均可读取响应头获取 token 并存储到 `auth-token`，其后业务请求继续带 `Authorization`。
- 受保护接口维持基于 token 的鉴权；后端支持可配置有效期。

## 后端改造
- 在 `cors` 中间件增加：`Access-Control-Expose-Headers: Authorization, X-Token`（保证 H5 浏览器能读取这两个响应头）。
- `/api/auth/weapp`：
  - 保留用户信息 `{ user, expiresAt }` 在响应体；从响应体移除 token。
  - 在响应头设置 `Authorization: Bearer <token>` 与同值 `X-Token`。
- 保持现有 `sessions` 表与 TTL（`TOKEN_TTL_SECONDS`）逻辑不变；其他受保护接口无需变动。

## 前端改造
- 登录页（`src/pages/profile/index.vue`）：
  - 登录请求使用 `Taro.request`（或在现有封装增加一个“返回原始响应”的选项）以便读取 `res.header.Authorization`/`res.header['X-Token']`。
  - 解析出 token 后写入 `Taro.setStorageSync('auth-token', token)`；仍保存 `user-info`。
- 请求封装（`src/utils/request.ts`）：
  - 现状已会自动添加 `Authorization` 头（读取 `auth-token`）；保持不变。
  - 可选：在 401 错误时触发统一登录弹窗（目前页面级已有处理，可暂不改）。

## 验证与兼容
- 小程序端：`Taro.request` 可直接读响应头，无跨域暴露限制。
- H5 端：通过 `Access-Control-Expose-Headers` 保证能读到响应头。
- 验证用例：
  1) 登录获取响应头 token 并入库；
  2) 携带 token 调用 `/api/records/create` 与 `/api/records/list`；
  3) 未携带 token 返回 401 并弹窗登录；
  4) 排行榜与统计接口在登录后正常返回。

## 交付项
- 后端：增加暴露响应头、登录接口在响应头下发 token；移除响应体的 token。
- 前端：登录流程读取响应头并保存 token；其余业务调用沿用现有封装。

确认后将开始实施上述改造并联调验证。