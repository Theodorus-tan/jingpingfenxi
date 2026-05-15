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
          <a-form-item field="mode" label="底层执行引擎 (Agent Engine)">
            <a-radio-group v-model="form.mode" type="button">
              <a-radio value="eino">「观潮」Agent (多路并发3D分析)</a-radio>
              <a-radio value="python">Python Agent (传统单线流式)</a-radio>
            </a-radio-group>
          </a-form-item>
          <a-form-item field="project" label="所属选品项目">
            <a-input
              v-model="form.project"
              allow-clear
              placeholder="例如：运动相机出海项目"
            />
          </a-form-item>
          <a-form-item field="scenario" label="分析场景">
            <a-radio-group v-model="form.scenario" type="button">
              <a-radio value="Product_Improvement">已有产品求改进</a-radio>
              <a-radio value="Market_Entry">无产品求入局</a-radio>
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

      <a-card
        v-if="running || finished"
        class="card"
        :title="form.mode === 'eino' ? '多 Agent 实时执行过程' : '实时执行过程'"
      >
        <a-tabs type="capsule" size="small">
          <a-tab-pane
            v-for="dimension in dimensionOrder"
            :key="dimension"
            :title="`${dimensionMeta[dimension].label} (${logsByDimension[dimension].length})`"
          >
            <div class="timeline-pane">
              <a-empty
                v-if="!logsByDimension[dimension].length && !running"
                description="当前维度暂无执行记录"
              />
              <a-timeline v-else class="process-timeline">
                <a-timeline-item
                  v-for="(log, idx) in logsByDimension[dimension]"
                  :key="`${dimension}-${idx}`"
                >
                  <div class="timeline-item-head">
                    <div class="timeline-item-meta">
                      <a-tag :color="dimensionMeta[dimension].color" size="small">
                        {{ dimensionMeta[dimension].label }}
                      </a-tag>
                      <a-tag :color="getLogTagColor(log.type)" size="small">
                        {{ getLogTypeLabel(log.type) }}
                      </a-tag>
                      <span class="timeline-time">
                        {{ formatLogTime(log.timestamp) }}
                      </span>
                    </div>
                  </div>
                  <div class="timeline-item-body">
                    <div class="timeline-main">
                      {{ getLogPrimaryText(log) }}
                    </div>
                    <div
                      v-if="getLogEvidences(log).length"
                      class="timeline-titles"
                    >
                      <div
                        v-for="(evidence, titleIndex) in getLogEvidences(log)"
                        :key="titleIndex"
                        class="timeline-evidence-card"
                      >
                        <div class="timeline-evidence-index">
                          证据 {{ titleIndex + 1 }}
                        </div>
                        <div class="timeline-evidence-text">{{ evidence.title }}</div>
                        <div v-if="evidence.snippet" class="timeline-evidence-snippet">
                          {{ evidence.snippet }}
                        </div>
                        <a
                          v-if="evidence.url"
                          class="timeline-evidence-link"
                          :href="evidence.url"
                          target="_blank"
                          rel="noreferrer"
                        >
                          查看原文
                        </a>
                      </div>
                    </div>
                  </div>
                </a-timeline-item>
                <a-timeline-item v-if="running && activeTimelineDimension === dimension">
                  <div class="timeline-item-head">
                    <div class="timeline-item-meta">
                      <a-tag :color="dimensionMeta[dimension].color" size="small">
                        {{ dimensionMeta[dimension].label }}
                      </a-tag>
                      <a-tag color="gold" size="small">进行中</a-tag>
                    </div>
                  </div>
                  <div class="timeline-main">
                    正在持续接收 {{ dimensionMeta[dimension].label }} 执行事件...
                  </div>
                </a-timeline-item>
              </a-timeline>
            </div>
          </a-tab-pane>
        </a-tabs>

        <a-divider v-if="finished" />
        <a-space v-if="finished" :size="12">
          <a-button type="primary" @click="goReport">查看报告</a-button>
        </a-space>
      </a-card>
    </a-space>
  </div>
</template>

<script setup lang="ts">
  import { computed, reactive, ref, watch } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { Message } from '@arco-design/web-vue';
  import {
    type AgentLogRecord,
    appendHistoryRecord,
    getLatestAnalysisRecord,
    saveLatestAnalysisLogs,
    saveLatestAnalysisRecord,
  } from '@/utils/analysis-storage';

  const router = useRouter();
  const route = useRoute();

  const form = reactive({
    competitor: (route.query.competitor as string) || '',
    mode: (route.query.mode as string) || 'eino',
    project: (route.query.project as string) || getLatestAnalysisRecord().project,
    scenario:
      (route.query.scenario as string) ||
      getLatestAnalysisRecord().scenario ||
      'Product_Improvement',
  });

  const running = ref(false);
  const finished = ref(false);
  const currentReport = ref('');
  const agentLogs = ref<AgentLogRecord[]>([]);

  const dimensionMeta = {
    master: { label: 'Master', color: 'orangered' },
    review: { label: 'Review', color: 'arcoblue' },
    macro: { label: '宏观', color: 'purple' },
    finance: { label: '财务', color: 'green' },
  } as const;

  const dimensionOrder = ['master', 'review', 'macro', 'finance'] as const;

  type TimelineDimension = (typeof dimensionOrder)[number];

  const getTimelineDimension = (dimension?: string): TimelineDimension => {
    if (dimension === 'review') return 'review';
    if (dimension === 'macro') return 'macro';
    if (dimension === 'finance') return 'finance';
    return 'master';
  };

  const logsByDimension = computed<Record<TimelineDimension, AgentLogRecord[]>>(
    () => {
      return dimensionOrder.reduce(
        (acc, dimension) => {
          acc[dimension] = agentLogs.value.filter(
            (log) => getTimelineDimension(log.dimension) === dimension
          );
          return acc;
        },
        {
          master: [],
          review: [],
          macro: [],
          finance: [],
        } as Record<TimelineDimension, AgentLogRecord[]>
      );
    }
  );

  const activeTimelineDimension = computed<TimelineDimension>(() => {
    if (agentLogs.value.length === 0) return 'master';
    return getTimelineDimension(agentLogs.value[agentLogs.value.length - 1].dimension);
  });

  const getLogTypeLabel = (type: string) => {
    if (type === 'searching') return '检索';
    if (type === 'search_result') return '结果';
    if (type === 'thinking') return '推理';
    if (type === 'writing') return '成稿';
    if (type === 'error') return '异常';
    if (type === 'done') return '完成';
    return '事件';
  };

  const getLogTagColor = (type: string) => {
    if (type === 'searching') return 'arcoblue';
    if (type === 'search_result') return 'green';
    if (type === 'thinking') return 'purple';
    if (type === 'writing') return 'orange';
    if (type === 'error') return 'red';
    if (type === 'done') return 'green';
    return 'gray';
  };

  const getLogPrimaryText = (log: AgentLogRecord) => {
    if (log.type === 'searching') {
      return log.query ? `发起检索：${log.query}` : '发起检索';
    }
    if (log.type === 'search_result') {
      return `获取到 ${log.results_count || 0} 条搜索结果`;
    }
    return log.message || '执行中';
  };

  const getLogEvidences = (log: AgentLogRecord) => {
    if (log.evidences?.length) return log.evidences;
    return (log.titles || []).map((title) => ({ title, snippet: '', url: '' }));
  };

  const formatLogTime = (value?: string) => {
    if (!value) return '--:--:--';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return '--:--:--';
    return date.toLocaleTimeString('zh-CN', { hour12: false });
  };

  const persistAgentLogs = () => {
    saveLatestAnalysisLogs(agentLogs.value);
  };

  const pushAgentLog = (log: AgentLogRecord) => {
    agentLogs.value.push({
      ...log,
      timestamp: log.timestamp || new Date().toISOString(),
    });
    persistAgentLogs();
  };

  const reset = () => {
    form.competitor = '';
    form.project = '';
    form.scenario = 'Product_Improvement';
    form.mode = 'eino';
    running.value = false;
    finished.value = false;
    currentReport.value = '';
    agentLogs.value = [];
    persistAgentLogs();
  };

  watch(
    () => route.query,
    (query) => {
      if (typeof query.competitor === 'string') {
        form.competitor = query.competitor;
      }
      if (typeof query.project === 'string') {
        form.project = query.project;
      }
      if (typeof query.scenario === 'string') {
        form.scenario = query.scenario;
      }
    },
    { deep: true }
  );

  const saveHistory = (name: string) => {
    appendHistoryRecord({
      competitor: name,
      mode: form.mode,
      project: form.project,
      scenario: form.scenario,
      time: new Date().toISOString().slice(0, 16).replace('T', ' '),
    });
  };

  const persistLatestAnalysisMeta = (report: string) => {
    saveLatestAnalysisRecord({
      report,
      competitor: form.competitor,
      project: form.project,
      scenario: form.scenario,
    });
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
    persistAgentLogs();

    if (form.mode === 'eino') {
      const { submitAnalysisTaskEino } = await import('@/api/analysis');
      await submitAnalysisTaskEino(
        {
          competitor_name: form.competitor,
          project: form.project,
          scenario: form.scenario,
        },
        (event) => {
          pushAgentLog({ ...event });
        },
        (report) => {
          currentReport.value = report;
          running.value = false;
          finished.value = true;
          persistLatestAnalysisMeta(report);
          saveHistory(form.competitor);
          Message.success('分析完成！');
        },
        (err) => {
          running.value = false;
          pushAgentLog({ type: 'error', message: `出错：${err}` });
          Message.error(`分析失败：${err}`);
        }
      );
      return;
    }

    const { submitAnalysisTaskStream } = await import('@/api/analysis');
    await submitAnalysisTaskStream(
      {
        competitor_name: form.competitor,
        project: form.project,
        scenario: form.scenario,
      },
      (event) => {
        pushAgentLog({ ...event });
      },
      (report) => {
        currentReport.value = report;
        running.value = false;
        finished.value = true;
        persistLatestAnalysisMeta(report);
        saveHistory(form.competitor);
        Message.success('分析完成！');
      },
      (err) => {
        running.value = false;
        pushAgentLog({ type: 'error', message: `出错：${err}` });
        Message.error(`分析失败：${err}`);
      }
    );
  };

  const goReport = () => {
    router.push({
      name: 'analysisReportView',
      query: {
        competitor: form.competitor,
        project: form.project,
        scenario: form.scenario,
      },
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

  .timeline-pane {
    min-height: 280px;
    max-height: 520px;
    overflow: auto;
    padding-right: 4px;
  }

  .process-timeline {
    padding-top: 4px;
  }

  .timeline-item-head {
    margin-bottom: 8px;
  }

  .timeline-item-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .timeline-time {
    color: var(--color-text-3);
    font-size: 12px;
    line-height: 1;
  }

  .timeline-item-body {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .timeline-main {
    color: var(--color-text-1);
    font-size: 13px;
    line-height: 1.6;
  }

  .timeline-titles {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .timeline-evidence-card {
    padding: 10px 12px;
    border: 1px solid var(--color-border-2);
    border-radius: 10px;
    background: linear-gradient(180deg, #f7faff 0%, #f2f6ff 100%);
    box-shadow: 0 4px 14px rgba(17, 36, 74, 0.05);
  }

  .timeline-evidence-index {
    margin-bottom: 6px;
    color: rgb(var(--arcoblue-6));
    font-size: 12px;
    font-weight: 600;
    line-height: 1.2;
  }

  .timeline-evidence-text {
    color: var(--color-text-2);
    font-size: 12px;
    line-height: 1.6;
  }
  .timeline-evidence-snippet {
    margin-top: 6px;
    color: var(--color-text-3);
    font-size: 12px;
    line-height: 1.6;
  }
  .timeline-evidence-link {
    display: inline-flex;
    margin-top: 8px;
    color: rgb(var(--arcoblue-6));
    font-size: 12px;
    text-decoration: none;
  }
</style>
