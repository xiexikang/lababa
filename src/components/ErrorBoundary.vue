<!-- 错误边界组件 -->
<template>
  <view class="error-boundary">
    <template v-if="hasError">
      <view class="error-container">
        <text class="error-icon">😅</text>
        <text class="error-title">哎呀，出错了</text>
        <text class="error-message">{{ errorMessage }}</text>
        <button class="retry-btn" @click="retry">重新加载</button>
        <view class="error-details" v-if="showDetails">
          <text class="details-title">详细信息：</text>
          <text class="details-content">{{ errorDetails }}</text>
        </view>
        <text class="show-details-btn" @click="showDetails = !showDetails">
          {{ showDetails ? '隐藏详情' : '显示详情' }}
        </text>
      </view>
    </template>
    <template v-else>
      <slot></slot>
    </template>
  </view>
</template>

<script setup>
import { ref, onErrorCaptured } from 'vue';

const hasError = ref(false);
const errorMessage = ref('');
const errorDetails = ref('');
const showDetails = ref(false);

onErrorCaptured((error, instance, info) => {
  console.error('错误边界捕获到错误:', error);
  console.error('错误信息:', info);
  console.error('错误实例:', instance);
  
  hasError.value = true;
  errorMessage.value = error.message || '未知错误';
  errorDetails.value = `
错误信息: ${error.message || '无'}
错误堆栈: ${error.stack || '无'}
组件信息: ${info || '无'}
时间: ${new Date().toLocaleString()}
  `.trim();
  
  // 返回false表示错误已被处理
  return false;
});

const retry = () => {
  hasError.value = false;
  errorMessage.value = '';
  errorDetails.value = '';
  showDetails.value = false;
  
  // 重新加载页面
  if (typeof window !== 'undefined') {
    window.location.reload();
  }
};
</script>

<style>
.error-boundary {
  width: 100%;
  height: 100%;
}

.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

.error-icon {
  font-size: 64px;
  margin-bottom: 20px;
}

.error-title {
  font-size: 24px;
  font-weight: bold;
  color: #333;
  margin-bottom: 12px;
}

.error-message {
  font-size: 16px;
  color: #666;
  margin-bottom: 24px;
  text-align: center;
  max-width: 300px;
}

.retry-btn {
  background: #4CAF50;
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: 25px;
  font-size: 16px;
  cursor: pointer;
  margin-bottom: 16px;
}

.retry-btn:hover {
  background: #45a049;
}

.show-details-btn {
  color: #999;
  font-size: 14px;
  text-decoration: underline;
  cursor: pointer;
}

.error-details {
  margin-top: 20px;
  padding: 16px;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 8px;
  max-width: 90%;
  max-height: 200px;
  overflow-y: auto;
}

.details-title {
  font-size: 14px;
  font-weight: bold;
  color: #333;
  margin-bottom: 8px;
}

.details-content {
  font-size: 12px;
  color: #666;
  white-space: pre-line;
  line-height: 1.4;
}
</style>
