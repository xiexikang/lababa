<template>
  <view class="statistics-root">
    <!-- 页面标题 -->
  <view class="header-section">
    <view class="title-wrapper">
      <text class="main-title">📊 粑粑统计</text>
    </view>
    <view style="margin-top: 16rpx; display:flex; align-items:center; gap:16rpx; justify-content:center;">
      <text>当前猫咪：</text>
      <text v-if="activeCatId && cats.length" style="font-weight:600;">{{ (cats.find(c=>String(c.id)===String(activeCatId))||{}).name || '未命名' }}</text>
      <button v-if="env==='WEAPP'" class="nut-button" @tap="openCatSelector">切换</button>
      <nut-button v-else @click="openCatSelector">切换</nut-button>
    </view>
  </view>

    <!-- 统计内容 -->
  <view class="main-content">
      <view class="calendar-section">
        <nut-calendar-card
          :model-value="selectedDateObj"
          :first-day-of-week="1"
          @day-click="onDayClick"
        >
          <template #bottom="{ day }">
            <view v-if="statusForDay(day)" :class="['cal-dot', 'dot-' + statusForDay(day)]"></view>
          </template>
        </nut-calendar-card>
        <view class="calendar-legend">
          <view class="legend-item legend-success"><text>正常</text></view>
          <view class="legend-item legend-warning"><text>拉肚子</text></view>
          <view class="legend-item legend-danger"><text>便秘</text></view>
        </view>
      </view>

      <view class="weekly-section">
        <view class="section-header" style="display:flex;justify-content:space-between;align-items:center;">
          <text class="section-title">⚖️ 体重趋势</text>
          <view>
            <template v-if="env==='WEAPP'">
              <button class="nut-button" @tap="openAddWeight">新增体重</button>
              <button class="nut-button" @tap="openReminderSettings">提醒设置</button>
            </template>
            <template v-else>
              <nut-button @click="openAddWeight">新增体重</nut-button>
              <nut-button @click="openReminderSettings">提醒设置</nut-button>
            </template>
          </view>
        </view>
        <view v-if="weightItems.length === 0" style="margin-top:8rpx;">暂无体重记录</view>
        <view v-else class="color-grid" style="grid-template-columns: repeat(5, 1fr);">
          <view v-for="w in weightItems" :key="w.id" style="text-align:center;">
            <view style="height:120rpx;display:flex;align-items:flex-end;justify-content:center;">
              <view :style="{width:'24rpx',height:Math.max(6,Math.round((Number(w.weightKg||0)-minWeight)/(maxWeight-minWeight||1)*100))+'rpx',background:'#8BCE92',borderRadius:'8rpx'}"></view>
            </view>
            <text style="font-size:22rpx;color:#666;">{{ formatDateShort(w.date) }}</text>
            <text style="display:block;font-size:24rpx;color:#333;">{{ w.weightKg }}kg</text>
          </view>
        </view>
      </view>
      <view class="weekly-section">
        <view class="section-header">
          <text class="section-title">📅 本周概览</text>
          <view class="week-row">
            <view class="kpi-card">
              <text class="kpi-number">{{ weeklyDaysCount }}</text>
              <text class="kpi-label">本周打卡天数</text>
            </view>
            <view class="kpi-card">
              <text class="kpi-number">{{ weeklyRecordsCount }}</text>
              <text class="kpi-label">本周粑粑次数</text>
            </view>
          </view>
        </view>

        <view class="color-section">
          <text class="section-subtitle">🎨 本周色谱</text>
          <view class="color-grid">
            <view class="color-item" v-for="c in colorKeys" :key="c">
              <text class="color-name">{{ colorLabels[c] }}</text>
              <text class="color-count">{{ weeklyColorDist[c] || 0 }}次</text>
            </view>
          </view>
        </view>

        <view class="score-section">
          <text class="section-subtitle">🟢 本周状态评分</text>
          <view class="score-bar">
            <view class="score-fill" :style="{ width: weeklyScore + '%' }"></view>
            <text class="score-text">{{ weeklyScore }}</text>
          </view>
        </view>
      </view>
      <view class="stats-overview">
        <view class="stat-card">
          <text class="stat-number">{{ totalRecords }}</text>
          <text class="stat-label">总记录次数</text>
        </view>
        <view class="stat-card">
          <text class="stat-number">{{ averageDuration }}</text>
          <text class="stat-label">平均时长</text>
        </view>
        <view class="stat-card">
          <text class="stat-number">{{ longestDuration }}</text>
          <text class="stat-label">最长记录</text>
        </view>
      </view>

      <!-- 最近记录 -->
      <view class="recent-records">
        <view class="section-header">
          <text class="section-title">🕐 最近记录</text>
        </view>
        <view class="records-list">
          <view v-if="recentRecords.length === 0" class="empty-state">
            <text class="empty-text">还没有记录哦～</text>
            <text class="empty-subtext">快去主页打卡吧！</text>
          </view>
          <view 
            v-for="record in recentRecords" 
            :key="record.id"
            class="record-item"
            @tap="openDetail(record.id)"
          >
            <view class="record-header">
              <text class="record-time">{{ formatTime(record.startTime) }}</text>
              <text class="record-duration">{{ formatDuration(record.duration) }}</text>
            </view>
            <view class="record-details">
              <view class="detail-item">
                <text class="detail-label">颜色:</text>
                <text class="detail-value">{{ getColorLabel(record.color) }}</text>
              </view>
              <view class="detail-item">
                <text class="detail-label">状态:</text>
                <text class="detail-value">{{ getStatusLabel(record.status) }}</text>
              </view>
              <view class="detail-item">
                <text class="detail-label">形状:</text>
                <text class="detail-value">{{ getShapeLabel(record.shape) }}</text>
              </view>
              <view class="detail-item">
                <text class="detail-label">把量:</text>
                <text class="detail-value">{{ getAmountLabel(record.amount) }}</text>
              </view>
            </view>
          </view>
        </view>
        <view class="load-more" v-if="recentRecords.length < total">
          <nut-button type="primary" @click="loadMore">加载更多</nut-button>
        </view>
      </view>
    </view>

    <!-- 底部导航栏 -->
  <nut-popup 
    v-model:visible="detailVisible" 
      position="bottom" 
      round 
      class="bottom-popup"
      :overlay-style="{ background: 'rgba(0,0,0,0.4)' }"
    >
      <view class="record-detail-popup">
        <view class="popup-header"><text>记录详情</text></view>
        <view class="popup-content" v-if="detail">
          <text>时间：{{ new Date(detail.endTime).toLocaleString() }}</text>
          <text>时长：{{ Math.floor((detail.duration||0)/60) }}分{{ (detail.duration||0)%60 }}秒</text>
          <text>颜色：{{ detail.color }}</text>
          <text>状态：{{ detail.status }}</text>
          <text>形状：{{ detail.shape }}</text>
          <text>把量：{{ detail.amount }}</text>
          <text>备注：{{ detail.note || '无' }}</text>
        </view>
      </view>
  </nut-popup>

  <nut-popup position="bottom" v-model:visible="addWeightVisible" round class="bottom-popup" :overlay-style="{ background: 'rgba(0,0,0,0.4)' }">
    <view class="record-detail-popup">
      <view class="popup-header"><text>新增体重</text></view>
      <view class="popup-content">
        <template v-if="env==='WEAPP'">
          <input class="mp-input" type="digit" :value="addWeightValue" placeholder="kg" @input="e=>addWeightValue=e.detail.value" />
        </template>
        <template v-else>
          <nut-input v-model="addWeightValue" type="number" placeholder="kg" />
        </template>
      </view>
      <view class="popup-actions">
        <template v-if="env==='WEAPP'">
          <button class="nut-button" @tap="addWeightVisible=false">取消</button>
          <button class="nut-button" @tap="submitAddWeight">保存</button>
        </template>
        <template v-else>
          <nut-button @click="addWeightVisible=false">取消</nut-button>
          <nut-button @click="submitAddWeight">保存</nut-button>
        </template>
      </view>
    </view>
  </nut-popup>

  <nut-popup position="bottom" v-model:visible="reminderVisible" round class="bottom-popup" :overlay-style="{ background: 'rgba(0,0,0,0.4)' }">
    <view class="record-detail-popup">
      <view class="popup-header"><text>提醒设置</text></view>
      <view class="popup-content">
        <view style="display:flex;align-items:center;justify-content:space-between;padding:8rpx 0;">
          <text>总提醒开关</text>
          <template v-if="env==='WEAPP'">
            <switch :checked="remindEnabled" @change="e=>remindEnabled=e.detail.value" />
          </template>
          <template v-else>
            <nut-switch v-model="remindEnabled" />
          </template>
        </view>
        <view style="display:flex;align-items:center;justify-content:space-between;padding:8rpx 0;">
          <text>48小时未记录</text>
          <template v-if="env==='WEAPP'">
            <switch :checked="remindNoRecord" @change="e=>remindNoRecord=e.detail.value" />
          </template>
          <template v-else>
            <nut-switch v-model="remindNoRecord" />
          </template>
        </view>
        <view style="display:flex;align-items:center;justify-content:space-between;padding:8rpx 0;">
          <text>24小时腹泻≥2次</text>
          <template v-if="env==='WEAPP'">
            <switch :checked="remindDiarrhea" @change="e=>remindDiarrhea=e.detail.value" />
          </template>
          <template v-else>
            <nut-switch v-model="remindDiarrhea" />
          </template>
        </view>
        <view style="margin-top:12rpx;display:flex;gap:12rpx;">
          <template v-if="env==='WEAPP'">
            <button class="nut-button" @tap="subscribeTemplates">订阅消息</button>
          </template>
          <template v-else>
            <nut-button @click="subscribeTemplates">订阅消息</nut-button>
          </template>
        </view>
      </view>
      <view class="popup-actions">
        <template v-if="env==='WEAPP'">
          <button class="nut-button" @tap="reminderVisible=false">取消</button>
          <button class="nut-button" @tap="submitReminderSettings">保存</button>
        </template>
        <template v-else>
          <nut-button @click="reminderVisible=false">取消</nut-button>
          <nut-button @click="submitReminderSettings">保存</nut-button>
        </template>
      </view>
    </view>
  </nut-popup>
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
</view>
</template>

<script setup lang="ts" name="Statistics">
import { ref, reactive, computed, onMounted } from 'vue';
import Taro from '@tarojs/taro';
import { useSimpleStore } from '@/store/simple';
import { getRecordDetail } from '@/store/simple'
import { get, ensureAuth } from '@/utils/request'


// 使用简单的状态管理
const store = useSimpleStore();

// 统计摘要（总数、平均、最长）
const summary = ref<{ totalRecords: number; averageDuration: number; longestDuration: number }>({ totalRecords: 0, averageDuration: 0, longestDuration: 0 })
const totalRecords = computed(() => {
  const v = Number(summary.value?.totalRecords || 0)
  return v > 0 ? v : store.totalRecords
})
const averageDuration = computed(() => {
  const avg = Number(summary.value?.averageDuration || store.averageDuration || 0)
  if (!avg) return '0分钟'
  const minutes = Math.floor(avg / 60)
  const seconds = avg % 60
  return minutes > 0 ? `${minutes}分钟${seconds}秒` : `${seconds}秒`
})
const longestDuration = computed(() => {
  const longest = Number(summary.value?.longestDuration || store.longestDuration || 0)
  if (!longest) return '0分钟'
  const minutes = Math.floor(longest / 60)
  const seconds = longest % 60
  return minutes > 0 ? `${minutes}分钟${seconds}秒` : `${seconds}秒`
})
const recentRecords = ref<any[]>([])
const pageNum = ref(1)
const pageSize = ref(10)
const total = ref(0)

const loadRecent = async (reset: boolean = false) => {
  if (reset) {
    pageNum.value = 1
    recentRecords.value = []
  }
  const token = Taro.getStorageSync('auth-token') || ''
  if (!token) return
  const data: any = await get('/api/records/list', { pageNum: pageNum.value, pageSize: pageSize.value, catId: activeCatId.value || undefined })
  const items = Array.isArray(data?.items) ? data.items : []
  total.value = Number(data?.total || 0)
  recentRecords.value = recentRecords.value.concat(items)
}

const loadSummary = async () => {
  const token = Taro.getStorageSync('auth-token') || ''
  if (!token) return
  try {
    const data: any = await get('/api/statistics/summary', { catId: activeCatId.value || undefined })
    const sum = data?.summary
    if (sum) {
      summary.value = {
        totalRecords: Number(sum.totalRecords || 0),
        averageDuration: Number(sum.averageDuration || 0),
        longestDuration: Number(sum.longestDuration || 0)
      }
    }
  } catch (_) { /* ignore */ }
}

const loadMore = async () => {
  if (recentRecords.value.length >= total.value) return
  pageNum.value += 1
  await loadRecent(false)
}

// 记录详情弹窗
const detailVisible = ref(false)
const detail = ref<any>(null)
const openDetail = async (id: string) => {
  const token = Taro.getStorageSync('auth-token') || ''
  if (!token) {
    const res = await Taro.showModal({ title: '提示', content: '请先登录再查看详情', confirmText: '去登录' })
    if (res && res.confirm) { ensureAuth() }
    return
  }
  try {
    const r = await getRecordDetail(id)
    detail.value = r
    detailVisible.value = true
  } catch (e) {
    Taro.showToast({ title: '加载详情失败', icon: 'error' })
  }
}

const pad2 = (n) => String(n).padStart(2, '0');
const fmtDate = (ts) => {
  const d = new Date(ts);
  return `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}`;
};
const today = computed(() => fmtDate(Date.now()));
const selectedDateObj = ref(new Date());
const monthStartMs = computed(() => {
  const d = new Date();
  d.setDate(1);
  d.setHours(0,0,0,0);
  return d.getTime();
});
const monthEndMs = computed(() => {
  const d = new Date();
  d.setMonth(d.getMonth() + 1);
  d.setDate(1);
  d.setHours(0,0,0,0);
  return d.getTime();
});
const monthDays = ref<string[]>([])
const monthDayStatusMap = ref<Record<string, { normal: number; diarrhea: number; constipation: number }>>({})
const monthMarkDays = computed(() => monthDays.value)
const monthDayStatus = computed(() => monthDayStatusMap.value)
const statusForDay = (day) => {
  const d = `${day.year}-${pad2(day.month)}-${pad2(day.date)}`;
  const s = monthDayStatus.value[d];
  if (!s) return '';
  if (s.constipation > 0) return 'danger';
  if (s.diarrhea > 0) return 'warning';
  if (s.normal > 0) return 'success';
  return '';
};
const selectedDate = ref('');
const onDayClick = (day) => {
  if (!day || typeof day !== 'object') return;
  const clicked = new Date(day.year, day.month - 1, day.date);
  clicked.setHours(0, 0, 0, 0);
  const now = new Date();
  now.setHours(0, 0, 0, 0);

  if (clicked.getTime() > now.getTime()) {
    Taro.showToast({ title: '这一天还没到哦', icon: 'none' });
    return;
  }

  const s = `${day.year}-${pad2(day.month)}-${pad2(day.date)}`;
  selectedDate.value = s;
  selectedDateObj.value = new Date(day.year, day.month - 1, day.date);
  Taro.navigateTo({ url: `/pages/statistics/detail/index?date=${s}` });
};

// ===== 本周统计 =====
const startOfWeek = () => {
  const d = new Date();
  const day = d.getDay();
  const diff = (day === 0 ? -6 : 1 - day);
  d.setHours(0,0,0,0);
  d.setDate(d.getDate() + diff);
  return d.getTime();
};
const endOfWeek = () => {
  const s = new Date(startOfWeek());
  s.setDate(s.getDate() + 7);
  return s.getTime();
};

const weekData = ref<{ daysCount: number; recordsCount: number; colorDist: Record<string, number>; score: number }>({
  daysCount: 0,
  recordsCount: 0,
  colorDist: {},
  score: 0
})
const weeklyDaysCount = computed(() => Number(weekData.value?.daysCount || 0))
const weeklyRecordsCount = computed(() => Number(weekData.value?.recordsCount || 0))

const colorLabels = {
  'yellow-brown': '黄褐色',
  'brown': '棕色',
  'black': '黑色',
  'green': '绿色',
  'red': '红色',
  'gray-white': '灰白色'
};
const colorKeys = Object.keys(colorLabels);
const weeklyColorDist = computed(() => {
  const dist: Record<string, number> = {}
  colorKeys.forEach(k => (dist[k] = 0))
  const src = (weekData.value?.colorDist) || {}
  Object.keys(src).forEach(k => { if (dist[k] !== undefined) dist[k] = Number(src[k] || 0) })
  return dist
})

const weeklyScore = computed(() => Number(weekData.value?.score || 0))

// 格式化时间
const formatTime = (timestamp) => {
  const date = new Date(timestamp);
  return `${date.getMonth() + 1}月${date.getDate()}日 ${date.getHours()}:${date.getMinutes().toString().padStart(2, '0')}`;
};

// 格式化时长
const formatDuration = (seconds) => {
  const minutes = Math.floor(seconds / 60);
  const secs = seconds % 60;
  if (minutes > 0) {
    return `${minutes}分钟${secs}秒`;
  }
  return `${secs}秒`;
};

// 获取标签
const getColorLabel = (value) => {
  const colorMap = {
    'yellow-brown': '黄褐色',
    'brown': '棕色',
    'black': '黑色',
    'green': '绿色',
    'red': '红色',
    'gray-white': '灰白色'
  };
  return colorMap[value] || value;
};

const getStatusLabel = (value) => {
  const statusMap = {
    'normal': '正常',
    'constipation': '便秘',
    'diarrhea': '拉肚子'
  };
  return statusMap[value] || value;
};

const getShapeLabel = (value) => {
  const shapeMap = {
    'banana': '香蕉状',
    'granular': '颗粒状',
    'soft': '软糊糊',
    'cracked': '裂块条纹',
    'watery': '水样便便'
  };
  return shapeMap[value] || value;
};

const getAmountLabel = (value) => {
  const amountMap = {
    'very-little': '非常少',
    'little': '少量',
    'moderate': '适中',
    'lot': '大量'
  };
  return amountMap[value] || value;
};

// 加载统计数据
const loadStatistics = async () => {
  console.log('加载统计数据');
  const token = Taro.getStorageSync('auth-token') || ''
  if (!token) {
    const res = await Taro.showModal({ title: '提示', content: '请先登录以查看统计', confirmText: '去登录' })
    if (res && res.confirm) { ensureAuth() }
    return
  }
  // 初始化与并发拉取
  try {
    store.init();
    await Promise.all([
      (async () => {
        const resp: any = await get('/api/records/list', { catId: activeCatId.value || undefined })
        const items: any[] = Array.isArray(resp?.items) ? resp.items : []
        const s = startOfWeek()
        const e = endOfWeek()
        const filtered = items.filter(r => Number(r?.endTime || 0) >= s && Number(r?.endTime || 0) < e)
        const daySet = new Set(filtered.map(r => new Date(Number(r.endTime)).toDateString()))
        const colorDist: Record<string, number> = {}
        filtered.forEach(r => {
          const c = String(r?.color || '')
          if (!colorDist[c]) colorDist[c] = 0
          colorDist[c] += 1
        })
        let total = filtered.length
        let normal = filtered.filter(r => String(r?.status || '') === 'normal').length
        const score = total > 0 ? Math.round((normal / total) * 100) : 0
        weekData.value = { daysCount: daySet.size, recordsCount: total, colorDist, score }
      })(),
      (async () => {
        const y = new Date().getFullYear()
        const mth = new Date().getMonth() + 1
        const m: any = await get('/api/statistics/month-days', { year: y, month: mth, catId: activeCatId.value || undefined })
        monthDays.value = Array.isArray(m?.days) ? m.days : []
        monthDayStatusMap.value = m?.dayStatusMap || {}
      })(),
      loadRecent(true),
      loadSummary()
    ])
  } catch (e) {
    Taro.showToast({ title: '加载失败', icon: 'error' })
  }
};

onMounted(() => {
  loadCats();
  loadStatistics();
  loadWeights();
});
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
const selectCat = async (id: string) => { activeCatId.value = String(id); showCatSelector.value = false; await loadStatistics() }

const weightItems = ref<any[]>([])
const addWeightVisible = ref(false)
const addWeightValue = ref<string>('')
const minWeight = computed(() => {
  const vals = weightItems.value.map(w => Number(w.weightKg || 0)).filter(n => !Number.isNaN(n))
  return vals.length ? Math.min(...vals) : 0
})
const maxWeight = computed(() => {
  const vals = weightItems.value.map(w => Number(w.weightKg || 0)).filter(n => !Number.isNaN(n))
  return vals.length ? Math.max(...vals) : 1
})
const formatDateShort = (dateStr: string) => {
  const d = new Date(dateStr)
  return `${d.getMonth() + 1}/${d.getDate()}`
}
const loadWeights = async () => {
  const token = Taro.getStorageSync('auth-token') || ''
  if (!token) return
  try {
    const data: any = await get('/api/weights/list', { catId: activeCatId.value || undefined, pageSize: 20 })
    weightItems.value = Array.isArray(data?.items) ? data.items : []
  } catch (_) { }
}
const openAddWeight = () => {
  addWeightValue.value = ''
  addWeightVisible.value = true
}
const submitAddWeight = async () => {
  const v = Number(addWeightValue.value)
  if (!v || Number.isNaN(v)) {
    Taro.showToast({ title: '请输入有效体重', icon: 'none' })
    return
  }
  const token = Taro.getStorageSync('auth-token') || ''
  if (!token) { ensureAuth(); return }
  try {
    await get('/api/weights/add', { weightKg: v, date: new Date().toISOString(), catId: activeCatId.value || undefined })
    addWeightVisible.value = false
    await loadWeights()
  } catch (_) {
    Taro.showToast({ title: '保存失败', icon: 'error' })
  }
}

const reminderVisible = ref(false)
const remindEnabled = ref(true)
const remindNoRecord = ref(false)
const remindDiarrhea = ref(false)
const openReminderSettings = () => { reminderVisible.value = true }
const submitReminderSettings = async () => {
  const token = Taro.getStorageSync('auth-token') || ''
  if (!token) { ensureAuth(); return }
  try {
    await get('/api/reminder/save', { enabled: remindEnabled.value, noRecord48h: remindNoRecord.value, diarrhea2x24h: remindDiarrhea.value, catId: activeCatId.value || undefined })
    reminderVisible.value = false
    Taro.showToast({ title: '已保存', icon: 'success' })
  } catch (_) {
    Taro.showToast({ title: '保存失败', icon: 'error' })
  }
}
const subscribeTemplates = async () => {
  if (env === 'WEAPP') {
    try {
      // @ts-ignore
      await Taro.requestSubscribeMessage({ tmplIds: [] })
      Taro.showToast({ title: '订阅成功', icon: 'success' })
    } catch (_) {
      Taro.showToast({ title: '订阅失败', icon: 'error' })
    }
  } else {
    Taro.showToast({ title: '仅微信小程序支持订阅', icon: 'none' })
  }
}
</script>

<style lang="scss">
.statistics-root {
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
  
  .main-content {
    padding: 0 40rpx;
    .calendar-section { background: rgba(255,255,255,0.9); border-radius: 20rpx; padding: 20rpx; margin-bottom: 20rpx; box-shadow: 0 4rpx 12rpx rgba(0,0,0,0.08); }
    .calendar-legend { display:flex; gap:16rpx; margin-top:12rpx; }
    .legend-item { display:flex; align-items:center; gap:8rpx; font-size:24rpx; color:#333; }
    .legend-success::before { content:''; width:16rpx; height:16rpx; background:#6ecb6d; border-radius:50%; display:inline-block; }
    .legend-warning::before { content:''; width:16rpx; height:16rpx; background:#ffb60d; border-radius:50%; display:inline-block; }
    .legend-danger::before { content:''; width:16rpx; height:16rpx; background:#ff6b6b; border-radius:50%; display:inline-block; }
    .cal-dot { width: 10rpx; height: 10rpx; border-radius: 50%; margin-top: 6rpx; }
    .dot-success { background:#6ecb6d; }
    .dot-warning { background:#ffb60d; }
    .dot-danger { background:#ff6b6b; }
    .weekly-section {
      background: rgba(255,255,255,0.9);
      border-radius: 20rpx;
      padding: 30rpx;
      margin-bottom: 30rpx;
      box-shadow: 0 4rpx 12rpx rgba(0,0,0,0.1);
      .section-header {
        margin-bottom: 20rpx;
        .section-title { font-size: 32rpx; font-weight: 600; color: #333; }
        .week-row { display: flex; gap: 20rpx; margin-top: 16rpx; }
        .kpi-card { flex: 1; background: #f7fff7; border: 2rpx solid #8BCE92; border-radius: 16rpx; padding: 24rpx; text-align: center; }
        .kpi-number { display:block; font-size: 40rpx; font-weight: 700; color: #4a7c59; }
        .kpi-label { font-size: 24rpx; color: #666; }
      }
      .color-section { margin-top: 20rpx; }
      .section-subtitle { font-size: 28rpx; color: #333; margin-bottom: 12rpx; display:block; }
      .color-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16rpx; }
      .color-item { background: #fff; border-radius: 12rpx; padding: 16rpx; border: 2rpx solid #e6f5ea; }
      .color-name { font-size: 24rpx; color: #2d5a3d; }
      .color-count { float: right; font-size: 24rpx; color: #666; }
      .score-section { margin-top: 24rpx; }
      .score-bar { position: relative; height: 28rpx; border-radius: 20rpx; background: #e6f5ea; overflow: hidden; }
      .score-fill { position:absolute; left:0; top:0; bottom:0; background: linear-gradient(90deg,#8BCE92,#6ecb6d); }
      .score-text { position: absolute; right: 12rpx; top: -36rpx; font-size: 24rpx; color:#2d5a3d; }
    }
    
    .stats-overview {
      display: flex;
      justify-content: space-between;
      margin-bottom: 40rpx;
      
      .stat-card {
        flex: 1;
        background: rgba(255, 255, 255, 0.9);
        border-radius: 20rpx;
        padding: 30rpx 20rpx;
        text-align: center;
        margin: 0 10rpx;
        box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);
        
        .stat-number {
          display: block;
          font-size: 36rpx;
          font-weight: bold;
          color: #4a7c59;
          margin-bottom: 8rpx;
        }
        
        .stat-label {
          font-size: 24rpx;
          color: #666;
        }
      }
    }
    
    .recent-records {
      .section-header {
        margin-bottom: 20rpx;
        
        .section-title {
          font-size: 32rpx;
          font-weight: 600;
          color: #fff;
        }
      }
      
      .records-list {
        .empty-state {
          text-align: center;
          padding: 60rpx 0;
          background: rgba(255, 255, 255, 0.9);
          border-radius: 20rpx;
          
          .empty-text {
            display: block;
            font-size: 28rpx;
            color: #666;
            margin-bottom: 12rpx;
          }
          
          .empty-subtext {
            font-size: 24rpx;
            color: #999;
          }
        }
        
        .record-item {
          background: rgba(255, 255, 255, 0.9);
          border-radius: 20rpx;
          padding: 30rpx;
          margin-bottom: 20rpx;
          box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);
          
          .record-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20rpx;
            padding-bottom: 16rpx;
            border-bottom: 1rpx solid #eee;
            
            .record-time {
              font-size: 28rpx;
              font-weight: 500;
              color: #333;
            }
            
            .record-duration {
              font-size: 24rpx;
              color: #666;
            }
          }
          
          .record-details {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: 16rpx;
            
            .detail-item {
              display: flex;
              align-items: center;
              
              .detail-label {
                font-size: 24rpx;
                color: #666;
                margin-right: 8rpx;
              }
              
              .detail-value {
                font-size: 24rpx;
                color: #333;
                font-weight: 500;
              }
            }
          }
        }
      }
    }
  }
}
</style>
