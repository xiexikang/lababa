## 目标
- 消除 Dart Sass 弃用警告，兼容未来 3.0。
- 保持现有样式输出不变，最小改动完成迁移。

## 改动范围
- 定位到两处 `@import`：
  - `src/pages/index/index.vue` 的 `<style lang="scss">@import "@/pages/index/styles/Index.scss"`。
  - `src/styles/reset.scss` 的 `@import "normalize.css/normalize.css"`。

## 具体修改
### 1) 组件样式引入改用 @use
- 位置：`src/pages/index/index.vue` 样式块
- 变更：
```scss
// before
@import "@/pages/index/styles/Index.scss";

// after
@use "@/pages/index/styles/Index.scss";
```
- 说明：`Index.scss` 本身包含选择器规则，不需要命名空间引用；`@use`会直接包含模块生成的 CSS，行为与 `@import`等价但无弃用警告。

### 2) reset 中的第三方 CSS 改为标准 CSS @import 或转到入口 JS
- 位置：`src/styles/reset.scss`
- 方案A（推荐，最小改动）：改为标准 CSS 语法，避免 Sass @import：
```scss
// before
@import "normalize.css/normalize.css";

// after
@import url("normalize.css/normalize.css");
```
- 方案B（可选）：将 normalize 交给入口 JS 管理（Tree-Shaking更清晰）：
```ts
// src/app.ts
import 'normalize.css/normalize.css';
```
并从 `reset.scss` 删除该行。
- 说明：`@use/@forward` 不支持导入纯 CSS；第三方 CSS 应使用标准 `@import url(...)` 或 JS 入口引入。

## 可选优化（按需）
- 若需要全局主题变量复用：
  - 将 `src/styles/custom_theme.scss` 作为模块，通过 `@use` 引入并用 `as *` 暴露变量：
```scss
// 例：在需要的 scss 文件中
@use "@/styles/custom_theme" as *;
```
- 若有多个基础样式模块，可新增聚合文件（如 `styles/_index.scss`）并用 `@forward` 汇总，统一入口：
```scss
// styles/_index.scss
@forward "custom_theme";
@forward "reset";
// 使用方
@use "@/styles/index" as *;
```
（此优化非必需，本次迁移可不新建文件。）

## 验证步骤
- H5 构建 `npm run dev:h5` 确认样式无变化、控制台无 Sass 弃用警告。
- 小程序构建（如需）验证同上。

## 变更影响
- 无业务逻辑改动；样式加载顺序与输出保持一致。
- `normalize.css` 的引入方式改变但效果不变。

## 执行计划
1. 按上述两处修改完成最小迁移。
2. 运行开发构建并检查控制台是否还出现 `@import` 弃用警告。
3. 如需变量复用与聚合，按“可选优化”执行。

请确认采用方案A（最小改动）还是同时执行方案B（将 normalize 移到入口 JS）。确认后我将直接完成改动并验证。