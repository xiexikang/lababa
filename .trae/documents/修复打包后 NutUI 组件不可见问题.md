## 问题研判
- 构建日志出现组件自动导入告警：`export 'Radiogroup' was not found in '@nutui/nutui-taro'`，说明自动按需导入对 `nut-radio-group` 进行了错误的大小写映射（Radiogroup → RadioGroup），导致组件未正确注册。
- 代码使用了自动导入（`config/index.ts:53-62, 107-116`），当前 resolver 来源为 `@nutui/auto-import-resolver`（`config/index.ts:1-2`），版本为 `^1.0.0`（`package.json:63`），可能存在上述映射问题。
- Mini 端样式转换配置中 `pxtransform.selectorBlackList` 被注释（`config/index.ts:82-87`）。NutUI 官方建议屏蔽 `nut-` 前缀样式的 px→rpx 转换，否则可能导致组件样式异常或“看不见”。
- 项目未进行全局样式注册，完全依赖自动导入的组件样式（`src/app.ts:1-24` 未注册 NutUI，也未全局引入 `@nutui/nutui-taro` 样式）。

## 修改方案
1) 修正自动导入 resolver
- 将 `import NutUIResolver from '@nutui/auto-import-resolver'` 替换为 `import NutUIResolver from '@nutui/nutui-taro/dist/resolver'`（兼容 Taro，对 `*-group` 等复合组件名映射更稳定）。
- 保持 `importStyle: 'sass', taro: true` 不变。

2) 禁用对 NutUI 选择器的 px→rpx 转换
- 在 `mini.postcss.pxtransform.config` 中启用：`selectorBlackList: ['nut-']`（`config/index.ts:82-87`）。
- 可选：在 H5 也同样配置，保持端内样式一致性。

3) 全局注册基础样式与关键组件（兜底）
- 在 `src/app.ts` 引入 `@nutui/nutui-taro/dist/style.css` 或 `@nutui/nutui-taro/dist/styles/themes/default.scss`（若需自定义主题，保留 `sass.resource` 的变量覆盖）。
- 显式全局注册常用组件以规避自动导入误映射：`Button, Popup, Input, Switch, Textarea, Radio, RadioGroup, CalendarCard`。

4) 可选健壮性优化
- 调整 `src/utils/request.ts:191` 的默认 BASE_URL 回退顺序，优先 `http://localhost:8081`，避免打包后连到不可达的内网 IP 导致页面整体无数据而误判“组件不显示”。

## 验证步骤
- 执行 `pnpm run build:weapp` 重新打包；检查是否消除 `Radiogroup` 告警。
- 打开微信开发者工具导入 `dist/weapp`，逐页验证 NutUI 组件（按钮、弹窗、输入、单选组、日历卡）是否正常显示与交互。
- 重点页面与代码：
  - 单选组使用：`src/pages/cats/index.vue:25-28`
  - 统计日历等：`src/pages/statistics/index.vue:18-26, 169-185, 187-250`
  - 弹窗与按钮：`src/pages/index/index.vue:93-126, 126-156`

## 变更清单
- `config/index.ts`：替换 resolver 导入源；开启 `selectorBlackList: ['nut-']`。
- `src/app.ts`：引入全局样式；手动注册若干 NutUI 组件。
- （可选）`src/utils/request.ts`：调整默认 BASE_URL 回退。

请确认，我将按以上方案修改并重新打包验证。