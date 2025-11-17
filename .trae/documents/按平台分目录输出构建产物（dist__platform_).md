**目标**

* 构建时根据平台分别输出到 `dist/<platform>`，例如：
  * 微信小程序：`dist/weapp`
  * H5：`dist/h5`
  * 头条小程序：`dist/tt`

**实现方案**

* 修改 `config/index.ts` 的 `outputRoot` 为按平台动态生成：
  * 在文件顶部计算：`const outRoot = `dist/${process.env.TARO_ENV || 'weapp'}``
  * 将 `outputRoot: 'dist'` 改为 `outputRoot: outRoot`
  * Taro 在 `taro build --type weapp|h5|tt` 时会设置 `process.env.TARO_ENV`，从而输出到对应目录

**兼容与注意**

* 其他配置保持不变（`alias`、`sass`、`mini/h5` 等），无需改脚本
* H5 的 `publicPath` 仍为 `'/'`，不会受目录变化影响
* 开发者工具导入路径变更：
  * 微信开发者工具导入目录改为 `dist/weapp`
  * H5 访问路径不变（开发时用 `npm run dev:h5`；打包时在 `dist/h5`）

**验证**

* 执行：
  * `npm run build:weapp` → 产物在 `dist/weapp`
  * `npm run build:h5` → 产物在 `dist/h5`
  * `npm run build:tt` → 产物在 `dist/tt`
* 打开微信开发者工具导入 `dist/weapp`；H5 打包目录 `dist/h5` 可由任意静态服务器部署

**后续可选**

* 如需更细分（比如 `dist/weapp/dev` 与 `dist/weapp/prod`），可根据 `process.env.NODE_ENV` 追加子目录：`dist/${TARO_ENV}/${NODE_ENV}`。

如果确认，我将更新 `config/index.ts` 的 `outputRoot` 为按平台动态输出，并重新打包验证。