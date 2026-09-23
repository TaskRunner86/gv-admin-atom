package handlers

import (
	"net/http"
	"time"

	"gv-dashboard/internal/database"
)

// DashboardOverview 返回物联网看板首页所需聚合数据：
// 统计卡片、上报趋势、设备类型分布、设备状态分布、最新告警。
func DashboardOverview(w http.ResponseWriter, r *http.Request) {
	stats, err := loadStats()
	if err != nil {
		Fail(w, http.StatusInternalServerError, 500, "统计查询失败: "+err.Error())
		return
	}
	trend, err := loadMetricTrend(7)
	if err != nil {
		Fail(w, http.StatusInternalServerError, 500, "趋势查询失败: "+err.Error())
		return
	}
	monthTrend, err := loadMetricTrend(30)
	if err != nil {
		Fail(w, http.StatusInternalServerError, 500, "月趋势查询失败: "+err.Error())
		return
	}
	typeDist, err := loadTypeDist()
	if err != nil {
		Fail(w, http.StatusInternalServerError, 500, "设备类型查询失败: "+err.Error())
		return
	}
	statusDist, err := loadStatusDist()
	if err != nil {
		Fail(w, http.StatusInternalServerError, 500, "设备状态查询失败: "+err.Error())
		return
	}
	recentAlerts, err := loadRecentAlerts(10)
	if err != nil {
		Fail(w, http.StatusInternalServerError, 500, "告警查询失败: "+err.Error())
		return
	}

	OK(w, map[string]any{
		"stats":        stats,
		"trend":        trend,
		"monthTrend":   monthTrend,
		"typeDist":     typeDist,
		"statusDist":   statusDist,
		"recentAlerts": recentAlerts,
	})
}

type stats struct {
	TotalDevices  int64 `json:"totalDevices"`
	OnlineDevices int64 `json:"onlineDevices"`
	TodayMessages int64 `json:"todayMessages"`
	TodayAlerts   int64 `json:"todayAlerts"`
}

func loadStats() (stats, error) {
	var s stats
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM devices").Scan(&s.TotalDevices); err != nil {
		return s, err
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM devices WHERE status = '在线'").
		Scan(&s.OnlineDevices); err != nil {
		return s, err
	}
	today := time.Now().Format("2006-01-02")
	if err := database.DB.QueryRow("SELECT COALESCE(SUM(messages), 0) FROM daily_metrics WHERE day = ?", today).
		Scan(&s.TodayMessages); err != nil {
		return s, err
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM alerts WHERE date(created_at) = ?", today).
		Scan(&s.TodayAlerts); err != nil {
		return s, err
	}
	return s, nil
}

type trendPoint struct {
	Day      string `json:"day"`
	Messages int64  `json:"messages"`
	Devices  int64  `json:"devices"`
}

// loadMetricTrend 读取最近 days 天的上报量与在线设备数（按日期升序）。
func loadMetricTrend(days int) ([]trendPoint, error) {
	start := time.Now().AddDate(0, 0, -(days - 1)).Format("2006-01-02")
	rows, err := database.DB.Query(
		"SELECT day, messages, devices FROM daily_metrics WHERE day >= ? ORDER BY day ASC", start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := make([]trendPoint, 0, days)
	for rows.Next() {
		var p trendPoint
		if err := rows.Scan(&p.Day, &p.Messages, &p.Devices); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

type categoryItem struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// loadTypeDist 统计各类型设备数量（按数量降序）。
func loadTypeDist() ([]categoryItem, error) {
	rows, err := database.DB.Query(
		"SELECT type, COUNT(*) FROM devices GROUP BY type ORDER BY COUNT(*) DESC, type")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]categoryItem, 0, 8)
	for rows.Next() {
		var it categoryItem
		var n int64
		if err := rows.Scan(&it.Name, &n); err != nil {
			return nil, err
		}
		it.Value = float64(n)
		items = append(items, it)
	}
	return items, rows.Err()
}

// loadStatusDist 统计设备运行状态分布，按在线 / 告警 / 离线 / 维护排序。
func loadStatusDist() ([]categoryItem, error) {
	rows, err := database.DB.Query(
		`SELECT status, COUNT(*) FROM devices GROUP BY status
		 ORDER BY CASE status WHEN '在线' THEN 1 WHEN '告警' THEN 2 WHEN '离线' THEN 3 ELSE 4 END`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]categoryItem, 0, 4)
	for rows.Next() {
		var it categoryItem
		var n int64
		if err := rows.Scan(&it.Name, &n); err != nil {
			return nil, err
		}
		it.Value = float64(n)
		items = append(items, it)
	}
	return items, rows.Err()
}

type alertItem struct {
	ID         int64  `json:"id"`
	AlertNo    string `json:"alertNo"`
	DeviceNo   string `json:"deviceNo"`
	DeviceName string `json:"deviceName"`
	Location   string `json:"location"`
	Metric     string `json:"metric"`
	Level      string `json:"level"`
	Value      string `json:"value"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
}

// loadRecentAlerts 读取最新 limit 条告警（含设备名称与点位）。
func loadRecentAlerts(limit int) ([]alertItem, error) {
	rows, err := database.DB.Query(
		`SELECT a.id, a.alert_no, d.device_no, d.name, d.location, a.metric, a.level, a.value, a.status, a.created_at
		 FROM alerts a JOIN devices d ON d.device_no = a.device_no
		 ORDER BY a.created_at DESC, a.id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]alertItem, 0, limit)
	for rows.Next() {
		var it alertItem
		if err := rows.Scan(&it.ID, &it.AlertNo, &it.DeviceNo, &it.DeviceName, &it.Location,
			&it.Metric, &it.Level, &it.Value, &it.Status, &it.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}
