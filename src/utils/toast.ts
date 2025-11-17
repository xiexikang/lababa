// H5兼容的toast工具函数
export const showToast = (options: { title: string; icon?: string; duration?: number; }): void => {
  const { title, icon = 'none', duration = 2000 } = options;
  
  // 检查是否在H5环境
  const isH5 = process.env.TARO_ENV === 'h5' || typeof window !== 'undefined';
  
  if (isH5) {
    // H5环境下使用自定义toast或原生alert
    showH5Toast(title, icon, duration);
  } else {
    // 小程序环境下使用Taro.showToast
    try {
      const Taro = require('@tarojs/taro');
      Taro.showToast({
        title,
        icon: icon as any,
        duration
      });
    } catch (error) {
      console.warn('Taro.showToast调用失败:', error);
      showH5Toast(title, icon, duration);
    }
  }
};

// H5环境下的toast实现
const showH5Toast = (message: string, icon: string, duration: number): void => {
  // 创建toast元素
  const toast = document.createElement('div');
  toast.className = 'h5-toast';
  toast.style.cssText = `
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    background: rgba(0, 0, 0, 0.8);
    color: white;
    padding: 12px 24px;
    border-radius: 8px;
    font-size: 14px;
    z-index: 9999;
    display: flex;
    align-items: center;
    gap: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    max-width: 80%;
    text-align: center;
    line-height: 1.4;
  `;
  
  // 添加图标
  let iconHtml = '';
  if (icon === 'success') {
    iconHtml = '<span style="color: #4CAF50;">✓</span>';
  } else if (icon === 'error') {
    iconHtml = '<span style="color: #f44336;">✗</span>';
  } else if (icon === 'loading') {
    iconHtml = '<span style="animation: spin 1s linear infinite;">⟳</span>';
  }
  
  toast.innerHTML = iconHtml + message;
  
  // 添加动画样式
  const style = document.createElement('style');
  style.textContent = `
    @keyframes spin {
      from { transform: rotate(0deg); }
      to { transform: rotate(360deg); }
    }
    .h5-toast {
      animation: fadeIn 0.3s ease;
    }
    @keyframes fadeIn {
      from { opacity: 0; transform: translate(-50%, -50%) scale(0.8); }
      to { opacity: 1; transform: translate(-50%, -50%) scale(1); }
    }
  `;
  
  document.head.appendChild(style);
  document.body.appendChild(toast);
  
  // 自动移除
  setTimeout(() => {
    if (toast.parentNode) {
      toast.style.animation = 'fadeIn 0.3s ease reverse';
      setTimeout(() => {
        if (toast.parentNode) {
          toast.parentNode.removeChild(toast);
        }
        if (style.parentNode) {
          style.parentNode.removeChild(style);
        }
      }, 300);
    }
  }, duration);
};