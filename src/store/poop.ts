import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { storageManager } from '@/utils/storage';

// 排便记录接口
export interface PoopRecord {
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

// 用户信息接口
export interface UserInfo {
  id: string;
  nickName: string;
  avatarUrl?: string;
  openId?: string;
}

export const usePoopStore = defineStore('poop', () => {
  // 状态
  const currentRecord = ref<PoopRecord | null>(null);
  const isRecording = ref(false);
  const startTime = ref<number>(0);
  const elapsedTime = ref(0);
  const records = ref<PoopRecord[]>([]);
  const userInfo = ref<UserInfo | null>(null);
  const lastRecordTime = ref<number>(0);

  // 计算属性
  const totalRecords = computed(() => records.value.length);
  const averageDuration = computed(() => {
    if (records.value.length === 0) return 0;
    const total = records.value.reduce((sum, record) => sum + record.duration, 0);
    return Math.floor(total / records.value.length);
  });
  const longestDuration = computed(() => {
    if (records.value.length === 0) return 0;
    return Math.max(...records.value.map(record => record.duration));
  });
  const timeSinceLastRecord = computed(() => {
    if (!lastRecordTime.value) return '';
    const now = Date.now();
    const diff = now - lastRecordTime.value;
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    
    if (hours > 0) {
      return `${hours}小时${minutes}分钟`;
    }
    return `${minutes}分钟`;
  });

  // 方法
  const startRecording = () => {
    isRecording.value = true;
    startTime.value = Date.now();
    elapsedTime.value = 0;
  };

  const stopRecording = () => {
    isRecording.value = false;
    elapsedTime.value = 0;
  };

  const updateElapsedTime = () => {
    if (isRecording.value && startTime.value) {
      elapsedTime.value = Math.floor((Date.now() - startTime.value) / 1000);
    }
  };

  const saveRecord = (details: Omit<PoopRecord, 'id' | 'userId' | 'startTime' | 'endTime' | 'duration' | 'isCompleted' | 'createdAt'>) => {
    const endTime = Date.now();
    const duration = Math.floor((endTime - startTime.value) / 1000);
    
    const record: PoopRecord = {
      id: generateId(),
      userId: userInfo.value?.id || 'default-user',
      startTime: startTime.value,
      endTime,
      duration,
      color: details.color,
      status: details.status,
      shape: details.shape,
      amount: details.amount,
      note: details.note,
      isCompleted: true,
      createdAt: Date.now()
    };

    records.value.unshift(record);
    lastRecordTime.value = endTime;
    
    // 保存到本地存储
    saveToLocalStorage();
    
    // 重置状态
    stopRecording();
    
    return record;
  };

  const loadRecords = () => {
    try {
      const stored = storageManager.getRecords();
      if (stored && stored.length > 0) {
        records.value = stored;
        if (records.value.length > 0) {
          lastRecordTime.value = Math.max(...records.value.map(r => r.endTime));
        }
      }
    } catch (error) {
      console.error('加载记录失败:', error);
    }
  };

  const saveToLocalStorage = () => {
    try {
      storageManager.saveRecords(records.value);
      storageManager.saveLastRecordTime(lastRecordTime.value);
    } catch (error) {
      console.error('保存记录失败:', error);
    }
  };

  const clearRecords = () => {
    records.value = [];
    lastRecordTime.value = 0;
    saveToLocalStorage();
  };

  const generateId = () => {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
  };

  // 初始化
  const init = () => {
    loadRecords();
  };

  return {
    // 状态
    currentRecord,
    isRecording,
    startTime,
    elapsedTime,
    records,
    userInfo,
    lastRecordTime,
    
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
    loadRecords,
    saveToLocalStorage,
    clearRecords,
    init
  };
});