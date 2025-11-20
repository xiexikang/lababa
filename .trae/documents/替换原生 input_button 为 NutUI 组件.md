## 改造范围
- 将项目中残留的原生 `<input>`、`<button>` 全量替换为 NutUI Taro 组件：`<nut-input>`、`<nut-button>`。
- 目标文件（示例包含行号，以便定位）：
  - `src/pages/cats/index.vue`：第 10、47 行为 `<input>`；第 76 行为 `<button>` 保存。
  - `src/pages/login/index.vue`：第 8 行为 `<button>` 微信授权登录。
  - `src/pages/index/components/DetailRecordPopup.vue`：第 120、121 行为确认/取消 `<button>`。
  - `src/pages/statistics/index.vue`：第 174 行为 `<input>` 添加体重。
  - `src/pages/profile/index.vue`：第 21、68、81、82、107、108 行为操作 `<button>`；第 99 行为 `<input>` 昵称。
  - `src/components/ErrorBoundary.vue`：第 9 行为重试 `<button>`。
- 项目已存在依赖：`@nutui/nutui-taro`、`@nutui/icons-vue-taro`、`@nutui/auto-import-resolver`，无需新增依赖。

## 组件替换规则
- 输入框：
  - 原生：`<input :value="x" type="digit" placeholder="..." @input="e=>x=e.detail.value" />`
  - 替换：`<nut-input v-model="x" type="number" placeholder="..." clearable />`
  - 说明：`type="digit"` 统一映射为 `type="number"`；事件改为 `v-model` 或 `@change`（若需拦截值可用 `@change="(v)=>x=v"`）。
- 按钮：
  - 原生：`<button :disabled="disabled" @tap="onClick">文本</button>`
  - 替换：`<nut-button :disabled="disabled" type="primary" block @click="onClick">文本</nut-button>`
  - 说明：统一使用 `@click`（Taro/NutUI 默认 click），可按场景设置 `type`（`primary`、`default`、`success`、`danger`）、`size`、`loading`、`shape`、`plain` 等。
- 事件与双向绑定：
  - 输入：优先 `v-model`；需要校验时配合 `@change`、`@blur`。
  - 按钮：统一 `@click`，移除小程序特有的 `@tap`。

## 样式与主题
- 保留页面已有容器样式（如 `Index-root`、布局容器等）。
- 对仍使用 `class="nut-button"` 的原生按钮，迁移为真正的 `<nut-button>` 后，清理冗余自定义样式，只在必要处通过外层容器或 NutUI 提供的属性控制尺寸与布局（如 `block`）。
- 如需自定义主题，优先通过 NutUI 变量或组件属性而非覆盖内部元素。

## 典型替换示例
- 保存按钮：
  - 原：`<button class="submit-btn nut-button" :disabled="!canSubmit" @tap="submit">保存</button>`
  - 新：`<nut-button class="submit-btn" type="primary" block :disabled="!canSubmit" @click="submit">保存</nut-button>`
- 数字输入：
  - 原：`<input class="mp-input" type="digit" :value="weightKgStr" placeholder="可选" @input="e=>weightKgStr=e.detail.value" />`
  - 新：`<nut-input class="mp-input" type="number" v-model="weightKgStr" placeholder="可选" clearable />`

## 实施步骤
1. 在上述文件中逐一替换原生标签为 NutUI 组件（保持现有 `class` 与布局容器不变）。
2. 将所有 `@tap` 事件改为 `@click`，并校正输入框的绑定方式为 `v-model`。
3. 清理对原生按钮的样式适配（如针对 `button` 的选择器），保留通用外层样式；必要时改为针对 `.submit-btn` 等外层类名的样式。
4. 按页面自测并修正：禁用态、加载态、数字输入合法性、弹窗内按钮交互、登录授权按钮点击行为等。

## 验证清单
- cats 页面：填写必填项后保存按钮可点击；数字输入仅接受数字；清空按钮生效。
- login 页面：NutUI 按钮触发授权逻辑；`disabled/loading` 状态正常。
- 统计与资料页面：新增/编辑流程不受影响；弹窗确认/取消按钮交互正常。
- ErrorBoundary：点击重试能恢复流程。

## 风险与兼容
- 小程序特有 `type="digit"` 与 web 的差异已统一为 `type="number"`，如需更严格数字控制可在 `@change` 中进行过滤。
- 若存在依赖原生 `<button>` 的选择器（如 `button + button`），需同步调整为类选择器，避免样式失效。

请确认以上计划，确认后我将开始逐文件替换并完成自测。