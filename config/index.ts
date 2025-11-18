import Components from 'unplugin-vue-components/webpack';
import NutUIResolver from '@nutui/auto-import-resolver';
const path = require('path');
const customWebpackConfig = require('../webpack.config.js');
const webpack = require('webpack');

const outRoot = `dist/${process.env.TARO_ENV || 'weapp'}`
const config = {
  projectName: 'myApp',
  date: '2024-3-6',
  designWidth (input) {
    if (input?.file?.replace(/\\+/g, '/').indexOf('@nutui') > -1) {
      return 375
    }
    return 750
  },
  deviceRatio: {
    640: 2.34 / 2,
    750: 1,
    828: 1.81 / 2,
    375: 2 / 1
  },
  sourceRoot: 'src',
  outputRoot: outRoot,
  alias: {
    '@': path.resolve(__dirname, '..', 'src'),
  },
  // https://docs.taro.zone/docs/config-detail/#cache
  cache:{
    enable: true,
  },
  plugins: ['@tarojs/plugin-html'],
  defineConstants: {
  },
  copy: {
    patterns: [
    ],
    options: {
    }
  },
  framework: 'vue3',
  compiler: {
    type: 'webpack5',
    prebundle: { enable: false }
  },
  sass:{
    resource: [
      path.resolve(__dirname, '..', 'src/styles/custom_theme.scss')
    ],
    data: `@import "@nutui/nutui-taro/dist/styles/variables.scss";`
  },
  mini: {
    webpackChain(chain) {
      chain.plugin('unplugin-vue-components').use(Components({
        resolvers: [
          NutUIResolver({
            importStyle: 'sass',
            taro: true
          })
        ]
      }))

      chain.plugin('vue-feature-flags').use(webpack.DefinePlugin, [
        {
          __VUE_OPTIONS_API__: true,
          __VUE_PROD_DEVTOOLS__: false,
          __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: false
        }
      ])

    },
    postcss: {
      pxtransform: {
        enable: true,
        config: {
          // selectorBlackList: ['nut-']
        }
      },
      url: {
        enable: true,
        config: {
          limit: 1024 // 设定转换尺寸上限
        }
      },
      cssModules: {
        enable: false, // 默认为 false，如需使用 css modules 功能，则设为 true
        config: {
          namingPattern: 'module', // 转换模式，取值为 global/module
          generateScopedName: '[name]__[local]___[hash:base64:5]'
        }
      }
    }
  },
  h5: {
    router:{
      mode:'browser'
    },
    webpackChain(chain) {
      chain.plugin('unplugin-vue-components').use(Components({
        resolvers: [
          NutUIResolver({
            importStyle: 'sass',
            taro: true
          })
        ]
      }))
      
      // 配置sass-loader以静默NutUI的警告
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
      
      // 应用自定义webpack配置
      chain.merge(customWebpackConfig({}))

      chain.plugin('vue-feature-flags').use(webpack.DefinePlugin, [
        {
          __VUE_OPTIONS_API__: true,
          __VUE_PROD_DEVTOOLS__: false,
          __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: false
        }
      ])
    },
    publicPath: '/',
    staticDirectory: 'static',
    esnextModules: ['nutui-taro', 'icons-vue-taro'],
    postcss: {
      autoprefixer: {
        enable: true,
        config: {
        }
      },
      cssModules: {
        enable: false, // 默认为 false，如需使用 css modules 功能，则设为 true
        config: {
          namingPattern: 'module', // 转换模式，取值为 global/module
          generateScopedName: '[name]__[local]___[hash:base64:5]'
        }
      }
    }
  }
}

module.exports = function (merge) {
  if (process.env.NODE_ENV === 'development') {
    return merge({}, config, require('./dev'))
  }
  return merge({}, config, require('./prod'))
}
