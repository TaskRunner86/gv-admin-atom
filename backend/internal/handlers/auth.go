package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"gv-admin-atom/internal/database"
	"gv-admin-atom/internal/middleware"
)

const tokenTTL = 24 * time.Hour

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type userInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Login 登录接口：校验用户名密码，签发 token。
func Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Username == "" || req.Password == "" {
		Fail(w, http.StatusBadRequest, 400, "用户名和密码不能为空")
		return
	}

	var u userInfo
	var hash string
	err := database.DB.QueryRow(
		"SELECT id, username, password, role FROM users WHERE username = ?",
		req.Username,
	).Scan(&u.ID, &u.Username, &hash, &u.Role)
	if err != nil {
		Fail(w, http.StatusUnauthorized, 401, "用户名或密码错误")
		return
	}
	if !database.CheckPassword(hash, req.Password) {
		Fail(w, http.StatusUnauthorized, 401, "用户名或密码错误")
		return
	}

	token, err := newToken()
	if err != nil {
		Fail(w, http.StatusInternalServerError, 500, "签发 token 失败")
		return
	}
	expires := time.Now().Add(tokenTTL).Format("2006-01-02 15:04:05")
	if _, err := database.DB.Exec(
		"INSERT INTO tokens (token, user_id, expires_at) VALUES (?, ?, ?)", token, u.ID, expires); err != nil {
		Fail(w, http.StatusInternalServerError, 500, "保存 token 失败")
		return
	}

	OK(w, map[string]any{"token": token, "user": u})
}

// Profile 返回当前登录用户信息。
func Profile(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	var u userInfo
	err := database.DB.QueryRow(
		"SELECT id, username, role FROM users WHERE id = ?", uid,
	).Scan(&u.ID, &u.Username, &u.Role)
	if err != nil {
		Fail(w, http.StatusNotFound, 404, "用户不存在")
		return
	}
	OK(w, u)
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
