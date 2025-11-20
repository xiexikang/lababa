module.exports = {
  env: {
    NODE_ENV: '"development"',
    API_BASE_URL: '"http://10.30.1.53:8081"'
  },
  defineConstants: {
  },
  // 开发服务器配置
  devServer: {
    host: '0.0.0.0',
    port: 9999,
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
      chain.module
        .rule('scss')
        .use('sass-loader')
        .tap(options => ({
          ...options,
          sassOptions: {
            ...(options?.sassOptions || {}),
            silenceDeprecations: ['import', 'legacy-js-api'],
            quietDeps: true
          }
        }))
    }
  },
  h5: {
  devServer: {
    host: '0.0.0.0',
    port: 9999,
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
      chain.module
        .rule('scss')
        .use('sass-loader')
        .tap(options => ({
          ...options,
          sassOptions: {
            ...(options?.sassOptions || {}),
            silenceDeprecations: ['import', 'legacy-js-api'],
            quietDeps: true
          }
        }))
    }
  }
}
