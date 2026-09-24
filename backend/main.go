package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gv-admin-atom/internal/database"
	"gv-admin-atom/internal/handlers"
	"gv-admin-atom/internal/middleware"
)

func main() {
	// 初始化数据库（含建表与种子数据）
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = filepath.Join("data", "dashboard.db")
	}
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.Close()

	mux := http.NewServeMux()

	// 公开接口
	mux.HandleFunc("POST /api/login", handlers.Login)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		handlers.OK(w, map[string]any{"status": "ok"})
	})

	// 受保护接口
	mux.Handle("GET /api/user/profile", middleware.Auth(http.HandlerFunc(handlers.Profile)))
	mux.Handle("GET /api/dashboard/overview", middleware.Auth(http.HandlerFunc(handlers.DashboardOverview)))
	// 用户管理接口：查看、新增、编辑、删除仅限管理员
	// 注意 Auth 必须在最外层，先注入 userID，RequireAdmin 才能读到当前用户角色
	adminOnly := func(h http.HandlerFunc) http.Handler {
		return middleware.Auth(middleware.RequireAdmin(h))
	}
	mux.Handle("GET /api/users", adminOnly(handlers.ListUsers))
	mux.Handle("POST /api/users", adminOnly(handlers.CreateUser))
	mux.Handle("PUT /api/users/{id}", adminOnly(handlers.UpdateUser))
	mux.Handle("DELETE /api/users/{id}", adminOnly(handlers.DeleteUser))

	// 若前端已构建（frontend/dist 存在），由 Go 直接托管静态资源
	dist := filepath.Join("..", "frontend", "dist")
	if _, err := os.Stat(dist); err == nil {
		fileServer := http.FileServer(http.Dir(dist))
		mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// SPA 回退：深层路由（如 /dashboard）刷新时回退到 index.html
			if r.URL.Path != "/" {
				candidate := filepath.Join(dist, filepath.Clean(r.URL.Path))
				if info, err := os.Stat(candidate); err != nil || info.IsDir() {
					r.URL.Path = "/"
				}
			}
			fileServer.ServeHTTP(w, r)
		}))
		log.Printf("前端静态资源托管目录: %s", dist)
	}

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("Dashboard 服务已启动: http://localhost%s", addr)
	if err := http.ListenAndServe(addr, middleware.CORS(mux)); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
