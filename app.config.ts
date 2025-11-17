//app.config.ts
const PRIMARY_COLOR = '#E28A43'
export default defineAppConfig({
  pages: ['pages/index'],
  window: {
    backgroundTextStyle: 'light',
    navigationBarBackgroundColor: PRIMARY_COLOR,
    navigationBarTitleText: 'ERP',
    navigationBarTextStyle: 'white',
    pullRefresh: true // 可下拉刷新
  }
})