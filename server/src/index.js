import express from 'express'
import cors from 'cors'
import { readFileSync, writeFileSync, existsSync, mkdirSync } from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'
import { nanoid } from 'nanoid'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const dataDir = path.join(__dirname, '..', 'data')
const dbFile = path.join(dataDir, 'db.json')

if (!existsSync(dataDir)) mkdirSync(dataDir, { recursive: true })
if (!existsSync(dbFile)) writeFileSync(dbFile, JSON.stringify({ users: [], records: [] }, null, 2))

const loadDB = () => JSON.parse(readFileSync(dbFile, 'utf-8'))
const saveDB = (db) => writeFileSync(dbFile, JSON.stringify(db, null, 2))

const app = express()
app.use(cors())
app.use(express.json())

const getUser = (db, id) => db.users.find(u => u.id === id)
const upsertUser = (db, payload) => {
  let user = payload.id ? getUser(db, payload.id) : null
  if (!user) {
    user = { id: payload.id || nanoid(), nickName: payload.nickName || '', avatarUrl: payload.avatarUrl || '', openId: payload.openId || '' }
    db.users.push(user)
  }
  Object.assign(user, payload)
  return user
}

const ensureRecord = (payload, userId) => {
  const now = Date.now()
  const endTime = payload.endTime || now
  const duration = payload.duration || Math.max(1, Math.floor(((endTime) - (payload.startTime || endTime - 300000)) / 1000))
  const startTime = payload.startTime || endTime - duration * 1000
  return {
    id: nanoid(),
    userId: userId,
    startTime,
    endTime,
    duration,
    color: payload.color || 'brown',
    status: payload.status || 'normal',
    shape: payload.shape || 'banana',
    amount: payload.amount || 'moderate',
    note: payload.note || '',
    isCompleted: payload.isCompleted !== false,
    createdAt: now
  }
}

const filterRecords = (records, query) => {
  let r = records
  if (query.userId) r = r.filter(x => x.userId === query.userId)
  if (query.start) r = r.filter(x => x.endTime >= Number(query.start))
  if (query.end) r = r.filter(x => x.endTime < Number(query.end))
  return r
}

const computeSummary = (records) => {
  const totalRecords = records.length
  const totalDuration = records.reduce((s, r) => s + (r.duration || 0), 0)
  const averageDuration = totalRecords ? Math.floor(totalDuration / totalRecords) : 0
  const longestDuration = totalRecords ? Math.max(...records.map(r => r.duration || 0)) : 0
  return { totalRecords, averageDuration, longestDuration }
}

const isInPeriod = (ts, period) => {
  const d = new Date(ts)
  const now = new Date()
  if (period === 'day') {
    return d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth() && d.getDate() === now.getDate()
  }
  if (period === 'week') {
    const day = (now.getDay() + 6) % 7
    const start = new Date(now)
    start.setHours(0, 0, 0, 0)
    start.setDate(now.getDate() - day)
    const end = new Date(start)
    end.setDate(start.getDate() + 7)
    return ts >= start.getTime() && ts < end.getTime()
  }
  if (period === 'month') {
    return d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth()
  }
  return true
}

app.get('/api/health/ping', (req, res) => {
  res.json({ status: 'ok' })
})

app.post('/api/auth/weapp', (req, res) => {
  const db = loadDB()
  const code = String(req.body?.code || '')
  const nickName = req.body?.nickName || ''
  const avatarUrl = req.body?.avatarUrl || ''
  const openId = code ? `mock_${code.slice(0, 16)}` : ''
  const user = upsertUser(db, { openId, nickName, avatarUrl })
  saveDB(db)
  res.json({ user })
})

app.get('/api/users/detail/:id', (req, res) => {
  const db = loadDB()
  const user = getUser(db, req.params.id)
  if (!user) return res.status(404).json({ error: 'not_found' })
  res.json({ user })
})

// 旧接口 /api/users/update/:id 已移除（使用 server-go 的鉴权版本）

app.get('/api/records/list', (req, res) => {
  const db = loadDB()
  const filtered = filterRecords(db.records, req.query)
  const total = filtered.length
  const offset = Number(req.query.offset || 0)
  const limit = Number(req.query.limit || 50)
  const items = filtered.slice(offset, offset + limit)
  res.json({ total, items })
})

app.post('/api/records/create', (req, res) => {
  const db = loadDB()
  const userId = req.body?.userId || 'default-user'
  const record = ensureRecord(req.body || {}, userId)
  db.records.unshift(record)
  saveDB(db)
  res.json({ record })
})

app.put('/api/records/update/:id', (req, res) => {
  const db = loadDB()
  const idx = db.records.findIndex(r => r.id === req.params.id)
  if (idx === -1) return res.status(404).json({ error: 'not_found' })
  const r = db.records[idx]
  const merged = { ...r, ...req.body, id: r.id, userId: r.userId }
  db.records[idx] = merged
  saveDB(db)
  res.json({ record: merged })
})

app.delete('/api/records/delete/:id', (req, res) => {
  const db = loadDB()
  const idx = db.records.findIndex(r => r.id === req.params.id)
  if (idx === -1) return res.status(404).json({ error: 'not_found' })
  const removed = db.records.splice(idx, 1)[0]
  saveDB(db)
  res.json({ record: removed })
})

app.get('/api/statistics/summary', (req, res) => {
  const db = loadDB()
  const filtered = filterRecords(db.records, req.query)
  const summary = computeSummary(filtered)
  res.json({ summary })
})

app.get('/api/ranking/list', (req, res) => {
  const db = loadDB()
  const period = String(req.query.period || 'total')
  const statsByUser = new Map()
  db.records.forEach(r => {
    const ts = r.endTime || r.createdAt || Date.now()
    if (!isInPeriod(ts, period)) return
    const k = r.userId
    if (!statsByUser.has(k)) statsByUser.set(k, { id: k, totalCount: 0, totalDuration: 0 })
    const s = statsByUser.get(k)
    s.totalCount += 1
    s.totalDuration += r.duration || 0
  })
  const ranking = Array.from(statsByUser.values()).sort((a, b) => b.totalCount - a.totalCount).slice(0, 50)
  res.json({ list: ranking })
})

app.get('/api/index/list', (req, res) => {
  const db = loadDB()
  const filtered = filterRecords(db.records, req.query)
  const total = filtered.length
  const offset = Number(req.query.offset || 0)
  const limit = Number(req.query.limit || 10)
  const items = filtered.slice(offset, offset + limit)
  const summary = computeSummary(filtered)
  res.json({ total, items, summary })
})

const port = Number(process.env.PORT || 8080)
app.listen(port, () => {})
