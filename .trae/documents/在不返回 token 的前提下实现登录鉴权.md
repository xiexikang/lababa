## 目标
- /api/auth/weapp 不在响应体返回 token
- 前端仍可获得并持久化可用的登录凭证
- 后端可配置 token 有效期，所有业务接口需登录

## 推荐方案A：响应头下发 token（不进响应体）
- 后端：登录成功时在响应头设置 `Authorization: Bearer <token>` 或 `X-Token: <token>`，并返回用户信息到响应体（不包含 token）
- 前端：在登录请求的响应中读取响应头并存储到 `auth-token`
- 业务请求：沿用现有 `Authorization: Bearer <token>` 头部鉴权
- 有效期：后端通过 `TOKEN_TTL_SECONDS` 控制，服务端校验过期
- 优点：与当前前端结构兼容，不改 Cookie/CORS 行为；满足“接口不返回 token”的约束

## 备选方案B：HttpOnly Cookie 会话
- 后端：登录成功 `Set-Cookie: sid=<random>; HttpOnly; Secure; SameSite=Lax`，服务端通过 Cookie 识别用户
- 前端：不需要读写 token，本地无需存储；仅保证后续请求携带 Cookie
- CORS 要求：
  - 不能使用 `Access-Control-Allow-Origin: *`，需精确来源域
  - 添加 `Access-Control-Allow-Credentials: true`
  - H5 端需开启 `withCredentials/credentials: 'include'`
- 注意：微信小程序对 Cookie 支持与跨域限制较多，本地联调与正式域名需分别验证

## 方案选择建议
- 开发期/联调优先采用方案A（响应头下发 token），实现成本低且与现有前端适配；正式环境可评估切换方案B提升安全性

## 具体改造点
- 后端：
  1. /api/auth/weapp 登录成功时，在响应头附加 `Authorization: Bearer <token>`（或 `X-Token`），响应体保留 `{ user, expiresAt }`（无 token）
  2. 中间件支持从 `Authorization` 或 Cookie 中解析会话，统一注入 `userId`
  3. token 有效期通过环境变量 `TOKEN_TTL_SECONDS` 配置
- 前端：
  1. 登录处（`src/pages/profile/index.vue`）读取登录响应的响应头，若存在 `Authorization/X-Token` 则写入 `auth-token`
  2. 请求工具保持在所有业务请求自动附带 `Authorization` 头；对 401 弹窗提示并触发登录
  3. 首页首屏不拉列表，按需加载；个人记录/统计/排行榜均基于服务端身份识别，无需传 `userId`

## 交付与验证
- 后端：完成响应头下发与鉴权中间件；保障排行榜与记录接口仅在登录后可用
- 前端：完成读取响应头存储 token、401 弹窗与重试逻辑；全页面联调通过

确认后我将按方案A实施，并保留切换到方案B的可扩展点（CORS、Cookie 中间件）。