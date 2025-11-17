<template>
  <view class="login-root">
    <view class="login-card">
      <text class="title">🔐 登录授权</text>
      <text class="subtitle">为保证功能正常使用，请完成登录授权</text>

      <view class="actions">
        <button v-if="env==='WEAPP'" class="login-btn" @tap="doLogin" @click="doLogin">微信授权登录</button>
        <nut-button v-else type="primary" class="login-btn" @click="doLogin">登录</nut-button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import Taro from '@tarojs/taro'
import { postRaw } from '@/utils/request'
import { showToast } from '@/utils/toast'

const env = Taro.getEnv()
let redirect = '/pages/index/index'

try {
  const inst: any = Taro.getCurrentInstance?.()
  const params: any = inst?.router?.params || {}
redirect = decodeURIComponent(params?.redirect || redirect)
} catch {}

const getPrevUrl = (): string => {
  try {
    const pages: any[] = (Taro as any).getCurrentPages?.() || []
    if (pages.length >= 2) {
      const prev = pages[pages.length - 2]
      const route: string = prev?.route || prev?.$taroPath || ''
      if (route) return `/${route.startsWith('/') ? route.slice(1) : route}`
    }
  } catch {}
  return ''
}
if (!redirect) {
  const prev = getPrevUrl()
  if (prev) redirect = prev
}

const tabPages = new Set([
  '/pages/index/index',
  '/pages/statistics/index',
  '/pages/profile/index',
  '/pages/ranking/index'
])

const goBack = () => {
  try {
    const prev = getPrevUrl()
    if (redirect && tabPages.has(redirect)) {
      Taro.switchTab({ url: redirect })
      return
    }
    if (redirect && !tabPages.has(redirect)) {
      Taro.redirectTo({ url: redirect })
      return
    }
    if (prev && tabPages.has(prev)) {
      Taro.switchTab({ url: prev })
      return
    }
    if (prev && !tabPages.has(prev)) {
      Taro.navigateBack({ delta: 1 })
      return
    }
    Taro.switchTab({ url: '/pages/index/index' })
  } catch {
    Taro.switchTab({ url: '/pages/index/index' })
  }
}

const doLogin = async () => {
  try {
    showToast({ title: '正在登录...', icon: 'none' })
    const codeRes = await Taro.login()
    const code = codeRes?.code || ''
    let info: any = null
    try {
      if (typeof Taro.getUserProfile === 'function') {
        const res = await Taro.getUserProfile({ desc: '用于完善个人资料', lang: 'zh_CN' })
        info = res?.userInfo || null
      } else {
        const res = await Taro.getUserInfo({ lang: 'zh_CN' })
        info = res?.userInfo || res || null
      }
    } catch (_) {}
    const raw = await postRaw('/api/auth/weapp', { code, nickName: info?.nickName, avatarUrl: info?.avatarUrl })
    const apiRes: any = raw?.data || {}
    const u = apiRes?.data?.user
    const h: any = raw?.header || {}
    const t = (h.Authorization || h.authorization || h['X-Token'] || h['x-token'] || '') as string
    if (u) {
      Taro.setStorageSync('user-info', u)
      if (t) {
        const tokenVal = t.startsWith('Bearer') ? t.split(' ')[1] : t
        Taro.setStorageSync('auth-token', tokenVal)
      }
      try { Taro.eventCenter.trigger('user-updated', { user: u }) } catch {}
      showToast({ title: '登录成功', icon: 'success' })
      goBack()
      return
    }
    // 回退兼容
    if (info) Taro.setStorageSync('user-info', info)
    showToast({ title: '登录成功', icon: 'success' })
    goBack()
  } catch (e) {
    showToast({ title: '登录失败或取消', icon: 'none' })
  }
}
</script>

<style lang="scss">
.login-root {
  min-height: 100vh;
  background: linear-gradient(135deg, #8BCE92 0%, #6ecb6d 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40rpx;
}
.login-card {
  width: 90%;
  max-width: 640rpx;
  background: #fff;
  border-radius: 24rpx;
  padding: 40rpx;
  box-shadow: 0 8rpx 24rpx rgba(0,0,0,0.12);
  text-align: center;
}
.title {
  font-size: 36rpx;
  font-weight: 700;
  color: #333;
}
.subtitle {
  display: block;
  margin-top: 12rpx;
  font-size: 26rpx;
  color: #666;
}
.actions {
  margin-top: 32rpx;
}
.login-btn {
  width: 100%;
  height: 88rpx;
  border-radius: 44rpx;
  background: linear-gradient(135deg, #4CAF50 0%, #2E7D32 100%);
  color: #fff;
  font-size: 30rpx;
  font-weight: 600;
  border: none;
}
button.login-btn::after { border: none; }
</style>
