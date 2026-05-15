import { describe, expect, it } from 'vitest';

import { parseAnalysisReport } from '../analysis-report';

describe('parseAnalysisReport', () => {
  it('解析新版本综述报告结构', () => {
    const payload = JSON.stringify({
      version: 3,
      summary: '# 一句话结论\n- 可以打',
      review: 'review',
      macro: 'macro',
      finance: 'finance',
      evidence_pool: [
        {
          dimension: 'review',
          title: '评测 A',
          snippet: '用户吐槽内存贵',
          url: 'https://example.com/a',
        },
      ],
    });

    expect(parseAnalysisReport(payload)).toEqual({
      rawReportContext: payload,
      hasStructuredReport: true,
      summary: '# 一句话结论\n- 可以打',
      review: 'review',
      macro: 'macro',
      finance: 'finance',
      evidencePool: [
        {
          dimension: 'review',
          title: '评测 A',
          snippet: '用户吐槽内存贵',
          url: 'https://example.com/a',
        },
      ],
    });
  });

  it('兼容旧版三维报告结构并回退到 review 作为综述', () => {
    const payload = JSON.stringify({
      review: 'review only',
      macro: 'macro',
      finance: 'finance',
    });

    const parsed = parseAnalysisReport(payload);

    expect(parsed.hasStructuredReport).toBe(true);
    expect(parsed.summary).toBe('review only');
    expect(parsed.review).toBe('review only');
    expect(parsed.macro).toBe('macro');
    expect(parsed.finance).toBe('finance');
    expect(parsed.evidencePool).toEqual([]);
  });

  it('兼容纯字符串旧报告', () => {
    const payload = '纯文本报告';

    expect(parseAnalysisReport(payload)).toEqual({
      rawReportContext: payload,
      hasStructuredReport: false,
      summary: payload,
      review: payload,
      macro: '',
      finance: '',
      evidencePool: [],
    });
  });
});
