<template>
  <view class="cat-form-root">
    <view class="form-card">
      <text class="title">🐱 录入我的猫咪</text>
      <text class="subtitle">完善基本信息，便于后续健康记录</text>

      <view class="form-item">
        <text class="label">名称</text>
        <template v-if="env==='WEAPP'">
          <input class="mp-input" :value="name" placeholder="必填" @input="e=>name=e.detail.value" />
        </template>
        <template v-else>
          <nut-input v-model="name" placeholder="必填" />
        </template>
      </view>

      <view class="form-item">
        <text class="label">性别</text>
        <template v-if="env==='WEAPP'">
          <picker mode="selector" :range="genderRange" @change="onGenderChange">
            <view class="picker-display">{{ genderLabel }}</view>
          </picker>
        </template>
        <template v-else>
          <nut-radiogroup v-model="gender">
            <nut-radio label="female">母</nut-radio>
            <nut-radio label="male">公</nut-radio>
          </nut-radiogroup>
        </template>
      </view>

      <view class="form-item">
        <text class="label">生日</text>
        <template v-if="env==='WEAPP'">
          <picker mode="date" @change="onBirthChange">
            <view class="picker-display">{{ birthDateDisplay }}</view>
          </picker>
        </template>
        <template v-else>
          <nut-input v-model="birthDateStr" placeholder="YYYY-MM-DD" />
        </template>
      </view>

      <view class="form-item">
        <text class="label">体重(kg)</text>
        <template v-if="env==='WEAPP'">
          <input class="mp-input" type="digit" :value="weightKgStr" placeholder="可选" @input="e=>weightKgStr=e.detail.value" />
        </template>
        <template v-else>
          <nut-input v-model="weightKgStr" type="number" placeholder="可选" />
        </template>
      </view>

      <view class="form-item">
        <text class="label">绝育</text>
        <template v-if="env==='WEAPP'">
          <switch :checked="neutered" @change="e=>neutered=e.detail.value" />
        </template>
        <template v-else>
          <nut-switch v-model="neutered" />
        </template>
      </view>

      <view class="form-item">
        <text class="label">备注</text>
        <template v-if="env==='WEAPP'">
          <textarea class="mp-textarea" :value="notes" placeholder="可选" @input="e=>notes=e.detail.value" />
        </template>
        <template v-else>
          <nut-textarea v-model="notes" placeholder="可选" :autosize="{ minHeight: 100 }" />
        </template>
      </view>

      <view class="actions">
        <template v-if="env==='WEAPP'">
          <button class="submit-btn nut-button" :disabled="!canSubmit" @tap="submit">保存</button>
        </template>
        <template v-else>
          <nut-button type="primary" class="submit-btn" :disabled="!canSubmit" @click="submit">保存</nut-button>
        </template>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import Taro from '@tarojs/taro'
import { ref, computed } from 'vue'
import { post, put, get } from '@/utils/request'
import { showToast } from '@/utils/toast'

const env = Taro.getEnv()
const name = ref('')
const gender = ref<string>('')
const genderRange = ['female','male']
const genderLabel = computed(()=> gender.value===''? '可选': (gender.value==='female'?'母':'公'))
const birthDateStr = ref('')
const birthDateDisplay = computed(()=> birthDateStr.value || '可选')
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

const onGenderChange = (e: any) => {
  const idx = Number(e?.detail?.value || 0)
  gender.value = genderRange[idx] || ''
}
const onBirthChange = (e: any) => {
  birthDateStr.value = String(e?.detail?.value || '')
}

const canSubmit = computed(()=> name.value.trim().length > 0)

const fmtDate = (ts: number): string => {
  if (!ts || ts <= 0) return ''
  try { const d = new Date(ts); const y=d.getFullYear(); const m=String(d.getMonth()+1).padStart(2,'0'); const dd=String(d.getDate()).padStart(2,'0'); return `${y}-${m}-${dd}` } catch { return '' }
}

const loadForEdit = async () => {
  if (!editId) return
  try {
    const res: any = await get('/api/cats/list')
    const items = res?.items || []
    const c = items.find((x: any) => String(x.id) === String(editId))
    if (c) {
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
    const res: any = isEdit.value ? await put(`/api/cats/update/${editId}`, payload) : await post('/api/cats/create', payload)
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
