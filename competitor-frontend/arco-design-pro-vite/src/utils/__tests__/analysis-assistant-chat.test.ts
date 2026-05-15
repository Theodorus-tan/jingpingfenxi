import { describe, expect, it } from 'vitest';

import { buildPreAnalysisChatReply } from '../analysis-assistant-chat';

describe('buildPreAnalysisChatReply', () => {
  it('在学生自我怀疑时给出承接式建议', () => {
    const reply = buildPreAnalysisChatReply(
      '这个竞品是不是很强大，我只是一个大学生',
      {
        competitor: 'iPhone 17 Pro',
        scenario: 'Product_Improvement',
      }
    );

    expect(reply).toContain('iPhone 17 Pro');
    expect(reply).toContain('不代表你没有机会');
    expect(reply).toContain('找细分人群');
  });

  it('围绕已知竞品回答强弱判断', () => {
    const reply = buildPreAnalysisChatReply('它是不是很强', {
      competitor: 'Mac mini',
      scenario: 'Product_Improvement',
    });

    expect(reply).toContain('Mac mini');
    expect(reply).toContain('强竞品');
    expect(reply).toContain('强点 / 弱点 / 你的切入位');
  });

  it('无上下文时返回更安全的引导语', () => {
    const reply = buildPreAnalysisChatReply('我有点担心', {});

    expect(reply).toContain('这句我先不触发分析');
    expect(reply).toContain('分析一下 iPhone 17 Pro');
  });
});
