package database

import (
	"crypto/rand"
	"math/big"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 对密码做哈希。
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword 校验明文密码与哈希是否匹配。
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// seedUsers 创建管理员与若干演示用户。
func seedUsers() error {
	adminHash, err := HashPassword("admin123")
	if err != nil {
		return err
	}
	_, err = DB.Exec(`INSERT INTO users (username, password, role) VALUES (?, ?, 'admin')`,
		"admin", adminHash)
	if err != nil {
		return err
	}

	names := []string{"zhangsan", "lisi", "wangwu", "zhaoliu", "sunqi", "zhouba"}
	for _, name := range names {
		hash, err := HashPassword("123456")
		if err != nil {
			return err
		}
		_, err = DB.Exec(`INSERT INTO users (username, password, role) VALUES (?, ?, 'user')`,
			name, hash)
		if err != nil {
			return err
		}
	}
	return nil
}

// deviceSeed 描述一台演示设备。
type deviceSeed struct {
	no string
	name string
	typ string
	loc string
	status string
}

// seedDevices 演示设备台账：8 类设备共 24 台，分布于园区各点位。
var deviceSeeds = []deviceSeed{
	{"IOT-ENV-1001", "1号厂房温湿度传感器", "环境传感器", "1号厂房A区", "在线"},
	{"IOT-ENV-1002", "1号厂房烟感传感器", "环境传感器", "1号厂房A区", "在线"},
	{"IOT-ENV-1003", "2号仓库温湿度传感器", "环境传感器", "2号仓库", "在线"},
	{"IOT-ENV-1004", "冷链仓库温湿度传感器", "环境传感器", "冷链仓库", "在线"},
	{"IOT-ENV-1005", "污水站气体传感器", "环境传感器", "污水处理站", "离线"},
	{"IOT-MET-2001", "配电房智能电表", "智能电表", "园区配电房", "在线"},
	{"IOT-MET-2002", "1号厂房智能电表", "智能电表", "1号厂房B区", "在线"},
	{"IOT-MET-2003", "办公楼智能电表", "智能电表", "综合办公楼3F", "在线"},
	{"IOT-MET-2004", "光伏电站智能电表", "智能电表", "光伏电站", "维护"},
	{"IOT-GW-3001", "1号厂房边缘网关", "边缘网关", "1号厂房A区", "在线"},
	{"IOT-GW-3002", "2号仓库边缘网关", "边缘网关", "2号仓库", "在线"},
	{"IOT-GW-3003", "办公楼边缘网关", "边缘网关", "综合办公楼1F", "在线"},
	{"IOT-CAM-4001", "1号厂房监控摄像头", "高清摄像头", "1号厂房B区", "在线"},
	{"IOT-CAM-4002", "仓库周界摄像头", "高清摄像头", "2号仓库", "在线"},
	{"IOT-CAM-4003", "园区北门摄像头", "高清摄像头", "园区北门", "离线"},
	{"IOT-THM-5001", "机房精密空调温控器", "温控器", "数据中心机房", "在线"},
	{"IOT-THM-5002", "冷链仓库温控器", "温控器", "冷链仓库", "在线"},
	{"IOT-WTR-6001", "1号厂房智能水表", "智能水表", "1号厂房A区", "在线"},
	{"IOT-WTR-6002", "食堂智能水表", "智能水表", "食堂后厨", "维护"},
	{"IOT-AIR-7001", "园区空气监测站", "空气监测站", "园区北门", "在线"},
	{"IOT-AIR-7002", "厂房空气监测站", "空气监测站", "1号厂房B区", "在线"},
	{"IOT-ACS-8001", "办公楼门禁控制器", "门禁控制器", "综合办公楼1F", "在线"},
	{"IOT-ACS-8002", "机房门禁控制器", "门禁控制器", "数据中心机房", "在线"},
	{"IOT-ACS-8003", "仓库门禁控制器", "门禁控制器", "2号仓库", "离线"},
}

// typeFirmware 各类型设备的固件版本。
var typeFirmware = map[string]string{
	"环境传感器": "v2.3.1",
	"智能电表":  "v3.0.2",
	"边缘网关":  "v4.1.0",
	"高清摄像头": "v1.8.3",
	"温控器":   "v2.5.0",
	"智能水表":  "v1.9.2",
	"空气监测站": "v3.2.4",
	"门禁控制器": "v2.0.1",
}

// seedDevices 写入演示设备台账，最后上报时间按设备状态推演。
func seedDevices() error {
	now := time.Now()
	for i, d := range deviceSeeds {
		fw := typeFirmware[d.typ]
		// 在线设备最近仍在心跳，离线/维护设备已失联较久
		var lastSeen time.Time
		switch d.status {
		case "在线":
			lastSeen = now.Add(-time.Duration(1+i%5) * time.Minute)
		default:
			lastSeen = now.Add(-time.Duration(6+i) * time.Hour)
		}
		_, err := DB.Exec(
			`INSERT INTO devices (device_no, name, type, location, status, firmware, last_seen_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			d.no, d.name, d.typ, d.loc, d.status, fw, lastSeen.Format("2006-01-02 15:04:05"))
		if err != nil {
			return err
		}
	}
	return nil
}

// seedMetrics 按最近 30 天滚动补齐数据上报量与在线设备数（随时间递增）。
// 已存在的日期会被保留，只填充缺失的日期，因此可重复执行。
func seedMetrics() error {
	now := time.Now()
	for i := 29; i >= 0; i-- {
		day := now.AddDate(0, 0, -i).Format("2006-01-02")
		// 上报量随时间增长
		messages := 12000 + int64(29-i)*450 + randN(1800)
		// 在线设备数 14~16
		devices := 14 + int(randN(3))
		_, err := DB.Exec(
			`INSERT OR IGNORE INTO daily_metrics (day, messages, devices) VALUES (?, ?, ?)`,
			day, messages, devices)
		if err != nil {
			return err
		}
	}
	return nil
}

func randN(max int64) int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return max / 2
	}
	return n.Int64()
}
