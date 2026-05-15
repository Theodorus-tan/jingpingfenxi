import type { AnalysisScenario } from './analysis-storage';

export type ParsedAssistantIntent = {
  competitor: string;
  project: string;
  scenario: AnalysisScenario;
  shouldStartAnalysis: boolean;
};

const PROJECT_PATTERN = /项目[:：\s]*([^\n，。；,;]+)/;
const COMPETITOR_PATTERNS = [
  /分析(?:一下|下)?\s*([A-Za-z0-9\u4e00-\u9fa5\s\-+]+?)(?:的|，|。|$)/,
  /竞品[:：]\s*([A-Za-z0-9\u4e00-\u9fa5\s\-+]+?)(?:，|。|$)/,
  /看一下\s*([A-Za-z0-9\u4e00-\u9fa5\s\-+]+?)(?:的|，|。|$)/,
];
const ANALYSIS_TRIGGER_PATTERN =
  /(分析|看一下|看下|研究一下|研究下|评估一下|评估下|拆解一下|拆解下|怎么改进|如何改进|值得入局|值不值得入局|进入市场|做这个品类)/;

function normalizeSegment(value: string) {
  return value.replace(/\s+/g, ' ').trim();
}

function extractByPatterns(input: string, patterns: RegExp[]) {
  for (let index = 0; index < patterns.length; index += 1) {
    const match = input.match(patterns[index]);
    if (match?.[1]) return normalizeSegment(match[1]);
  }
  return '';
}

export function parseAssistantIntent(message: string): ParsedAssistantIntent {
  const normalized = normalizeSegment(message);
  const lower = normalized.toLowerCase();

  const scenario: AnalysisScenario =
    lower.includes('入局') || lower.includes('进入市场') || lower.includes('做这个品类')
      ? 'Market_Entry'
      : 'Product_Improvement';

  const projectMatch = normalized.match(PROJECT_PATTERN);
  const project = projectMatch?.[1] ? normalizeSegment(projectMatch[1]) : '';
  const competitor = extractByPatterns(normalized, COMPETITOR_PATTERNS);

  const shouldStartAnalysis =
    Boolean(competitor) && ANALYSIS_TRIGGER_PATTERN.test(normalized);

  return {
    competitor,
    project,
    scenario,
    shouldStartAnalysis,
  };
}

export function buildIntentSummary(intent: ParsedAssistantIntent) {
  const scenarioLabel =
    intent.scenario === 'Market_Entry' ? '无产品求入局' : '已有产品求改进';

  return [
    `识别到竞品：${intent.competitor || '未识别'}`,
    `识别到项目：${intent.project || '未识别'}`,
    `识别到场景：${scenarioLabel}`,
  ].join('\n');
}
