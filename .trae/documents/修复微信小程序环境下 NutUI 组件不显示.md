**问题定位**

* 项目使用 `vue3 + Taro + @nutui/nutui-taro`，样式与组件在入口已全局引入与注册：`src/app.ts:6-7,24`。

* 多个页面在微信环境(`env==='WEAPP'`)下采用条件渲染，分支里大量使用原生组件(`picker/switch/textarea`)，导致在小程序开发工具里看不到对应的 NutUI 组件：

  * `src/pages/cats/index.vue:19,34,46,56,66,75`

  * `src/pages/index/components/DetailRecordPopup.vue:97,119`

  * `src/pages/statistics/index.vue:173,193,202,211`

  * `src/pages/profile/index.vue:20,67,80,98,106,267`

  * `src/pages/login/index.vue:8`

* 构建配置已按 NutUI 官方建议设置：`designWidth` 对 NutUI 指定 `375`，`pxtransform.selectorBlackList: ['nut-']`，按需引入与 `sass` 已开启：`config/index.ts:11-16,82-87,46-51,52-62`。

**修复思路**

* 目标：在微信小程序端也统一展示 NutUI 组件，避免被 `env==='WEAPP'` 分支替换成原生控件。

* 方案：

  * 第一阶段：将上述页面里对 `env==='WEAPP'` 的 UI 分支调整为一致使用 NutUI 组件（删除或改写 `v-if`/`v-else`，保留原有 `v-model` 与事件）。

  * 第二阶段：保留必要的微信特有逻辑（如分享、订阅消息）不变，仅对 UI 控件层面统一。

  * 第三阶段：核验全局样式与注册（入口已正确）并在真机/开发者工具验证视觉与交互。

**拟修改文件与要点**

* `src/pages/cats/index.vue`

  * 将性别/生日/开关/备注等原生控件替换为对应的 `<nut-radio-group>/<nut-input>/<nut-switch>/<nut-textarea>`，删除 `v-if="env==='WEAPP'"` 分支，仅保留 NutUI。

* `src/pages/index/components/DetailRecordPopup.vue`

  * 将备注输入与底部按钮统一为 NutUI 版本，删除 `WEAPP` 条件分支。

* `src/pages/statistics/index.vue`

  * 将新增体重输入、多个提醒开关统一为 `<nut-input>/<nut-switch>`，移除条件分支。

* `src/pages/profile/index.vue`

  * 授权按钮、猫咪列表操作按钮、编辑弹窗按钮统一保持 NutUI，移除条件分支。

* `src/pages/login/index.vue`

  * 登录按钮保持 NutUI（按文案差异仅保留一版或用同一组件根据环境改变文案）。

**验证步骤**

* 启动小程序开发模式：`pnpm dev:weapp` 或 `npm run dev:weapp`，在微信开发者工具中加载 `dist/weapp`。

* 快速核验页面：

  * 猫咪创建页 `cats`：检查输入、选择、开关与备注的 NutUI 渲染与交互。

  * 统计页 `statistics`：检查弹窗输入与开关。

  * 详情弹窗 `DetailRecordPopup`：检查文本域与底部按钮。

  * 个人中心与登录页：检查按钮显示与点击逻辑。

* 真机预览一次，确认样式（`default.scss` 主题与 `custom_theme.scss`）与 `pxtransform` 效果正常。

**风险与回滚**

* 风险：部分原生控件的行为与 NutUI 在小程序端仍有细微差异，需做适配（如 `picker` 的选择器交互改为 NutUI 组件组合）。

* 回滚：保留当前原生分支代码作为注释或通过 `git` 回滚；如因交互不满足再对个别点恢复原生实现。

**备注**

* 入口已正确全局注册与样式引入：`src/app.ts:6-7,24`，无需额外变更。

* 构建链对 NutUI 的 `sass`、按需与尺寸适配已就绪：`config/index.ts:11-16,46-62,82-87,150-156`。

