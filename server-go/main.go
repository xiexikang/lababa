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
	CatID       string `json:"catId"`
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

type CatWeightItem struct {
	ID       string  `json:"id"`
	CatID    string  `json:"catId"`
	UserID   string  `json:"userId"`
	WeightKg float64 `json:"weightKg"`
	Date     int64   `json:"date"`
	Note     string  `json:"note"`
}

type Reminder struct {
	CatID   string `json:"catId"`
	CatName string `json:"catName"`
	Type    string `json:"type"`
	Message string `json:"message"`
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
            catId VARCHAR(64),
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
            INDEX idx_records_user_end (userId, endTime),
            INDEX idx_records_user_cat_end (userId, catId, endTime)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
    `)
    if err != nil {
        log.Fatal("create tables:", err)
    }

    ensureRecordsSchema(db)

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

	// 猫咪表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cats (
			id VARCHAR(64) PRIMARY KEY,
			userId VARCHAR(64) NOT NULL,
			name VARCHAR(255) NOT NULL,
			breedId VARCHAR(64),
			avatarUrl VARCHAR(512),
			gender VARCHAR(16),
			birthDate BIGINT,
			weightKg DECIMAL(5,2),
			neutered TINYINT(1),
			notes TEXT,
			createdAt BIGINT,
			INDEX idx_cats_user (userId)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		log.Fatal("create cats table:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cat_weights (
			id VARCHAR(64) PRIMARY KEY,
			catId VARCHAR(64) NOT NULL,
			userId VARCHAR(64) NOT NULL,
			weightKg DECIMAL(5,2) NOT NULL,
			date BIGINT NOT NULL,
			note TEXT,
			INDEX idx_weights_cat_date (catId, date)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		log.Fatal("create cat_weights table:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cat_settings (
			catId VARCHAR(64) PRIMARY KEY,
			userId VARCHAR(64) NOT NULL,
			remindEnabled TINYINT(1) DEFAULT 1,
			remindNoRecord TINYINT(1) DEFAULT 1,
			remindDiarrhea TINYINT(1) DEFAULT 1,
			quietStart INT DEFAULT 0,
			quietEnd INT DEFAULT 0
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		log.Fatal("create cat_settings table:", err)
	}

    return db
}

func ensureRecordsSchema(db *sql.DB) {
    // detect and add missing columns
    cols := map[string]bool{}
    var dbname string
    _ = db.QueryRow("SELECT DATABASE()").Scan(&dbname)
    rows, err := db.Query("SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=? AND TABLE_NAME='records'", dbname)
    if err == nil {
        defer rows.Close()
        for rows.Next() {
            var c string
            _ = rows.Scan(&c)
            cols[c] = true
        }
    }
    addIfMissing := func(column string, ddl string) {
        if !cols[column] {
            _, _ = db.Exec("ALTER TABLE records ADD COLUMN " + ddl)
        }
    }
    addIfMissing("userId", "userId VARCHAR(64)")
    addIfMissing("catId", "catId VARCHAR(64)")
    addIfMissing("startTime", "startTime BIGINT")
    addIfMissing("endTime", "endTime BIGINT")
    addIfMissing("duration", "duration BIGINT")
    addIfMissing("color", "color VARCHAR(32)")
    addIfMissing("status", "status VARCHAR(32)")
    addIfMissing("shape", "shape VARCHAR(32)")
    addIfMissing("amount", "amount VARCHAR(32)")
    addIfMissing("note", "note TEXT")
    addIfMissing("isCompleted", "isCompleted TINYINT(1)")
    addIfMissing("createdAt", "createdAt BIGINT")
    // ensure indexes
    _, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_records_user_end ON records(userId, endTime)")
    _, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_records_user_cat_end ON records(userId, catId, endTime)")
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

func readBody(r *http.Request) map[string]any {
    var m map[string]any
    _ = json.NewDecoder(r.Body).Decode(&m)
    if m == nil {
        m = map[string]any{}
    }
    return m
}

func getBodyString(b map[string]any, k string, def string) string {
    v, ok := b[k]
    if !ok || v == nil {
        return def
    }
    if s, ok := v.(string); ok {
        return s
    }
    return def
}

func getBodyInt64(b map[string]any, k string, def int64) int64 {
    v, ok := b[k]
    if !ok || v == nil {
        return def
    }
    switch t := v.(type) {
    case float64:
        return int64(t)
    case int64:
        return t
    case int:
        return int64(t)
    case string:
        i, err := strconv.ParseInt(t, 10, 64)
        if err != nil {
            return def
        }
        return i
    default:
        return def
    }
}

func getBodyInt(b map[string]any, k string, def int) int {
    return int(getBodyInt64(b, k, int64(def)))
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

	mux.HandleFunc("/api/users/update", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
		if r.Method != http.MethodPut {
			writeErr(w, http.StatusMethodNotAllowed, 500, "服务异常")
			return
		}
		var body struct {
			NickName  string `json:"nickName"`
			AvatarUrl string `json:"avatarUrl"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		res, _ := db.Exec("UPDATE users SET nickName=IFNULL(NULLIF(?,''),nickName), avatarUrl=IFNULL(NULLIF(?,''),avatarUrl) WHERE id=?", body.NickName, body.AvatarUrl, userId)
		n, _ := res.RowsAffected()
		if n == 0 {
			writeErr(w, http.StatusNotFound, 500, "服务异常")
			return
		}
		row := db.QueryRow("SELECT id,nickName,avatarUrl,openId FROM users WHERE id= ?", userId)
		var u User
		_ = row.Scan(&u.ID, &u.NickName, &u.AvatarUrl, &u.OpenID)
		writeOK(w, map[string]any{"user": u})
	}))

    mux.HandleFunc("/api/records/list", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
        body := readBody(r)
        start := getBodyString(body, "start", r.URL.Query().Get("start"))
        end := getBodyString(body, "end", r.URL.Query().Get("end"))
		var where []string
		var args []any
		if userId != "" {
			where = append(where, "userId=?")
			args = append(args, userId)
		}
		// 可选猫咪筛选
        catId := getBodyString(body, "catId", r.URL.Query().Get("catId"))
        if catId == "" { catId = getBodyString(body, "id", r.URL.Query().Get("id")) }
        if catId == "" {
            catId = getBodyString(body, "id", r.URL.Query().Get("id"))
        }
		if catId != "" {
			where = append(where, "catId=?")
			args = append(args, catId)
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
        pnAlias := getBodyInt(body, "paNum", -1)
        pageNum := getBodyInt(body, "pageNum", int(getQueryInt64(r, "pageNum", int64(getQueryInt64(r, "page", 0)))))
        if pnAlias > 0 { pageNum = pnAlias }
        pageSize := getBodyInt(body, "pageSize", int(getQueryInt64(r, "pageSize", int64(getQueryInt64(r, "limit", 0)))))
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
			q := "SELECT id,userId,catId,startTime,endTime,duration,color,status,shape,amount,note,isCompleted,createdAt FROM records" + cond + " ORDER BY createdAt DESC LIMIT ? OFFSET ?"
			rows, err = db.Query(q, argsItems...)
		} else {
			q := "SELECT id,userId,catId,startTime,endTime,duration,color,status,shape,amount,note,isCompleted,createdAt FROM records" + cond + " ORDER BY createdAt DESC"
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
			_ = rows.Scan(&rcd.ID, &rcd.UserID, &rcd.CatID, &rcd.StartTime, &rcd.EndTime, &rcd.Duration, &rcd.Color, &rcd.Status, &rcd.Shape, &rcd.Amount, &rcd.Note, &ic, &rcd.CreatedAt)
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
        catId := toString(body["catId"], "")
        if strings.TrimSpace(catId) == "" {
            writeErr(w, http.StatusBadRequest, 400, "缺少猫咪ID")
            return
        }
        var owner string
        _ = db.QueryRow("SELECT userId FROM cats WHERE id=?", catId).Scan(&owner)
        if owner == "" {
            writeErr(w, http.StatusNotFound, 404, "猫不存在")
            return
        }
        if owner != userID {
            writeErr(w, http.StatusForbidden, 403, "猫咪不属于当前用户")
            return
        }
        res, err := db.Exec("INSERT INTO records(id,userId,catId,startTime,endTime,duration,color,status,shape,amount,note,isCompleted,createdAt) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)",
            id, userID, catId, start, end, dur, color, status, shape, amount, note, 1, now)
        if err != nil {
            writeErr(w, http.StatusInternalServerError, 500, "服务异常")
            return
        }
        if n, _ := res.RowsAffected(); n == 0 {
            writeErr(w, http.StatusInternalServerError, 500, "服务异常")
            return
        }
		// 更新排行榜表：当日计数 +1
		day := time.UnixMilli(end).Format("2006-01-02")
		var userName string
		_ = db.QueryRow("SELECT nickName FROM users WHERE id=?", userID).Scan(&userName)
		_, _ = db.Exec("INSERT INTO leaderboard(id,userId,userName,day,count) VALUES(?,?,?,?,1) ON DUPLICATE KEY UPDATE count=count+1", newID(), userID, userName, day)
		rcd := Record{ID: id, UserID: userID, CatID: catId, StartTime: start, EndTime: end, Duration: dur, Color: color, Status: status, Shape: shape, Amount: amount, Note: note, IsCompleted: true, CreatedAt: now}
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
		if v, ok := body["catId"].(string); ok {
			sets = append(sets, "catId=?")
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
		row := db.QueryRow("SELECT id,userId,catId,startTime,endTime,duration,color,status,shape,amount,note,isCompleted,createdAt FROM records WHERE id=?", id)
		var rcd Record
		var ic int
		_ = row.Scan(&rcd.ID, &rcd.UserID, &rcd.CatID, &rcd.StartTime, &rcd.EndTime, &rcd.Duration, &rcd.Color, &rcd.Status, &rcd.Shape, &rcd.Amount, &rcd.Note, &ic, &rcd.CreatedAt)
		rcd.IsCompleted = ic != 0
		writeOK(w, map[string]any{"record": rcd})
	}))

	mux.HandleFunc("/api/records/detail/", withAuth(func(w http.ResponseWriter, r *http.Request, userID string) {
		id := strings.TrimPrefix(r.URL.Path, "/api/records/detail/")
		row := db.QueryRow("SELECT id,userId,catId,startTime,endTime,duration,color,status,shape,amount,note,isCompleted,createdAt FROM records WHERE id=? AND userId=?", id, userID)
		var rcd Record
		var ic int
		if err := row.Scan(&rcd.ID, &rcd.UserID, &rcd.CatID, &rcd.StartTime, &rcd.EndTime, &rcd.Duration, &rcd.Color, &rcd.Status, &rcd.Shape, &rcd.Amount, &rcd.Note, &ic, &rcd.CreatedAt); err != nil {
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
        body := readBody(r)
        start := getBodyString(body, "start", r.URL.Query().Get("start"))
        end := getBodyString(body, "end", r.URL.Query().Get("end"))
		var where []string
		var args []any
		if userId != "" {
			where = append(where, "userId=?")
			args = append(args, userId)
		}
        catId := getBodyString(body, "catId", r.URL.Query().Get("catId"))
		if catId != "" {
			where = append(where, "catId=?")
			args = append(args, catId)
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
        body := readBody(r)
        y := getBodyInt(body, "year", int(getQueryInt64(r, "year", int64(time.Now().Year()))))
        m := getBodyInt(body, "month", int(getQueryInt64(r, "month", int64(int(time.Now().Month())))))
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
        catId := getBodyString(body, "catId", r.URL.Query().Get("catId"))
        if catId == "" { catId = getBodyString(body, "id", r.URL.Query().Get("id")) }
		var rows *sql.Rows
		var err error
		if catId != "" {
			rows, err = db.Query("SELECT DATE(FROM_UNIXTIME(endTime/1000)) AS d, status, COUNT(*) AS cnt FROM records WHERE userId=? AND catId=? AND endTime>=? AND endTime<? GROUP BY d, status ORDER BY d", userId, catId, startTs, endTs)
		} else {
			rows, err = db.Query("SELECT DATE(FROM_UNIXTIME(endTime/1000)) AS d, status, COUNT(*) AS cnt FROM records WHERE userId=? AND endTime>=? AND endTime<? GROUP BY d, status ORDER BY d", userId, startTs, endTs)
		}
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
        body := readBody(r)
        period := getBodyString(body, "period", r.URL.Query().Get("period"))
        if period == "" {
            period = "total"
        }
        pageNum := getBodyInt(body, "pageNum", int(getQueryInt64(r, "pageNum", int64(getQueryInt64(r, "page", 0)))))
        pageSize := getBodyInt(body, "pageSize", int(getQueryInt64(r, "pageSize", int64(getQueryInt64(r, "limit", 0)))))
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
        ID            string `json:"id"`
        Nickname      string `json:"nickname"`
        CatName       string `json:"catName"`
        TotalCount    int64  `json:"totalCount"`
        TotalDuration int64  `json:"totalDuration"`
    }
    var list []Item
    for rows.Next() {
        var id string
        var name string
        var c int64
        _ = rows.Scan(&id, &name, &c)
        var nick string
        _ = db.QueryRow("SELECT nickName FROM users WHERE id=?", id).Scan(&nick)
        var catName string
        if startTs >= 0 && endTs > 0 {
            var cn sql.NullString
            _ = db.QueryRow("SELECT c.name FROM records r LEFT JOIN cats c ON c.id=r.catId WHERE r.userId=? AND r.endTime>=? AND r.endTime<? GROUP BY r.catId, c.name ORDER BY COUNT(*) DESC LIMIT 1", id, startTs, endTs).Scan(&cn)
            if cn.Valid {
                catName = cn.String
            }
        }
        var sumDur int64
        _ = db.QueryRow("SELECT IFNULL(SUM(duration),0) FROM records WHERE userId=? AND endTime>=? AND endTime<?", id, startTs, endTs).Scan(&sumDur)
        list = append(list, Item{ID: id, Nickname: nick, CatName: catName, TotalCount: c, TotalDuration: sumDur})
    }
    writeOK(w, map[string]any{"list": list, "total": total, "pageNum": pageNum, "pageSize": pageSize})
    }))


    mux.HandleFunc("/api/cats/list", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
        body := readBody(r)
        q := strings.TrimSpace(getBodyString(body, "q", r.URL.Query().Get("q")))
        pageNum := getBodyInt(body, "pageNum", int(getQueryInt64(r, "pageNum", int64(getQueryInt64(r, "page", 0)))))
        pageSize := getBodyInt(body, "pageSize", int(getQueryInt64(r, "pageSize", int64(getQueryInt64(r, "limit", 0)))))
		cond := " WHERE userId=?"
		args := []any{userId}
		if q != "" {
			cond += " AND name LIKE ?"
			args = append(args, "%"+q+"%")
		}
		row := db.QueryRow("SELECT COUNT(*) FROM cats"+cond, args...)
		var total int
		_ = row.Scan(&total)
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
			args2 := append(append([]any{}, args...), limit, offset)
			rows, err = db.Query("SELECT id,userId,name,breedId,avatarUrl,gender,birthDate,weightKg,neutered,notes,createdAt FROM cats"+cond+" ORDER BY createdAt DESC LIMIT ? OFFSET ?", args2...)
		} else {
			rows, err = db.Query("SELECT id,userId,name,breedId,avatarUrl,gender,birthDate,weightKg,neutered,notes,createdAt FROM cats"+cond+" ORDER BY createdAt DESC", args...)
		}
		if err != nil {
			writeOK(w, map[string]any{"total": 0, "items": []any{}})
			return
		}
		defer rows.Close()
		type Cat struct {
			ID        string  `json:"id"`
			UserID    string  `json:"userId"`
			Name      string  `json:"name"`
			BreedID   string  `json:"breedId"`
			AvatarUrl string  `json:"avatarUrl"`
			Gender    string  `json:"gender"`
			BirthDate int64   `json:"birthDate"`
			WeightKg  float64 `json:"weightKg"`
			Neutered  bool    `json:"neutered"`
			Notes     string  `json:"notes"`
			CreatedAt int64   `json:"createdAt"`
		}
		var items []Cat
		for rows.Next() {
			var c Cat
			var neuterInt int
			var w sql.NullFloat64
			_ = rows.Scan(&c.ID, &c.UserID, &c.Name, &c.BreedID, &c.AvatarUrl, &c.Gender, &c.BirthDate, &w, &neuterInt, &c.Notes, &c.CreatedAt)
			c.Neutered = neuterInt != 0
			if w.Valid {
				c.WeightKg = w.Float64
			} else {
				c.WeightKg = 0
			}
			items = append(items, c)
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

	mux.HandleFunc("/api/cats/create", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, 500, "服务异常")
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		name := toString(body["name"], "")
		if strings.TrimSpace(name) == "" {
			writeErr(w, http.StatusBadRequest, 500, "服务异常")
			return
		}
		breedId := toString(body["breedId"], "")
		avatarUrl := toString(body["avatarUrl"], "")
		gender := toString(body["gender"], "")
		birthDate := int64(0)
		if v, ok := body["birthDate"].(float64); ok {
			birthDate = int64(v)
		}
		weight := 0.0
		if v, ok := body["weightKg"].(float64); ok {
			weight = v
		}
		neutered := 0
		if v, ok := body["neutered"].(bool); ok {
			if v {
				neutered = 1
			}
		}
		notes := toString(body["notes"], "")
		id := newID()
		now := time.Now().UnixMilli()
		_, _ = db.Exec("INSERT INTO cats(id,userId,name,breedId,avatarUrl,gender,birthDate,weightKg,neutered,notes,createdAt) VALUES(?,?,?,?,?,?,?,?,?,?,?)",
			id, userId, name, breedId, avatarUrl, gender, birthDate, weight, neutered, notes, now)
		row := db.QueryRow("SELECT id,userId,name,breedId,avatarUrl,gender,birthDate,weightKg,neutered,notes,createdAt FROM cats WHERE id=?", id)
		var cID, uID, nm, bID, av, gd, nts string
		var bd, cr int64
		var wVal sql.NullFloat64
		var nt int
		_ = row.Scan(&cID, &uID, &nm, &bID, &av, &gd, &bd, &wVal, &nt, &nts, &cr)
		writeOK(w, map[string]any{"cat": map[string]any{
			"id": cID, "userId": uID, "name": nm, "breedId": bID, "avatarUrl": av, "gender": gd, "birthDate": bd, "weightKg": func() float64 {
				if wVal.Valid {
					return wVal.Float64
				} else {
					return 0
				}
			}(), "neutered": nt != 0, "notes": nts, "createdAt": cr,
		}})
	}))



	mux.HandleFunc("/api/cats/detail/", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
		id := strings.TrimPrefix(r.URL.Path, "/api/cats/detail/")
		row := db.QueryRow("SELECT id,userId,name,breedId,avatarUrl,gender,birthDate,weightKg,neutered,notes,createdAt FROM cats WHERE id=? AND userId=?", id, userId)
		var cID, uID, nm, bID, av, gd, nts string
		var bd, cr int64
		var wVal sql.NullFloat64
		var nt int
		if err := row.Scan(&cID, &uID, &nm, &bID, &av, &gd, &bd, &wVal, &nt, &nts, &cr); err != nil {
			writeErr(w, http.StatusNotFound, 500, "服务异常")
			return
		}
		writeOK(w, map[string]any{"cat": map[string]any{
			"id": cID, "userId": uID, "name": nm, "breedId": bID, "avatarUrl": av, "gender": gd, "birthDate": bd, "weightKg": func() float64 {
				if wVal.Valid { return wVal.Float64 } else { return 0 }
			}(), "neutered": nt != 0, "notes": nts, "createdAt": cr,
		}})
	}))

	mux.HandleFunc("/api/cats/update", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, 500, "服务异常")
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		id := toString(body["id"], "")
		if strings.TrimSpace(id) == "" {
			writeErr(w, http.StatusBadRequest, 500, "服务异常")
			return
		}
		var owner string
		_ = db.QueryRow("SELECT userId FROM cats WHERE id=?", id).Scan(&owner)
		if owner == "" || owner != userId {
			writeErr(w, http.StatusForbidden, 500, "服务异常")
			return
		}
		var sets []string
		var args []any
		if v, ok := body["name"].(string); ok { sets = append(sets, "name=?"); args = append(args, v) }
		if v, ok := body["breedId"].(string); ok { sets = append(sets, "breedId=?"); args = append(args, v) }
		if v, ok := body["avatarUrl"].(string); ok { sets = append(sets, "avatarUrl=?"); args = append(args, v) }
		if v, ok := body["gender"].(string); ok { sets = append(sets, "gender=?"); args = append(args, v) }
		if v, ok := body["birthDate"].(float64); ok { sets = append(sets, "birthDate=?"); args = append(args, int64(v)) }
		if v, ok := body["weightKg"].(float64); ok { sets = append(sets, "weightKg=?"); args = append(args, v) }
		if v, ok := body["neutered"].(bool); ok { sets = append(sets, "neutered=?"); if v { args = append(args, 1) } else { args = append(args, 0) } }
		if v, ok := body["notes"].(string); ok { sets = append(sets, "notes=?"); args = append(args, v) }
		if len(sets) == 0 {
			writeErr(w, http.StatusBadRequest, 500, "服务异常")
			return
		}
		q := "UPDATE cats SET " + strings.Join(sets, ", ") + " WHERE id=? AND userId=?"
		args = append(args, id, userId)
		res, _ := db.Exec(q, args...)
		n, _ := res.RowsAffected()
		if n == 0 {
			writeErr(w, http.StatusNotFound, 500, "服务异常")
			return
		}
		row := db.QueryRow("SELECT id,userId,name,breedId,avatarUrl,gender,birthDate,weightKg,neutered,notes,createdAt FROM cats WHERE id=?", id)
		var cID, uID, nm, bID, av, gd, nts string
		var bd, cr int64
		var wVal sql.NullFloat64
		var nt int
		_ = row.Scan(&cID, &uID, &nm, &bID, &av, &gd, &bd, &wVal, &nt, &nts, &cr)
		writeOK(w, map[string]any{"cat": map[string]any{
			"id": cID, "userId": uID, "name": nm, "breedId": bID, "avatarUrl": av, "gender": gd, "birthDate": bd, "weightKg": func() float64 { if wVal.Valid { return wVal.Float64 } else { return 0 } }(), "neutered": nt != 0, "notes": nts, "createdAt": cr,
		}})
	}))

	mux.HandleFunc("/api/cats/delete/", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
		if r.Method != http.MethodDelete {
			writeErr(w, http.StatusMethodNotAllowed, 500, "服务异常")
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/cats/delete/")
		res, _ := db.Exec("DELETE FROM cats WHERE id=? AND userId=?", id, userId)
		n, _ := res.RowsAffected()
		if n == 0 {
			writeErr(w, http.StatusNotFound, 500, "服务异常")
			return
		}
		writeOK(w, map[string]any{"cat": map[string]string{"id": id}})
	}))

    mux.HandleFunc("/api/index/list", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
        body := readBody(r)
        start := getBodyString(body, "start", r.URL.Query().Get("start"))
        end := getBodyString(body, "end", r.URL.Query().Get("end"))
		var where []string
		var args []any
		if userId != "" {
			where = append(where, "userId=?")
			args = append(args, userId)
		}
        catId := getBodyString(body, "catId", r.URL.Query().Get("catId"))
		if catId != "" {
			where = append(where, "catId=?")
			args = append(args, catId)
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
        pageNum := getBodyInt(body, "pageNum", int(getQueryInt64(r, "pageNum", int64(getQueryInt64(r, "page", 0)))))
        pageSize := getBodyInt(body, "pageSize", int(getQueryInt64(r, "pageSize", int64(getQueryInt64(r, "limit", 0)))))
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
			q := "SELECT id,userId,catId,startTime,endTime,duration,color,status,shape,amount,note,isCompleted,createdAt FROM records" + cond + " ORDER BY createdAt DESC LIMIT ? OFFSET ?"
			rows, err = db.Query(q, argsItems...)
		} else {
			q := "SELECT id,userId,catId,startTime,endTime,duration,color,status,shape,amount,note,isCompleted,createdAt FROM records" + cond + " ORDER BY createdAt DESC"
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
			_ = rows.Scan(&rcd.ID, &rcd.UserID, &rcd.CatID, &rcd.StartTime, &rcd.EndTime, &rcd.Duration, &rcd.Color, &rcd.Status, &rcd.Shape, &rcd.Amount, &rcd.Note, &ic, &rcd.CreatedAt)
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

    mux.HandleFunc("/api/cats/weights/list", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
        body := readBody(r)
        catId := getBodyString(body, "catId", r.URL.Query().Get("catId"))
        if catId == "" {
            catId = getBodyString(body, "id", r.URL.Query().Get("id"))
        }
        pageNum := getBodyInt(body, "paNum", getBodyInt(body, "pageNum", 0))
        pageSize := getBodyInt(body, "pageSize", 20)
		if catId == "" {
			writeErr(w, http.StatusBadRequest, 500, "服务异常")
			return
		}
		var owner string
		_ = db.QueryRow("SELECT userId FROM cats WHERE id=?", catId).Scan(&owner)
		if owner == "" || owner != userId {
			writeErr(w, http.StatusForbidden, 500, "服务异常")
			return
		}
        var rows *sql.Rows
        var err error
        if pageNum > 0 {
            if pageNum < 1 { pageNum = 1 }
            if pageSize <= 0 { pageSize = 20 }
            offset := (pageNum - 1) * pageSize
            rows, err = db.Query("SELECT id, catId, userId, weightKg, date, note FROM cat_weights WHERE catId=? ORDER BY date DESC LIMIT ? OFFSET ?", catId, pageSize, offset)
        } else {
            rows, err = db.Query("SELECT id, catId, userId, weightKg, date, note FROM cat_weights WHERE catId=? ORDER BY date DESC", catId)
        }
		if err != nil {
			writeOK(w, map[string]any{"items": []any{}})
			return
		}
		defer rows.Close()
		var list []CatWeightItem
		for rows.Next() {
			var it CatWeightItem
			var wv sql.NullFloat64
			_ = rows.Scan(&it.ID, &it.CatID, &it.UserID, &wv, &it.Date, &it.Note)
			if wv.Valid {
				it.WeightKg = wv.Float64
			}
			list = append(list, it)
		}
        data := map[string]any{"items": list}
        if pageNum > 0 { data["pageNum"] = pageNum; data["pageSize"] = pageSize }
        writeOK(w, data)
    }))

	mux.HandleFunc("/api/cats/weights/create", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, 500, "服务异常")
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		catId := toString(body["catId"], "")
		if catId == "" {
			writeErr(w, http.StatusBadRequest, 500, "服务异常")
			return
		}
		var owner string
		_ = db.QueryRow("SELECT userId FROM cats WHERE id=?", catId).Scan(&owner)
		if owner == "" || owner != userId {
			writeErr(w, http.StatusForbidden, 500, "服务异常")
			return
		}
		var wkg float64
		if v, ok := body["weightKg"].(float64); ok {
			wkg = v
		} else {
			writeErr(w, http.StatusBadRequest, 500, "服务异常")
			return
		}
		var dt int64
		if v, ok := body["date"].(float64); ok {
			dt = int64(v)
		} else {
			dt = time.Now().UnixMilli()
		}
		note := toString(body["note"], "")
		id := newID()
		_, _ = db.Exec("INSERT INTO cat_weights(id,catId,userId,weightKg,date,note) VALUES(?,?,?,?,?,?)", id, catId, userId, wkg, dt, note)
		row := db.QueryRow("SELECT id, catId, userId, weightKg, date, note FROM cat_weights WHERE id=?", id)
		var outId, outCat, outUser, outNote string
		var outDate int64
		var outW sql.NullFloat64
		_ = row.Scan(&outId, &outCat, &outUser, &outW, &outDate, &outNote)
		writeOK(w, map[string]any{"item": map[string]any{"id": outId, "catId": outCat, "userId": outUser, "weightKg": func() float64 {
			if outW.Valid {
				return outW.Float64
			} else {
				return 0
			}
		}(), "date": outDate, "note": outNote}})
	}))

    mux.HandleFunc("/api/cats/settings/get", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
        body := readBody(r)
        catId := getBodyString(body, "catId", r.URL.Query().Get("catId"))
        if catId == "" { catId = getBodyString(body, "id", r.URL.Query().Get("id")) }
		if catId == "" {
			writeErr(w, http.StatusBadRequest, 500, "服务异常")
			return
		}
		var owner string
		_ = db.QueryRow("SELECT userId FROM cats WHERE id=?", catId).Scan(&owner)
		if owner == "" || owner != userId {
			writeErr(w, http.StatusForbidden, 500, "服务异常")
			return
		}
		row := db.QueryRow("SELECT remindEnabled, remindNoRecord, remindDiarrhea, quietStart, quietEnd FROM cat_settings WHERE catId=? AND userId=?", catId, userId)
		var re, rn, rd int
		var qs, qe int
		err := row.Scan(&re, &rn, &rd, &qs, &qe)
		if err != nil {
			writeOK(w, map[string]any{"settings": map[string]any{"remindEnabled": 1, "remindNoRecord": 1, "remindDiarrhea": 1, "quietStart": 0, "quietEnd": 0}})
			return
		}
		writeOK(w, map[string]any{"settings": map[string]any{"remindEnabled": re, "remindNoRecord": rn, "remindDiarrhea": rd, "quietStart": qs, "quietEnd": qe}})
	}))

	mux.HandleFunc("/api/cats/settings/update/", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
		if r.Method != http.MethodPut {
			writeErr(w, http.StatusMethodNotAllowed, 500, "服务异常")
			return
		}
		catId := strings.TrimPrefix(r.URL.Path, "/api/cats/settings/update/")
		if catId == "" {
			writeErr(w, http.StatusBadRequest, 500, "服务异常")
			return
		}
		var owner string
		_ = db.QueryRow("SELECT userId FROM cats WHERE id=?", catId).Scan(&owner)
		if owner == "" || owner != userId {
			writeErr(w, http.StatusForbidden, 500, "服务异常")
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		re := 1
		rn := 1
		rd := 1
		qs := 0
		qe := 0
		if v, ok := body["remindEnabled"].(bool); ok {
			if v {
				re = 1
			} else {
				re = 0
			}
		}
		if v, ok := body["remindNoRecord"].(bool); ok {
			if v {
				rn = 1
			} else {
				rn = 0
			}
		}
		if v, ok := body["remindDiarrhea"].(bool); ok {
			if v {
				rd = 1
			} else {
				rd = 0
			}
		}
		if v, ok := body["quietStart"].(float64); ok {
			qs = int(v)
		}
		if v, ok := body["quietEnd"].(float64); ok {
			qe = int(v)
		}
		_, _ = db.Exec("INSERT INTO cat_settings(catId,userId,remindEnabled,remindNoRecord,remindDiarrhea,quietStart,quietEnd) VALUES(?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE remindEnabled=VALUES(remindEnabled), remindNoRecord=VALUES(remindNoRecord), remindDiarrhea=VALUES(remindDiarrhea), quietStart=VALUES(quietStart), quietEnd=VALUES(quietEnd)", catId, userId, re, rn, rd, qs, qe)
		writeOK(w, map[string]any{"settings": map[string]any{"remindEnabled": re, "remindNoRecord": rn, "remindDiarrhea": rd, "quietStart": qs, "quietEnd": qe}})
	}))

	mux.HandleFunc("/api/reminders/list", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
		rows, err := db.Query("SELECT id, name FROM cats WHERE userId=?", userId)
		if err != nil {
			writeOK(w, map[string]any{"items": []any{}})
			return
		}
		defer rows.Close()
		var out []Reminder
		for rows.Next() {
			var cid, cname string
			_ = rows.Scan(&cid, &cname)
			var re, rn, rd int
			var qs, qe int
			_ = db.QueryRow("SELECT remindEnabled, remindNoRecord, remindDiarrhea, quietStart, quietEnd FROM cat_settings WHERE catId=? AND userId=?", cid, userId).Scan(&re, &rn, &rd, &qs, &qe)
			if re == 0 {
				continue
			}
			now := time.Now()
			h := now.Hour()*60 + now.Minute()
			if qs > 0 || qe > 0 {
				if qs <= qe {
					if h >= qs && h < qe {
						continue
					}
				} else {
					if h >= qs || h < qe {
						continue
					}
				}
			}
			var lastEnd int64
			_ = db.QueryRow("SELECT IFNULL(MAX(endTime),0) FROM records WHERE userId=? AND catId=?", userId, cid).Scan(&lastEnd)
			if rn != 0 {
				if lastEnd == 0 || now.UnixMilli()-lastEnd >= 48*60*60*1000 {
					out = append(out, Reminder{CatID: cid, CatName: cname, Type: "no_record_48h", Message: "48小时未记录"})
				}
			}
			if rd != 0 {
				start := now.Add(-24 * time.Hour).UnixMilli()
				var cnt int
				_ = db.QueryRow("SELECT COUNT(*) FROM records WHERE userId=? AND catId=? AND endTime>=? AND endTime<? AND status=?", userId, cid, start, now.UnixMilli(), "diarrhea").Scan(&cnt)
				if cnt >= 2 {
					out = append(out, Reminder{CatID: cid, CatName: cname, Type: "diarrhea_24h", Message: "24小时腹泻次数较多"})
				}
			}
		}
		writeOK(w, map[string]any{"items": out})
	}))

	mux.HandleFunc("/api/reminders/templates", func(w http.ResponseWriter, r *http.Request) {
		env := os.Getenv("WEAPP_SUBSCRIBE_TEMPLATES")
		var list []string
		if env != "" {
			parts := strings.Split(env, ",")
			for _, p := range parts {
				s := strings.TrimSpace(p)
				if s != "" {
					list = append(list, s)
				}
			}
		}
		writeOK(w, map[string]any{"templates": list})
	})

	// 个人中心：统计卡片
	mux.HandleFunc("/api/profile/stats", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
		var totalCnt int64
		_ = db.QueryRow("SELECT COUNT(*) FROM records WHERE userId=?", userId).Scan(&totalCnt)
		var totalDur int64
		_ = db.QueryRow("SELECT IFNULL(SUM(duration),0) FROM records WHERE userId=?", userId).Scan(&totalDur)
		var friends int
		_ = db.QueryRow("SELECT COUNT(*) FROM friend_relations WHERE inviterUserId=? OR inviteeUserId=?", userId, userId).Scan(&friends)
		writeOK(w, map[string]any{"stats": map[string]any{"totalCount": totalCnt, "totalMinutes": totalDur / 60, "friendsCount": friends}})
	}))

	// 个人中心：我的成就
	mux.HandleFunc("/api/profile/achievements", withAuth(func(w http.ResponseWriter, r *http.Request, userId string) {
		// 连续记录天数（到今天）
		now := time.Now()
		start := now.AddDate(0, 0, -60) // 取近 60 天用于连续计算
		rows, err := db.Query("SELECT DISTINCT DATE(FROM_UNIXTIME(endTime/1000)) AS d FROM records WHERE userId=? AND endTime>=? AND endTime<? ORDER BY d DESC", userId, start.UnixMilli(), now.AddDate(0, 0, 1).UnixMilli())
		if err != nil {
			writeOK(w, map[string]any{"list": []any{}})
			return
		}
		defer rows.Close()
		daySet := make(map[string]struct{})
		for rows.Next() {
			var d string
			_ = rows.Scan(&d)
			daySet[d] = struct{}{}
		}
		streak := 0
		cur := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		for {
			d := cur.Format("2006-01-02")
			if _, ok := daySet[d]; ok {
				streak++
				cur = cur.AddDate(0, 0, -1)
			} else {
				break
			}
		}
		var totalCnt int64
		_ = db.QueryRow("SELECT COUNT(*) FROM records WHERE userId=?", userId).Scan(&totalCnt)
		list := []map[string]any{
			{"id": "streak", "title": "坚持记录", "desc": "已连续记录" + strconv.Itoa(streak) + "天", "achieved": streak >= 7, "value": streak},
			{"id": "milestone_30", "title": "健康达人", "desc": "记录超过30次", "achieved": totalCnt >= 30, "value": totalCnt},
		}
		writeOK(w, map[string]any{"list": list})
	}))

	h := cors(mux)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	log.Fatal(http.ListenAndServe(":"+port, h))
}
