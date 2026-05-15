export type StructuredAnalysisReportPayload = {
  version?: number;
  competitor?: string;
  scenario?: string;
  generated_at?: string;
  summary?: string;
  review?: string;
  macro?: string;
  finance?: string;
  evidence_pool?: Array<{
    dimension?: string;
    title?: string;
    snippet?: string;
    url?: string;
  }>;
};

export type ParsedAnalysisReport = {
  rawReportContext: string;
  hasStructuredReport: boolean;
  summary: string;
  review: string;
  macro: string;
  finance: string;
  evidencePool: Array<{
    dimension: string;
    title: string;
    snippet: string;
    url: string;
  }>;
};

function pickString(value: unknown): string {
  return typeof value === 'string' ? value.trim() : '';
}

export function parseAnalysisReport(rawReport: string): ParsedAnalysisReport {
  const fallback = {
    rawReportContext: rawReport || '',
    hasStructuredReport: false,
    summary: rawReport || '',
    review: rawReport || '',
    macro: '',
    finance: '',
    evidencePool: [],
  };

  if (!rawReport) {
    return fallback;
  }

  try {
    const parsed = JSON.parse(rawReport) as StructuredAnalysisReportPayload;
    const summary = pickString(parsed.summary);
    const review = pickString(parsed.review);
    const macro = pickString(parsed.macro);
    const finance = pickString(parsed.finance);
    const evidencePool = Array.isArray(parsed.evidence_pool)
      ? parsed.evidence_pool
          .map((item) => ({
            dimension: pickString(item.dimension),
            title: pickString(item.title),
            snippet: pickString(item.snippet),
            url: pickString(item.url),
          }))
          .filter((item) => item.title)
      : [];
    const hasStructuredReport = Boolean(summary || review || macro || finance);

    if (!hasStructuredReport) {
      return fallback;
    }

    return {
      rawReportContext: rawReport,
      hasStructuredReport,
      summary: summary || review || rawReport,
      review: review || '',
      macro,
      finance,
      evidencePool,
    };
  } catch {
    return fallback;
  }
}
