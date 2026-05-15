import { describe, expect, it } from 'vitest';

import {
  buildReportExportFileName,
  calculatePdfPageCount,
} from '../report-export';

describe('buildReportExportFileName', () => {
  it('生成包含竞品名和时间的 pdf 文件名', () => {
    expect(buildReportExportFileName('Mac mini', '2026-05-14 10:00')).toBe(
      'Mac mini-综述报告-2026-05-14-10-00.pdf'
    );
  });

  it('清理非法文件名字符', () => {
    expect(buildReportExportFileName('Mac/mini:*?', '2026:05:14')).toBe(
      'Mac-mini----综述报告-2026-05-14.pdf'
    );
  });
});

describe('calculatePdfPageCount', () => {
  it('单页内容返回 1 页', () => {
    expect(calculatePdfPageCount(1000, 1200)).toBe(1);
  });

  it('长内容返回多页', () => {
    expect(calculatePdfPageCount(1000, 5000)).toBeGreaterThan(1);
  });
});
