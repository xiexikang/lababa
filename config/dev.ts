module.exports = {
  env: {
    NODE_ENV: '"development"'
  },
  defineConstants: {
  },
  // 开发服务器配置
  devServer: {
    host: '0.0.0.0',
    port: 3000,
    open: false,
    allowedHosts: 'all',
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, PATCH, OPTIONS',
      'Access-Control-Allow-Headers': 'X-Requested-With, content-type, Authorization'
    }
  },
  mini: {
    webpackChain(chain) {
      // 静默第三方库的警告
      chain.module
        .rule('scss')
        .use('1')
        .tap(options => {
          return {
            ...options,
            sassOptions: {
              ...(options?.sassOptions || {}),
              silenceDeprecations: ['import', 'legacy-js-api'],
              quietDeps: true
            }
          }
        })
    }
  },
  h5: {
  devServer: {
    host: '0.0.0.0',
    port: 3000,
    open: true,
    allowedHosts: 'all',
    compress: false,
    client: {
      overlay: false
    },
    historyApiFallback: {
      rewrites: [
        { from: /^\/pages\/index\/index$/, to: '/' }
      ]
    }
  },
    webpackChain(chain) {
      // 静默第三方库的警告
      chain.module
        .rule('scss')
        .use('1')
        .tap(options => {
          return {
            ...options,
            sassOptions: {
              ...(options?.sassOptions || {}),
              silenceDeprecations: ['import', 'legacy-js-api'],
              quietDeps: true
            }
          }
        })

      // 保持默认的 stats 配置，避免与 webpack 校验冲突
    }
  }
}
