import { describe, expect, it } from 'vitest';

import {
  buildIntentSummary,
  parseAssistantIntent,
} from '../analysis-assistant-intent';

describe('parseAssistantIntent', () => {
  it('解析竞品、项目和改进场景', () => {
    const intent = parseAssistantIntent(
      '帮我分析一下 Mac mini，项目: 电脑，看看怎么改进'
    );

    expect(intent).toEqual({
      competitor: 'Mac mini',
      project: '电脑',
      scenario: 'Product_Improvement',
      shouldStartAnalysis: true,
    });
  });

  it('解析入局场景', () => {
    const intent = parseAssistantIntent('想分析 GoPro，这个品类值不值得入局');

    expect(intent.competitor).toBe('GoPro');
    expect(intent.scenario).toBe('Market_Entry');
    expect(intent.shouldStartAnalysis).toBe(true);
  });

  it('无明显竞品时不触发分析', () => {
    const intent = parseAssistantIntent('帮我总结一下这份报告最该打的点');

    expect(intent.competitor).toBe('');
    expect(intent.shouldStartAnalysis).toBe(false);
  });

  it('普通追问不应误识别成新的竞品分析请求', () => {
    const intent = parseAssistantIntent(
      '这个竞品是不是很强大，我只是一个大学生'
    );

    expect(intent.competitor).toBe('');
    expect(intent.shouldStartAnalysis).toBe(false);
  });

  it('支持显式竞品字段写法', () => {
    const intent = parseAssistantIntent(
      '竞品：iPhone 17 Pro，项目：手机，看看怎么改进'
    );

    expect(intent.competitor).toBe('iPhone 17 Pro');
    expect(intent.project).toBe('手机');
    expect(intent.shouldStartAnalysis).toBe(true);
  });
});

describe('buildIntentSummary', () => {
  it('输出用户可读的识别摘要', () => {
    const summary = buildIntentSummary({
      competitor: 'Mac mini',
      project: '电脑',
      scenario: 'Product_Improvement',
      shouldStartAnalysis: true,
    });

    expect(summary).toContain('识别到竞品：Mac mini');
    expect(summary).toContain('识别到项目：电脑');
    expect(summary).toContain('识别到场景：已有产品求改进');
  });
});
