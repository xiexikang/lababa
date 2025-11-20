## 执行步骤
- 设置生产环境变量并启动构建：`NODE_ENV=production` + `npm run build:weapp`。
- 等待 Taro 完成产物输出到 `dist/weapp`。
- 验证打包结果：检查 `dist/weapp` 下页面与 `app.json` 等文件存在。
- 使用微信开发者工具打开 `dist/weapp` 目录重新预览/上传。