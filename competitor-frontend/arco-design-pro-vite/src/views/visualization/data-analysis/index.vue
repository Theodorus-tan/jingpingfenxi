<template>
  <div class="page">
    <a-space direction="vertical" :size="16" fill>
      <a-card class="card" title="分析报告">
        <a-space direction="vertical" :size="8" fill>
          <div class="headline">
            <span class="title">{{ competitorName }}</span>
            <a-tag color="arcoblue" size="small">AI 生成报告</a-tag>
          </div>
          <div class="muted">生成时间：{{ generatedAt }}</div>
        </a-space>
      </a-card>

      <a-card class="card" title="AI 深度洞察">
        <div v-if="reportHtml" class="markdown-body" v-html="reportHtml"></div>
        <a-empty v-else description="暂无分析报告，请先去「新建分析」生成" />
      </a-card>
    </a-space>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { marked } from 'marked';

  const route = useRoute();

  const competitorName = ref('');
  const generatedAt = ref('');
  const reportHtml = ref('');

  onMounted(async () => {
    // 优先取 url 参数中的竞品，如果没有，取 localstorage 中的最新一次分析记录
    const nameFromQuery = route.query.competitor as string;
    const localName = localStorage.getItem('latest_analysis_competitor') || '未知竞品';
    competitorName.value = nameFromQuery || localName;

    generatedAt.value = new Date().toISOString().slice(0, 16).replace('T', ' ');

    const rawReport = localStorage.getItem('latest_analysis_report');
    if (rawReport) {
      reportHtml.value = await marked.parse(rawReport);
    }
  });
</script>

<style lang="less" scoped>
  .page {
    background: var(--color-fill-2);
    padding: 16px 20px;
    min-height: calc(100vh - 120px);
  }

  .card {
    border-radius: 8px;
  }

  .headline {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .title {
    font-weight: 600;
    color: var(--color-text-1);
    font-size: 20px;
  }

  .muted {
    color: var(--color-text-3);
    font-size: 12px;
  }

  /* 简单的 Markdown 样式适配 */
  .markdown-body {
    font-size: 14px;
    line-height: 1.8;
    color: var(--color-text-2);

    :deep(h1), :deep(h2), :deep(h3), :deep(h4) {
      color: var(--color-text-1);
      margin-top: 24px;
      margin-bottom: 16px;
      font-weight: 600;
    }
    
    :deep(h1) { font-size: 22px; }
    :deep(h2) { font-size: 18px; }
    :deep(h3) { font-size: 16px; }

    :deep(p) {
      margin-bottom: 16px;
    }

    :deep(ul), :deep(ol) {
      padding-left: 20px;
      margin-bottom: 16px;
    }

    :deep(li) {
      margin-bottom: 8px;
    }

    :deep(strong) {
      color: var(--color-text-1);
      font-weight: 600;
    }
  }
</style>
