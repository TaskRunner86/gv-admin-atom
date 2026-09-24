package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"gv-admin-atom/internal/database"
	"gv-admin-atom/internal/middleware"
)

type userRow struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
}

// ListUsers 分页 + 关键字查询用户列表。
func ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))

	where := "WHERE 1=1"
	args := []any{}
	if keyword != "" {
		where += " AND username LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	var total int64
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM users "+where, args...).Scan(&total); err != nil {
		Fail(w, http.StatusInternalServerError, 500, "查询失败: "+err.Error())
		return
	}

	offset := (page - 1) * pageSize
	rows, err := database.DB.Query(
		"SELECT id, username, role, created_at FROM users "+where+
			" ORDER BY id DESC LIMIT ? OFFSET ?", append(args, pageSize, offset)...)
	if err != nil {
		Fail(w, http.StatusInternalServerError, 500, "查询失败: "+err.Error())
		return
	}
	defer rows.Close()

	list := make([]userRow, 0, pageSize)
	for rows.Next() {
		var u userRow
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt); err != nil {
			Fail(w, http.StatusInternalServerError, 500, "读取数据失败: "+err.Error())
			return
		}
		list = append(list, u)
	}
	OK(w, map[string]any{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

type userPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// CreateUser 新建用户。
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var p userPayload
	if !decodeJSON(w, r, &p) {
		return
	}
	p.Username = strings.TrimSpace(p.Username)
	if p.Username == "" || p.Password == "" {
		Fail(w, http.StatusBadRequest, 400, "用户名和密码不能为空")
		return
	}
	hash, err := database.HashPassword(p.Password)
	if err != nil {
		Fail(w, http.StatusInternalServerError, 500, "密码加密失败")
		return
	}
	res, err := database.DB.Exec(
		`INSERT INTO users (username, password, role) VALUES (?, ?, ?)`,
		p.Username, hash, p.Role)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			Fail(w, http.StatusConflict, 409, "用户名已存在")
			return
		}
		Fail(w, http.StatusInternalServerError, 500, "创建失败: "+err.Error())
		return
	}
	id, _ := res.LastInsertId()
	OK(w, map[string]any{"id": id})
}

// UpdateUser 更新用户资料（密码为空则保持不变）。
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, 400, "无效的用户ID")
		return
	}
	var p userPayload
	if !decodeJSON(w, r, &p) {
		return
	}

	// 先读取现有记录
	var existing struct {
		role     string
		password string
	}
	err = database.DB.QueryRow(
		"SELECT role, password FROM users WHERE id = ?", id).
		Scan(&existing.role, &existing.password)
	if err != nil {
		Fail(w, http.StatusNotFound, 404, "用户不存在")
		return
	}

	role := firstNonEmpty(p.Role, existing.role)

	args := []any{role, id}
	query := "UPDATE users SET role = ? WHERE id = ?"
	if p.Password != "" {
		hash, err := database.HashPassword(p.Password)
		if err != nil {
			Fail(w, http.StatusInternalServerError, 500, "密码加密失败")
			return
		}
		query = "UPDATE users SET role = ?, password = ? WHERE id = ?"
		args = []any{role, hash, id}
	}
	if _, err := database.DB.Exec(query, args...); err != nil {
		Fail(w, http.StatusInternalServerError, 500, "更新失败: "+err.Error())
		return
	}
	OK(w, map[string]any{"id": id})
}

// DeleteUser 删除用户（禁止删除自己与 admin）。
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, 400, "无效的用户ID")
		return
	}
	selfID := middleware.UserID(r.Context())
	if id == selfID {
		Fail(w, http.StatusForbidden, 403, "不能删除当前登录账号")
		return
	}
	var username string
	if err := database.DB.QueryRow("SELECT username FROM users WHERE id = ?", id).Scan(&username); err != nil {
		Fail(w, http.StatusNotFound, 404, "用户不存在")
		return
	}
	if username == "admin" {
		Fail(w, http.StatusForbidden, 403, "不能删除内置管理员")
		return
	}
	if _, err := database.DB.Exec("DELETE FROM users WHERE id = ?", id); err != nil {
		Fail(w, http.StatusInternalServerError, 500, "删除失败: "+err.Error())
		return
	}
	OK(w, map[string]any{"id": id})
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
