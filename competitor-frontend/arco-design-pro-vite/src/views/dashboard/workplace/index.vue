<template>
  <div class="page">
    <a-space direction="vertical" :size="16" fill>
      <a-card class="card" title="工作台">
        <a-space :size="12" wrap>
          <a-button type="primary" @click="go('/analysis/new')">
            新建分析
          </a-button>
          <a-button @click="go('/competitors/list')">竞品管理</a-button>
          <a-button @click="go('/analysis-report/view')">查看报告</a-button>
        </a-space>
      </a-card>

      <a-card class="card" title="概览">
        <a-grid :cols="24" :col-gap="16" :row-gap="16">
          <a-grid-item :span="{ xs: 24, sm: 12, md: 12, lg: 6 }">
            <a-statistic title="竞品数" :value="stats.competitors" />
          </a-grid-item>
          <a-grid-item :span="{ xs: 24, sm: 12, md: 12, lg: 6 }">
            <a-statistic title="分析次数" :value="stats.analyses" />
          </a-grid-item>
          <a-grid-item :span="{ xs: 24, sm: 12, md: 12, lg: 6 }">
            <a-statistic
              title="平均评分"
              :value="stats.avgScore"
              :precision="1"
            />
          </a-grid-item>
          <a-grid-item :span="{ xs: 24, sm: 12, md: 12, lg: 6 }">
            <a-statistic title="本月新增" :value="stats.newThisMonth" />
          </a-grid-item>
        </a-grid>
      </a-card>

      <a-card class="card" title="最近分析">
        <a-list :bordered="false">
          <a-list-item v-for="item in recent" :key="item.id">
            <a-space :size="10" wrap>
              <span class="name">{{ item.name }}</span>
              <a-tag size="small" color="arcoblue">{{ item.category }}</a-tag>
              <a-tag
                size="small"
                :color="item.status === 'completed' ? 'green' : 'orange'"
              >
                {{ item.status === 'completed' ? '已完成' : '进行中' }}
              </a-tag>
              <span class="muted">{{ item.time }}</span>
              <a-button size="mini" type="text" @click="viewReport(item.name)">
                查看
              </a-button>
            </a-space>
          </a-list-item>
          <a-list-item v-if="recent.length === 0">
            <span class="muted"
              >暂无分析记录，去「新建分析」开始第一次分析</span
            >
          </a-list-item>
        </a-list>
      </a-card>
    </a-space>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import { useRouter } from 'vue-router';
  import {
    getCompetitorStorageRecords,
    getHistoryRecords,
  } from '@/utils/analysis-storage';

  const router = useRouter();

  type HistoryRecord = {
    competitor?: string;
    mode?: string;
    project?: string;
    scenario?: string;
    time?: string;
  };

  type CompetitorRecord = {
    id?: string;
    name?: string;
    category?: string;
    status?: string;
    updatedAt?: string;
  };

  const historyRecords = computed<HistoryRecord[]>(() => {
    return getHistoryRecords();
  });

  const competitorRecords = computed<CompetitorRecord[]>(() => {
    return getCompetitorStorageRecords();
  });

  const stats = computed(() => {
    const history = historyRecords.value;
    const competitorNames = new Set<string>();

    competitorRecords.value.forEach((item) => {
      if (item.name?.trim()) competitorNames.add(item.name.trim());
    });
    history.forEach((item) => {
      if (item.competitor?.trim()) competitorNames.add(item.competitor.trim());
    });

    const now = new Date();
    const newThisMonth = history.filter((item) => {
      if (!item.time) return false;
      const date = new Date(item.time.replace(' ', 'T'));
      return (
        !Number.isNaN(date.getTime()) &&
        date.getFullYear() === now.getFullYear() &&
        date.getMonth() === now.getMonth()
      );
    }).length;

    const avgScore = history.length > 0 ? 8.6 : 0;

    return {
      competitors: competitorNames.size,
      analyses: history.length,
      avgScore,
      newThisMonth,
    };
  });

  const getRecentCategory = (scenario?: string) => {
    if (scenario === 'Market_Entry') return '入局研判';
    if (scenario === 'Product_Improvement') return '产品优化';
    return '竞品分析';
  };

  const recent = computed(() => {
    const records = historyRecords.value;
    if (records.length === 0) {
      return [];
    }
    return records.slice(0, 8).map((r: HistoryRecord, i: number) => ({
      id: `h_${i}`,
      name: r.competitor || '未命名竞品',
      category: getRecentCategory(r.scenario),
      status: 'completed' as const,
      time: r.time || '',
    }));
  });

  const go = (path: string) => {
    router.push(path);
  };

  const viewReport = (name: string) => {
    router.push({
      name: 'analysisReportView',
      query: { competitor: name },
    });
  };
</script>

<script lang="ts">
  export default {
    name: 'Dashboard',
  };
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

  .name {
    font-weight: 500;
    color: var(--color-text-1);
  }

  .muted {
    color: var(--color-text-3);
    font-size: 12px;
  }
</style>
