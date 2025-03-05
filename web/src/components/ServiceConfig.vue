<template>
  <div class="config-container">
    <el-table :data="services" style="width: 100%">
      <el-table-column prop="displayName" label="服务名称" width="180" />
      <el-table-column prop="port" label="端口" width="180">
        <template #default="{ row }">
          <el-input v-model="row.port" placeholder="请输入端口号" />
        </template>
      </el-table-column>
      <el-table-column prop="user" label="用户名" width="180">
        <template #default="{ row }">
          <el-input v-model="row.user" placeholder="请输入用户名" />
        </template>
      </el-table-column>
      <el-table-column prop="pass" label="密码">
        <template #default="{ row }">
          <el-input v-model="row.pass" placeholder="请输入密码" show-pass />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120">
        <template #default="{ row }">
          <el-button type="primary" size="small" @click="saveConfig(row)">保存</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'

const services = ref([
  { name: 'ssh', displayName: 'SSH服务', port: '', user: '', pass: '' },
  { name: 'redis', displayName: 'Redis服务', port: '', user: '', pass: '' },
  { name: 'mysql', displayName: 'MySQL服务', port: '', user: '', pass: '' },
  { name: 'postgres', displayName: 'PostgreSQL服务', port: '', user: '', pass: '' },
  { name: 'telnet', displayName: 'Telnet服务', port: '', user: '', pass: '' },
  { name: 'ftp', displayName: 'FTP服务', port: '', user: '', pass: '' }
])

const fetchConfig = async () => {
  try {
    const response = await axios.get('/api/service/config')
    services.value = services.value.map(service => ({
      ...service,
      ...response.data[service.name]
    }))
  } catch (error) {
    ElMessage.error('获取配置信息失败')
    console.error('Error fetching config:', error)
  }
}

const saveConfig = async (service) => {
  try {
    await axios.post('/api/service/config', {
      service_name: service.name,
      port: service.port,
      user: service.user,
      pass: service.pass
    })
    ElMessage.success(`${service.displayName}配置保存成功`)
  } catch (error) {
    ElMessage.error(`${service.displayName}配置保存失败`)
    console.error('Error saving config:', error)
  }
}

onMounted(() => {
  fetchConfig()
})
</script>

<style scoped>
.config-container {
  padding: 20px;
}
</style>