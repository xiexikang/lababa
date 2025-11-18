<template>
  <view class="profile-root">

    <!-- 用户信息 -->
    <view class="user-section">
      <view class="user-info" @tap="handleUserCardTap">
        <view class="avatar-section">
          <image v-if="userInfo.avatarUrl" class="avatar-img" :src="userInfo.avatarUrl" mode="aspectFill" />
          <view v-else class="avatar-placeholder">😊</view>
        </view>
        <view class="user-details">
          <text class="user-name">{{ userInfo.nickName || '未授权用户' }}</text>
          <text class="user-desc">记录健康生活，关注身体状况</text>
        </view>
      </view>

      <!-- 授权登录按钮 -->
      <view class="auth-section" v-if="!userInfo.avatarUrl">
        <template v-if="env === 'WEAPP'">
          <button class="auth-btn" @tap="authorizeWeapp">申请微信授权登录</button>
          <view class="auth-tip">点击用户卡片也可授权</view>
        </template>
        <template v-else>
          <nut-button type="primary" color="#8BCE92" class="auth-btn" @click="authorizeWeapp">申请微信授权登录</nut-button>
          <view class="auth-tip">请在微信小程序中打开以授权头像与昵称</view>
        </template>
      </view>
    </view>

    <!-- 统计卡片 -->
    <view class="stats-card">
      <view class="stat-item">
        <view class="stat-title">粑粑次数</view>
        <view class="stat-value"><text class="num">{{ totalCount }}</text><text class="unit">次</text></view>
      </view>
      <view class="stat-item">
        <view class="stat-title">时长总计</view>
        <view class="stat-value"><text class="num">{{ totalMinutes }}</text><text class="unit">分</text></view>
      </view>
      <view class="stat-item">
        <view class="stat-title">粑友数量</view>
        <view class="stat-value"><text class="num">{{ friendsCount }}</text><text class="unit">位</text></view>
      </view>
    </view>

    <!-- 成就展示 -->
    <view class="achievement-section">
      <view class="section-header">
        <text class="section-title">🏆 我的成就</text>
      </view>
      <view class="achievement-list">
        <view class="achievement-item">
          <view class="achievement-icon">🌟</view>
          <view class="achievement-info">
            <text class="achievement-title">坚持记录</text>
            <text class="achievement-desc">已连续记录7天</text>
          </view>
        </view>
        <view class="achievement-item">
          <view class="achievement-icon">💪</view>
          <view class="achievement-info">
            <text class="achievement-title">健康达人</text>
            <text class="achievement-desc">记录超过30次</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 我的猫咪管理 -->
    <view class="cats-section">
      <view class="section-header">
        <text class="section-title">🐱 我的猫咪</text>
        <template v-if="env==='WEAPP'">
          <button class="nut-button" @tap="goCreateCat">新增</button>
        </template>
        <template v-else>
          <nut-button type="primary" @click="goCreateCat">新增</nut-button>
        </template>
      </view>
      <view v-if="loadingCats" class="cat-item"><text>加载中...</text></view>
      <view v-else-if="!cats.length" class="cat-item"><text>暂无猫咪，点击新增创建</text></view>
      <view v-else>
        <view v-for="c in cats" :key="c.id" class="cat-item">
          <text class="cat-name">{{ c.name || '未命名' }}</text>
          <view class="cat-actions">
            <template v-if="env==='WEAPP'">
              <button class="nut-button" @tap="() => goEditCat(String(c.id))">编辑</button>
              <button class="nut-button" @tap="() => removeCat(String(c.id))">删除</button>
            </template>
            <template v-else>
              <nut-button type="primary" @click="() => goEditCat(String(c.id))">编辑</nut-button>
              <nut-button type="danger" @click="() => removeCat(String(c.id))">删除</nut-button>
            </template>
          </view>
        </view>
      </view>
    </view>

    <!-- 底部导航栏 -->
  </view>
</template>

<script setup lang="ts" name="Profile">
import { reactive, computed, ref, onMounted } from 'vue';
import Taro from '@tarojs/taro';
import { useSimpleStore } from '@/store/simple';
import { showToast } from '@/utils/toast';
import { post, postRaw, get, del } from '@/utils/request'

// 使用简单的状态管理
const store = useSimpleStore();

// 运行环境
const env = Taro.getEnv();

// 用户数据
const userInfo = reactive({
  nickName: '',
  avatarUrl: '',
  recordCount: 0,
  continuousDays: 0
});

// 统计数据
const totalCount = computed(() => store.globalState.records.length);
const totalMinutes = computed(() => {
  const totalSeconds = store.globalState.records.reduce((sum, r) => sum + (r?.duration || 0), 0);
  return Math.floor(totalSeconds / 60);
});
const friendsCount = ref(0);



// 加载与保存用户信息
const loadUserInfo = () => {
  try {
    const cached = Taro.getStorageSync('user-info');
    if (cached) {
      const data = typeof cached === 'string' ? JSON.parse(cached) : cached;
      userInfo.nickName = data.nickName || '';
      userInfo.avatarUrl = data.avatarUrl || '';
    }
  } catch (e) {
    console.warn('读取用户信息失败', e);
  }
};

const saveUserInfo = () => {
  try {
    Taro.setStorageSync('user-info', { nickName: userInfo.nickName, avatarUrl: userInfo.avatarUrl });
  } catch (e) {
    console.warn('保存用户信息失败', e);
  }
};

// 微信授权
  const applyUserInfo = (info) => {
    if (!info) return;
    const { nickName, avatarUrl } = info;
    userInfo.nickName = nickName || '';
    userInfo.avatarUrl = avatarUrl || '';
    saveUserInfo();
    showToast({ title: '授权成功', icon: 'success' });
    try { Taro.eventCenter.trigger('user-updated', { user: { nickName: userInfo.nickName, avatarUrl: userInfo.avatarUrl } }) } catch {}
  };

const authorizeWeapp = async () => {
  console.log('authorizeWeapp');
  showToast({ title: '正在调起授权...', icon: 'none' })
  try {
    const codeRes = await Taro.login();
    console.log('codeRes', codeRes);
    const code = codeRes.code;
    try {
      const res = await Taro.getUserProfile({ desc: '用于完善个人资料', lang: 'zh_CN' });
      const raw = await postRaw('/api/auth/weapp', { code, nickName: res.userInfo?.nickName, avatarUrl: res.userInfo?.avatarUrl });
      const apiRes: any = raw?.data || {};
      const u = apiRes?.data?.user;
      const h: any = raw?.header || {};
      const t = (h.Authorization || h.authorization || h['X-Token'] || h['x-token'] || '') as string;
      console.log('raw', raw);
      if (u) {
        userInfo.nickName = u.nickName || '';
        userInfo.avatarUrl = u.avatarUrl || '';
        Taro.setStorageSync('user-info', u);
        if (t) {
          const tokenVal = t.startsWith('Bearer') ? t.split(' ')[1] : t;
          Taro.setStorageSync('auth-token', tokenVal);
        }
        showToast({ title: '授权成功', icon: 'success' });
        try { Taro.eventCenter.trigger('user-updated', { user: u }) } catch {}
      } else {
        applyUserInfo(res.userInfo);
      }
    } catch (_) {
      const res = await Taro.getUserInfo({ lang: 'zh_CN' });
      console.log('getUserInfo', res);
      const info = (res && res.userInfo) ? res.userInfo : res;
      const raw = await postRaw('/api/auth/weapp', { code, nickName: info?.nickName, avatarUrl: info?.avatarUrl });
      const apiRes: any = raw?.data || {};
      const u = apiRes?.data?.user;
      const h: any = raw?.header || {};
      const t = (h.Authorization || h.authorization || h['X-Token'] || h['x-token'] || '') as string;
      if (u) {
        userInfo.nickName = u.nickName || '';
        userInfo.avatarUrl = u.avatarUrl || '';
        Taro.setStorageSync('user-info', u);
        if (t) {
          const tokenVal = t.startsWith('Bearer') ? t.split(' ')[1] : t;
          Taro.setStorageSync('auth-token', tokenVal);
        }
        showToast({ title: '授权成功', icon: 'success' });
        try { Taro.eventCenter.trigger('user-updated', { user: u }) } catch {}
      } else {
        applyUserInfo(info);
      }
    }
  } catch (error) {
    console.log('error',error);
    showToast({ title: '用户取消授权', icon: 'none' });
  }
};

onMounted(() => {
  loadUserInfo();
  try {
    const cachedFriends = Taro.getStorageSync('friends-count');
    friendsCount.value = Number(cachedFriends || 0);
  } catch (e) { /* ignore */ }
  loadCats();
});
// 卡片点击触发授权（未授权时）
const handleUserCardTap = () => {
  if (env === 'WEAPP' && !userInfo.avatarUrl) {
    authorizeWeapp();
  }
};

// 我的猫咪管理
const cats = ref<any[]>([])
const loadingCats = ref(false)
const loadCats = async () => {
  try {
    loadingCats.value = true
    const res: any = await get('/api/cats/list')
    cats.value = res?.items || []
  } catch {
    cats.value = []
  } finally {
    loadingCats.value = false
  }
}
const goCreateCat = () => {
  const target = encodeURIComponent('/pages/profile/index')
  try { Taro.navigateTo({ url: `/pages/cats/index?redirect=${target}` }) } catch {}
}
const removeCat = async (id: string) => {
  try {
    await del(`/api/cats/delete/${id}`)
    cats.value = cats.value.filter(c => String(c.id) !== String(id))
    showToast({ title: '已删除', icon: 'success' })
  } catch {
    showToast({ title: '删除失败', icon: 'none' })
  }
}
const goEditCat = (id: string) => {
  const target = encodeURIComponent('/pages/profile/index')
  try { Taro.navigateTo({ url: `/pages/cats/index?id=${encodeURIComponent(id)}&redirect=${target}` }) } catch {}
}
</script>

<style lang="scss">
.profile-root {
  min-height: 100vh;
  background: linear-gradient(135deg, #8BCE92 0%, #6ecb6d 100%);
  padding-bottom: 140rpx;
  
  .header-section {
    padding: 60rpx 40rpx 40rpx;
    text-align: center;
    
    .title-wrapper {
      .main-title {
        font-size: 48rpx;
        font-weight: bold;
        color: #fff;
        text-shadow: 2rpx 2rpx 4rpx rgba(0, 0, 0, 0.2);
      }
    }
  }
  
  .user-section {
    padding: 0 40rpx 40rpx;
    
    .user-info {
      background: rgba(255, 255, 255, 0.9);
      border-radius: 20rpx;
      padding: 40rpx;
      display: flex;
      align-items: center;
      box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);
      transition: transform .08s ease;
      &:active { transform: scale(.98); }
      
      .avatar-section {
        margin-right: 30rpx;
        
        .avatar-img {
          width: 120rpx;
          height: 120rpx;
          border-radius: 50%;
          background: #f7f7f7;
        }
        .avatar-placeholder {
          width: 120rpx;
          height: 120rpx;
          border-radius: 50%;
          background: #8BCE92;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 60rpx;
        }
      }
      
      .user-details {
        flex: 1;
        
        .user-name {
          display: block;
          font-size: 32rpx;
          font-weight: 600;
          color: #333;
          margin-bottom: 8rpx;
        }
        
        .user-desc {
          font-size: 24rpx;
          color: #666;
        }
      }
    }
  }
  .auth-section {
    padding: 20rpx 40rpx;
    display: flex;
    justify-content: center;
    .auth-btn {
      min-width: 260rpx;
      height: 96rpx;
      border-radius: 48rpx;
      font-weight: 700;
      background: #8BCE92;
      color: #fff;
      display: flex;
      align-items: center;
      justify-content: center;
      border: none;
    }
    .auth-tip {
      font-size: 26rpx;
      color: #fff;
      opacity: 0.9;
      background: rgba(255,255,255,0.2);
      padding: 16rpx 24rpx;
      border-radius: 24rpx;
    }
  }
  
  // 统计卡片样式
  .stats-card {
    margin: 20rpx 40rpx;
    padding: 24rpx 28rpx;
    border: 6rpx solid #2d5a3d;
    border-radius: 28rpx;
    background: #ffffe8;
    box-shadow: 0 6rpx 16rpx rgba(0,0,0,0.12);
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    align-items: center;
    text-align: center;
  }
  .stat-item { padding: 8rpx 0; }
  .stat-title { font-size: 26rpx; color: #2d5a3d; opacity: .9; }
  .stat-value { margin-top: 10rpx; display:inline-flex; align-items: baseline; gap: 8rpx; }
  .num { font-size: 36rpx; font-weight: 800; color: #2d5a3d; }
  .unit { font-size: 26rpx; color: #2d5a3d; opacity: .8; }
  
  .achievement-section {
    padding: 0 40rpx;
    
    .section-header {
      margin-bottom: 20rpx;
      
      .section-title {
        font-size: 32rpx;
        font-weight: 600;
        color: #fff;
      }
    }
    
    .achievement-list {
      .achievement-item {
        background: rgba(255, 255, 255, 0.9);
        border-radius: 20rpx;
        padding: 30rpx;
        margin-bottom: 20rpx;
        display: flex;
        align-items: center;
        box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);
        
        .achievement-icon {
          font-size: 48rpx;
          margin-right: 20rpx;
          width: 60rpx;
          text-align: center;
        }
        
        .achievement-info {
          flex: 1;
          
          .achievement-title {
            display: block;
            font-size: 28rpx;
            font-weight: 600;
            color: #333;
            margin-bottom: 4rpx;
          }
          
          .achievement-desc {
            font-size: 24rpx;
            color: #666;
          }
        }
      }
    }
  }
  }

  // 猫咪管理
  .cats-section {
    padding: 0 40rpx 40rpx;
    .section-header { display:flex; justify-content: space-between; align-items:center; margin-bottom: 20rpx; }
    .section-title { font-size: 32rpx; font-weight: 600; color: #fff; }
    .cat-item {
      background: rgba(255, 255, 255, 0.9);
      border-radius: 20rpx;
      padding: 24rpx;
      margin-bottom: 16rpx;
      display: flex;
      align-items: center;
      justify-content: space-between;
      box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);
    }
    .cat-name { font-size: 28rpx; color: #333; font-weight: 600; }
    .cat-actions { display:flex; gap: 12rpx; }
  }
</style>
