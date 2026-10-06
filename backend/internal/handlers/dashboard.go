package handlers

import (
	"net/http"
	"time"

	"gv-admin-atom/internal/database"
)

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

	OK(w, map[string]any{
		"stats": stats,
		"trend": trend,
		"monthTrend": monthTrend,
		"typeDist": typeDist,
		"statusDist": statusDist,
	})
}

type stats struct {
	TotalDevices int64 `json:"totalDevices"`
	OnlineDevices int64 `json:"onlineDevices"`
	TodayMessages int64 `json:"todayMessages"`
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
	return s, nil
}

type trendPoint struct {
	Day string `json:"day"`
	Messages int64 `json:"messages"`
	Devices int64 `json:"devices"`
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
	Name string `json:"name"`
	Value float64 `json:"value"`
}

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
		err := rows.Scan(&it.Name, &n)
		if err != nil {
			return nil, err
		}
		it.Value = float64(n)
		items = append(items, it)
	}
	return items, rows.Err()
}

// loadStatusDist 统计设备运行状态分布，按在线 / 离线 / 维护排序。
func loadStatusDist() ([]categoryItem, error) {
	rows, err := database.DB.Query(
		`SELECT status, COUNT(*) FROM devices GROUP BY status
		 ORDER BY CASE status WHEN '在线' THEN 1 WHEN '离线' THEN 2 ELSE 3 END`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]categoryItem, 0, 4)
	for rows.Next() {
		var it categoryItem
		var n int64
		err := rows.Scan(&it.Name, &n)
		if err != nil {
			return nil, err
		}
		it.Value = float64(n)
		items = append(items, it)
	}
	return items, rows.Err()
}
