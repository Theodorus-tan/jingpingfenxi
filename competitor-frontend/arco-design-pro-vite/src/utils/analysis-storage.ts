export type AnalysisScenario = 'Product_Improvement' | 'Market_Entry' | '';

export type HistoryRecord = {
  competitor?: string;
  mode?: string;
  project?: string;
  scenario?: AnalysisScenario | string;
  time?: string;
};

export type LatestAnalysisRecord = {
  report: string;
  competitor: string;
  project: string;
  scenario: AnalysisScenario | string;
};

export type AnalysisWorkspaceRecord = {
  reportPaneWidth: number;
  processPaneWidth: number;
  chatPaneWidth: number;
  assistantVisible: boolean;
};

export type UploadedAnalysisMaterialRecord = {
  id: string;
  name: string;
  content: string;
  mimeType: string;
  uploadedAt: string;
};

export type AgentLogRecord = {
  type: string;
  dimension?: string;
  message?: string;
  query?: string;
  results_count?: number;
  titles?: string[];
  evidences?: Array<{
    title: string;
    snippet?: string;
    url?: string;
  }>;
  timestamp?: string;
};

export type CompetitorStorageRecord = {
  id?: string;
  name?: string;
  category?: string;
  status?: string;
  project?: string;
  scenario?: AnalysisScenario | string;
  analysisCount?: number;
  updatedAt?: string;
};

const KEYS = {
  history: 'analysis_history',
  competitorList: 'competitor_list',
  latestReport: 'latest_analysis_report',
  latestCompetitor: 'latest_analysis_competitor',
  latestProject: 'latest_analysis_project',
  latestScenario: 'latest_analysis_scenario',
  latestLogs: 'latest_analysis_logs',
  workspace: 'analysis_workspace',
  materials: 'analysis_uploaded_materials',
  assistantVisible: 'global_assistant_visible',
} as const;

function readJSON<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key);
    return raw ? (JSON.parse(raw) as T) : fallback;
  } catch {
    return fallback;
  }
}

function writeJSON<T>(key: string, value: T) {
  localStorage.setItem(key, JSON.stringify(value));
}

export function getHistoryRecords(): HistoryRecord[] {
  return readJSON<HistoryRecord[]>(KEYS.history, []);
}

export function saveHistoryRecords(records: HistoryRecord[]) {
  writeJSON(KEYS.history, records);
}

export function appendHistoryRecord(record: HistoryRecord) {
  const history = getHistoryRecords();
  history.unshift(record);
  if (history.length > 20) history.length = 20;
  saveHistoryRecords(history);
}

export function getCompetitorStorageRecords(): CompetitorStorageRecord[] {
  return readJSON<CompetitorStorageRecord[]>(KEYS.competitorList, []);
}

export function saveCompetitorStorageRecords(records: CompetitorStorageRecord[]) {
  writeJSON(KEYS.competitorList, records);
}

export function getLatestAnalysisRecord(): LatestAnalysisRecord {
  return {
    report: localStorage.getItem(KEYS.latestReport) || '',
    competitor: localStorage.getItem(KEYS.latestCompetitor) || '',
    project: localStorage.getItem(KEYS.latestProject) || '',
    scenario: (localStorage.getItem(KEYS.latestScenario) || '') as AnalysisScenario,
  };
}

export function saveLatestAnalysisRecord(record: LatestAnalysisRecord) {
  localStorage.setItem(KEYS.latestReport, record.report);
  localStorage.setItem(KEYS.latestCompetitor, record.competitor);
  localStorage.setItem(KEYS.latestProject, record.project);
  localStorage.setItem(KEYS.latestScenario, record.scenario);
}

export function getLatestAnalysisLogs(): AgentLogRecord[] {
  return readJSON<AgentLogRecord[]>(KEYS.latestLogs, []);
}

export function saveLatestAnalysisLogs(records: AgentLogRecord[]) {
  writeJSON(KEYS.latestLogs, records);
}

export function getAnalysisWorkspaceRecord(): AnalysisWorkspaceRecord | null {
  return readJSON<AnalysisWorkspaceRecord | null>(KEYS.workspace, null);
}

export function saveAnalysisWorkspaceRecord(record: AnalysisWorkspaceRecord) {
  writeJSON(KEYS.workspace, record);
}

export function getUploadedAnalysisMaterials(): UploadedAnalysisMaterialRecord[] {
  return readJSON<UploadedAnalysisMaterialRecord[]>(KEYS.materials, []);
}

export function saveUploadedAnalysisMaterials(
  records: UploadedAnalysisMaterialRecord[]
) {
  writeJSON(KEYS.materials, records);
}

export function getGlobalAssistantVisible() {
  return readJSON<boolean>(KEYS.assistantVisible, false);
}

export function saveGlobalAssistantVisible(visible: boolean) {
  writeJSON(KEYS.assistantVisible, visible);
}
