// 测试页面，验证基本功能
<template>
  <view class="test-root">
    <view class="test-header">
      <text class="test-title">🧪 功能测试</text>
    </view>
    
    <view class="test-content">
      <view class="test-section">
        <text class="section-title">状态管理测试</text>
        <view class="test-item">
          <text>记录数量: {{ totalRecords }}</text>
        </view>
        <view class="test-item">
          <text>是否记录中: {{ isRecording }}</text>
        </view>
        <view class="test-item">
          <text>计时: {{ elapsedTime }}秒</text>
        </view>
      </view>
      
      <view class="test-section">
        <text class="section-title">功能测试</text>
        <NutButton @click="testStartRecording" size="small">开始计时</NutButton>
        <NutButton @click="testStopRecording" size="small">停止计时</NutButton>
        <NutButton @click="testAddRecord" size="small">添加测试记录</NutButton>
      </view>
      
      <view class="test-section">
        <text class="section-title">本地存储测试</text>
        <NutButton @click="testStorage" size="small">测试存储</NutButton>
        <NutButton @click="clearRecords" size="small">清空记录</NutButton>
      </view>
    </view>
  </view>
</template>

<script setup>
import { Button as NutButton } from '@nutui/nutui-taro'
import { ref, computed, onMounted } from 'vue';
import { usePoopStore } from '@/store/poop';
import { storageManager } from '@/utils/storage';

const poopStore = usePoopStore();

// 计算属性
const totalRecords = computed(() => poopStore.totalRecords);
const isRecording = computed(() => poopStore.isRecording);
const elapsedTime = computed(() => poopStore.elapsedTime);

// 测试方法
const testStartRecording = () => {
  console.log('测试开始记录');
  poopStore.startRecording();
};

const testStopRecording = () => {
  console.log('测试停止记录');
  poopStore.stopRecording();
};

const testAddRecord = () => {
  console.log('添加测试记录');
  const testRecord = {
    color: 'brown',
    status: 'normal',
    shape: 'banana',
    amount: 'moderate',
    note: '测试记录'
  };
  poopStore.saveRecord(testRecord);
};

const testStorage = () => {
  console.log('测试本地存储');
  const records = storageManager.getRecords();
  console.log('存储的记录:', records);
  
  const storageInfo = storageManager.getStorageInfo();
  console.log('存储信息:', storageInfo);
};

const clearRecords = () => {
  console.log('清空所有记录');
  poopStore.clearRecords();
};

onMounted(() => {
  console.log('测试页面加载完成');
  poopStore.init();
});
</script>

<style lang="scss">
.test-root {
  padding: 40rpx;
  min-height: 100vh;
  background: #f5f5f5;
  
  .test-header {
    text-align: center;
    margin-bottom: 40rpx;
    
    .test-title {
      font-size: 36rpx;
      font-weight: bold;
      color: #333;
    }
  }
  
  .test-content {
    .test-section {
      background: #fff;
      border-radius: 16rpx;
      padding: 30rpx;
      margin-bottom: 20rpx;
      
      .section-title {
        font-size: 28rpx;
        font-weight: 600;
        color: #333;
        margin-bottom: 20rpx;
        display: block;
      }
      
      .test-item {
        margin-bottom: 16rpx;
        
        text {
          font-size: 24rpx;
          color: #666;
        }
      }
      
      .nut-button {
        margin-right: 16rpx;
        margin-bottom: 16rpx;
      }
    }
  }
}
</style>
