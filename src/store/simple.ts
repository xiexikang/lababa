// 简单的状态管理，不使用Pinia
import { reactive, computed } from 'vue';
import { storageManager } from '@/utils/storage';
import { get, post } from '@/utils/request'
import Taro from '@tarojs/taro'

// 定义记录类型
interface PoopRecord {
  id: string;
  userId: string;
  startTime: number;
  endTime: number;
  duration: number;
  color: string;
  status: string;
  shape: string;
  amount: string;
  note?: string;
  isCompleted: boolean;
  createdAt: number;
}

// 全局状态
const globalState = reactive({
  records: [] as PoopRecord[],
  isRecording: false,
  startTime: 0,
  elapsedTime: 0,
  lastRecordTime: 0
});

// 计算属性
const totalRecords = computed(() => globalState.records.length);
const averageDuration = computed(() => {
  if (globalState.records.length === 0) return 0;
  const total = globalState.records.reduce((sum, record) => sum + record.duration, 0);
  return Math.floor(total / globalState.records.length);
});
const longestDuration = computed(() => {
  if (globalState.records.length === 0) return 0;
  return Math.max(...globalState.records.map(record => record.duration));
});
const timeSinceLastRecord = computed(() => {
  if (!globalState.lastRecordTime) return '';
  const now = Date.now();
  const diff = now - globalState.lastRecordTime;
  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
  
  if (hours > 0) {
    return `${hours}小时${minutes}分钟`;
  }
  return `${minutes}分钟`;
});

// 方法
export const startRecording = () => {
  globalState.isRecording = true;
  globalState.startTime = Date.now();
  globalState.elapsedTime = 0;
};

export const stopRecording = () => {
  globalState.isRecording = false;
  globalState.elapsedTime = 0;
};

export const updateElapsedTime = () => {
  if (globalState.isRecording && globalState.startTime) {
    globalState.elapsedTime = Math.floor((Date.now() - globalState.startTime) / 1000);
  }
};

export const saveRecord = async (details: any) => {
  const endTime = Date.now();
  const duration = Math.floor((endTime - globalState.startTime) / 1000);
  const cid = await ensureValidCatId(details.catId);
  const payload = {
    startTime: globalState.startTime,
    endTime,
    duration,
    color: details.color,
    status: details.status,
    shape: details.shape,
    amount: details.amount,
    note: details.note,
    catId: cid
  };
  const res = await post<{ record: PoopRecord }>("/api/records/create", payload);
  const record = res.record as PoopRecord;
  globalState.records.unshift(record);
  globalState.lastRecordTime = record.endTime;
  saveToLocalStorage();
  stopRecording();
  return record;
};

export const addRecordForDate = async (params: any) => {
  const { date, time, durationMinutes, color, status, shape, amount, note } = params;
  const [y, m, d] = String(date).split('-').map((x) => parseInt(x, 10));
  const [hh, mm] = String(time || '12:00').split(':').map((x) => parseInt(x, 10));
  const end = new Date(y, (m - 1), d, hh, mm, 0).getTime();
  const durSec = Math.max(1, parseInt(durationMinutes || 5, 10)) * 60;
  const start = end - durSec * 1000;
  const cid = await ensureValidCatId(params.catId);
  const payload = { startTime: start, endTime: end, duration: durSec, color, status, shape, amount, note, catId: cid };
  const res = await post<{ record: PoopRecord }>("/api/records/create", payload);
  const record = res.record as PoopRecord;
  globalState.records.unshift(record);
  globalState.lastRecordTime = end;
  saveToLocalStorage();
  return record;
};

interface CatItem { id: string; name?: string }
const fetchCatsList = async (): Promise<CatItem[]> => {
  try {
    const res = await get<{ total: number; items: CatItem[] }>("/api/cats/list")
    return (res?.items || []) as CatItem[]
  } catch {
    return []
  }
}

const ensureValidCatId = async (catId?: string): Promise<string> => {
  const cats = await fetchCatsList()
  if (!cats || cats.length === 0) {
    try { await Taro.showToast({ title: '请先创建猫咪', icon: 'none' }) } catch {}
    throw new Error('NO_CAT')
  }
  const listIds = new Set(cats.map(c => c.id))
  if (catId && listIds.has(catId)) return catId
  const first = cats[0]
  try { await Taro.showToast({ title: '未选择猫咪，已默认选择', icon: 'none' }) } catch {}
  return first.id
}

export const loadRecords = async () => {
  try {
    const res = await get<{ total: number; items: PoopRecord[] }>("/api/records/list");
    if (res && res.items) {
      globalState.records = res.items as PoopRecord[];
      if (globalState.records.length > 0) {
        globalState.lastRecordTime = Math.max(...globalState.records.map(r => r.endTime));
      }
    }
  } catch (error) {
    try {
      const stored = storageManager.getRecords();
      if (stored && stored.length > 0) {
        globalState.records = stored as PoopRecord[];
        globalState.lastRecordTime = Math.max(...globalState.records.map(r => r.endTime));
      }
    } catch (_) {}
  }
};

export const saveToLocalStorage = () => {
  try {
    storageManager.saveRecords(globalState.records);
    storageManager.saveLastRecordTime(globalState.lastRecordTime);
  } catch (error) {
    console.error('保存记录失败:', error);
  }
};

export const clearRecords = () => {
  globalState.records = [];
  globalState.lastRecordTime = 0;
  saveToLocalStorage();
};

// 初始化
export const init = () => {
  loadRecords();
};

// 导出状态和方法
export const useSimpleStore = () => {
  return {
    // 状态
    globalState,
    // 计算属性
    totalRecords,
    averageDuration,
    longestDuration,
    timeSinceLastRecord,
    // 方法
    startRecording,
    stopRecording,
    updateElapsedTime,
    saveRecord,
    addRecordForDate,
    loadRecords,
    saveToLocalStorage,
    clearRecords,
    init
  };
};

export const getRecordDetail = async (id: string) => {
  const res = await get<{ record: PoopRecord }>(`/api/records/detail/${id}`)
  return res.record as PoopRecord
}
