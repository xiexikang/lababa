import { storageManager } from './storage';
import { usePoopStore } from '@/store/poop';

// 数据管理工具
export class DataManager {
  private poopStore: ReturnType<typeof usePoopStore>;

  constructor() {
    this.poopStore = usePoopStore();
  }

  // 导出数据
  exportData(): string {
    try {
      const backupData = storageManager.backupData();
      const blob = new Blob([backupData], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      
      // 创建下载链接
      const link = document.createElement('a');
      link.href = url;
      link.download = `poop-records-${new Date().toISOString().split('T')[0]}.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      // 清理URL
      URL.revokeObjectURL(url);
      
      return backupData;
    } catch (error) {
      console.error('导出数据失败:', error);
      throw error;
    }
  }

  // 导入数据
  importData(jsonString: string): boolean {
    try {
      const success = storageManager.restoreData(jsonString);
      if (success) {
        // 重新加载store数据
        this.poopStore.loadRecords();
        return true;
      }
      return false;
    } catch (error) {
      console.error('导入数据失败:', error);
      return false;
    }
  }

  // 从文件导入
  importFromFile(file: File): Promise<boolean> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      
      reader.onload = (e) => {
        try {
          const content = e.target?.result as string;
          const success = this.importData(content);
          resolve(success);
        } catch (error) {
          reject(error);
        }
      };
      
      reader.onerror = () => {
        reject(new Error('读取文件失败'));
      };
      
      reader.readAsText(file);
    });
  }

  // 清理数据
  clearAllData(): boolean {
    try {
      // 清理store数据
      this.poopStore.clearRecords();
      
      // 清理本地存储
      storageManager.clear();
      
      return true;
    } catch (error) {
      console.error('清理数据失败:', error);
      return false;
    }
  }

  // 获取数据统计
  getDataStats() {
    const records = this.poopStore.records;
    const storageInfo = storageManager.getStorageInfo();
    
    return {
      recordCount: records.length,
      storageUsed: storageInfo.used,
      storageRemaining: storageInfo.remaining,
      storageTotal: storageInfo.total,
      oldestRecord: records.length > 0 ? Math.min(...records.map(r => r.startTime)) : null,
      newestRecord: records.length > 0 ? Math.max(...records.map(r => r.startTime)) : null
    };
  }

  // 数据验证
  validateData(data: any): { isValid: boolean; errors: string[] } {
    const errors: string[] = [];
    
    try {
      if (!data || typeof data !== 'object') {
        errors.push('数据格式不正确');
        return { isValid: false, errors };
      }

      // 验证记录数据
      if (data.records && Array.isArray(data.records)) {
        data.records.forEach((record: any, index: number) => {
          const requiredFields = ['id', 'startTime', 'endTime', 'duration', 'color', 'status', 'shape', 'amount'];
          const missingFields = requiredFields.filter(field => !record[field]);
          
          if (missingFields.length > 0) {
            errors.push(`记录 ${index + 1} 缺少字段: ${missingFields.join(', ')}`);
          }
          
          if (record.duration < 0) {
            errors.push(`记录 ${index + 1} 时长不能为负数`);
          }
          
          if (record.startTime > record.endTime) {
            errors.push(`记录 ${index + 1} 开始时间不能晚于结束时间`);
          }
        });
      }

      return {
        isValid: errors.length === 0,
        errors
      };
    } catch (error) {
      errors.push('数据验证过程中发生错误');
      return { isValid: false, errors };
    }
  }

  // 数据压缩
  compressData(data: any): string {
    try {
      return JSON.stringify(data);
    } catch (error) {
      console.error('数据压缩失败:', error);
      throw error;
    }
  }

  // 数据解压
  decompressData(compressedData: string): any {
    try {
      return JSON.parse(compressedData);
    } catch (error) {
      console.error('数据解压失败:', error);
      throw error;
    }
  }
}

// 创建单例实例
export const dataManager = new DataManager();

export default DataManager;