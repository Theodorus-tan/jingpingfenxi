export type AnalysisWorkspaceState = {
  reportPaneWidth: number;
  processPaneWidth: number;
  chatPaneWidth: number;
  assistantVisible: boolean;
};

export const DEFAULT_WORKSPACE_STATE: AnalysisWorkspaceState = {
  reportPaneWidth: 52,
  processPaneWidth: 20,
  chatPaneWidth: 28,
  assistantVisible: true,
};

const MIN_REPORT_WIDTH = 38;
const MIN_PROCESS_WIDTH = 16;
const MIN_CHAT_WIDTH = 24;

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max);
}

export function normalizeWorkspaceState(
  input?: Partial<AnalysisWorkspaceState>
): AnalysisWorkspaceState {
  const next = {
    ...DEFAULT_WORKSPACE_STATE,
    ...input,
  };

  const processPaneWidth = clamp(next.processPaneWidth, MIN_PROCESS_WIDTH, 32);

  if (!next.assistantVisible) {
    return {
      reportPaneWidth: 100 - processPaneWidth,
      processPaneWidth,
      chatPaneWidth: 0,
      assistantVisible: false,
    };
  }

  const chatPaneWidth = clamp(next.chatPaneWidth, MIN_CHAT_WIDTH, 38);
  const minReportWidth = MIN_REPORT_WIDTH;
  const maxProcessWidth = 100 - minReportWidth - chatPaneWidth;
  const boundedProcessWidth = clamp(processPaneWidth, MIN_PROCESS_WIDTH, maxProcessWidth);
  const reportPaneWidth = 100 - boundedProcessWidth - chatPaneWidth;

  return {
    reportPaneWidth,
    processPaneWidth: boundedProcessWidth,
    chatPaneWidth,
    assistantVisible: true,
  };
}

export function resizeWorkspacePane(
  state: AnalysisWorkspaceState,
  pane: 'report-process' | 'process-chat',
  deltaPercent: number
): AnalysisWorkspaceState {
  const current = normalizeWorkspaceState(state);

  if (pane === 'report-process') {
    return normalizeWorkspaceState({
      ...current,
      reportPaneWidth: current.reportPaneWidth + deltaPercent,
      processPaneWidth: current.processPaneWidth - deltaPercent,
    });
  }

  return normalizeWorkspaceState({
    ...current,
    processPaneWidth: current.processPaneWidth + deltaPercent,
    chatPaneWidth: current.chatPaneWidth - deltaPercent,
  });
}
