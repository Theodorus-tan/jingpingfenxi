import axios from 'axios';

export interface AnalysisTaskParams {
  competitor_name: string;
  scenario?: string;
  project?: string;
}

export interface AnalysisTaskResponse {
  report: string;
}

export interface CompetitorImageCandidate {
  title: string;
  snippet: string;
  page_url: string;
  image_url: string;
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

function resolveApiUrl(path: string) {
  if (!API_BASE_URL) return path;
  return `${API_BASE_URL.replace(/\/$/, '')}${path}`;
}

function extractJSONPayloads(buffer: string): {
  events: Array<Record<string, any>>;
  rest: string;
} {
  const events: Array<Record<string, any>> = [];
  let cursor = 0;

  const findJsonEnd = (input: string, startIndex: number) => {
    let depth = 0;
    let inString = false;
    let escaped = false;

    for (let i = startIndex; i < input.length; i += 1) {
      const char = input[i];

      if (escaped) {
        escaped = false;
        continue; // eslint-disable-line no-continue
      }

      if (char === '\\') {
        escaped = true;
        continue; // eslint-disable-line no-continue
      }

      if (char === '"') {
        inString = !inString;
        continue; // eslint-disable-line no-continue
      }

      if (inString) continue; // eslint-disable-line no-continue

      if (char === '{') depth += 1;
      if (char === '}') {
        depth -= 1;
        if (depth === 0) return i;
      }
    }

    return -1;
  };

  while (cursor < buffer.length) {
    const dataIndex = buffer.indexOf('data:', cursor);
    if (dataIndex === -1) {
      return { events, rest: '' };
    }

    const jsonStart = buffer.indexOf('{', dataIndex);
    if (jsonStart === -1) {
      return { events, rest: buffer.slice(dataIndex) };
    }

    const jsonEnd = findJsonEnd(buffer, jsonStart);
    if (jsonEnd === -1) {
      return { events, rest: buffer.slice(dataIndex) };
    }

    try {
      events.push(JSON.parse(buffer.slice(jsonStart, jsonEnd + 1)));
    } catch {
      // skip malformed payload
    }

    cursor = jsonEnd + 1;
  }

  return { events, rest: '' };
}

export async function submitAnalysisTaskStream(
  data: AnalysisTaskParams,
  onEvent: (event: {
    type: string;
    message?: string;
    query?: string;
    titles?: string[];
    evidences?: Array<{ title: string; snippet?: string; url?: string }>;
    results_count?: number;
  }) => void,
  onReport: (report: string) => void,
  onError: (err: string) => void
): Promise<void> {
  try {
    const response = await fetch(resolveApiUrl('/analyze/stream'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      let errMsg = `请求失败: ${response.status}`;
      try {
        const errBody = await response.json();
        if (errBody.error) errMsg = errBody.error;
      } catch {
        /* skip parse error */
      }
      onError(errMsg);
      return;
    }

    const { body } = response;
    if (!body) {
      onError('无法读取响应流');
      return;
    }

    const reader = body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    let reading = true;

    while (reading) {
      // eslint-disable-next-line no-await-in-loop
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const parsed = extractJSONPayloads(buffer);
      buffer = parsed.rest;

      for (let i = 0; i < parsed.events.length; i += 1) {
        const event = parsed.events[i];
        if (event.type === 'done') {
          onReport(event.report || '');
          reading = false;
          break;
        }
        if (event.type === 'error') {
          onError(event.message || '未知错误');
          reading = false;
          break;
        }
        onEvent(event as any);
      }
    }

    if (reading) {
      onError('分析流已结束，但没有收到完成事件');
    }
  } catch (err: any) {
    onError(err.message || '网络错误');
  }
}

export async function chatWithReportStream(
  data: { report: string; message: string },
  onChunk: (chunk: string) => void,
  onDone: () => void,
  onError: (err: string) => void
): Promise<void> {
  try {
    const response = await fetch(resolveApiUrl('/api/analysis/chat'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      let errMsg = `请求失败: ${response.status}`;
      try {
        const errBody = await response.json();
        if (errBody.error) errMsg = errBody.error;
      } catch {
        /* skip parse error */
      }
      onError(errMsg);
      return;
    }

    const { body } = response;
    if (!body) {
      onError('无法读取响应流');
      return;
    }

    const reader = body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    let reading = true;

    while (reading) {
      // eslint-disable-next-line no-await-in-loop
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const parsed = extractJSONPayloads(buffer);
      buffer = parsed.rest;

      for (let i = 0; i < parsed.events.length; i += 1) {
        const event = parsed.events[i];
        if (event.done) {
          onDone();
          reading = false;
          break;
        }
        if (event.error) {
          onError(event.error);
          reading = false;
          break;
        }
        if (event.chunk) {
          onChunk(event.chunk);
        }
      }
    }

    if (reading) {
      onError('对话流已结束，但没有收到完成事件');
    }
  } catch (err: any) {
    onError(err.message || '网络错误');
  }
}

export function submitAnalysisTask(data: AnalysisTaskParams) {
  return axios.post<AnalysisTaskResponse>('/api/analysis/task', data);
}

export async function fetchCompetitorImages(competitorName: string) {
  const response = await fetch(
    resolveApiUrl(
      `/api/analysis/images?competitor_name=${encodeURIComponent(competitorName)}`
    )
  );
  if (!response.ok) {
    throw new Error(`请求失败: ${response.status}`);
  }
  const payload = await response.json();
  return (payload.data || []) as CompetitorImageCandidate[];
}

export async function submitAnalysisTaskEino(
  data: AnalysisTaskParams,
  onEvent: (event: {
    type: string;
    dimension?: string;
    message?: string;
    query?: string;
    report?: string;
  }) => void,
  onReport: (report: string) => void,
  onError: (err: string) => void
): Promise<void> {
  try {
    const response = await fetch(resolveApiUrl('/api/analysis/eino'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      let errMsg = `请求失败: ${response.status}`;
      try {
        const errBody = await response.json();
        if (errBody.error) errMsg = errBody.error;
      } catch {
        /* skip parse error */
      }
      onError(errMsg);
      return;
    }

    const { body } = response;
    if (!body) {
      onError('无法读取响应流');
      return;
    }

    const reader = body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    let reading = true;

    while (reading) {
      // eslint-disable-next-line no-await-in-loop
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const parsed = extractJSONPayloads(buffer);
      buffer = parsed.rest;

      for (let i = 0; i < parsed.events.length; i += 1) {
        const event = parsed.events[i];
        if (event.type === 'done') {
          onReport(event.report || '');
          reading = false;
          break;
        }
        if (event.type === 'error') {
          onError(event.message || '未知错误');
          reading = false;
          break;
        }
        onEvent(event as any);
      }
    }

    if (reading) {
      onError('分析流已结束，但没有收到完成事件');
    }
  } catch (err: any) {
    onError(err.message || '网络错误');
  }
}
