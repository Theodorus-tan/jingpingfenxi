import { describe, expect, it } from 'vitest';

import {
  DEFAULT_WORKSPACE_STATE,
  normalizeWorkspaceState,
  resizeWorkspacePane,
} from '../analysis-workspace';

describe('normalizeWorkspaceState', () => {
  it('关闭助手时自动让左中区域吃满空间', () => {
    expect(normalizeWorkspaceState({ assistantVisible: false })).toEqual({
      reportPaneWidth: 80,
      processPaneWidth: 20,
      chatPaneWidth: 0,
      assistantVisible: false,
    });
  });

  it('约束聊天区和过程区最小宽度', () => {
    const normalized = normalizeWorkspaceState({
      reportPaneWidth: 10,
      processPaneWidth: 5,
      chatPaneWidth: 90,
      assistantVisible: true,
    });

    expect(normalized.processPaneWidth).toBeGreaterThanOrEqual(16);
    expect(normalized.chatPaneWidth).toBeLessThanOrEqual(38);
    expect(normalized.reportPaneWidth).toBeGreaterThanOrEqual(38);
  });
});

describe('resizeWorkspacePane', () => {
  it('拖拽左侧分隔线时更新报告区和过程区宽度', () => {
    const resized = resizeWorkspacePane(
      DEFAULT_WORKSPACE_STATE,
      'report-process',
      -6
    );

    expect(resized.reportPaneWidth).toBe(46);
    expect(resized.processPaneWidth).toBe(26);
  });

  it('拖拽右侧分隔线时更新过程区和聊天区宽度', () => {
    const resized = resizeWorkspacePane(
      DEFAULT_WORKSPACE_STATE,
      'process-chat',
      4
    );

    expect(resized.processPaneWidth).toBe(24);
    expect(resized.chatPaneWidth).toBe(24);
  });
});
