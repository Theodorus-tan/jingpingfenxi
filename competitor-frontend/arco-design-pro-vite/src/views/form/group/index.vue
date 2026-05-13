<template>
  <div class="page">
    <a-space direction="vertical" :size="16" fill>
      <a-card class="card" title="新建分析">
        <a-form :model="form" layout="vertical">
          <a-form-item
            field="competitor"
            label="竞品"
            :rules="[{ required: true, message: '请输入或选择竞品' }]"
          >
            <a-input
              v-model="form.competitor"
              allow-clear
              placeholder="例如：DJI Osmo Action 4"
            />
          </a-form-item>
          <a-form-item field="mode" label="分析引擎">
            <a-radio-group v-model="form.mode" type="button">
              <a-radio value="python">Python Agent（流式）</a-radio>
              <a-radio value="eino">Eino Agent（快速）</a-radio>
            </a-radio-group>
          </a-form-item>
          <a-space :size="12">
            <a-button type="primary" :loading="running" @click="start"
              >开始分析</a-button
            >
            <a-button :disabled="running" @click="reset">重置</a-button>
          </a-space>
        </a-form>
      </a-card>

      <a-card v-if="running || finished" class="card" title="分析师工作台">
        <div class="agent-logs">
          <div v-for="(log, idx) in agentLogs" :key="idx" class="agent-log">
            <template v-if="log.type === 'thinking'">
              <icon-send class="icon thinking" />
              <span class="thinking-text">{{ log.message }}</span>
            </template>
            <template v-else-if="log.type === 'searching'">
              <icon-search class="icon searching" />
              <span class="search-text">搜索：</span>
              <span class="query-text">{{ log.query }}</span>
            </template>
            <template v-else-if="log.type === 'search_result'">
              <icon-check-circle class="icon result" />
              <span class="result-text"
                >找到 {{ log.results_count }} 条结果</span
              >
              <div v-if="log.titles && log.titles.length" class="result-titles">
                <div
                  v-for="(t, ti) in log.titles"
                  :key="ti"
                  class="result-title"
                >
                  {{ t }}
                </div>
              </div>
            </template>
            <template v-else-if="log.type === 'writing'">
              <icon-edit class="icon writing" />
              <span class="writing-text">{{ log.message }}</span>
            </template>
          </div>
          <div v-if="running" class="agent-log cursor">
            <icon-loading class="icon" />
            <span class="waiting-text">分析中...</span>
          </div>
        </div>

        <a-divider v-if="finished" />
        <a-space v-if="finished" :size="12">
          <a-button type="primary" @click="goReport">查看报告</a-button>
        </a-space>
      </a-card>
    </a-space>
  </div>
</template>

<script setup lang="ts">
  import { reactive, ref } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { Message } from '@arco-design/web-vue';
  import {
    IconSend,
    IconSearch,
    IconCheckCircle,
    IconEdit,
    IconLoading,
  } from '@arco-design/web-vue/es/icon';

  const router = useRouter();
  const route = useRoute();

  const form = reactive({
    competitor: (route.query.competitor as string) || '',
    mode: (route.query.mode as string) || 'python',
  });

  const running = ref(false);
  const finished = ref(false);
  const currentReport = ref('');
  const agentLogs = ref<
    Array<{
      type: string;
      message?: string;
      query?: string;
      results_count?: number;
      titles?: string[];
    }>
  >([]);

  const reset = () => {
    form.competitor = '';
    running.value = false;
    finished.value = false;
    currentReport.value = '';
    agentLogs.value = [];
  };

  const saveHistory = (name: string) => {
    try {
      const history = JSON.parse(
        localStorage.getItem('analysis_history') || '[]'
      );
      history.unshift({
        competitor: name,
        mode: form.mode,
        time: new Date().toISOString().slice(0, 16).replace('T', ' '),
      });
      if (history.length > 20) history.length = 20;
      localStorage.setItem('analysis_history', JSON.stringify(history));
    } catch {
      // ignore
    }
  };

  const start = async () => {
    if (!form.competitor.trim()) {
      Message.warning('请输入竞品名称');
      return;
    }

    running.value = true;
    finished.value = false;
    currentReport.value = '';
    agentLogs.value = [];

    if (form.mode === 'eino') {
      const { submitAnalysisTaskEino } = await import('@/api/analysis');
      await submitAnalysisTaskEino(
        { competitor_name: form.competitor },
        (event) => {
          agentLogs.value.push({ ...event });
        },
        (report) => {
          currentReport.value = report;
          running.value = false;
          finished.value = true;
          localStorage.setItem('latest_analysis_report', report);
          localStorage.setItem('latest_analysis_competitor', form.competitor);
          saveHistory(form.competitor);
          Message.success('分析完成！');
        },
        (err) => {
          running.value = false;
          agentLogs.value.push({ type: 'thinking', message: `出错：${err}` });
          Message.error(`分析失败：${err}`);
        }
      );
      return;
    }

    const { submitAnalysisTaskStream } = await import('@/api/analysis');
    await submitAnalysisTaskStream(
      { competitor_name: form.competitor },
      (event) => {
        agentLogs.value.push({ ...event });
      },
      (report) => {
        currentReport.value = report;
        running.value = false;
        finished.value = true;
        localStorage.setItem('latest_analysis_report', report);
        localStorage.setItem('latest_analysis_competitor', form.competitor);
        saveHistory(form.competitor);
        Message.success('分析完成！');
      },
      (err) => {
        running.value = false;
        agentLogs.value.push({ type: 'thinking', message: `出错：${err}` });
        Message.error(`分析失败：${err}`);
      }
    );
  };

  const goReport = () => {
    router.push({
      name: 'analysisReportView',
      query: { competitor: form.competitor },
    });
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

  .agent-logs {
    display: flex;
    flex-direction: column;
    gap: 12px;
    max-height: 400px;
    overflow-y: auto;
  }

  .agent-log {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
    gap: 6px;
    padding: 6px 0;
    border-bottom: 1px solid var(--color-fill-2);
    font-size: 14px;
    line-height: 1.6;

    &.cursor {
      opacity: 0.7;
    }

    .icon {
      font-size: 16px;
      flex-shrink: 0;
      margin-top: 2px;

      &.thinking {
        color: rgb(var(--arcoblue-6));
      }
      &.searching {
        color: rgb(var(--orange-6));
      }
      &.result {
        color: rgb(var(--green-6));
      }
      &.writing {
        color: rgb(var(--purple-6));
      }
    }

    .thinking-text {
      color: var(--color-text-2);
    }
    .search-text {
      color: var(--color-text-3);
    }
    .query-text {
      color: var(--color-text-1);
      font-weight: 500;
    }
    .result-text {
      color: var(--color-text-2);
    }
    .writing-text {
      color: var(--color-text-2);
    }
    .waiting-text {
      color: var(--color-text-3);
    }

    .result-titles {
      width: 100%;
      padding-left: 22px;
      margin-top: 4px;
      display: flex;
      flex-direction: column;
      gap: 2px;
    }

    .result-title {
      font-size: 13px;
      color: var(--color-text-3);
      padding: 2px 8px;
      background: var(--color-fill-1);
      border-radius: 4px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
</style>
