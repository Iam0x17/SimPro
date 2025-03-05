<template>
  <div class="container">
    <p>SimPro - 轻量级协议模拟器</p>
    <el-tabs v-model="activeTab">
      <el-tab-pane label="服务管理" name="services">
        <service-management />
      </el-tab-pane>
      <el-tab-pane label="服务配置" name="config">
        <service-config />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import ServiceConfig from './components/ServiceConfig.vue'
import ServiceManagement from './components/ServiceManagement.vue'
const activeTab = ref('services')
</script>

<style scoped>
.container {
  max-width: 1200px;
  margin: 20px auto;
  padding: 0 20px;
}
</style>