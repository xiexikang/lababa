package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID        string `json:"id"`
	NickName  string `json:"nickName"`
	AvatarUrl string `json:"avatarUrl"`
	OpenID    string `json:"openId"`
}

type Record struct {
	ID          string `json:"id"`
	UserID      string `json:"userId"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
	Duration    int64  `json:"duration"`
	Color       string `json:"color"`
	Status      string `json:"status"`
	Shape       string `json:"shape"`
	Amount      string `json:"amount"`
	Note        string `json:"note"`
	IsCompleted bool   `json:"isCompleted"`
	CreatedAt   int64  `json:"createdAt"`
}

type JSONDB struct {
	Users   []User   `json:"users"`
	Records []Record `json:"records"`
}

type Resp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func ensureDir(p string) error { return os.MkdirAll(p, 0755) }

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeOK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, Resp{Code: 0, Msg: "成功", Data: data})
}

func writeErr(w http.ResponseWriter, httpCode int, code int, msg string) {
	writeJSON(w, httpCode, Resp{Code: code, Msg: msg, Data: nil})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization, X-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getQueryInt64(r *http.Request, key string, def int64) int64 {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return i
}

func computeSummaryRecords(recs []Record) map[string]int64 {
	total := int64(len(recs))
	var sum int64
	var longest int64
	for _, r := range recs {
		sum += r.Duration
		if r.Duration > longest {
			longest = r.Duration
		}
	}
	var avg int64
	if total > 0 {
		avg = int64(math.Floor(float64(sum) / float64(total)))
	}
	return map[string]int64{"totalRecords": total, "averageDuration": avg, "longestDuration": longest}
}

func isInPeriod(ts int64, period string) bool {
	d := time.UnixMilli(ts)
	now := time.Now()
	if period == "day" {
		return d.Year() == now.Year() && d.Month() == now.Month() && d.Day() == now.Day()
	}
	if period == "week" {
		day := (int(now.Weekday()) + 6) % 7
		start := time.Date(now.Year(), now.Month(), now.Day()-day, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 0, 7)
		return ts >= start.UnixMilli() && ts < end.UnixMilli()
	}
	if period == "month" {
		return d.Year() == now.Year() && d.Month() == now.Month()
	}
	return true
}

func openMySQL() *sql.DB {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:password@tcp(localhost:3306)/lababa?charset=utf8mb4&parseTime=true"
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("open mysql:", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("ping mysql:", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(64) PRIMARY KEY,
			nickName VARCHAR(255),
			avatarUrl VARCHAR(512),
			openId VARCHAR(128) UNIQUE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		log.Fatal("create users table:", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS records (
			id VARCHAR(64) PRIMARY KEY,
			userId VARCHAR(64),
			startTime BIGINT,
			endTime BIGINT,
			duration BIGINT,
			color VARCHAR(32),
			status VARCHAR(32),
			shape VARCHAR(32),
			amount VARCHAR(32),
			note TEXT,
			isCompleted TINYINT(1),
			createdAt BIGINT,
			INDEX idx_records_user_end (userId, endTime)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		log.Fatal("create tables:", err)
	}

	// 会话表：保存登录 token 及过期时间
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id VARCHAR(64) PRIMARY KEY,
			userId VARCHAR(64) NOT NULL,
			token VARCHAR(128) UNIQUE NOT NULL,
			expiresAt BIGINT NOT NULL,
			INDEX idx_sessions_token (token)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		log.Fatal("create sessions table:", err)
	}

	// 排行榜表：按天聚合次数
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS leaderboard (
            id VARCHAR(64) PRIMARY KEY,
            userId VARCHAR(64) NOT NULL,
            userName VARCHAR(255),
            day VARCHAR(16) NOT NULL,
            count INT NOT NULL,
            UNIQUE KEY uniq_leaderboard (userId, day)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
    `)
	if err != nil {
		log.Fatal("create leaderboard table:", err)
	}

	// 朋友邀请表
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS friend_invites (
            id VARCHAR(64) PRIMARY KEY,
            inviterUserId VARCHAR(64) NOT NULL,
            createdAt BIGINT NOT NULL,
            INDEX idx_friend_invites_inviter (inviterUserId)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
    `)
	if err != nil {
		log.Fatal("create friend_invites table:", err)
	}

	// 朋友关系表（接受邀请后建立关系）
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS friend_relations (
            id VARCHAR(64) PRIMARY KEY,
            inviterUserId VARCHAR(64) NOT NULL,
            inviteeUserId VARCHAR(64) NOT NULL,
            createdAt BIGINT NOT NULL,
            UNIQUE KEY uniq_relation (inviterUserId, inviteeUserId),
            INDEX idx_relation_inviter (inviterUserId),
            INDEX idx_relation_invitee (inviteeUserId)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
    `)
	if err != nil {
		log.Fatal("create friend_relations table:", err)
	}
	return db
}

func migrateFromJSON(db *sql.DB, jsonPath string) {
	f, err := os.Open(jsonPath)
	if err != nil {
		return
	}
	defer f.Close()
	var jd JSONDB
	if err := json.NewDecoder(f).Decode(&jd); err != nil {
		return
	}
	tx, err := db.Begin()
	if err != nil {
		return
	}
	for _, u := range jd.Users {
		_, _ = tx.Exec("INSERT OR IGNORE INTO users(id,nickName,avatarUrl,openId) VALUES(?,?,?,?)", u.ID, u.NickName, u.AvatarUrl, u.OpenID)
	}
	for _, r := range jd.Records {
		_, _ = tx.Exec("INSERT OR IGNORE INTO records(id,userId,startTime,endTime,duration,color,status,shape,amount,note,isCompleted,createdAt) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)",
			r.ID, r.UserID, r.StartTime, r.EndTime, r.Duration, r.Color, r.Status, r.Shape, r.Amount, r.Note, boolToInt(r.IsCompleted), r.CreatedAt)
	}
	_ = tx.Commit()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func toString(v any, def string) string {
	if v == nil {
		return def
	}
	if s, ok := v.(string); ok {
		return s
	}
	return def
}

func main() {
	db := openMySQL()
	// 可选：如需迁移旧 JSON，可保留 migrateFromJSON(db, "server-go/data/db.json")

	mux := http.NewServeMux()

	// 读取 token TTL（秒）
	getTokenTTL := func() int64 {
		ttlStr := os.Getenv("TOKEN_TTL_SECONDS")
		if ttlStr == "" {
			return 7 * 24 * 3600
		}
		v, err := strconv.ParseInt(ttlStr, 10, 64)
		if err != nil || v <= 0 {
			return 7 * 24 * 3600
		}
		return v
	}

	// 鉴权辅助：依据 Authorization Bearer token 返回 userId
	getUserIDFromAuth := func(r *http.Request) (string, bool) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			return "", false
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return "", false
		}
		tok := parts[1]
		row := db.QueryRow("SELECT userId, expiresAt FROM sessions WHERE token=?", tok)
		var uid string
		var exp int64
		if err := row.Scan(&uid, &exp); err != nil {
			return "", false
		}
		if time.Now().UnixMilli() >= exp {
			return "", false
		}
		return uid, true
	}

	withAuth := func(handler func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			uid, ok := getUserIDFromAuth(r)
			if !ok {
				writeErr(w, http.StatusUnauthorized, 401, "暂未登录")
				return
			}
			handler(w, r, uid)
		}
	}

	mux.HandleFunc("/api/health/ping", func(w http.ResponseWriter, r *http.Request) {
		writeOK(w, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/auth/weapp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, 500, "服务异常")
			return
		}
		var body struct {
			Code      string `json:"code"`
			NickName  string `json:"nickName"`
			AvatarUrl string `json:"avatarUrl"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		appid := os.Getenv("WECHAT_APPID")
		secret := os.Getenv("WECHAT_SECRET")
		openID := ""
		if appid != "" && secret != "" && body.Code != "" {
			u := "https://api.weixin.qq.com/sns/jscode2session?appid=" + appid + "&secret=" + secret + "&js_code=" + body.Code + "&grant_type=authorization_code"
			resp, err := http.Get(u)
			if err == nil && resp != nil {
				defer resp.Body.Close()
				var sess struct {
					OpenID     string `json:"openid"`
					SessionKey string `json:"session_key"`
					ErrCode    int    `json:"errcode"`
					ErrMsg     string `json:"errmsg"`
				}
				_ = json.NewDecoder(resp.Body).Decode(&sess)
				if sess.OpenID != "" {
					openID = sess.OpenID
				}
			}
		}
		if openID == "" && body.Code != "" {
			if len(body.Code) > 16 {
				openID = "mock_" + body.Code[:16]
			} else {
				openID = "mock_" + body.Code
			}
		}
		var id string
		row := db.QueryRow("SELECT id FROM users WHERE openId=? LIMIT 1", openID)
		_ = row.Scan(&id)
		if id == "" {
			id = newID()
			_, _ = db.Exec("INSERT INTO users(id,nickName,avatarUrl,openId) VALUES(?,?,?,?)", id, body.NickName, body.AvatarUrl, openID)
		} else {
			_, _ = db.Exec("UPDATE users SET nickName=IFNULL(NULLIF(?,''),nickName), avatarUrl=IFNULL(NULLIF(?,''),avatarUrl) WHERE id=?", body.NickName, body.AvatarUrl, id)
		}
		// 生成并保存会话 token
		tok := newID()
		expires := time.Now().UnixMilli() + getTokenTTL()*1000
		_, _ = db.Exec("INSERT INTO sessions(id,userId,token,expiresAt) VALUES(?,?,?,?)", newID(), id, tok, expires)

		w.Header().Set("Authorization", "Bearer "+tok)
		w.Header().Set("X-Token", tok)
		row2 := db.QueryRow("SELECT id,nickName,avatarUrl,openId FROM users WHERE id=?", id)
		var u User
		_ = row2.Scan(&u.ID, &u.NickName, &u.AvatarUrl, &u.OpenID)
		writeOK(w, map[string]any{"user": u, "expiresAt": expires})
	})

	mux.HandleFunc("/api/users/detail/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/users/detail/")
		row := db.QueryRow("SELECT id,nickName,avatarUrl,openId FROM users WHERE id=?", id)
		var u User
		if err := row.Scan(&u.ID, &u.NickName, &u.AvatarUrl, &u.OpenID); err != nil {
			writeErr(w, http.StatusNotFound, 500, "服务异常")
			return
		}
		writeOK(w, map[string]any{"user": u})
	})

	mux.HandleFunc("/api/users/update/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			writeErr(w, http.StatusMethodNotAllowed, 500, "服务异常")
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/users/update/")
		var body struct {
			NickName  string `json:"nickName"`
			AvatarUrl string `json:"avatarUrl"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		res, _ := db.Exec("UPDATE users SET nickName=IFNULL(NULLIF(?,''),nickName), avatarUrl=IFNULL(NULLIF(?,''),avatarUrl) WHERE id=?", body.NickName, body.AvatarUrl, id)
		n, _ := res.RowsAffected()
		if n == 0 {
			writeErr(w, http.StatusNotFound, 500, "服务异常")
			return
		}
		row := db.QueryRow("SELECT id,nickName,avatarUrl,openId FROM users WHERE id=?", id)
		var u User
		_ = row.Scan(&u.ID, &u.NickName, &u.AvatarUrl, &u.OpenID)
		writeOK(w, map[string]any{"user": u})
	})

	mux.HandleFunc("/api/records/list", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
		// userId 来自 token
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")
		var where []string
		var args []any
		if userId != "" {
			where = append(where, "userId=?")
			args = append(args, userId)
		}
		if start != "" {
			where = append(where, "endTime>=?")
			args = append(args, start)
		}
		if end != "" {
			where = append(where, "endTime<?")
			args = append(args, end)
		}
		cond := ""
		if len(where) > 0 {
			cond = " WHERE " + strings.Join(where, " AND ")
		}
		row := db.QueryRow("SELECT COUNT(*) FROM records"+cond, args...)
		var total int
		_ = row.Scan(&total)
		pageNum := int(getQueryInt64(r, "pageNum", int64(getQueryInt64(r, "page", 0))))
		pageSize := int(getQueryInt64(r, "pageSize", int64(getQueryInt64(r, "limit", 0))))
		var rows *sql.Rows
		var err error
		if pageNum > 0 || pageSize > 0 {
			if pageNum < 1 {
				pageNum = 1
			}
			limit := pageSize
			if limit <= 0 {
				limit = 20
			}
			offset := (pageNum - 1) * limit
			argsItems := append(append([]any{}, args...), limit, offset)
			q := "SELECT id,userId,startTime,endTime,duration,color,status,shape,amount,note,isCompleted,createdAt FROM records" + cond + " ORDER BY createdAt DESC LIMIT ? OFFSET ?"
			rows, err = db.Query(q, argsItems...)
		} else {
			q := "SELECT id,userId,startTime,endTime,duration,color,status,shape,amount,note,isCompleted,createdAt FROM records" + cond + " ORDER BY createdAt DESC"
			rows, err = db.Query(q, args...)
		}
		if err != nil {
			writeOK(w, map[string]any{"total": 0, "items": []Record{}})
			return
		}
		defer rows.Close()
		var items []Record
		for rows.Next() {
			var rcd Record
			var ic int
			_ = rows.Scan(&rcd.ID, &rcd.UserID, &rcd.StartTime, &rcd.EndTime, &rcd.Duration, &rcd.Color, &rcd.Status, &rcd.Shape, &rcd.Amount, &rcd.Note, &ic, &rcd.CreatedAt)
			rcd.IsCompleted = ic != 0
			items = append(items, rcd)
		}
		data := map[string]any{"total": total, "items": items}
		if pageNum > 0 || pageSize > 0 {
			data["pageNum"] = pageNum
			data["pageSize"] = func() int {
				if pageSize <= 0 {
					return 20
				} else {
					return pageSize
				}
			}()
		}
		writeOK(w, data)
	}))

	mux.HandleFunc("/api/records/create", withAuth(func(w http.ResponseWriter, r *http.Request, userID string) {
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, 500, "服务异常")
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		now := time.Now().UnixMilli()
		end := now
		if v, ok := body["endTime"].(float64); ok {
			end = int64(v)
		}
		var dur int64
		if v, ok := body["duration"].(float64); ok {
			dur = int64(v)
		} else {
			dur = 300
		}
		start := end - dur*1000
		if v, ok := body["startTime"].(float64); ok {
			start = int64(v)
		}
		id := newID()
		color := toString(body["color"], "brown")
		status := toString(body["status"], "normal")
		shape := toString(body["shape"], "banana")
		amount := toString(body["amount"], "moderate")
		note := toString(body["note"], "")
		_, _ = db.Exec("INSERT INTO records(id,userId,startTime,endTime,duration,color,status,shape,amount,note,isCompleted,createdAt) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)",
			id, userID, start, end, dur, color, status, shape, amount, note, 1, now)
		// 更新排行榜表：当日计数 +1
		day := time.UnixMilli(end).Format("2006-01-02")
		var userName string
		_ = db.QueryRow("SELECT nickName FROM users WHERE id=?", userID).Scan(&userName)
		_, _ = db.Exec("INSERT INTO leaderboard(id,userId,userName,day,count) VALUES(?,?,?,?,1) ON DUPLICATE KEY UPDATE count=count+1", newID(), userID, userName, day)
		row := db.QueryRow("SELECT id,userId,startTime,endTime,duration,color,status,shape,amount,note,isCompleted,createdAt FROM records WHERE id=?", id)
		var rcd Record
		var ic int
		_ = row.Scan(&rcd.ID, &rcd.UserID, &rcd.StartTime, &rcd.EndTime, &rcd.Duration, &rcd.Color, &rcd.Status, &rcd.Shape, &rcd.Amount, &rcd.Note, &ic, &rcd.CreatedAt)
		rcd.IsCompleted = ic != 0
		writeOK(w, map[string]any{"record": rcd})
	}))

	mux.HandleFunc("/api/records/update/", withAuth(func(w http.ResponseWriter, r *http.Request, userID string) {
		if r.Method != http.MethodPut {
			writeErr(w, http.StatusMethodNotAllowed, 500, "服务异常")
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/records/update/")
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		var sets []string
		var args []any
		if v, ok := body["startTime"].(float64); ok {
			sets = append(sets, "startTime=?")
			args = append(args, int64(v))
		}
		if v, ok := body["endTime"].(float64); ok {
			sets = append(sets, "endTime=?")
			args = append(args, int64(v))
		}
		if v, ok := body["duration"].(float64); ok {
			sets = append(sets, "duration=?")
			args = append(args, int64(v))
		}
		if v, ok := body["color"].(string); ok {
			sets = append(sets, "color=?")
			args = append(args, v)
		}
		if v, ok := body["status"].(string); ok {
			sets = append(sets, "status=?")
			args = append(args, v)
		}
		if v, ok := body["shape"].(string); ok {
			sets = append(sets, "shape=?")
			args = append(args, v)
		}
		if v, ok := body["amount"].(string); ok {
			sets = append(sets, "amount=?")
			args = append(args, v)
		}
		if v, ok := body["note"].(string); ok {
			sets = append(sets, "note=?")
			args = append(args, v)
		}
		if v, ok := body["isCompleted"].(bool); ok {
			sets = append(sets, "isCompleted=?")
			if v {
				args = append(args, 1)
			} else {
				args = append(args, 0)
			}
		}
		if len(sets) == 0 {
			writeErr(w, http.StatusBadRequest, 500, "服务异常")
			return
		}
		q := "UPDATE records SET " + strings.Join(sets, ", ") + " WHERE id=? AND userId=?"
		args = append(args, id, userID)
		res, _ := db.Exec(q, args...)
		n, _ := res.RowsAffected()
		if n == 0 {
			writeErr(w, http.StatusNotFound, 500, "服务异常")
			return
		}
		row := db.QueryRow("SELECT id,userId,startTime,endTime,duration,color,status,shape,amount,note,isCompleted,createdAt FROM records WHERE id=?", id)
		var rcd Record
		var ic int
		_ = row.Scan(&rcd.ID, &rcd.UserID, &rcd.StartTime, &rcd.EndTime, &rcd.Duration, &rcd.Color, &rcd.Status, &rcd.Shape, &rcd.Amount, &rcd.Note, &ic, &rcd.CreatedAt)
		rcd.IsCompleted = ic != 0
		writeOK(w, map[string]any{"record": rcd})
	}))

	mux.HandleFunc("/api/records/detail/", withAuth(func(w http.ResponseWriter, r *http.Request, userID string) {
		id := strings.TrimPrefix(r.URL.Path, "/api/records/detail/")
		row := db.QueryRow("SELECT id,userId,startTime,endTime,duration,color,status,shape,amount,note,isCompleted,createdAt FROM records WHERE id=? AND userId=?", id, userID)
		var rcd Record
		var ic int
		if err := row.Scan(&rcd.ID, &rcd.UserID, &rcd.StartTime, &rcd.EndTime, &rcd.Duration, &rcd.Color, &rcd.Status, &rcd.Shape, &rcd.Amount, &rcd.Note, &ic, &rcd.CreatedAt); err != nil {
			writeErr(w, http.StatusNotFound, 500, "服务异常")
			return
		}
		rcd.IsCompleted = ic != 0
		writeOK(w, map[string]any{"record": rcd})
	}))

	mux.HandleFunc("/api/records/delete/", withAuth(func(w http.ResponseWriter, r *http.Request, userID string) {
		if r.Method != http.MethodDelete {
			writeErr(w, http.StatusMethodNotAllowed, 500, "服务异常")
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/records/delete/")
		res, _ := db.Exec("DELETE FROM records WHERE id=? AND userId=?", id, userID)
		n, _ := res.RowsAffected()
		if n == 0 {
			writeErr(w, http.StatusNotFound, 500, "服务异常")
			return
		}
		writeOK(w, map[string]any{"record": map[string]string{"id": id}})
	}))

	mux.HandleFunc("/api/statistics/summary", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")
		var where []string
		var args []any
		if userId != "" {
			where = append(where, "userId=?")
			args = append(args, userId)
		}
		if start != "" {
			where = append(where, "endTime>=?")
			args = append(args, start)
		}
		if end != "" {
			where = append(where, "endTime<?")
			args = append(args, end)
		}
		cond := ""
		if len(where) > 0 {
			cond = " WHERE " + strings.Join(where, " AND ")
		}
		row := db.QueryRow("SELECT COUNT(*), IFNULL(SUM(duration),0), IFNULL(MAX(duration),0) FROM records"+cond, args...)
		var cnt, sum, max int64
		_ = row.Scan(&cnt, &sum, &max)
		var avg int64
		if cnt > 0 {
			avg = int64(math.Floor(float64(sum) / float64(cnt)))
		}
		writeOK(w, map[string]any{"summary": map[string]int64{"totalRecords": cnt, "averageDuration": avg, "longestDuration": max}})
	}))

	// 月度按天统计
	mux.HandleFunc("/api/statistics/month-days", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
		y := int(getQueryInt64(r, "year", int64(time.Now().Year())))
		m := int(getQueryInt64(r, "month", int64(int(time.Now().Month()))))
		if m < 1 {
			m = 1
		}
		if m > 12 {
			m = 12
		}
		start := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.Now().Location())
		end := start.AddDate(0, 1, 0)
		startTs := start.UnixMilli()
		endTs := end.UnixMilli()
		rows, err := db.Query("SELECT DATE(FROM_UNIXTIME(endTime/1000)) AS d, status, COUNT(*) AS cnt FROM records WHERE userId=? AND endTime>=? AND endTime<? GROUP BY d, status ORDER BY d", userId, startTs, endTs)
		if err != nil {
			writeOK(w, map[string]any{"year": y, "month": m, "days": []any{}, "totalDays": 0, "totalRecords": 0})
			return
		}
		defer rows.Close()
		totalRecords := 0
		dayMap := make(map[string]map[string]int)
		for rows.Next() {
			var day string
			var status string
			var c int
			_ = rows.Scan(&day, &status, &c)
			if _, ok := dayMap[day]; !ok {
				dayMap[day] = map[string]int{"normal": 0, "diarrhea": 0, "constipation": 0}
			}
			if status == "normal" || status == "diarrhea" || status == "constipation" {
				dayMap[day][status] += c
			}
			totalRecords += c
		}
		var days = make([]map[string]any, 0)
		for d, counts := range dayMap {
			days = append(days, map[string]any{"date": d, "normal": counts["normal"], "diarrhea": counts["diarrhea"], "constipation": counts["constipation"], "total": counts["normal"] + counts["diarrhea"] + counts["constipation"]})
		}
		writeOK(w, map[string]any{"year": y, "month": m, "days": days, "totalDays": len(dayMap), "totalRecords": totalRecords})
	}))

	mux.HandleFunc("/api/ranking/list", withAuth(func(w http.ResponseWriter, r *http.Request, _ string) {
		period := r.URL.Query().Get("period")
		if period == "" {
			period = "total"
		}
		pageNum := int(getQueryInt64(r, "pageNum", int64(getQueryInt64(r, "page", 0))))
		pageSize := int(getQueryInt64(r, "pageSize", int64(getQueryInt64(r, "limit", 0))))
		if pageNum < 1 {
			pageNum = 1
		}
		if pageSize < 1 {
			pageSize = 20
		}
		now := time.Now()
		var startTs, endTs int64
		if period == "day" {
			s := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			startTs = s.UnixMilli()
			endTs = s.AddDate(0, 0, 1).UnixMilli()
		} else if period == "week" {
			day := (int(now.Weekday()) + 6) % 7
			s := time.Date(now.Year(), now.Month(), now.Day()-day, 0, 0, 0, 0, now.Location())
			startTs = s.UnixMilli()
			endTs = s.AddDate(0, 0, 7).UnixMilli()
		} else if period == "month" {
			s := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			startTs = s.UnixMilli()
			endTs = s.AddDate(0, 1, 0).UnixMilli()
		} else {
			startTs = 0
			// 为了包含当天数据，将结束边界设为明天的日期
			endTs = time.Now().AddDate(0, 0, 1).UnixMilli()
		}
		// 从 leaderboard 聚合排名
		startDay := time.UnixMilli(startTs).Format("2006-01-02")
		endDay := time.UnixMilli(endTs).Format("2006-01-02")
		var total int
		if period == "day" {
			row := db.QueryRow("SELECT COUNT(DISTINCT userId) FROM leaderboard WHERE day=?", startDay)
			_ = row.Scan(&total)
		} else {
			row := db.QueryRow("SELECT COUNT(DISTINCT userId) FROM leaderboard WHERE day>=? AND day<?", startDay, endDay)
			_ = row.Scan(&total)
		}
		offset := (pageNum - 1) * pageSize
		var rows *sql.Rows
		var err error
		if period == "day" {
			rows, err = db.Query("SELECT userId, userName, SUM(count) AS c FROM leaderboard WHERE day=? GROUP BY userId, userName ORDER BY c DESC LIMIT ? OFFSET ?", startDay, pageSize, offset)
		} else {
			rows, err = db.Query("SELECT userId, userName, SUM(count) AS c FROM leaderboard WHERE day>=? AND day<? GROUP BY userId, userName ORDER BY c DESC LIMIT ? OFFSET ?", startDay, endDay, pageSize, offset)
		}
		if err != nil {
			writeOK(w, map[string]any{"list": []any{}, "total": 0, "pageNum": pageNum, "pageSize": pageSize})
			return
		}
		defer rows.Close()
		type Item struct {
			ID         string `json:"id"`
			TotalCount int64  `json:"totalCount"`
		}
		var list []Item
		for rows.Next() {
			var id string
			var name string
			var c int64
			_ = rows.Scan(&id, &name, &c)
			list = append(list, Item{ID: id, TotalCount: c})
		}
		writeOK(w, map[string]any{"list": list, "total": total, "pageNum": pageNum, "pageSize": pageSize})
	}))

	mux.HandleFunc("/api/index/list", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
		// userId 来自 token
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")
		var where []string
		var args []any
		if userId != "" {
			where = append(where, "userId=?")
			args = append(args, userId)
		}
		if start != "" {
			where = append(where, "endTime>=?")
			args = append(args, start)
		}
		if end != "" {
			where = append(where, "endTime<?")
			args = append(args, end)
		}
		cond := ""
		if len(where) > 0 {
			cond = " WHERE " + strings.Join(where, " AND ")
		}
		row := db.QueryRow("SELECT COUNT(*), IFNULL(SUM(duration),0), IFNULL(MAX(duration),0) FROM records"+cond, args...)
		var cnt, sum, max int64
		_ = row.Scan(&cnt, &sum, &max)
		var avg int64
		if cnt > 0 {
			avg = int64(math.Floor(float64(sum) / float64(cnt)))
		}
		pageNum := int(getQueryInt64(r, "pageNum", int64(getQueryInt64(r, "page", 0))))
		pageSize := int(getQueryInt64(r, "pageSize", int64(getQueryInt64(r, "limit", 0))))
		var rows *sql.Rows
		var err error
		if pageNum > 0 || pageSize > 0 {
			if pageNum < 1 {
				pageNum = 1
			}
			limit := pageSize
			if limit <= 0 {
				limit = 10
			}
			offset := (pageNum - 1) * limit
			argsItems := append(append([]any{}, args...), limit, offset)
			q := "SELECT id,userId,startTime,endTime,duration,color,status,shape,amount,note,isCompleted,createdAt FROM records" + cond + " ORDER BY createdAt DESC LIMIT ? OFFSET ?"
			rows, err = db.Query(q, argsItems...)
		} else {
			q := "SELECT id,userId,startTime,endTime,duration,color,status,shape,amount,note,isCompleted,createdAt FROM records" + cond + " ORDER BY createdAt DESC"
			rows, err = db.Query(q, args...)
		}
		if err != nil {
			writeOK(w, map[string]any{"total": 0, "items": []Record{}, "summary": map[string]int64{"totalRecords": 0, "averageDuration": 0, "longestDuration": 0}})
			return
		}
		defer rows.Close()
		var items []Record
		for rows.Next() {
			var rcd Record
			var ic int
			_ = rows.Scan(&rcd.ID, &rcd.UserID, &rcd.StartTime, &rcd.EndTime, &rcd.Duration, &rcd.Color, &rcd.Status, &rcd.Shape, &rcd.Amount, &rcd.Note, &ic, &rcd.CreatedAt)
			rcd.IsCompleted = ic != 0
			items = append(items, rcd)
		}
		data := map[string]any{"total": cnt, "items": items, "summary": map[string]int64{"totalRecords": cnt, "averageDuration": avg, "longestDuration": max}}
		if pageNum > 0 || pageSize > 0 {
			data["pageNum"] = pageNum
			data["pageSize"] = pageSize
		}
		writeOK(w, data)
	}))

	mux.HandleFunc("/api/friends/invite", withAuth(func(w http.ResponseWriter, r *http.Request, userID string) {
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, 500, "服务异常")
			return
		}
		id := newID()
		now := time.Now().UnixMilli()
		_, _ = db.Exec("INSERT INTO friend_invites(id,inviterUserId,createdAt) VALUES(?,?,?)", id, userID, now)
		writeOK(w, id)
	}))

	mux.HandleFunc("/api/friends/accept", withAuth(func(w http.ResponseWriter, r *http.Request, userID string) {
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, 500, "服务异常")
			return
		}
		var body struct {
			InviteID string `json:"inviteId"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.InviteID == "" {
			writeErr(w, http.StatusBadRequest, 500, "服务异常")
			return
		}
		var inviter string
		err := db.QueryRow("SELECT inviterUserId FROM friend_invites WHERE id=?", body.InviteID).Scan(&inviter)
		if err != nil || inviter == "" {
			writeErr(w, http.StatusNotFound, 500, "服务异常")
			return
		}
		if inviter == userID {
			writeErr(w, http.StatusBadRequest, 500, "服务异常")
			return
		}
		now := time.Now().UnixMilli()
		_, _ = db.Exec("INSERT IGNORE INTO friend_relations(id,inviterUserId,inviteeUserId,createdAt) VALUES(?,?,?,?)", newID(), inviter, userID, now)
		writeOK(w, map[string]any{"inviterUserId": inviter, "inviteeUserId": userID})
	}))

	// 个人周期概览：默认本周，可拓展当天/当月/当年
	mux.HandleFunc("/api/overview/personal", func(w http.ResponseWriter, r *http.Request) {
		userId := r.URL.Query().Get("userId")
		if userId == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_userId"})
			return
		}
		period := r.URL.Query().Get("period") // week(默认) / day / month / year
		if period == "" {
			period = "week"
		}
		today := time.Now()
		var startT, endT int64
		switch period {
		case "day":
			startT = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location()).UnixMilli()
			endT = startT + 24*60*60*1000
		case "week":
			dow := int(today.Weekday())
			if dow == 0 {
				dow = 7
			}
			startT = time.Date(today.Year(), today.Month(), today.Day()-(dow-1), 0, 0, 0, 0, today.Location()).UnixMilli()
			endT = startT + 7*24*60*60*1000
		case "month":
			startT = time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location()).UnixMilli()
			endT = time.Date(today.Year(), today.Month()+1, 1, 0, 0, 0, 0, today.Location()).UnixMilli()
		case "year":
			startT = time.Date(today.Year(), 1, 1, 0, 0, 0, 0, today.Location()).UnixMilli()
			endT = time.Date(today.Year()+1, 1, 1, 0, 0, 0, 0, today.Location()).UnixMilli()
		default:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_period"})
			return
		}

		// 1) 打卡天数 = 有记录的日期数
		rows, err := db.Query("SELECT DISTINCT DATE(FROM_UNIXTIME(endTime/1000)) AS d FROM records WHERE userId=? AND endTime>=? AND endTime<?", userId, startT, endT)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{"打卡天数": 0, "排便次数": 0, "色谱": map[string]int{}, "状态评分": 0})
			return
		}
		defer rows.Close()
		days := 0
		for rows.Next() {
			days++
		}

		// 2) 排便次数 & 色谱统计
		rows2, err := db.Query("SELECT color, COUNT(*) AS cnt FROM records WHERE userId=? AND endTime>=? AND endTime<? GROUP BY color", userId, startT, endT)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{"打卡天数": days, "排便次数": 0, "色谱": map[string]int{}, "状态评分": 0})
			return
		}
		defer rows2.Close()
		colorMap := make(map[string]int)
		totalCount := 0
		for rows2.Next() {
			var c string
			var n int
			_ = rows2.Scan(&c, &n)
			colorMap[c] = n
			totalCount += n
		}

		// 3) 状态评分：简单模型——正常次数占比*100
		rows3, err := db.Query("SELECT status, COUNT(*) AS cnt FROM records WHERE userId=? AND endTime>=? AND endTime<? GROUP BY status", userId, startT, endT)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{"打卡天数": days, "排便次数": totalCount, "色谱": colorMap, "状态评分": 0})
			return
		}
		defer rows3.Close()
		normal := 0
		for rows3.Next() {
			var s string
			var n int
			_ = rows3.Scan(&s, &n)
			if s == "normal" {
				normal = n
			}
		}
		score := 0
		if totalCount > 0 {
			score = int(float64(normal) / float64(totalCount) * 100)
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"打卡天数": days,
			"排便次数": totalCount,
			"色谱":   colorMap,
			"状态评分": score,
		})
	})

	h := cors(mux)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	log.Fatal(http.ListenAndServe(":"+port, h))
}
