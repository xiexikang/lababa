**问题**
- 在 `src/utils/request.ts` 顶层默认初始化中使用了 `process.env.API_BASE_URL`；微信小程序环境没有 `process`，导致“ReferenceError: process is not defined”。

**修复方案**
- 改为安全取值：仅在 `typeof process !== 'undefined'` 时读取 `process.env.API_BASE_URL`，否则使用默认值。
- 代码：
  - `const envApi = (typeof process !== 'undefined' && (process as any).env?.API_BASE_URL) || ''`
  - `setBaseURL(envApi || 'http://localhost:8081')`
- 其它逻辑不变，仍可在外部覆盖 `setBaseURL` 和 `setAuthConfig`。

**步骤**
1. 更新 `src/utils/request.ts` 默认初始化块：移除直接访问 `process` 的写法，改为带 `typeof` 判空的安全读取。
2. 重新打包小程序，验证不再报错。

**验证**
- 构建通过；在微信开发者工具启动不再出现 `process is not defined`。
- 公开接口匹配 `/api/auth/**` 仍不鉴权；其他接口自动加 Authorization。

如确认，我将立即修改 `request.ts` 并打包验证。