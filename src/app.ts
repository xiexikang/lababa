import { createApp } from 'vue'
import { createPinia } from 'pinia';
// import './app.scss'
import '@/styles/reset.scss';
import { Button, Popup, Calendar, CalendarCard } from '@nutui/nutui-taro'
import './app.scss'

// 创建Pinia实例
const pinia = createPinia();

const App = createApp({
  onShow (options) {
    console.log('App onShow');
  },
  onLaunch (options) {
    console.log('App onLaunch');
  },
  // 入口组件不需要实现 render 方法，即使实现了也会被 taro 所覆盖
})

// 安装Pinia
App.use(pinia);
App.use(Button)
App.use(Popup)
App.use(Calendar)
App.use(CalendarCard)


export default App
