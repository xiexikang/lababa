export default defineAppConfig({
  pages: [
    'pages/index/index',
    'pages/statistics/index',
    'pages/statistics/detail/index',
    'pages/profile/index',
    'pages/login/index',
    'pages/ranking/index',
    'pages/test/index'
  ],
  entryPagePath: 'pages/index/index',
  window: {
    backgroundTextStyle: 'light',
    navigationBarBackgroundColor: '#fff',
    navigationBarTitleText: '粑粑星人',
    navigationBarTextStyle: 'black'
  },
  tabBar: {
    color: '#666',
    selectedColor: '#8BCE92',
    backgroundColor: '#fff',
    borderStyle: 'black',
    list: [
      {
        pagePath: 'pages/index/index',
        text: '打卡'
      },
      {
        pagePath: 'pages/statistics/index',
        text: '统计'
      },
      {
        pagePath: 'pages/ranking/index',
        text: '粑王'
      },
      {
        pagePath: 'pages/profile/index',
        text: '我的'
      }
    ]
  }
})
