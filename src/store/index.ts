import { createPinia } from 'pinia';
import type { App } from 'vue';

// 创建Pinia实例
export const pinia = createPinia();

// 安装Pinia插件
export function installPinia(app: App) {
  app.use(pinia);
}