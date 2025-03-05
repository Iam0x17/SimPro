<template>
  <div class="service-grid">
    <el-card v-for="service in services" :key="service.name" class="service-card">
      <template #header>
        <div class="card-header">
          <span>{{ service.displayName }}</span>
          <el-tag :type="service.status === 'Running' ? 'success' : 'info'" size="small">
            {{ service.status }}
          </el-tag>
        </div>
      </template>
      <div class="card-content">
        <el-button
          :type="service.status === 'Running' ? 'danger' : 'primary'"
          @click="toggleService(service)"
          :loading="service.loading"
        >
          {{ service.status === 'Running' ? '停止' : '启动' }}
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'

const services = ref([
  { name: 'ssh', displayName: 'SSH服务', status: 'Stopped', loading: false },
  { name: 'redis', displayName: 'Redis服务', status: 'Stopped', loading: false },
  { name: 'mysql', displayName: 'MySQL服务', status: 'Stopped', loading: false },
  { name: 'postgres', displayName: 'PostgreSQL服务', status: 'Stopped', loading: false },
  { name: 'telnet', displayName: 'Telnet服务', status: 'Stopped', loading: false },
  { name: 'ftp', displayName: 'FTP服务', status: 'Stopped', loading: false }
])

const fetchServiceStatus = async () => {
  try {
    for (const service of services.value) {
      const response = await axios.get(`/api/service?service_name=${service.name}&action=status`)
      service.status = response.data.status
    }
  } catch (error) {
    ElMessage.error('获取服务状态失败')
    console.error('Error fetching service status:', error)
  }
}

const toggleService = async (service) => {
  service.loading = true
  try {
    const action = service.status === 'Running' ? 'stop' : 'start'
    await axios.post(`/api/service?service_name=${service.name}&action=${action}`)
    await fetchServiceStatus()
    ElMessage.success(`${service.displayName}${action === 'start' ? '启动' : '停止'}成功`)
  } catch (error) {
    ElMessage.error(`${service.displayName}操作失败`)
    console.error('Error toggling service:', error)
  } finally {
    service.loading = false
  }
}

onMounted(() => {
  fetchServiceStatus()
  // 定期刷新服务状态
  setInterval(fetchServiceStatus, 5000) 
})
</script>

<style scoped>
.service-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
  margin-top: 20px;
}

.service-card {
  margin-bottom: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-content {
  text-align: center;
}
</style>