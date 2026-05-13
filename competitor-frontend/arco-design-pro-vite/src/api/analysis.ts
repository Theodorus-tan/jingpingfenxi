import axios from 'axios';

export interface AnalysisTaskParams {
  competitor_name: string;
}

export interface AnalysisTaskResponse {
  report: string;
}

export async function submitAnalysisTaskStream(
  data: AnalysisTaskParams,
  onEvent: (event: {
    type: string;
    message?: string;
    query?: string;
    titles?: string[];
    results_count?: number;
  }) => void,
  onReport: (report: string) => void,
  onError: (err: string) => void
): Promise<void> {
  try {
    const response = await fetch('/analyze/stream', {
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
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';

      for (let i = 0; i < lines.length; i += 1) {
        const line = lines[i];
        if (!line.startsWith('data:')) continue; // eslint-disable-line no-continue

        try {
          const event = JSON.parse(line.slice(5).trim());
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
          onEvent(event);
        } catch {
          // skip malformed events
        }
      }
    }
  } catch (err: any) {
    onError(err.message || '网络错误');
  }
}

export function submitAnalysisTask(data: AnalysisTaskParams) {
  return axios.post<AnalysisTaskResponse>('/api/analysis/task', data);
}

export async function submitAnalysisTaskEino(
  data: AnalysisTaskParams,
  onEvent: (event: {
    type: string;
    message?: string;
    query?: string;
    report?: string;
  }) => void,
  onReport: (report: string) => void,
  onError: (err: string) => void
): Promise<void> {
  try {
    const response = await fetch('/api/analysis/eino', {
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
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';

      for (let i = 0; i < lines.length; i += 1) {
        const line = lines[i];
        if (!line.startsWith('data:')) continue; // eslint-disable-line no-continue

        try {
          const event = JSON.parse(line.slice(5).trim());
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
          onEvent(event);
        } catch {
          /* skip malformed events */
        }
      }
    }
  } catch (err: any) {
    onError(err.message || '网络错误');
  }
}
