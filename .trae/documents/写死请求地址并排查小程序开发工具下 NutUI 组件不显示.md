## 目标
- 将小程序请求基地址临时写死为 `http://10.30.1.53:8081`。
- 排查并恢复 NutUI 组件在微信小程序开发工具中的显示问题。

## 改动
- 修改 `src/utils/request.ts` 的默认地址，从 `http://localhost:8081` 改为 `http://10.30.1.53:8081`，保留环境变量覆盖能力（若未来再用 `API_BASE_URL`）。

## 排查与修复 NutUI
- 已配置：
  - 自动导入与样式：`unplugin-vue-components` + `NutUIResolver({ importStyle: 'sass', taro: true })`；全局样式变量已注入。
  - Mini 端 `pxtransform` 黑名单：`selectorBlackList: ['nut-']`，避免样式被错误转换。
  - 测试页显式导入 `NutButton` 组件，绕过自动导入链路并验证组件本身。
- 建议操作：
  - 完整重启 Taro 构建与微信开发者工具，并清理项目缓存后重新编译（避免旧产物导致组件不显示）。
  - 确认依赖版本：`@nutui/nutui-taro@^4.3.13` 与 `@tarojs/*@4.1.1` 已兼容。
  - 若仍不显示，可作为下一步尝试开启 `compiler.prebundle.enable = true` 稳定 node_modules 打包（我可为你加上）。

## 验证
- 手机访问 `http://10.30.1.53:8081/api/health/ping` 应返回 `{"code":0,...}`。
- 在 `pages/test/index` 里按钮应正常显示与点击；若仍不显示，继续按上述步骤重编与升级基础库后反馈。