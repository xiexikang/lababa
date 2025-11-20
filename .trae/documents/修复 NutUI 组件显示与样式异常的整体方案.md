## 问题诊断结论
- H5 构建未正确编译 NutUI 包：`config/index.ts:136` 使用了错误的包名，`esnextModules` 应使用带作用域的真实包名。
- 组件注册方式混杂：同时存在手动注册与自动按需导入，易导致部分页面未正确引入组件（类型文件也显示只识别了 Button/Popup）。
- 主题样式注入已配置，但样式异常可能源于 H5 侧未编译依赖或按需样式未被正确解析。

## 具体修复项
1. 修正 H5 侧编译白名单
- 将 `config/index.ts:136` 的 `esnextModules: ['nutui-taro', 'icons-vue-taro']` 改为 `['@nutui/nutui-taro', '@nutui/icons-vue-taro']`，确保 H5 正确编译 NutUI 依赖。

2. 统一组件注册方式为“自动按需”
- 移除 `src/app.ts:23-26` 的手动注册：`App.use(Button/Popup/Calendar/CalendarCard)`，完全交由 `unplugin-vue-components` + `NutUIResolver` 管理（已在 `config/index.ts:54-61`、`98-106` 配置）。
- 可选：为按需类型生成开启 `dts`，让 `components.d.ts` 自动保持最新，便于在开发中查类型（在两处 `Components({...})` 中增加 `dts: true`）。

3. 样式与主题确认
- 保持 `sass.data` 注入变量：`config/index.ts:50` 引入 `@nutui/nutui-taro/dist/styles/variables.scss`；自定义主题集中在 `src/styles/custom_theme.scss:1-3`，无需额外全局样式引入。
- 若仍有尺寸不一致问题，可启用 `mini.postcss.pxtransform.config.selectorBlackList: ['nut-']` 作为补救，但当前 `designWidth` 差异化配置已正确（`config/index.ts:11-22`）。

## 验证步骤
- H5：运行 `npm run dev:h5`，逐页检查含 NutUI 的页面（如 `src/pages/index/index.vue`、`src/pages/cats/index.vue`、`src/pages/statistics/index.vue`）按钮、弹窗、输入类组件是否正常显示与交互。
- WEAPP：运行 `npm run dev:weapp`，确认公共组件（如 `<nut-button>`、`<nut-popup>`）正常。`src/pages/cats/index.vue` 在 WEAPP 分支使用原生组件，不受 NutUI 影响。
- 类型校验：确认 `components.d.ts:12-15` 能包含更多 NutUI 组件声明（开启 `dts` 后自动维护）。

## 风险与回滚
- 统一到自动按需导入后，如个别页面仍缺组件，可针对该页面临时显式导入对应组件；或恢复 `App.use(...)`（不推荐）。
- `esnextModules` 修改仅作用于 H5 构建，不影响小程序端。

## 我将执行的改动
- 编辑 `config/index.ts:136` 为正确的包名作用域。
- 删除 `src/app.ts:23-26` 的手动注册语句。
- 可选：在两处 `Components({...})` 增加 `dts: true`，保持类型文件更新。
- 本地启动 H5 与 WEAPP，逐页验证显示与样式。