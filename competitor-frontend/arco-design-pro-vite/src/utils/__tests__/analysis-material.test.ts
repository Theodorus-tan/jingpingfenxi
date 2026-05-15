import { describe, expect, it } from 'vitest';

import {
  buildUploadedMaterialOverview,
  buildChatContext,
  createUploadedAnalysisMaterial,
  isSupportedTextMaterial,
  isMaterialSummaryQuestion,
} from '../analysis-material';

describe('isSupportedTextMaterial', () => {
  it('识别可读取的文本文件', () => {
    expect(isSupportedTextMaterial('notes.md')).toBe(true);
    expect(isSupportedTextMaterial('metrics.csv')).toBe(true);
  });

  it('拒绝非文本材料', () => {
    expect(isSupportedTextMaterial('deck.pdf')).toBe(false);
    expect(isSupportedTextMaterial('spec.docx')).toBe(false);
  });
});

describe('buildChatContext', () => {
  it('无上传材料时直接返回报告内容', () => {
    expect(buildChatContext('report', [])).toBe('report');
  });

  it('拼接用户上传材料到聊天上下文', () => {
    const material = createUploadedAnalysisMaterial(
      { name: 'notes.md', type: 'text/markdown' } as File,
      '第一条证据',
      '2026-05-14T10:00:00.000Z'
    );

    const context = buildChatContext('主报告', [material]);

    expect(context).toContain('## 用户上传材料');
    expect(context).toContain('notes.md');
    expect(context).toContain('第一条证据');
  });
});

describe('material summary helpers', () => {
  it('识别材料总结类问题', () => {
    expect(isMaterialSummaryQuestion('这个文档讲什么的')).toBe(true);
    expect(isMaterialSummaryQuestion('帮我继续分析 Mac mini')).toBe(false);
  });

  it('生成上传材料概览', () => {
    const material = createUploadedAnalysisMaterial(
      { name: '用户需求原文.md', type: 'text/markdown' } as File,
      '这是第一段。\n\n这是第二段，讲用户最核心的诉求。\n\n这是第三段，讲想要的交互效果。',
      '2026-05-15T11:01:51.000Z'
    );

    const overview = buildUploadedMaterialOverview([material]);

    expect(overview).toContain('材料 1: 用户需求原文.md');
    expect(overview).toContain('这是第一段。');
    expect(overview).toContain('这是第二段，讲用户最核心的诉求。');
  });
});
