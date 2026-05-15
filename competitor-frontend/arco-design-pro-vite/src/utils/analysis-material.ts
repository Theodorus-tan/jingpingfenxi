export type UploadedAnalysisMaterial = {
  id: string;
  name: string;
  content: string;
  mimeType: string;
  uploadedAt: string;
};

const SUPPORTED_TEXT_EXTENSIONS = ['txt', 'md', 'csv', 'json'];

export function isSupportedTextMaterial(fileName: string) {
  const extension = fileName.split('.').pop()?.toLowerCase() || '';
  return SUPPORTED_TEXT_EXTENSIONS.includes(extension);
}

export async function readTextMaterial(file: File): Promise<string> {
  return file.text();
}

export function createUploadedAnalysisMaterial(
  file: Pick<File, 'name' | 'type'>,
  content: string,
  uploadedAt: string
): UploadedAnalysisMaterial {
  return {
    id: `${uploadedAt}-${file.name}`,
    name: file.name,
    content: content.trim(),
    mimeType: file.type || 'text/plain',
    uploadedAt,
  };
}

export function buildChatContext(
  report: string,
  materials: UploadedAnalysisMaterial[]
): string {
  const normalizedReport = report.trim();
  if (!materials.length) return normalizedReport;

  const materialBlocks = materials.map((material, index) => {
    const excerpt = material.content.slice(0, 4000);
    return [
      `### 材料 ${index + 1}: ${material.name}`,
      excerpt || '（空内容）',
    ].join('\n');
  });

  return [
    normalizedReport,
    '---',
    '## 用户上传材料',
    ...materialBlocks,
  ].join('\n\n');
}

function splitMaterialSections(content: string) {
  return content
    .split(/\n{2,}|\r\n{2,}/)
    .map((section) => section.replace(/\s+/g, ' ').trim())
    .filter(Boolean);
}

export function isMaterialSummaryQuestion(message: string) {
  return /这个文档|这份文档|这个文件|这份材料|上传材料|文档讲什么|材料讲什么|文件讲什么/.test(
    message
  );
}

export function buildUploadedMaterialOverview(
  materials: UploadedAnalysisMaterial[]
) {
  if (!materials.length) {
    return '当前还没有上传材料。';
  }

  return materials
    .map((material, index) => {
      const sections = splitMaterialSections(material.content).slice(0, 3);
      const bullets = sections.length
        ? sections.map((section) => `- ${section.slice(0, 120)}`).join('\n')
        : '- 这份材料当前没有可读正文。';

      return [`### 材料 ${index + 1}: ${material.name}`, bullets].join('\n');
    })
    .join('\n\n');
}
