<template>
  <view class="ranking-container">
    <!-- 排行榜标题 -->
  <view class="ranking-header">
    <text class="title">👑 粑王</text>
    <text class="subtitle">看看谁是拉粑达人</text>
  </view>
  <view style="display:flex; align-items:center; justify-content:center; gap: 16rpx; margin-bottom: 12rpx;">
    <text>当前猫咪：</text>
    <text v-if="activeCatId && cats.length" style="font-weight:600;">{{ (cats.find(c=>String(c.id)===String(activeCatId))||{}).name || '未命名' }}</text>
    <button v-if="env==='WEAPP'" class="nut-button" @tap="openCatSelector">切换</button>
    <nut-button v-else @click="openCatSelector">切换</nut-button>
  </view>

    <!-- 周期切换栏 -->
  <view class="period-tabs">
      <view 
        v-for="tab in periodTabs" 
        :key="tab.value" 
        class="period-tab" 
        :class="{ active: activePeriod === tab.value }"
        @tap="setActivePeriod(tab.value)"
      >
        {{ tab.label }}
      </view>
    </view>

    <!-- 排行榜列表 -->
    <view class="ranking-list">
      <view 
        v-for="(user, index) in rankingList" 
        :key="user.id"
        class="ranking-item"
        :class="{ 'top-three': index < 3 }"
      >
        <!-- 排名 -->
        <view class="rank-number" :class="getRankClass(index)">
          <text v-if="index < 3" class="rank-icon">{{ getRankIcon(index) }}</text>
          <text v-else>{{ index + 1 }}</text>
        </view>

        <!-- 用户信息 -->
        <view class="user-info">
          <image 
            class="avatar" 
            :src="user.avatar || defaultAvatar"
            mode="aspectFill"
          />
          <view class="user-details">
            <text class="nickname">{{ user.nickname || '匿名用户' }}</text>
            <text class="count-text">共拉 {{ user.totalCount }} 次</text>
          </view>
        </view>

        <!-- 次数 -->
        <view class="count">
          <text class="count-number">{{ user.totalCount }}</text>
          <text class="count-unit">次</text>
        </view>
      </view>
    </view>

    <!-- 空状态 -->
    <view v-if="rankingList.length === 0" class="empty-state">
      <text class="empty-text">🤔 还没有人上榜哦</text>
      <text class="empty-subtext">快去拉粑粑成为第一个上榜的人吧！</text>
    </view>

    <!-- 我的排名 -->
    <view v-if="myRanking" class="my-ranking">
      <view class="my-ranking-content">
        <text class="my-rank">我的排名: {{ myRanking.rank }}</text>
        <text class="my-count">共拉 {{ myRanking.totalCount }} 次</text>
      </view>
    </view>

    <!-- 加载更多 -->
    <view class="load-more" v-if="rankingList.length < total">
      <nut-button type="primary" @click="loadMore">加载更多</nut-button>
    </view>
  </view>

  <view class="my-profile" v-if="myNickname">
    <image class="avatar" :src="myAvatar" mode="aspectFill" />
    <view class="user-box">
      <text class="nickname">{{ myNickname }}</text>
      <text class="hint">我的粑王名片</text>
      <text class="hint">本期当前猫次数：{{ myCatPeriodCount }}</text>
    </view>
  </view>
  <nut-popup 
    position="bottom" 
    v-model:visible="showCatSelector"
    round
    class="bottom-popup"
    :overlay-style="{ background: 'rgba(0,0,0,0.4)' }"
  >
    <view class="record-detail-popup">
      <view class="popup-header"><text>选择猫咪</text></view>
      <view class="popup-content">
        <view v-if="!cats.length">暂无猫咪，请先在首页新增</view>
        <view v-else>
          <view v-for="c in cats" :key="c.id" style="display:flex;justify-content:space-between;align-items:center;padding:12rpx 0;">
            <text>{{ c.name || '未命名' }}</text>
            <template v-if="env==='WEAPP'">
              <button class="nut-button" @tap="() => selectCat(String(c.id))">选择</button>
            </template>
            <template v-else>
              <nut-button @click="() => selectCat(String(c.id))">选择</nut-button>
            </template>
          </view>
        </view>
      </view>
      <view class="popup-actions">
        <template v-if="env==='WEAPP'">
          <button class="nut-button" @tap="showCatSelector=false">取消</button>
        </template>
        <template v-else>
          <nut-button @click="showCatSelector=false">取消</nut-button>
        </template>
      </view>
    </view>
  </nut-popup>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import Taro from '@tarojs/taro'
import { usePoopStore } from '@/store/poop'
import { storageManager } from '@/utils/storage'
import { get, ensureAuth } from '@/utils/request'

interface RankingUser {
  id: string
  nickname: string
  avatar: string
  totalCount: number
  totalDuration: number
}

interface MyRankingInfo {
  rank: number
  totalCount: number
  totalDuration: number
}

const poopStore = usePoopStore()
const rankingList = ref<RankingUser[]>([])
const page = ref(1)
const limit = ref(10)
const total = ref(0)
const myRanking = ref<MyRankingInfo | null>(null)
const defaultAvatar = 'https://img.yzcdn.cn/vant/cat.jpeg'
const myNickname = ref('')
const myAvatar = ref(defaultAvatar)
const refreshMyUser = () => {
  const u = storageManager.getUserInfo()
  myNickname.value = u?.nickName || ''
  myAvatar.value = u?.avatarUrl || defaultAvatar
}

type Period = 'total' | 'day' | 'week' | 'month'
const activePeriod = ref<Period>('total')
const periodTabs = [
  { label: '总榜', value: 'total' as Period },
  { label: '日榜', value: 'day' as Period },
  { label: '周榜', value: 'week' as Period },
  { label: '月榜', value: 'month' as Period }
]

const setActivePeriod = (p: Period) => {
  activePeriod.value = p
  updateRanking()
  computeMyCatPeriodCount()
}

// 获取排名图标
const getRankIcon = (index: number): string => {
  const icons = ['🥇', '🥈', '🥉']
  return icons[index] || ''
}

// 获取排名样式类
const getRankClass = (index: number): string => {
  const classes = ['first', 'second', 'third']
  return classes[index] || 'other'
}


const updateRanking = async () => {
  try {
    const data: any = await get('/api/ranking/list', { period: activePeriod.value, pageNum: page.value, pageSize: limit.value })
    const list = Array.isArray(data?.list) ? data.list : []
    total.value = Number(data?.total || 0)
    const ui = storageManager.getUserInfo()
    rankingList.value = (list as any[]).map((it: any) => ({
      id: it.id,
      nickname: ui && ui.id === it.id ? (ui.nickName || '') : '',
      avatar: ui && ui.id === it.id ? (ui.avatarUrl || '') : '',
      totalCount: it.totalCount,
      totalDuration: it.totalDuration
    })) as any
    const userInfo = storageManager.getUserInfo()
    if (userInfo && userInfo.id) {
      const myIndex = rankingList.value.findIndex(user => user.id === userInfo.id)
      const myStats = rankingList.value.find(user => user.id === userInfo.id) as any
      if (myStats && myIndex !== -1) {
        myRanking.value = { rank: myIndex + 1, totalCount: myStats.totalCount, totalDuration: myStats.totalDuration }
      } else {
        myRanking.value = null
      }
    } else {
      myRanking.value = null
    }
  } catch (error) {
    Taro.showToast({ title: '加载失败', icon: 'error' })
  }
}

// 加载排行榜数据
const loadRankingData = async () => { page.value = 1; await updateRanking() }

const loadMore = async () => {
  if (rankingList.value.length >= total.value) return
  page.value += 1
  try {
    const data: any = await get('/api/ranking/list', { period: activePeriod.value, pageNum: page.value, pageSize: limit.value })
    const list = Array.isArray(data?.list) ? data.list : []
    const ui = storageManager.getUserInfo()
    const more = (list as any[]).map((it: any) => ({
      id: it.id,
      nickname: ui && ui.id === it.id ? (ui.nickName || '') : '',
      avatar: ui && ui.id === it.id ? (ui.avatarUrl || '') : '',
      totalCount: it.totalCount,
      totalDuration: it.totalDuration
    })) as any
    rankingList.value = rankingList.value.concat(more)
  } catch (e) {
    Taro.showToast({ title: '加载失败', icon: 'error' })
  }
}

onMounted(() => {
  const token = Taro.getStorageSync('auth-token') || ''
  if (!token) {
    Taro.showModal({ title: '提示', content: '登录后可查看排行榜，是否现在登录？', confirmText: '去登录' }).then(res => { if (res && res.confirm) { ensureAuth() } })
  } else {
    loadRankingData()
  }
  refreshMyUser()
  const onUserUpdated = () => { refreshMyUser(); updateRanking() }
  try { Taro.eventCenter.on('user-updated', onUserUpdated) } catch {}
  loadCats().then(computeMyCatPeriodCount)
})

onUnmounted(() => {
  try { Taro.eventCenter.off('user-updated') } catch {}
})

const env = Taro.getEnv()
const cats = ref<any[]>([])
const activeCatId = ref<string>('')
const showCatSelector = ref(false)
const loadCats = async () => {
  const token = Taro.getStorageSync('auth-token') || ''
  if (!token) return
  const res: any = await get('/api/cats/list')
  cats.value = res?.items || []
  if (!activeCatId.value && cats.value.length > 0) {
    activeCatId.value = String(cats.value[0]?.id || '')
  }
}
const openCatSelector = async () => { await loadCats(); if (cats.value.length) showCatSelector.value = true }
const selectCat = async (id: string) => { activeCatId.value = String(id); showCatSelector.value = false; await computeMyCatPeriodCount() }

const myCatPeriodCount = ref(0)
const periodBounds = (p: Period) => {
  const now = new Date()
  if (p === 'day') {
    const s = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0,0,0,0).getTime()
    const e = s + 24*60*60*1000
    return { s, e }
  } else if (p === 'week') {
    const d = new Date()
    const dow = d.getDay()
    const diff = (dow === 0 ? -6 : 1 - dow)
    d.setHours(0,0,0,0)
    d.setDate(d.getDate() + diff)
    const s = d.getTime()
    const e = s + 7*24*60*60*1000
    return { s, e }
  } else if (p === 'month') {
    const sDate = new Date(now.getFullYear(), now.getMonth(), 1, 0,0,0,0)
    const eDate = new Date(now.getFullYear(), now.getMonth()+1, 1, 0,0,0,0)
    return { s: sDate.getTime(), e: eDate.getTime() }
  } else {
    return { s: 0, e: new Date(now.getFullYear(), now.getMonth(), now.getDate()+1, 0,0,0,0).getTime() }
  }
}
const computeMyCatPeriodCount = async () => {
  const token = Taro.getStorageSync('auth-token') || ''
  if (!token || !activeCatId.value) { myCatPeriodCount.value = 0; return }
  const { s, e } = periodBounds(activePeriod.value)
  try {
    const data: any = await get('/api/records/list', { start: s, end: e, catId: activeCatId.value, pageNum: 1, pageSize: 1 })
    myCatPeriodCount.value = Number(data?.total || 0)
  } catch { myCatPeriodCount.value = 0 }
}
</script>

<style lang="scss">
.ranking-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  padding: 20rpx;
}

.ranking-header {
  text-align: center;
  margin-bottom: 40rpx;
  padding-top: 20rpx;

  .title {
    display: block;
    font-size: 48rpx;
    font-weight: bold;
    color: #333;
    margin-bottom: 10rpx;
  }

  .subtitle {
    display: block;
    font-size: 28rpx;
    color: #666;
  }
}

.period-tabs {
  display: flex;
  justify-content: space-around;
  background: #fff;
  border-radius: 16rpx;
  padding: 12rpx 8rpx;
  margin: 0 0 24rpx 0;
  box-shadow: 0 4rpx 16rpx rgba(0,0,0,0.06);
}

.period-tab {
  flex: 1;
  text-align: center;
  padding: 16rpx 0;
  font-size: 28rpx;
  color: #666;
  border-radius: 12rpx;
}

.period-tab.active {
  color: #fff;
  background: #8BCE92;
  font-weight: 600;
}

.ranking-list {
  background: white;
  border-radius: 20rpx;
  padding: 20rpx;
  margin-bottom: 20rpx;
  box-shadow: 0 4rpx 20rpx rgba(0, 0, 0, 0.1);
}

.ranking-item {
  display: flex;
  align-items: center;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f0f0f0;

  &:last-child {
    border-bottom: none;
  }

  &.top-three {
    background: linear-gradient(90deg, #fff9e6 0%, #fff 100%);
    margin: 10rpx 0;
    border-radius: 15rpx;
    padding: 25rpx 20rpx;
  }
}

.rank-number {
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  font-weight: bold;
  margin-right: 20rpx;

  &.first {
    background: linear-gradient(135deg, #ffd700 0%, #ffed4e 100%);
    color: #333;
  }

  &.second {
    background: linear-gradient(135deg, #c0c0c0 0%, #e8e8e8 100%);
    color: #333;
  }

  &.third {
    background: linear-gradient(135deg, #cd7f32 0%, #daa520 100%);
    color: white;
  }

  &.other {
    background: #f5f5f5;
    color: #666;
  }

  .rank-icon {
    font-size: 40rpx;
  }
}

.user-info {
  flex: 1;
  display: flex;
  align-items: center;

  .avatar {
    width: 80rpx;
    height: 80rpx;
    border-radius: 50%;
    margin-right: 20rpx;
    border: 2rpx solid #f0f0f0;
  }

  .user-details {
    display: flex;
    flex-direction: column;

    .nickname {
      font-size: 32rpx;
      font-weight: 500;
      color: #333;
      margin-bottom: 8rpx;
    }

    .count-text {
      font-size: 24rpx;
      color: #999;
    }
  }
}

.count {
  text-align: right;

  .count-number {
    display: block;
    font-size: 40rpx;
    font-weight: bold;
    color: #ff6b6b;
    margin-bottom: 4rpx;
  }

  .count-unit {
    font-size: 24rpx;
    color: #999;
  }
}

.empty-state {
  text-align: center;
  padding: 100rpx 40rpx;

  .empty-text {
    display: block;
    font-size: 36rpx;
    color: #999;
    margin-bottom: 20rpx;
  }

  .empty-subtext {
    display: block;
    font-size: 28rpx;
    color: #ccc;
  }
}

.my-ranking {
  position: fixed;
  bottom: 120rpx;
  left: 20rpx;
  right: 20rpx;
  background: white;
  border-radius: 20rpx;
  padding: 30rpx;
  box-shadow: 0 4rpx 20rpx rgba(0, 0, 0, 0.15);

  .my-ranking-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 16rpx;

    .avatar {
      width: 48rpx;
      height: 48rpx;
      border-radius: 50%;
      border: 2rpx solid #f0f0f0;
    }
    .my-rank {
      font-size: 32rpx;
      font-weight: bold;
      color: #333;
    }

    .my-count {
      font-size: 28rpx;
      color: #666;
    }
  }
}

.my-profile {
  display: flex;
  align-items: center;
  background: #fff;
  border-radius: 20rpx;
  padding: 20rpx;
  margin: 16rpx 20rpx;
  box-shadow: 0 4rpx 16rpx rgba(0,0,0,0.08);

  .avatar {
    width: 88rpx;
    height: 88rpx;
    border-radius: 50%;
    margin-right: 20rpx;
    border: 2rpx solid #f0f0f0;
  }

  .user-box {
    display: flex;
    flex-direction: column;
  }

  .nickname {
    font-size: 32rpx;
    font-weight: 600;
    color: #333;
  }

  .hint {
    font-size: 24rpx;
    color: #999;
  }
}
</style>
