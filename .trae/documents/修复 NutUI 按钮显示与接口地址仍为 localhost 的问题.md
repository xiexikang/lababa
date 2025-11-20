## 问题原因研判
- 接口地址：`src/utils/request.ts` 需要编译期替换 `process.env.API_BASE_URL`；若开发工具未重启或缓存未清，仍会回退到 `http://localhost:8081`。
- NutUI 按钮不显示：Mini 端未声明 `esnextModules`，以及 `pxtransform` 未排除 `nut-` 前缀，会导致 NutUI 组件样式/编译异常。

## 计划变更
- 在 `config/dev.ts` 已设置 `API_BASE_URL` 后，确保编译期替换生效：清理缓存并重新编译运行。
- 更新 `config/index.ts` 的 `mini` 配置：
  - 添加 `esnextModules: ['@nutui/nutui-taro', '@nutui/icons-vue-taro']`
  - 开启 `postcss.pxtransform.config.selectorBlackList = ['nut-']`
- 保持现有 `unplugin-vue-components` + `NutUIResolver({ importStyle: 'sass', taro: true })`，无需改动。

## 验证步骤
- 重新编译后在手机上访问：`http://10.30.1.53:8081/api/health/ping`，确保后端可达。
- 在测试页 `src/pages/test/index.vue` 点击 `<nut-button>`，确认显示与点击正常。
- 如果接口仍显示 `localhost`：
  - 完全重启 Taro 构建与微信开发者工具；
  - 检查 `config/dev.ts` 的 `env.API_BASE_URL` 是否为 `"http://10.30.1.53:8081"`；
  - 我可以增加一次性日志输出 `BASE_URL` 以定位（仅开发）。

## 预期结果
- Mini 端 NutUI 按钮正常渲染，样式正确；
- 小程序接口基地址使用 `http://10.30.1.53:8081`，不再回退到 `localhost`。