<template>
  <view class="detail-root">
    <view class="date-banner">{{ displayDate }}</view>
    <template v-if="dayRecords.length > 0">
      <view class="kpi-row">
        <view class="kpi-card">
          <text class="kpi-number">{{ dayRecords.length }}</text>
          <text class="kpi-label">本日粑粑次数</text>
        </view>
      </view>
      <view class="color-section">
        <view class="section-title">本日色谱</view>
        <view class="color-grid">
          <view class="color-item" v-for="c in colorKeys" :key="c">
            <text class="color-name">{{ colorLabels[c] }}</text>
            <text class="color-count">{{ dayColorDist[c] || 0 }}次</text>
          </view>
        </view>
      </view>
      <view class="score-section">
        <view class="section-title">本日状态评分</view>
        <view class="score-bar">
          <view class="score-fill" :style="{ width: dayScore + '%' }"></view>
          <text class="score-text">{{ dayScore }}</text>
        </view>
      </view>
      <view class="list-section">
        <view class="section-title">本日详细记录</view>
        <view class="record-item" v-for="r in dayRecords" :key="r.id">
          <view class="record-time">{{ formatHM(r.endTime) }}</view>
          <view class="record-tags">
            <text>{{ getColorLabel(r.color) }}</text>
            <text>{{ getStatusLabel(r.status) }}</text>
            <text>{{ getShapeLabel(r.shape) }}</text>
            <text>把量：{{ getAmountLabel(r.amount) }}</text>
            <text>{{ formatDuration(r.duration) }}</text>
          </view>
        </view>
      </view>
    </template>
    <template v-else>
      <view class="empty-wrap">
        <view class="empty-tip">没有打卡数据哦～</view>
        <nut-button class="fill-btn" color="#ffb60d" @click="showFill = true">补卡</nut-button>
      </view>
      <nut-popup v-model:visible="showFill" position="bottom" round closeable class="fill-popup-container" :overlay-style="{ background: 'rgba(0,0,0,0.4)' }">
        <view class="fill-popup">
          <view class="popup-header">
            <text class="popup-title">💩 补卡详情</text>
            <text class="popup-subtitle">记录你的便便状态，关注健康</text>
          </view>
          <scroll-view scroll-y class="fill-form">
            <view class="form-row">
              <text class="row-label">日期</text>
              <text class="row-value">{{ displayDate }}</text>
            </view>
            <view class="form-row">
              <text class="row-label">粑粑时间</text>
              <view class="row-choices">
                <view v-for="t in timeChoices" :key="t" class="nut-button" :class="{ active: fill.time===t }" @tap="fill.time=t">{{ timeIcons[t] }} {{ t }}</view>
              </view>
            </view>
            <view class="form-row">
              <text class="row-label">时长（分钟）</text>
              <view class="row-choices">
                <view v-for="m in durChoices" :key="m" class="nut-button" :class="{ active: fill.durationMinutes===m }" @tap="fill.durationMinutes=m">{{ durIcons[m] }} {{ m }}</view>
              </view>
            </view>
            <view class="form-row">
              <text class="row-label">颜色</text>
              <view class="row-choices">
                <view v-for="c in colorKeys" :key="c" class="nut-button" :class="{ active: fill.color===c }" @tap="fill.color=c">{{ colorIcons[c] }} {{ colorLabels[c] }}</view>
              </view>
            </view>
            <view class="form-row">
              <text class="row-label">状态</text>
              <view class="row-choices">
                <view v-for="s in statusKeys" :key="s" class="nut-button" :class="{ active: fill.status===s }" @tap="fill.status=s">{{ statusIcons[s] }} {{ statusLabels[s] }}</view>
              </view>
            </view>
            <view class="form-row">
              <text class="row-label">形状</text>
              <view class="row-choices">
                <view v-for="sh in shapeKeys" :key="sh" class="nut-button" :class="{ active: fill.shape===sh }" @tap="fill.shape=sh">{{ shapeIcons[sh] }} {{ shapeLabels[sh] }}</view>
              </view>
            </view>
            <view class="form-row">
              <text class="row-label">把量</text>
              <view class="row-choices">
                <view v-for="a in amountKeys" :key="a" class="nut-button" :class="{ active: fill.amount===a }" @tap="fill.amount=a">{{ amountIcons[a] }} {{ amountLabels[a] }}</view>
              </view>
            </view>
          </scroll-view>
          <view class="popup-footer">
            <nut-button color="#ccc" class="cancel-btn" @click="showFill = false">我再看看</nut-button>
            <nut-button color="#ffb60d" class="confirm-btn" :disabled="!isFillValid" @click="confirmFill">确定保存</nut-button>
          </view>
        </view>
      </nut-popup>
    </template>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue';
import Taro from '@tarojs/taro';

import { useSimpleStore } from '@/store/simple';

const store = useSimpleStore();
const router = Taro.getCurrentInstance()?.router;
const dateParam = router?.params?.date || '';
const catIdParam = router?.params?.id || '';
const pad2 = (n) => String(n).padStart(2, '0');
const displayDate = computed(() => {
  if (!dateParam) return '';
  const [y,m,d] = dateParam.split('-');
  return `${y}.${m}.${d}`;
});

const toDateStr = (ts) => {
  const d = new Date(ts);
  return `${d.getFullYear()}-${pad2(d.getMonth()+1)}-${pad2(d.getDate())}`;
};
const formatHM = (ts) => {
  const d = new Date(ts);
  return `${pad2(d.getHours())}:${pad2(d.getMinutes())}`;
};
const formatDuration = (sec) => {
  const m = Math.floor(sec/60), s = sec%60;
  return m>0?`${m}分钟${s}秒`:`${s}秒`;
};

const dayRecords = computed(() => {
  return store.globalState.records.filter(r => toDateStr(r.endTime) === dateParam && (!catIdParam || String(r.catId) === String(catIdParam)));
});

const colorLabels = { 'yellow-brown':'黄褐色','brown':'棕色','black':'黑色','green':'绿色','red':'红色','gray-white':'灰白色' };
const statusLabels = { 'normal':'正常','constipation':'便秘','diarrhea':'拉肚子' };
const shapeLabels = { 'banana':'香蕉状','granular':'颗粒状','soft':'软糊糊','cracked':'裂块条纹','watery':'水样便便' };
const amountLabels = { 'very-little':'非常少','little':'少量','moderate':'适中','lot':'大量' };
const colorKeys = Object.keys(colorLabels);
const statusKeys = Object.keys(statusLabels);
const shapeKeys = Object.keys(shapeLabels);
const amountKeys = Object.keys(amountLabels);

const dayColorDist = computed(() => {
  const dist = {}; colorKeys.forEach(k=>dist[k]=0);
  dayRecords.value.forEach(r => { if (r.color && dist[r.color]!==undefined) dist[r.color]++; });
  return dist;
});
const dayScore = computed(() => {
  let s = 100; dayRecords.value.forEach(r=>{ if(r.status==='constipation') s-=10; if(r.status==='diarrhea') s-=10; if(r.status==='normal') s+=2; });
  s = Math.max(0, Math.min(100, s)); return s;
});

const showFill = ref(false);
const fill = ref({ time:'12:00', durationMinutes:5, color:'yellow-brown', status:'normal', shape:'banana', amount:'moderate' });
// 图标数据
const timeIcons = { '06:00':'🌅', '08:00':'☕', '12:00':'🍽️', '16:00':'🍵', '20:00':'🌙' };
const durIcons = { 1:'⚡', 5:'⏱️', 10:'⏰', 15:'🕐', 20:'🕑', 30:'🕒', 60:'🕓' };
const colorIcons = { 'yellow-brown':'💩', 'brown':'💩', 'black':'🌑', 'green':'🍃', 'red':'🔴', 'gray-white':'⚪' };
const statusIcons = { 'normal':'😊', 'constipation':'😣', 'diarrhea':'😰' };
const shapeIcons = { 'banana':'🍌', 'granular':'⚪', 'soft':'🍮', 'cracked':'🍪', 'watery':'💧' };
const amountIcons = { 'very-little':'💧', 'little':'🥤', 'moderate':'🍺', 'lot':'🛁' };

const timeChoices = ['06:00','08:00','12:00','16:00','20:00'];
const durChoices = [1,5,10,15,20,30,60];
const isFillValid = computed(() => !!(fill.value.color && fill.value.status && fill.value.shape && fill.value.amount && fill.value.durationMinutes && fill.value.time));
const confirmFill = () => {
  if (!isFillValid.value) return;
  store.addRecordForDate({ date: dateParam, catId: catIdParam || undefined, ...fill.value });
  showFill.value = false;
  Taro.showToast({ title: '补卡成功', icon: 'success', duration: 1500 });
};

const getColorLabel = (v)=>colorLabels[v]||v;
const getStatusLabel = (v)=>statusLabels[v]||v;
const getShapeLabel = (v)=>shapeLabels[v]||v;
const getAmountLabel = (v)=>amountLabels[v]||v;
</script>

<style lang="scss">
// 按钮动画效果
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(20rpx); }
  to { opacity: 1; transform: translateY(0); }
}
.detail-root { min-height: 100vh; background: linear-gradient(135deg,#8BCE92 0%, #6ecb6d 100%); padding-bottom: 140rpx; }
.date-banner { margin: 20rpx auto; background: #e6f5ea; border-radius: 20rpx; padding: 16rpx 24rpx; width: 60%; text-align: center; font-size: 28rpx; }
.kpi-row { display:flex; gap:20rpx; padding: 0 40rpx; }
.kpi-card { flex:1; background:#fff; border-radius:20rpx; padding:30rpx 20rpx; text-align:center; box-shadow:0 4rpx 10rpx rgba(0,0,0,0.08); }
.kpi-number { display:block; font-size:36rpx; font-weight:700; color:#4a7c59; }
.kpi-label { font-size:24rpx; color:#666; }
.section-title { padding: 20rpx 0 10rpx; font-size: 32rpx; color:#333; font-weight: bold; }
.color-grid { display:grid; grid-template-columns: repeat(2,1fr); gap:16rpx; padding:0 40rpx; }
.color-item { background:#fff; border-radius:16rpx; padding:20rpx; }
.color-name { font-size:24rpx; color:#2d5a3d; }
.color-count { float:right; font-size:24rpx; color:#666; }
.score-section { padding: 0 40rpx; }
.score-bar { position:relative; height: 28rpx; border-radius:20rpx; background:#e6f5ea; }
.score-fill { position:absolute; left:0; top:0; bottom:0; background: linear-gradient(90deg,#8BCE92,#6ecb6d); }
.score-text { position:absolute; right:12rpx; top:-36rpx; font-size:24rpx; color:#2d5a3d; }
.list-section { padding: 0 40rpx; }
.record-item { background:#fff; border-radius:16rpx; padding:20rpx; margin-bottom:16rpx; }
.record-time { font-size:24rpx; color:#333; margin-bottom:12rpx; }
.record-tags text { margin-right: 12rpx; font-size:24rpx; color:#666; }
.empty-wrap { text-align:center; padding: 120rpx 0; animation: fadeIn 0.8s ease-out; }
.empty-tip { font-size:32rpx; color:#fff; margin-bottom: 40rpx; font-weight: bold; }
.fill-btn { width: 60%; margin: 0 auto; height: 100rpx; border-radius: 50rpx; box-shadow: 0 8rpx 20rpx rgba(0,0,0,0.15); transition: all 0.3s ease; animation: fadeIn 0.6s ease-out; 
  &:active { transform: scale(0.95); } }
.fill-popup { height: 100%; display: flex; flex-direction: column; background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%); }
.fill-form { flex: 1; padding: 20rpx 40rpx; }
.form-row { margin-bottom: 40rpx; animation: fadeIn 0.4s ease-out; 
  &:nth-child(2) { animation-delay: 0.1s; }
  &:nth-child(3) { animation-delay: 0.2s; }
  &:nth-child(4) { animation-delay: 0.3s; }
  &:nth-child(5) { animation-delay: 0.4s; }
  &:nth-child(6) { animation-delay: 0.5s; }
  &:nth-child(7) { animation-delay: 0.6s; }
  &:last-child { margin-bottom: 40rpx; } }
.row-label { display:block; font-size:28rpx; color:#333; margin-bottom: 24rpx; font-weight: 600; }
.row-value { font-size:28rpx; color:#666; background:#f8f9fa; padding:16rpx 24rpx; border-radius:12rpx; display:inline-block; }
.row-choices { display:grid; grid-template-columns: repeat(3, 1fr); gap: 20rpx; animation: fadeIn 0.4s ease-out; 
  .nut-button { 
    display: flex; flex-direction: column; align-items: center; padding: 24rpx 16rpx; 
    border-radius: 16rpx; background: #f8f9fa; border: 2rpx solid transparent; 
    transition: all 0.3s ease; font-size: 24rpx; color: #666; text-align: center; 
    &.active { 
      background: #fff8e1; border-color: #ffb60d; transform: scale(1.05); 
    } 
  } 
}

.fill-popup-container {
  .nut-popup__content {
    height: 80vh;
    border-top-left-radius: 24rpx;
    border-top-right-radius: 24rpx;
    overflow: hidden;
    background: #fff;
  }
}
// 弹窗头部样式
.popup-header {
  padding: 40rpx 40rpx 20rpx;
  text-align: center;
  background: #fff;
  border-radius: 0 0 20rpx 20rpx;
  box-shadow: 0 4rpx 12rpx rgba(0,0,0,0.1);
  
  .popup-title {
    display: block;
    font-size: 36rpx;
    font-weight: bold;
    color: #4a7c59;
    margin-bottom: 8rpx;
  }
  
  .popup-subtitle {
    font-size: 24rpx;
    color: #666;
  }
}

// 底部按钮区域
.popup-footer {
  padding: 20rpx 40rpx 40rpx;
  background: #fff;
  display: flex;
  gap: 20rpx;
  
  .cancel-btn, .confirm-btn {
    flex: 1;
    height: 80rpx;
    border-radius: 40rpx;
    font-size: 28rpx;
    font-weight: 500;
    
    &:disabled {
      opacity: 0.6;
    }
  }
  
  .confirm-btn {
    color: #fff;
  }
}

// 时间选项特殊处理
.form-row:nth-child(2) .row-choices {
  grid-template-columns: repeat(2, 1fr);
}

// 时长选项特殊处理  
.form-row:nth-child(3) .row-choices {
  grid-template-columns: repeat(4, 1fr);
}

// 把量选项特殊处理
.form-row:nth-child(7) .row-choices {
  grid-template-columns: repeat(2, 1fr);
}
</style>
