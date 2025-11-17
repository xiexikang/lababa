import Taro from '@tarojs/taro'
// 本地存储管理工具
export const STORAGE_KEYS = {
  RECORDS: 'poop-records',
  USER_INFO: 'user-info',
  SETTINGS: 'app-settings',
  LAST_RECORD_TIME: 'last-record-time',
  CURRENT_RECORD: 'current-record'
} as const;

export interface StorageData {
  records: any[];
  userInfo: any;
  settings: any;
  lastRecordTime: number;
  currentRecord: any;
}

class LocalStorageManager {
  private isWeapp(): boolean {
    try {
      return Taro.getEnv() === Taro.ENV_TYPE.WEAPP
    } catch {
      return false
    }
  }
  // 保存数据
  setItem(key: string, value: any): boolean {
    try {
      const data = JSON.stringify(value);
      if (this.isWeapp()) {
        Taro.setStorageSync(key, data)
      } else {
        localStorage.setItem(key, data);
      }
      return true;
    } catch (error) {
      console.error('保存数据失败:', error);
      return false;
    }
  }

  // 获取数据
  getItem<T>(key: string, defaultValue: T | null = null): T | null {
    try {
      let data: any = null
      if (this.isWeapp()) {
        const v = Taro.getStorageSync(key)
        data = typeof v === 'string' ? v : (v ? JSON.stringify(v) : null)
      } else {
        data = localStorage.getItem(key)
      }
      if (data === null || data === undefined || data === '') return defaultValue;
      return JSON.parse(data as string);
    } catch (error) {
      console.error('获取数据失败:', error);
      return defaultValue;
    }
  }

  // 删除数据
  removeItem(key: string): boolean {
    try {
      if (this.isWeapp()) {
        Taro.removeStorageSync(key)
      } else {
        localStorage.removeItem(key);
      }
      return true;
    } catch (error) {
      console.error('删除数据失败:', error);
      return false;
    }
  }

  // 清空所有数据
  clear(): boolean {
    try {
      if (this.isWeapp()) {
        Taro.clearStorageSync()
      } else {
        localStorage.clear();
      }
      return true;
    } catch (error) {
      console.error('清空数据失败:', error);
      return false;
    }
  }

  // 批量保存记录
  saveRecords(records: any[]): boolean {
    return this.setItem(STORAGE_KEYS.RECORDS, records);
  }

  // 获取记录
  getRecords(): any[] {
    return this.getItem(STORAGE_KEYS.RECORDS, []) || [];
  }

  // 保存用户信息
  saveUserInfo(userInfo: any): boolean {
    return this.setItem(STORAGE_KEYS.USER_INFO, userInfo);
  }

  // 获取用户信息
  getUserInfo(): any | null {
    return this.getItem(STORAGE_KEYS.USER_INFO);
  }

  // 保存设置
  saveSettings(settings: any): boolean {
    return this.setItem(STORAGE_KEYS.SETTINGS, settings);
  }

  // 获取设置
  getSettings(): any | null {
    return this.getItem(STORAGE_KEYS.SETTINGS);
  }

  // 保存上次记录时间
  saveLastRecordTime(time: number): boolean {
    return this.setItem(STORAGE_KEYS.LAST_RECORD_TIME, time);
  }

  // 获取上次记录时间
  getLastRecordTime(): number {
    return this.getItem(STORAGE_KEYS.LAST_RECORD_TIME, 0) || 0;
  }

  // 保存当前记录（用于恢复未完成的记录）
  saveCurrentRecord(record: any): boolean {
    return this.setItem(STORAGE_KEYS.CURRENT_RECORD, record);
  }

  // 获取当前记录
  getCurrentRecord(): any | null {
    return this.getItem(STORAGE_KEYS.CURRENT_RECORD);
  }

  // 删除当前记录
  removeCurrentRecord(): boolean {
    return this.removeItem(STORAGE_KEYS.CURRENT_RECORD);
  }

  // 备份数据
  backupData(): string {
    const data: Partial<StorageData> = {};
    
    const records = this.getRecords();
    if (records.length > 0) data.records = records;
    
    const userInfo = this.getUserInfo();
    if (userInfo) data.userInfo = userInfo;
    
    const settings = this.getSettings();
    if (settings) data.settings = settings;
    
    const lastRecordTime = this.getLastRecordTime();
    if (lastRecordTime > 0) data.lastRecordTime = lastRecordTime;
    
    const currentRecord = this.getCurrentRecord();
    if (currentRecord) data.currentRecord = currentRecord;
    
    return JSON.stringify(data, null, 2);
  }

  // 恢复数据
  restoreData(jsonString: string): boolean {
    try {
      const data = JSON.parse(jsonString) as StorageData;
      
      if (data.records) {
        this.saveRecords(data.records);
      }
      
      if (data.userInfo) {
        this.saveUserInfo(data.userInfo);
      }
      
      if (data.settings) {
        this.saveSettings(data.settings);
      }
      
      if (data.lastRecordTime) {
        this.saveLastRecordTime(data.lastRecordTime);
      }
      
      if (data.currentRecord) {
        this.saveCurrentRecord(data.currentRecord);
      }
      
      return true;
    } catch (error) {
      console.error('恢复数据失败:', error);
      return false;
    }
  }

  // 获取存储使用情况
  getStorageInfo(): { used: number; remaining: number; total: number } {
    try {
      if (this.isWeapp()) {
        const info = Taro.getStorageInfoSync()
        const used = (info.currentSize || 0) * 1024
        const total = (info.limitSize || 10 * 1024) * 1024
        const remaining = total - used
        return { used, remaining, total }
      } else {
        let used = 0;
        for (let i = 0; i < localStorage.length; i++) {
          const key = localStorage.key(i);
          if (key) {
            const value = localStorage.getItem(key);
            if (value) {
              used += key.length + value.length;
            }
          }
        }
        const total = 10 * 1024 * 1024;
        const remaining = total - used;
        return { used, remaining, total };
      }
    } catch (error) {
      console.error('获取存储信息失败:', error);
      return { used: 0, remaining: 0, total: 0 };
    }
  }
}

// 创建单例实例
export const storageManager = new LocalStorageManager();

export default LocalStorageManager;
