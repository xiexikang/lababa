// 自定义webpack配置来消除警告
const path = require('path');

module.exports = function(config) {
  return {
    ...config,
    performance: {
      hints: false
    },
    ignoreWarnings: [
      {
        message: /You don't need `webpackExports` if the usage of dynamic import is statically analyse-able/
      },
      /asset size limit/i,
      /entrypoint size limit/i
    ]
  };
};
