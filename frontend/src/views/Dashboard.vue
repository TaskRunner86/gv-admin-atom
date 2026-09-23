<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col v-for="card in statCards" :key="card.label" :xs="12" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-body">
            <div class="stat-icon" :style="{ backgroundColor: card.color + '1a', color: card.color }">
              <el-icon :size="26"><component :is="card.icon" /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ card.value }}</div>
              <div class="stat-label">{{ card.label }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 上报趋势折线 + 设备类型饼图 -->
    <el-row :gutter="16" class="chart-row">
      <el-col :xs="24" :md="16">
        <el-card shadow="hover">
          <template #header>近 7 天数据上报趋势</template>
          <EChart :option="trendOption" height="320px" />
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card shadow="hover">
          <template #header>设备类型占比</template>
          <EChart :option="typeOption" height="320px" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 30 天上报柱状 + 设备状态 -->
    <el-row :gutter="16" class="chart-row">
      <el-col :xs="24" :md="14">
        <el-card shadow="hover">
          <template #header>近 30 天数据上报量</template>
          <EChart :option="monthOption" height="300px" />
        </el-card>
      </el-col>
      <el-col :xs="24" :md="10">
        <el-card shadow="hover">
          <template #header>设备状态分布</template>
          <EChart :option="statusOption" height="300px" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 最新告警 -->
    <el-card shadow="hover" class="table-card">
      <template #header>最新告警</template>
      <el-table :data="recentAlerts" stripe style="width: 100%">
        <el-table-column prop="alertNo" label="告警编号" width="160" />
        <el-table-column prop="deviceName" label="设备" min-width="170" show-overflow-tooltip />
        <el-table-column prop="location" label="点位" width="130" show-overflow-tooltip />
        <el-table-column prop="metric" label="告警指标" min-width="140" show-overflow-tooltip />
        <el-table-column prop="level" label="级别" width="90">
          <template #default="{ row }">
            <el-tag :type="levelTagType(row.level)" size="small">{{ row.level }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="value" label="触发值" width="130" align="right">
          <template #default="{ row }">
            <span class="alert-value">{{ row.value }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="处理状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="告警时间" min-width="160" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getOverview } from '../api'
import EChart from '../components/EChart.vue'

const overview = ref({
  stats: { totalDevices: 0, onlineDevices: 0, todayMessages: 0, todayAlerts: 0 },
  trend: [],
  monthTrend: [],
  typeDist: [],
  statusDist: [],
  recentAlerts: []
})

const formatNumber = (n) => Number(n || 0).toLocaleString('zh-CN')
const formatCompact = (n) => {
  const v = Number(n || 0)
  return v >= 10000 ? (v / 10000).toFixed(1) + ' 万' : formatNumber(v)
}

const statCards = computed(() => [
  { label: '接入设备总数', value: formatNumber(overview.value.stats.totalDevices), icon: 'Monitor', color: '#409eff' },
  { label: '在线设备数', value: formatNumber(overview.value.stats.onlineDevices), icon: 'Connection', color: '#67c23a' },
  { label: '今日上报消息', value: formatCompact(overview.value.stats.todayMessages), icon: 'DataLine', color: '#e6a23c' },
  { label: '今日告警数', value: formatNumber(overview.value.stats.todayAlerts), icon: 'Warning', color: '#f56c6c' }
])

const trendOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: ['消息上报量', '在线设备数'] },
  grid: { left: 60, right: 55, top: 40, bottom: 30 },
  xAxis: { type: 'category', data: overview.value.trend.map((p) => p.day.slice(5)) },
  yAxis: [
    { type: 'value', name: '条' },
    { type: 'value', name: '台', min: 0, splitLine: { show: false } }
  ],
  series: [
    {
      name: '消息上报量',
      type: 'line',
      smooth: true,
      data: overview.value.trend.map((p) => p.messages),
      areaStyle: { opacity: 0.15 },
      itemStyle: { color: '#409eff' }
    },
    {
      name: '在线设备数',
      type: 'line',
      smooth: true,
      yAxisIndex: 1,
      data: overview.value.trend.map((p) => p.devices),
      itemStyle: { color: '#67c23a' }
    }
  ]
}))

const monthOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  grid: { left: 60, right: 20, top: 30, bottom: 30 },
  xAxis: { type: 'category', data: overview.value.monthTrend.map((p) => p.day.slice(5)) },
  yAxis: {
    type: 'value',
    axisLabel: { formatter: (v) => (v >= 10000 ? (v / 10000).toFixed(1) + '万' : v) }
  },
  series: [
    {
      name: '上报消息量',
      type: 'bar',
      barMaxWidth: 22,
      data: overview.value.monthTrend.map((p) => p.messages),
      itemStyle: { color: '#409eff', borderRadius: [3, 3, 0, 0] }
    }
  ]
}))

const typeOption = computed(() => ({
  tooltip: { trigger: 'item', formatter: '{b}: {c} 台 ({d}%)' },
  legend: { bottom: 0, type: 'scroll' },
  series: [
    {
      name: '设备类型',
      type: 'pie',
      radius: ['40%', '68%'],
      center: ['50%', '44%'],
      avoidLabelOverlap: true,
      itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
      label: { show: false },
      emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
      data: overview.value.typeDist.map((t) => ({ name: t.name, value: t.value }))
    }
  ]
}))

const statusOption = computed(() => ({
  tooltip: { trigger: 'item', formatter: '{b}: {c} 台 ({d}%)' },
  legend: { bottom: 0 },
  series: [
    {
      name: '设备状态',
      type: 'pie',
      radius: '68%',
      center: ['50%', '44%'],
      itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
      label: { show: true, formatter: '{b}: {c}' },
      data: overview.value.statusDist.map((s) => ({ name: s.name, value: s.value }))
    }
  ]
}))

const recentAlerts = computed(() => overview.value.recentAlerts)

const levelTagType = (level) => {
  const map = { 严重: 'danger', 警告: 'warning', 提示: 'info' }
  return map[level] || 'info'
}

const statusTagType = (status) => {
  const map = { 未处理: 'danger', 处理中: 'warning', 已恢复: 'success' }
  return map[status] || 'info'
}

onMounted(async () => {
  try {
    overview.value = await getOverview()
  } catch {
    /* 拦截器已提示 */
  }
})
</script>

<style scoped>
.stat-row {
  margin-bottom: 16px;
}

.stat-card :deep(.el-card__body) {
  padding: 18px;
}

.stat-body {
  display: flex;
  align-items: center;
  gap: 14px;
}

.stat-icon {
  width: 54px;
  height: 54px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #303133;
  line-height: 1.2;
}

.stat-label {
  margin-top: 4px;
  font-size: 13px;
  color: #909399;
}

.chart-row {
  margin-bottom: 16px;
}

.table-card {
  margin-bottom: 16px;
}

.alert-value {
  color: #f56c6c;
  font-weight: 600;
}
</style>
