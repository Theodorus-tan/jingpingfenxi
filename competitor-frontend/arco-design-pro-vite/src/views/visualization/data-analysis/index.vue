<template>
  <div class="page">
    <div class="workspace-toolbar">
      <a-tag color="arcoblue">聊天请使用全局助手入口</a-tag>
    </div>

    <div ref="workspaceRef" class="workspace-shell">
      <section class="workspace-pane report-pane full-width">
        <a-space id="report-container" direction="vertical" :size="16" fill>
          <a-card class="card" title="3D 独立分析工程看板">
            <template #extra>
              <a-button type="outline" size="small" @click="exportPDF">
                <template #icon><icon-download /></template>
                导出 PDF
              </a-button>
            </template>
            <a-space direction="vertical" :size="8" fill>
              <div class="headline">
                <span class="title">{{ competitorName }}</span>
                <a-tag color="arcoblue" size="small">AI 生成报告</a-tag>
                <a-tag
                  v-if="scenarioLabel"
                  :color="
                    scenario === 'Product_Improvement' ? 'orange' : 'green'
                  "
                  size="small"
                  >场景: {{ scenarioLabel }}</a-tag
                >
              </div>
              <div class="meta-line">
                <span>所属选品项目：{{ project || '未填写' }}</span>
                <span>生成时间：{{ generatedAt }}</span>
              </div>
            </a-space>
          </a-card>

          <a-card
            class="card"
            style="border-left: 4px solid rgb(var(--arcoblue-6))"
          >
            <template #title>
              <span style="color: rgb(var(--arcoblue-6)); font-weight: 600">
                <icon-bulb /> 分析解读
              </span>
            </template>
            <div class="advice-body">
              <template v-if="scenario === 'Product_Improvement'">
                <p>
                  这份报告优先用于找出竞品的功能短板、用户抱怨点和可借鉴亮点，帮助你形成下一轮产品迭代和卖点重写。
                </p>
              </template>
              <template v-else-if="scenario === 'Market_Entry'">
                <p>
                  这份报告优先用于判断是否值得入局，重点看市场缝隙、渠道壁垒、品牌势能和长期经营风险。
                </p>
              </template>
              <template v-else>
                <p>
                  当前报告未携带明确分析场景，建议结合三大维度交叉阅读，再决定是做产品优化还是市场进入判断。
                </p>
              </template>
            </div>
          </a-card>

          <a-card class="card" title="综述报告">
            <a-alert type="info" style="margin-bottom: 16px">
              当前报告默认以
              <strong>产品 Review 诊断</strong>
              为主视角汇总成一份精炼综述，宏观与财务内容只作为补充判断，不再并列干扰主阅读链路。
            </a-alert>
            <a-alert
              v-if="!hasStructuredReport"
              type="warning"
              style="margin-bottom: 16px"
            >
              当前为旧版报告格式，页面已自动兼容展示；后续新生成的报告将默认包含“综述
              + 三维原稿”结构。
            </a-alert>

            <div class="summary-section">
              <div class="summary-section-head">
                <span class="summary-index">01</span>
                <div>
                  <div class="summary-title">精炼综述</div>
                  <div class="summary-desc">
                    先讲结论，再讲痛点、优势、可攻击点和可执行动作，默认只保留一份主报告。
                  </div>
                </div>
              </div>
              <div
                v-if="reportSummaryHtml"
                class="markdown-body"
                v-html="reportSummaryHtml"
              ></div>
              <a-empty v-else description="暂无综述报告内容" />
            </div>

            <a-collapse
              v-if="hasRawDimensionReports"
              class="raw-dimension-collapse"
              :default-active-key="[]"
            >
              <a-collapse-item key="raw-dimensions" header="查看三维原始稿">
                <a-tabs type="rounded" size="large">
                  <a-tab-pane key="review" title="产品 Review 诊断">
                    <div class="tab-content">
                      <a-alert type="info" style="margin-bottom: 16px">
                        本维度由
                        <strong>Review 分析师 Agent</strong>
                        驱动，抓取电商与社交媒体反馈，提炼痛点与亮点。
                      </a-alert>
                      <div
                        v-if="reports.review"
                        class="markdown-body"
                        v-html="reports.review"
                      ></div>
                      <a-empty v-else description="暂无分析数据" />
                    </div>
                  </a-tab-pane>
                  <a-tab-pane key="macro" title="宏观商业战略">
                    <div class="tab-content">
                      <a-alert type="info" style="margin-bottom: 16px">
                        本维度由
                        <strong>战略分析师 Agent</strong>
                        驱动，抓取百科与官网，拆解商业模式与发展壁垒。
                      </a-alert>
                      <div
                        v-if="reports.macro"
                        class="markdown-body"
                        v-html="reports.macro"
                      ></div>
                      <a-empty v-else description="战略分析维度暂无独立数据" />
                    </div>
                  </a-tab-pane>
                  <a-tab-pane key="finance" title="财务与生命力评估">
                    <div class="tab-content">
                      <a-alert type="warning" style="margin-bottom: 16px">
                        本维度由
                        <strong>金融分析师 Agent</strong>
                        驱动，具有防幻觉熔断机制。如遇未上市企业，将自动降级或折叠。
                      </a-alert>
                      <div
                        v-if="reports.finance"
                        class="markdown-body"
                        v-html="reports.finance"
                      ></div>
                      <a-empty v-else description="财务分析维度暂无独立数据" />
                    </div>
                  </a-tab-pane>
                </a-tabs>
              </a-collapse-item>
            </a-collapse>
          </a-card>

          <a-card class="card" title="关键证据链">
            <a-alert type="info" style="margin-bottom: 16px">
              每条关键结论都应尽量能回到原始来源，这里统一展示来源名、摘要与跳转链接。
            </a-alert>
            <a-empty
              v-if="!displayEvidencePool.length"
              description="当前暂无可展示的证据链"
            />
            <div v-else class="report-evidence-grid">
              <div
                v-for="(evidence, evidenceIndex) in displayEvidencePool"
                :key="`${evidence.dimension}-${evidence.title}-${evidenceIndex}`"
                class="report-evidence-card"
              >
                <div class="report-evidence-head">
                  <a-tag
                    :color="
                      dimensionMeta[getTimelineDimension(evidence.dimension)]
                        .color
                    "
                    size="small"
                  >
                    {{
                      dimensionMeta[getTimelineDimension(evidence.dimension)]
                        .label
                    }}
                  </a-tag>
                  <span class="report-evidence-index"
                    >证据 {{ evidenceIndex + 1 }}</span
                  >
                </div>
                <div class="report-evidence-title">{{ evidence.title }}</div>
                <div v-if="evidence.snippet" class="report-evidence-snippet">
                  {{ evidence.snippet }}
                </div>
                <a
                  v-if="evidence.url"
                  class="report-evidence-link"
                  :href="evidence.url"
                  target="_blank"
                  rel="noreferrer"
                >
                  查看原文
                </a>
              </div>
            </div>
          </a-card>

          <a-card class="card" title="竞品图片与截图">
            <a-empty
              v-if="!imageLoading && !imageCandidates.length"
              description="暂未获取到图片候选"
            />
            <div v-else class="image-grid">
              <a-skeleton
                v-if="imageLoading"
                :loading="true"
                animation
                class="image-skeleton"
              >
                <a-skeleton-line :rows="4" />
              </a-skeleton>
              <div
                v-for="(image, imageIndex) in imageCandidates"
                v-else
                :key="`${image.page_url}-${imageIndex}`"
                class="image-card"
              >
                <img
                  :src="image.image_url"
                  :alt="image.title"
                  class="image-preview"
                />
                <div class="image-card-body">
                  <div class="image-card-title">{{ image.title }}</div>
                  <div class="image-card-desc">
                    {{
                      image.snippet ||
                      '该图片可作为竞品主图/卖点图候选，建议结合综述结论继续补图注。'
                    }}
                  </div>
                  <a
                    class="image-card-link"
                    :href="image.page_url"
                    target="_blank"
                    rel="noreferrer"
                  >
                    查看来源页面
                  </a>
                </div>
              </div>
            </div>
          </a-card>
        </a-space>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue';
  import { useRoute } from 'vue-router';
  import { marked } from 'marked';
  import { Message } from '@arco-design/web-vue';
  import { IconBulb, IconDownload } from '@arco-design/web-vue/es/icon';
  import {
    fetchCompetitorImages,
    type CompetitorImageCandidate,
  } from '@/api/analysis';
  import { parseAnalysisReport } from '@/utils/analysis-report';
  import { exportReportElementToPdf } from '@/utils/report-export';
  import {
    type AgentLogRecord,
    getLatestAnalysisLogs,
    getLatestAnalysisRecord,
  } from '@/utils/analysis-storage';

  const route = useRoute();

  const competitorName = ref('');
  const scenario = ref('');
  const project = ref('');
  const generatedAt = ref('');
  const rawReportContext = ref('');
  const hasStructuredReport = ref(false);
  const reportSummaryHtml = ref('');
  const reportEvidencePool = ref<
    Array<{ dimension: string; title: string; snippet: string; url: string }>
  >([]);
  const reports = ref({
    review: '',
    macro: '',
    finance: '',
  });
  const executionLogs = ref<AgentLogRecord[]>([]);
  const imageCandidates = ref<CompetitorImageCandidate[]>([]);
  const imageLoading = ref(false);

  const scenarioLabel = computed(() => {
    if (scenario.value === 'Product_Improvement') return '已有产品求改进';
    if (scenario.value === 'Market_Entry') return '无产品求入局';
    return '';
  });

  const dimensionMeta = {
    master: { label: 'Master', color: 'orangered' },
    review: { label: 'Review', color: 'arcoblue' },
    macro: { label: '宏观', color: 'purple' },
    finance: { label: '财务', color: 'green' },
  } as const;

  type TimelineDimension = 'master' | 'review' | 'macro' | 'finance';

  const getTimelineDimension = (dimension?: string): TimelineDimension => {
    if (dimension === 'review') return 'review';
    if (dimension === 'macro') return 'macro';
    if (dimension === 'finance') return 'finance';
    return 'master';
  };

  const hasRawDimensionReports = computed(() =>
    Boolean(
      reports.value.review || reports.value.macro || reports.value.finance
    )
  );

  const getLogEvidences = (log: AgentLogRecord) => {
    if (log.evidences?.length) return log.evidences;
    return (log.titles || []).map((title) => ({ title, snippet: '', url: '' }));
  };

  const displayEvidencePool = computed(() => {
    if (reportEvidencePool.value.length) {
      return reportEvidencePool.value;
    }

    return executionLogs.value
      .flatMap((log) =>
        getLogEvidences(log).map((evidence) => ({
          dimension: log.dimension || 'master',
          title: evidence.title,
          snippet: evidence.snippet || '',
          url: evidence.url || '',
        }))
      )
      .filter((evidence) => evidence.title)
      .slice(0, 12);
  });

  const exportPDF = async () => {
    const reportElement = document.getElementById('report-container');
    if (!reportElement) {
      Message.error('未找到可导出的报告内容');
      return;
    }

    try {
      await exportReportElementToPdf(
        reportElement,
        competitorName.value,
        generatedAt.value
      );
      Message.success('PDF 导出成功');
    } catch (error) {
      Message.error(`PDF 导出失败: ${String(error)}`);
    }
  };

  const loadCompetitorImages = async () => {
    if (!competitorName.value) return;
    imageLoading.value = true;
    try {
      imageCandidates.value = await fetchCompetitorImages(competitorName.value);
    } catch (error) {
      imageCandidates.value = [];
    } finally {
      imageLoading.value = false;
    }
  };

  onMounted(async () => {
    const latestAnalysis = getLatestAnalysisRecord();
    executionLogs.value = getLatestAnalysisLogs();
    // 优先取 url 参数，如果没有，取 localstorage
    competitorName.value =
      (route.query.competitor as string) ||
      latestAnalysis.competitor ||
      '未知竞品';
    scenario.value =
      (route.query.scenario as string) || latestAnalysis.scenario || '';
    project.value =
      (route.query.project as string) || latestAnalysis.project || '';

    generatedAt.value = new Date().toISOString().slice(0, 16).replace('T', ' ');

    const rawReport = latestAnalysis.report;
    if (rawReport) {
      const parsedReport = parseAnalysisReport(rawReport);
      rawReportContext.value = parsedReport.rawReportContext;
      hasStructuredReport.value = parsedReport.hasStructuredReport;
      reportSummaryHtml.value = await marked.parse(parsedReport.summary);
      reportEvidencePool.value = parsedReport.evidencePool;

      if (parsedReport.review) {
        reports.value.review = await marked.parse(parsedReport.review);
      }
      if (parsedReport.macro) {
        reports.value.macro = await marked.parse(parsedReport.macro);
      }
      if (parsedReport.finance) {
        reports.value.finance = await marked.parse(parsedReport.finance);
      }
    }

    await loadCompetitorImages();
  });
</script>

<style lang="less" scoped>
  .page {
    background: var(--color-fill-2);
    padding: 16px 20px;
    min-height: calc(100vh - 120px);
  }

  .workspace-toolbar {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    margin-bottom: 12px;
  }

  .workspace-shell {
    min-height: calc(100vh - 184px);
  }

  .workspace-pane {
    min-width: 0;
  }

  .report-pane {
    display: flex;
    flex-direction: column;
  }

  .full-width {
    width: 100%;
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

  .meta-line {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .advice-body {
    font-size: 14px;
    line-height: 1.7;
    color: var(--color-text-2);
  }

  .summary-section {
    padding: 18px 0 8px;
    border-top: 1px solid var(--color-border-2);
  }

  .summary-section:first-of-type {
    padding-top: 4px;
    border-top: none;
  }

  .summary-section-secondary {
    margin-top: 12px;
  }

  .summary-section-head {
    display: flex;
    gap: 14px;
    align-items: flex-start;
    margin-bottom: 16px;
  }

  .summary-index {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border-radius: 12px;
    background: linear-gradient(180deg, #ebf3ff 0%, #dfeaff 100%);
    color: rgb(var(--arcoblue-6));
    font-size: 14px;
    font-weight: 700;
    flex-shrink: 0;
  }

  .summary-title {
    color: var(--color-text-1);
    font-size: 18px;
    font-weight: 600;
    line-height: 1.4;
  }

  .summary-desc {
    margin-top: 4px;
    color: var(--color-text-3);
    font-size: 13px;
    line-height: 1.6;
  }

  .sub-insight-block + .sub-insight-block {
    margin-top: 20px;
  }

  .sub-insight-title {
    margin-bottom: 12px;
    color: var(--color-text-1);
    font-size: 15px;
    font-weight: 600;
  }

  .image-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 16px;
  }

  .report-evidence-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 14px;
  }

  .report-evidence-card {
    padding: 16px;
    border: 1px solid rgba(173, 181, 193, 0.18);
    border-radius: 16px;
    background: rgba(255, 255, 255, 0.9);
    box-shadow: 0 10px 24px rgba(149, 157, 165, 0.1);
  }

  .report-evidence-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  .report-evidence-index {
    color: var(--color-text-3);
    font-size: 12px;
  }

  .report-evidence-title {
    margin-top: 12px;
    color: var(--color-text-1);
    font-size: 14px;
    font-weight: 600;
    line-height: 1.6;
  }

  .report-evidence-snippet {
    margin-top: 8px;
    color: var(--color-text-2);
    font-size: 12px;
    line-height: 1.7;
  }

  .report-evidence-link {
    display: inline-flex;
    margin-top: 10px;
    color: rgb(var(--arcoblue-6));
    font-size: 12px;
    text-decoration: none;
  }

  .image-skeleton {
    grid-column: 1 / -1;
  }

  .image-card {
    overflow: hidden;
    border: 1px solid rgba(173, 181, 193, 0.18);
    border-radius: 16px;
    background: #fff;
    box-shadow: 0 12px 28px rgba(149, 157, 165, 0.12);
  }

  .image-preview {
    display: block;
    width: 100%;
    height: 180px;
    object-fit: cover;
    background: rgba(245, 247, 250, 0.8);
  }

  .image-card-body {
    padding: 14px;
  }

  .image-card-title {
    color: var(--color-text-1);
    font-size: 14px;
    font-weight: 600;
    line-height: 1.5;
  }

  .image-card-desc {
    margin-top: 8px;
    color: var(--color-text-2);
    font-size: 12px;
    line-height: 1.7;
  }

  .image-card-link {
    display: inline-flex;
    margin-top: 10px;
    color: rgb(var(--arcoblue-6));
    font-size: 12px;
    text-decoration: none;
  }

  .raw-dimension-collapse {
    margin-top: 20px;
  }

  .tab-content {
    padding: 8px 0;
  }

  /* 简单的 Markdown 样式适配 */
  .markdown-body {
    font-size: 14px;
    line-height: 1.8;
    color: var(--color-text-2);

    :deep(h1),
    :deep(h2),
    :deep(h3),
    :deep(h4) {
      color: var(--color-text-1);
      margin-top: 24px;
      margin-bottom: 16px;
      font-weight: 600;
    }

    :deep(h1) {
      font-size: 22px;
    }
    :deep(h2) {
      font-size: 18px;
    }
    :deep(h3) {
      font-size: 16px;
    }

    :deep(p) {
      margin-bottom: 16px;
    }

    :deep(ul),
    :deep(ol) {
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

  /* PDF Print Styles */
  @media print {
    body * {
      visibility: hidden;
    }
    #report-container,
    #report-container * {
      visibility: visible;
    }
    #report-container {
      position: absolute;
      left: 0;
      top: 0;
      width: 100%;
    }
    .chat-sider {
      display: none !important;
    }
    .arco-tabs-nav {
      display: none !important;
    }
    .arco-tabs-content {
      padding-top: 0 !important;
    }
    .tab-content {
      page-break-inside: avoid;
    }
    .arco-alert {
      display: none !important;
    }
    .arco-btn {
      display: none !important;
    }
  }
</style>
