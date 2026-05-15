import type { AnalysisScenario } from './analysis-storage';

export type PreAnalysisChatContext = {
  competitor?: string;
  project?: string;
  scenario?: AnalysisScenario | string;
};

const SELF_DOUBT_PATTERN =
  /(大学生|学生|小白|新手|没经验|没有经验|没资源|没有资源|一个人|打不过|不自信|害怕|怕做不过)/;
const STRENGTH_PATTERN =
  /(强大|很强|厉害|能打|竞争力|壁垒|优势|强不强)/;
const HOW_TO_START_PATTERN =
  /(怎么开始|如何开始|怎么做|如何做|怎么切入|如何切入|从哪开始|从哪里开始|怎么入手|如何入手)/;

function getScenarioLabel(scenario?: AnalysisScenario | string) {
  if (scenario === 'Market_Entry') return '入局判断';
  return '产品改进';
}

export function buildPreAnalysisChatReply(
  message: string,
  context: PreAnalysisChatContext
) {
  const normalized = message.trim();
  const competitor = (context.competitor || '').trim();
  const project = (context.project || '').trim();
  const scenarioLabel = getScenarioLabel(context.scenario);

  if (!normalized) return '';

  if (SELF_DOUBT_PATTERN.test(normalized)) {
    return [
      competitor
        ? `如果你是在看 \`${competitor}\`，它大概率确实是个强竞品，但这不代表你没有机会。`
        : '对手再强，也不代表你没有机会。',
      '',
      '你现在最值得做的不是正面硬碰，而是先找 3 个更现实的切入口：',
      '- 找细分人群：先服务一个更具体的小群体，而不是一上来打全市场。',
      '- 找高频痛点：先解决 1 个用户天天会遇到的问题，别试图一次做全。',
      `- 找差异打法：从速度、价格、体验、内容、渠道里至少挑 1 个方向做出明显差异。`,
      '',
      competitor
        ? `如果你愿意，我下一句可以直接按“${competitor} 为什么强 / 你作为学生怎么切入 / 最小可行产品怎么做”给你拆开。`
        : '如果你愿意，我下一句可以直接按“对手为什么强 / 你怎么切入 / 最小可行产品怎么做”给你拆开。',
    ].join('\n');
  }

  if (competitor && STRENGTH_PATTERN.test(normalized)) {
    return [
      `结论：\`${competitor}\` 大概率是强竞品，但强不等于没有破绽。`,
      '',
      '你先看 3 件事：',
      '- 它在哪些能力上明显领先：品牌、供应链、渠道、生态还是技术。',
      '- 它为了做大规模，在哪些体验上必须做取舍。',
      '- 有没有一群用户被它服务得还不够好。',
      '',
      `如果你要，我可以继续按“${competitor} 的强点 / 弱点 / 你的切入位”给你快速拆一版。`,
    ].join('\n');
  }

  if (competitor && HOW_TO_START_PATTERN.test(normalized)) {
    return [
      `如果你想围绕 \`${competitor}\` 找机会，我建议先按这个顺序推进：`,
      '- 先定义目标用户：不要说“所有人”，只选一类最具体的人群。',
      `- 先判断场景：你更偏“${scenarioLabel}”，还是想做全新替代方案。`,
      project
        ? `- 先收敛产品范围：既然项目是 \`${project}\`，先只做一个最小核心能力。`
        : '- 先收敛产品范围：先只做一个最小核心能力，不要把功能铺太大。',
      '- 先验证差异点：为什么用户要从现有强竞品切到你这里。',
      '',
      '如果你愿意，我可以下一句直接帮你列一版最小可行方案。',
    ].join('\n');
  }

  if (competitor) {
    return [
      `如果你现在是在继续聊 \`${competitor}\`，我可以先不跳分析页，直接陪你把思路聊清楚。`,
      '',
      '你可以继续问我这几类问题：',
      '- 它为什么强',
      '- 它最大的弱点是什么',
      '- 我该怎么切入',
      '- 先做什么 MVP',
    ].join('\n');
  }

  return [
    '这句我先不触发分析，避免误识别。',
    '',
    '你可以任选一种说法：',
    '- `分析一下 iPhone 17 Pro，项目: 手机，看看怎么改进`',
    '- `我想先聊聊这个竞品为什么强`',
  ].join('\n');
}
