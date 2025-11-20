<template>
  <nut-popup 
    v-model:visible="showModel" 
    position="bottom" 
    round
    closeable
    class="detail-popup-container"
    :overlay-style="{ background: 'rgba(0,0,0,0.45)' }"
  >
    <view class="detail-popup">
      <view class="popup-header">
        <text class="popup-title">💩 粑粑详情记录</text>
        <text class="popup-subtitle">记录你的便便状态，关注健康</text>
      </view>
      
      <scroll-view scroll-y class="popup-content">
        <!-- 颜色选择 -->
        <view class="form-section">
          <view class="section-header">
            <text class="section-title">🎨 颜色选择</text>
          </view>
          <view class="color-options">
            <view 
              v-for="color in colorOptions" 
              :key="color.value"
              :class="['color-option', { active: form.color === color.value }]"
              @click="form.color = color.value"
            >
              <view 
                class="color-circle" 
                :style="{ backgroundColor: color.color }"
              ></view>
              <text class="color-label">{{ color.label }}</text>
            </view>
          </view>
        </view>

        <!-- 状态选择 -->
        <view class="form-section">
          <view class="section-header">
            <text class="section-title">😊 状态选择</text>
          </view>
          <view class="status-options">
            <view 
              v-for="status in statusOptions" 
              :key="status.value"
              :class="['status-option', { active: form.status === status.value }]"
              @click="form.status = status.value"
            >
              <text class="status-icon">{{ status.icon }}</text>
              <text class="status-label">{{ status.label }}</text>
            </view>
          </view>
        </view>

        <!-- 形状选择 -->
        <view class="form-section">
          <view class="section-header">
            <text class="section-title">📏 形状选择</text>
          </view>
          <view class="shape-options">
            <view 
              v-for="shape in shapeOptions" 
              :key="shape.value"
              :class="['shape-option', { active: form.shape === shape.value }]"
              @click="form.shape = shape.value"
            >
              <text class="shape-icon">{{ shape.icon }}</text>
              <text class="shape-label">{{ shape.label }}</text>
            </view>
          </view>
        </view>

        <!-- 把量选择 -->
        <view class="form-section">
          <view class="section-header">
            <text class="section-title">📊 把量选择</text>
          </view>
          <view class="amount-options">
            <view 
              v-for="amount in amountOptions" 
              :key="amount.value"
              :class="['amount-option', { active: form.amount === amount.value }]"
              @click="form.amount = amount.value"
            >
              <text class="amount-icon">{{ amount.icon }}</text>
              <text class="amount-label">{{ amount.label }}</text>
            </view>
          </view>
        </view>

        <!-- 备注 -->
        <view class="form-section">
          <view class="section-header">
            <text class="section-title">📝 备注（选填）</text>
          </view>
          <nut-textarea 
            v-model="form.note" 
            placeholder="有什么特别想记录的吗？"
            :autosize="{ minHeight: 120 }"
            maxlength="200"
          />
        </view>
      </scroll-view>

      <!-- 底部按钮 -->
      <view class="popup-footer">
        <nut-button 
          color="#ccc" 
          class="cancel-btn"
          @click="dataInfo.close()"
        >
          我再看看
        </nut-button>
        <nut-button 
          color="#8BCE92" 
          class="confirm-btn"
          @click="dataInfo.confirm()"
          :disabled="!isFormValid"
        >
          确定保存
        </nut-button>
      </view>
    </view>
  </nut-popup>
</template>

<script setup name="DetailRecordPopup">
  import { ref, reactive, computed, toRefs, watch, defineProps } from 'vue';
  
  
  const props = defineProps({
    modelValue: { type: Boolean, default: false },
  });
  
  const emit = defineEmits(['update:modelValue', 'on-ok']);
  
  const showModel = computed({
    get() { return props?.modelValue; },
    set(value) {
      emit('update:modelValue', value);
    },
  });

  // 选项数据
  const colorOptions = [
    { value: 'yellow-brown', label: '黄褐色', color: '#D2691E', icon: '💩' },
    { value: 'brown', label: '棕色', color: '#8B4513', icon: '💩' },
    { value: 'black', label: '黑色', color: '#2F2F2F', icon: '🌑' },
    { value: 'green', label: '绿色', color: '#228B22', icon: '🍃' },
    { value: 'red', label: '红色', color: '#DC143C', icon: '🔴' },
    { value: 'gray-white', label: '灰白色', color: '#D3D3D3', icon: '⚪' }
  ];

  const statusOptions = [
    { value: 'normal', label: '正常', icon: '😊' },
    { value: 'constipation', label: '便秘', icon: '😣' },
    { value: 'diarrhea', label: '拉肚子', icon: '😰' }
  ];

  const shapeOptions = [
    { value: 'banana', label: '香蕉状', icon: '🍌' },
    { value: 'granular', label: '颗粒状', icon: '⚪' },
    { value: 'soft', label: '软糊糊', icon: '🍮' },
    { value: 'cracked', label: '裂块条纹', icon: '🍪' },
    { value: 'watery', label: '水样便便', icon: '💧' }
  ];

  const amountOptions = [
    { value: 'very-little', label: '非常少', icon: '💧' },
    { value: 'little', label: '少量', icon: '🥤' },
    { value: 'moderate', label: '适中', icon: '🍺' },
    { value: 'lot', label: '大量', icon: '🛁' }
  ];

  const dataInfo = reactive({
    form: {
      color: '',
      status: '',
      shape: '',
      amount: '',
      note: ''
    },
    
    // 关闭弹窗
    close() {
      showModel.value = false;
      this.resetForm();
    },
    
    // 确认保存
    confirm() {
      if (!isFormValid.value) {
        return;
      }
      
      console.log('保存记录:', dataInfo.form);
      emit('on-ok', { ...dataInfo.form });
      this.close();
    },
    
    // 重置表单
    resetForm() {
      this.form = {
        color: '',
        status: '',
        shape: '',
        amount: '',
        note: ''
      };
    }
  });

  // 计算属性：验证表单是否完整
  const isFormValid = computed(() => {
    return dataInfo.form.color && 
           dataInfo.form.status && 
           dataInfo.form.shape && 
           dataInfo.form.amount;
  });

  const { form } = toRefs(dataInfo);

  const onNoteInput = (e) => {
    dataInfo.form.note = e && e.detail ? e.detail.value : ''
  }
</script>

<style lang="scss">
.detail-popup {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  
  .popup-header {
    padding: 40rpx 40rpx 20rpx;
    text-align: center;
    background: #fff;
    border-radius: 0 0 20rpx 20rpx;
    box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);
    
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
  
  .popup-content {
    flex: 1;
    padding: 20rpx 40rpx;
    
    .form-section {
      margin-bottom: 40rpx;
      background: #fff;
      border-radius: 20rpx;
      padding: 30rpx;
      box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.05);
      
      .section-header {
        margin-bottom: 24rpx;
        
        .section-title {
          font-size: 28rpx;
          font-weight: 600;
          color: #333;
        }
      }
      
      // 颜色选项
      .color-options {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 20rpx;
        
        .color-option {
          display: flex;
          flex-direction: column;
          align-items: center;
          padding: 20rpx;
          border-radius: 16rpx;
          background: #f8f9fa;
          border: 2rpx solid transparent;
          transition: all 0.3s ease;
          cursor: pointer;
          
          &.active {
            background: #e8f5e8;
            border-color: #8BCE92;
            transform: scale(1.05);
          }
          
          .color-circle {
            width: 60rpx;
            height: 60rpx;
            border-radius: 50%;
            margin-bottom: 12rpx;
            border: 2rpx solid #ddd;
          }
          
          .color-label {
            font-size: 24rpx;
            color: #666;
            text-align: center;
          }
        }
      }
      
      // 状态、形状、把量选项
      .status-options,
      .shape-options,
      .amount-options {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 20rpx;
        
        .status-option,
        .shape-option,
        .amount-option {
          display: flex;
          flex-direction: column;
          align-items: center;
          padding: 24rpx 16rpx;
          border-radius: 16rpx;
          background: #f8f9fa;
          border: 2rpx solid transparent;
          transition: all 0.3s ease;
          cursor: pointer;
          
          &.active {
            background: #e8f5e8;
            border-color: #8BCE92;
            transform: scale(1.05);
          }
          
          .status-icon,
          .shape-icon,
          .amount-icon {
            font-size: 40rpx;
            margin-bottom: 12rpx;
          }
          
          .status-label,
          .shape-label,
          .amount-label {
            font-size: 24rpx;
            color: #666;
            text-align: center;
          }
        }
      }
      
      // 把量选项特殊布局
      .amount-options {
        grid-template-columns: repeat(2, 1fr);
      }
      
      // 备注区域
      .nut-textarea,
      .mp-textarea {
        background: #f8f9fa;
        border-radius: 12rpx;
        padding: 20rpx;
        font-size: 26rpx;
        border: 2rpx solid transparent;
        width: 100%;
        box-sizing: border-box;
      }
      .mp-textarea:focus {
        border-color: #8BCE92;
        background: #fff;
      }
    }
  }
  
  .popup-footer {
    padding: 20rpx 40rpx 40rpx;
    background: #fff;
    display: flex;
    gap: 20rpx;
    
    .cancel-btn,
    .confirm-btn {
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
}

.detail-popup-container {
  .nut-popup__content {
    height: 80vh;
    border-top-left-radius: 24rpx;
    border-top-right-radius: 24rpx;
    overflow: hidden;
    background: #fff;
  }
}

// 动画效果
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(20rpx); }
  to { opacity: 1; transform: translateY(0); }
}

.form-section {
  animation: fadeIn 0.4s ease-out;
  
  &:nth-child(2) { animation-delay: 0.1s; }
  &:nth-child(3) { animation-delay: 0.2s; }
  &:nth-child(4) { animation-delay: 0.3s; }
  &:nth-child(5) { animation-delay: 0.4s; }
}
</style>
