<template>
  <view class="tab-bar">
    <view class="tab-bar-inner">
      <view 
        v-for="tab in tabs" 
        :key="tab.key"
        :class="['tab-item', { active: activeTab === tab.key }]"
        @click="switchTab(tab.key)"
      >
        <view class="tab-icon">{{ tab.icon }}</view>
        <text class="tab-text">{{ tab.label }}</text>
      </view>
    </view>
  </view>
</template>

<script setup name="TabBar">
import { ref, computed } from 'vue';
import Taro from '@tarojs/taro';
  import { showToast } from '@/utils/toast';

const props = defineProps({
  activeTab: { type: String, default: 'index' }
});

const emit = defineEmits(['tab-change']);

const tabs = ref([
  { key: 'index', label: '打卡', icon: '💩', path: '/pages/index/index' },
  { key: 'statistics', label: '统计', icon: '📊', path: '/pages/statistics/index' },
  { key: 'profile', label: '我的', icon: '👤', path: '/pages/profile/index' }
]);

const handleTabClick = (tabKey) => {
  if (tabKey === props.activeTab) return;
  
  const tab = tabs.value.find(t => t.key === tabKey);
  if (tab) {
    try {
      // 使用Taro路由跳转
      Taro.switchTab({
        url: tab.path
      });
      emit('tab-change', tabKey);
    } catch (error) {
      console.error('Tab导航失败:', error);
      // H5环境下使用window.location.href作为备用方案
      if (typeof window !== 'undefined') {
        window.location.href = tab.path;
      } else {
        showToast({ title: '导航失败', icon: 'error' });
      }
    }
  }
};
</script>

<style lang="scss" scoped>
.tab-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: #fff;
  box-shadow: 0 -4rpx 12rpx rgba(0, 0, 0, 0.1);
  z-index: 1000;
  padding-bottom: env(safe-area-inset-bottom);
  
  .tab-bar-inner {
    display: flex;
    height: 120rpx;
    align-items: center;
    justify-content: space-around;
    padding: 0 40rpx;
    
    .tab-item {
      flex: 1;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      height: 100%;
      transition: all 0.3s ease;
      cursor: pointer;
      
      &.active {
        .tab-icon {
          transform: scale(1.2);
        }
        
        .tab-text {
          color: #8BCE92;
          font-weight: 600;
        }
      }
      
      &:active {
        transform: scale(0.95);
      }
      
      .tab-icon {
        font-size: 40rpx;
        margin-bottom: 8rpx;
        transition: all 0.3s ease;
      }
      
      .tab-text {
        font-size: 24rpx;
        color: #666;
        transition: all 0.3s ease;
      }
    }
  }
}
</style>