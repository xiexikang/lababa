<template>
  <view class="cat-form-root">
    <view class="form-card">
      <text class="title">🐱 录入我的猫咪</text>
      <text class="subtitle">完善基本信息，便于后续健康记录</text>

      <view class="form-item">
        <text class="label">名称</text>
        <nut-input v-model="name" placeholder="必填" />
      </view>

      <view class="form-item">
        <text class="label">性别</text>
        <nut-radio-group v-model="gender">
          <nut-radio label="female">母</nut-radio>
          <nut-radio label="male">公</nut-radio>
        </nut-radio-group>
      </view>

      <view class="form-item">
        <text class="label">生日</text>
        <nut-input v-model="birthDateStr" placeholder="YYYY-MM-DD" />
      </view>

      <view class="form-item">
        <text class="label">体重(kg)</text>
        <nut-input v-model="weightKgStr" type="number" placeholder="可选" />
      </view>

      <view class="form-item">
        <text class="label">绝育</text>
        <nut-switch v-model="neutered" />
      </view>

      <view class="form-item">
        <text class="label">备注</text>
        <nut-textarea v-model="notes" placeholder="可选" :autosize="{ minHeight: 100 }" />
      </view>

      <view class="actions">
        <nut-button type="primary" class="submit-btn" :disabled="!canSubmit" @click="submit">保存</nut-button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import Taro from '@tarojs/taro'
import { ref } from 'vue'
import { post, put, get } from '@/utils/request'
import { showToast } from '@/utils/toast'

const name = ref('')
const gender = ref<string>('')
const birthDateStr = ref('')
const weightKgStr = ref('')
const neutered = ref(false)
const notes = ref('')

let redirect = '/pages/index/index'
let editId = ''
const isEdit = ref(false)
try {
  const inst: any = Taro.getCurrentInstance?.()
  const params: any = inst?.router?.params || {}
  redirect = decodeURIComponent(params?.redirect || redirect)
  editId = String(params?.id || '')
} catch {}

// 性别与生日改为直接使用 NutUI 输入控件

const canSubmit = computed(()=> name.value.trim().length > 0)

const fmtDate = (ts: number): string => {
  if (!ts || ts <= 0) return ''
  try { const d = new Date(ts); const y=d.getFullYear(); const m=String(d.getMonth()+1).padStart(2,'0'); const dd=String(d.getDate()).padStart(2,'0'); return `${y}-${m}-${dd}` } catch { return '' }
}

const loadForEdit = async () => {
  if (!editId) return
  try {
    const res: any = await get(`/api/cats/detail/${encodeURIComponent(editId)}`)
    const c = res?.cat || null
    if (c && c.id) {
      isEdit.value = true
      name.value = c.name || ''
      gender.value = c.gender || ''
      birthDateStr.value = fmtDate(Number(c.birthDate || 0))
      weightKgStr.value = c.weightKg ? String(c.weightKg) : ''
      neutered.value = !!c.neutered
      notes.value = c.notes || ''
    }
  } catch {}
}

loadForEdit()

const parseBirthToTs = (s: string): number => {
  if (!s) return 0
  try { const [y,m,d] = s.split('-').map(n=>parseInt(n,10)); return new Date(y, (m-1), d).getTime() } catch { return 0 }
}
const submit = async () => {
  if (!canSubmit.value) return
  try {
    const payload: any = {
      name: name.value.trim(),
      gender: gender.value || '',
      birthDate: parseBirthToTs(birthDateStr.value),
      weightKg: weightKgStr.value ? Number(weightKgStr.value) : undefined,
      neutered: neutered.value,
      notes: notes.value || ''
    }
    const res: any = isEdit.value ? await post('/api/cats/update', { id: editId, ...payload }) : await post('/api/cats/create', payload)
    const c = (res && (res.cat || res))
    if (c && c.id) {
      showToast({ title: '保存成功', icon: 'success' })
      try {
        const tabPages = new Set([
          '/pages/index/index',
          '/pages/statistics/index',
          '/pages/profile/index',
          '/pages/ranking/index'
        ])
        if (redirect && tabPages.has(redirect)) {
          Taro.switchTab({ url: redirect })
        } else if (redirect) {
          Taro.redirectTo({ url: redirect })
        } else {
          Taro.switchTab({ url: '/pages/index/index' })
        }
      } catch {}
    } else {
      showToast({ title: '保存失败', icon: 'none' })
    }
  } catch (e) {
    showToast({ title: '保存失败', icon: 'none' })
  }
}
</script>

<style lang="scss">
.cat-form-root { min-height: 100vh; background: #f7f7f7; padding: 32rpx; }
.form-card { background: #fff; border-radius: 24rpx; padding: 32rpx; box-shadow: 0 8rpx 24rpx rgba(0,0,0,0.06); }
.title { font-size: 36rpx; font-weight: 700; color: #333; }
.subtitle { display: block; margin-top: 8rpx; font-size: 26rpx; color: #666; }
.form-item { margin-top: 24rpx; }
.label { display: inline-block; margin-bottom: 8rpx; color: #444; font-size: 26rpx; }
.mp-input { width: 100%; height: 80rpx; border: 1px solid #eee; border-radius: 12rpx; padding: 0 20rpx; background: #fafafa; }
.mp-textarea { width: 100%; min-height: 120rpx; border: 1px solid #eee; border-radius: 12rpx; padding: 12rpx 20rpx; background: #fafafa; }
.picker-display { height: 80rpx; display:flex; align-items:center; border:1px solid #eee; border-radius:12rpx; padding: 0 20rpx; background:#fafafa; }
.actions { margin-top: 32rpx; }
.submit-btn { width: 100%; }
</style>
