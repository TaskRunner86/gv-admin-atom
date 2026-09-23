// Package handlers 实现各 HTTP 接口的业务逻辑。
package handlers

import (
	"encoding/json"
	"net/http"
)

// OK 返回统一成功结构：{"code":0,"message":"ok","data":...}
func OK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, map[string]any{"code": 0, "message": "ok", "data": data})
}

// Fail 返回统一失败结构：{"code":非0,"message":错误信息}
func Fail(w http.ResponseWriter, status int, code int, message string) {
	writeJSON(w, status, map[string]any{"code": code, "message": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// decodeJSON 解析请求体 JSON，失败时写入错误响应并返回 false。
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		Fail(w, http.StatusBadRequest, 400, "请求体格式错误: "+err.Error())
		return false
	}
	return true
}
