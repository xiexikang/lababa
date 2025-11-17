<template>
  <ErrorBoundary>
    <view class="Index-root">
    <nut-button>测试</nut-button>

      <!-- 标题区域 -->
      <view class="header-section">
        <view class="title-wrapper">
          <text class="main-title">粑粑星人</text>
          <view class="flower-decoration">🌼</view>
        </view>
      </view>

    <!-- 主内容区域 -->
    <view class="main-content">
      <template v-if="!isStart">
        <!-- 初始状态 -->
        <view class="start-section">
          <view class="timer-display">
            <text class="timer-text">准备开始</text>
          </view>
          <view class="action-buttons">
            <button 
              v-if="env==='WEAPP'"
              class="start-btn"
              @tap="dataInfo.start()"
            >💩 我要拉了哦</button>
            <nut-button 
              v-else
              color="#8BCE92" 
              class="start-btn"
              @click="dataInfo.start()"
            >💩 我要拉了哦</nut-button>
          </view>
          <view class="last-record-tip">
            <text class="tip-text">🕐 距离上次拉粑粑已经是{{ time }}之前了</text>
          </view>
        </view>
      </template>

      <template v-else>
        <!-- 计时状态 -->
        <view class="recording-section">
          <view class="timer-display large">
            <text class="timer-text large">⏱️ {{ dataInfo.formatTime(beingTime) }}</text>
          </view>
          <view class="recording-status">
            <text class="status-text">正在记录中...</text>
          </view>
          
          <view class="action-buttons recording">
            <view class="btn-group">
              <template v-if="env==='WEAPP'">
                <button class="complete-btn nut-button" @tap="dataInfo.finally()">
                  <text class="btn-text" style="white-space: nowrap;">😌 拉完了</text>
                </button>
                <button class="help-btn nut-button" @tap="dataInfo.showHelp = true">
                  <text class="btn-text" style="white-space: nowrap;">🙏 请祈祷</text>
                </button>
              </template>
              <template v-else>
                <nut-button 
                  color="#6ecb6d" 
                  class="complete-btn"
                  @click="dataInfo.finally()"
                >
                  <text class="btn-text" style="white-space: nowrap;">😌 拉完了</text>
                </nut-button>
                <nut-button 
                  color="#ffb60d" 
                  class="help-btn"
                  @click="dataInfo.showHelp = true"
                >
                  <text class="btn-text" style="white-space: nowrap;">🙏 请祈祷</text>
                </nut-button>
              </template>
            </view>
          </view>
          
          <view class="give-up-section" @click="dataInfo.showGiveUpConfirm = true">
            <text class="give-up-text">😅 尽力了，没拉出来</text>
          </view>
        </view>
      </template>
    </view>

    <!-- 右侧悬浮分享按钮 -->
    <view class="floating-share">
      <button v-if="env === 'WEAPP'" class="share-inner" @tap="handleInviteAndShare">👥 一起拉</button>
      <nut-button v-else class="share-inner" color="#ff69b4" @click="dataInfo.shareTogether()">👥 一起拉</nut-button>
    </view>

    <!-- 底部装饰 -->
    <view class="bottom-decoration">
      <view class="decoration-item">🌱</view>
      <view class="decoration-item">🍃</view>
      <view class="decoration-item">💚</view>
    </view>

    <!-- 详情记录弹窗 -->
    <DetailRecordPopup v-model:visible="dataInfo.show" @on-ok="handleSaveRecord"></DetailRecordPopup>
    
    <!-- 放弃确认弹窗 -->
      <nut-popup 
        position="bottom" 
        v-model:visible="dataInfo.showGiveUpConfirm"
        round
        class="bottom-popup"
        :overlay-style="{ background: 'rgba(0,0,0,0.4)' }"
      >
        <view class="confirm-popup">
          <view class="popup-header">
            <text class="popup-title">确认放弃</text>
          </view>
          <view class="popup-content">
            <text class="popup-text">确定要放弃这次记录吗？</text>
          </view>
          <view class="popup-actions">
            <template v-if="env==='WEAPP'">
              <button 
                class="cancel-btn nut-button"
                @tap="dataInfo.showGiveUpConfirm = false"
              >取消</button>
              <button 
                class="confirm-btn nut-button"
                @tap="dataInfo.giveUp()"
              >确认放弃</button>
            </template>
            <template v-else>
              <nut-button 
                color="#ccc" 
                class="cancel-btn"
                @click="dataInfo.showGiveUpConfirm = false"
              >
                取消
              </nut-button>
              <nut-button 
                color="#ff6b6b" 
                class="confirm-btn"
                @click="dataInfo.giveUp()"
              >
                确认放弃
              </nut-button>
            </template>
          </view>
        </view>
      </nut-popup>
    
      <!-- 底部导航栏 -->

    </view>
  </ErrorBoundary>
</template>

<script setup lang="ts" name="Index">
  import { ref, reactive, onMounted, onUnmounted, computed } from 'vue';
  import { useSimpleStore } from '@/store/simple';
  import { showToast } from '@/utils/toast';
  import DetailRecordPopup from './components/DetailRecordPopup.vue';

  import ErrorBoundary from '@/components/ErrorBoundary.vue';
  import Taro, { useShareAppMessage, useShareTimeline } from '@tarojs/taro';
  import { post, ensureAuth } from '@/utils/request'
  
  // 运行环境
  const env = Taro.getEnv();
  
  // 分享与好友绑定（微信小程序）
  const inviteId = ref('')
  const createInvite = async () => {
    try {
      const id = await post<string>('/api/friends/invite', {})
      inviteId.value = id || ''
    } catch {
      showToast({ title: '邀请功能暂不可用', icon: 'none' })
    }
  }
  const acceptInvite = async (id: string) => {
    try {
      await post('/api/friends/accept', { inviteId: id })
      showToast({ title: '已成为粑友', icon: 'success' })
    } catch (e) {
      showToast({ title: '绑定失败', icon: 'none' })
    }
  }
  if (env === 'WEAPP') {
    useShareAppMessage(() => ({
      title: '粑粑星人：一起拉吧！',
      path: `/pages/index/index?inviteId=${inviteId.value}`
    }))
    useShareTimeline(() => ({
      title: '粑粑星人：一起拉挑战！',
      query: `inviteId=${inviteId.value}`
    }))
  }

  const handleInviteAndShare = async () => {
    if (!ensureAuth()) return
    await createInvite()
    try { Taro.showShareMenu({ withShareTicket: true, menus: ['shareAppMessage', 'shareTimeline'] }) } catch {}
  }
  
  // 控制顶部横幅显示
  const showPromo = ref(false);
  const closePromo = () => {
    showPromo.value = false;
  };
  
  // 显示粑粑庙
  const showTemple = () => {
    showToast({ title: '粑粑庙功能开发中...', icon: 'none' });
  };
  
  let store;
  try {
    store = useSimpleStore();
    console.log('状态管理初始化成功:', store);
  } catch (error) {
    console.error('状态管理初始化失败:', error);
    store = {
      globalState: { isRecording: false, elapsedTime: 0, records: [], lastRecordTime: 0 },
      startRecording: () => console.warn('状态管理初始化失败，使用备用方法'),
      stopRecording: () => console.warn('状态管理初始化失败，使用备用方法'),
      updateElapsedTime: () => console.warn('状态管理初始化失败，使用备用方法'),
      saveRecord: () => console.warn('状态管理初始化失败，使用备用方法'),
      timeSinceLastRecord: '未知'
    };
  }
  
  // 数据信息
  console.log('开始创建dataInfo...');
  const dataInfo = reactive({
    show: false, // 显示详情弹窗
    showGiveUpConfirm: false, // 显示放弃确认弹窗
    showHelp: false, // 显示帮助弹窗
    
    // 开始记录
    start() {
      console.log('开始记录');
      try {
        store.startRecording();
        this.startTimer();
        console.log('记录开始成功');
      } catch (error) {
        console.error('开始记录失败:', error);
        showToast({ title: '开始记录失败', icon: 'error', duration: 2000 });
      }
    },
    
    // 完成记录
    finally() {
      console.log('完成记录，显示详情弹窗');
      this.show = true;
    },

    // 一起拉分享功能
    shareTogether() {
      const currentEnv = Taro.getEnv();
      if (currentEnv === 'WEAPP') {
        Taro.showShareMenu({ withShareTicket: true, menus: ['shareAppMessage', 'shareTimeline'] });
      } else {
        showToast({ title: '请在微信小程序中分享', icon: 'none', duration: 2000 });
      }
    },
    
    // 放弃记录
    giveUp() {
      console.log('放弃记录');
      this.showGiveUpConfirm = false;
      store.stopRecording();
      this.stopTimer();
    },
    
    // 结束记录（未完成的放弃）
    end() {
      console.log('结束记录');
      store.stopRecording();
      this.stopTimer();
    },
    
    // 计时器
    timer: null,
    startTimer() {
      this.timer = setInterval(() => {
        store.updateElapsedTime();
      }, 1000);
    },
    
    stopTimer() {
      if (this.timer) {
        clearInterval(this.timer);
        this.timer = null;
      }
    },

    // 格式化时间
    formatTime(seconds) {
      const hours = Math.floor(seconds / 3600);
      const minutes = Math.floor((seconds % 3600) / 60);
      const secs = seconds % 60;
      let result = '';
      if (hours > 0) {
        result += `${hours}小时`;
      }
      if (minutes > 0) {
        result += `${minutes}分钟`;
      }
      if (secs > 0 || result === '') {
        result += `${secs}秒`;
      }
      return result;
    }
  });
  
  // 计算属性
  const isStart = computed(() => store.globalState.isRecording);
  const beingTime = computed(() => store.globalState.elapsedTime);
  const time = computed(() => store.timeSinceLastRecord || '1小时20分钟');
  
  // 处理保存记录
  const handleSaveRecord = async (recordDetails) => {
    console.log('保存记录详情:', recordDetails);
    
    try {
      if (!ensureAuth()) return
      // 保存记录到store
      await store.saveRecord(recordDetails);
      
      // 停止计时器
      dataInfo.stopTimer();
      
      // 显示成功提示
      showToast({
        title: '记录成功！',
        icon: 'success',
        duration: 2000
      });
    } catch (error) {
      console.error('保存记录失败:', error);
      showToast({
        title: '保存失败，请重试',
        icon: 'error',
        duration: 2000
      });
    }
  };

  // 生命周期
  onMounted(() => {
    console.log('主页面加载完成');
    try {
      // 首页不主动拉取列表，延迟到需要时
      console.log('跳过首屏列表请求');
      if (env === 'WEAPP') {
        // 接受邀请（仅当链接带参数时）
        const params = (Taro.getCurrentInstance() && (Taro.getCurrentInstance() as any).router && (Taro.getCurrentInstance() as any).router.params) || {}
        const qId = params.inviteId || ''
        if (qId) acceptInvite(String(qId))
      }
    } catch (error) {
      console.error('状态管理初始化失败:', error);
      showToast({ title: '初始化失败', icon: 'error', duration: 2000 });
    }
  });
  
  onUnmounted(() => {
    dataInfo.stopTimer();
    if (store.globalState.isRecording) {
      store.stopRecording();
    }
  });
</script>

<style lang="scss">

  .bottom-popup {
    .nut-popup__content {
      border-top-left-radius: 24rpx;
      border-top-right-radius: 24rpx;
      padding: 40rpx 40rpx 32rpx;
      background: #fff;
    }
  }
</style>

<style src="@/pages/index/styles/Index.scss" lang="scss"></style>
