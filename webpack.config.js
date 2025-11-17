// 自定义webpack配置来消除警告
const path = require('path');

module.exports = function(config) {
  return {
    ...config,
    ignoreWarnings: [
      {
        message: /You don't need `webpackExports` if the usage of dynamic import is statically analyse-able/
      }
    ]
  };
};
