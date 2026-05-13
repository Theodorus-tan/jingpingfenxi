<template>
  <div class="container">
    <div class="content">
      <div class="welcome-box">
        <h1 class="title">AI 竞品分析系统</h1>
        <p class="subtitle">深度洞察竞对动态，助力业务决策</p>
        <a-button
          type="primary"
          size="large"
          class="enter-btn"
          @click="handleEnter"
        >
          进入竞品分析
        </a-button>
      </div>
    </div>
    <div class="footer">
      <Footer />
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { useRouter } from 'vue-router';
  import { useUserStore } from '@/store';
  import Footer from '@/components/footer/index.vue';

  const router = useRouter();
  const userStore = useUserStore();

  const handleEnter = async () => {
    // 模拟一个默认账号登录状态以绕过路由守卫
    userStore.$patch({
      name: '分析师',
      role: 'admin',
    });
    router.push({ name: 'workbench' });
  };
</script>

<style lang="less" scoped>
  .container {
    display: flex;
    flex-direction: column;
    height: 100vh;
    background: linear-gradient(163.85deg, #1d2129 0%, #00308f 100%);
  }

  .content {
    display: flex;
    flex: 1;
    align-items: center;
    justify-content: center;
  }

  .welcome-box {
    text-align: center;
    background: var(--color-bg-2);
    padding: 60px 80px;
    border-radius: 12px;
    box-shadow: 0 4px 10px rgba(0, 0, 0, 0.1);

    .title {
      font-size: 32px;
      color: var(--color-text-1);
      margin-bottom: 16px;
      margin-top: 0;
    }

    .subtitle {
      font-size: 16px;
      color: var(--color-text-3);
      margin-bottom: 40px;
    }

    .enter-btn {
      width: 200px;
      height: 48px;
      font-size: 18px;
      border-radius: 24px;
    }
  }

  .footer {
    width: 100%;
    padding: 20px 0;
  }
</style>
