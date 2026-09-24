// Package middleware 提供 CORS 与登录鉴权中间件。
package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"gv-admin-atom/internal/database"
)

// CORS 允许跨域访问（开发时前端 5173 直连后端）。
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type ctxKey string

// CtxUserID 是上下文中注入用户 ID 使用的键（类型化，避免与字符串键混淆）。
const CtxUserID ctxKey = "userID"

// UserID 从请求上下文读取当前登录用户 ID。
func UserID(ctx context.Context) int64 {
	id, _ := ctx.Value(CtxUserID).(int64)
	return id
}

// Auth 校验 Authorization: Bearer <token>，通过后在上下文中注入 userID。
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		token := ""
		if strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		}
		if token == "" {
			http.Error(w, `{"code":401,"message":"未登录或登录已过期"}`, http.StatusUnauthorized)
			return
		}

		var userID int64
		var expires string
		err := database.DB.QueryRow(
			"SELECT user_id, expires_at FROM tokens WHERE token = ?", token).Scan(&userID, &expires)
		if err == sql.ErrNoRows {
			http.Error(w, `{"code":401,"message":"未登录或登录已过期"}`, http.StatusUnauthorized)
			return
		}
		if err != nil {
			http.Error(w, `{"code":500,"message":"鉴权查询失败"}`, http.StatusInternalServerError)
			return
		}
		if exp, perr := time.ParseInLocation("2006-01-02 15:04:05", expires, time.Local); perr == nil && time.Now().After(exp) {
			_, _ = database.DB.Exec("DELETE FROM tokens WHERE token = ?", token)
			http.Error(w, `{"code":401,"message":"登录已过期，请重新登录"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), CtxUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin 需置于 Auth 之后，校验当前用户是否为管理员，否则返回 403。
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var role string
		err := database.DB.QueryRow("SELECT role FROM users WHERE id = ?", UserID(r.Context())).Scan(&role)
		if err == sql.ErrNoRows {
			http.Error(w, `{"code":401,"message":"账号不存在或已被删除"}`, http.StatusUnauthorized)
			return
		}
		if err != nil {
			http.Error(w, `{"code":500,"message":"权限校验失败"}`, http.StatusInternalServerError)
			return
		}
		if role != "admin" {
			http.Error(w, `{"code":403,"message":"无权限，用户管理仅限管理员使用"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
