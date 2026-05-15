import html2canvas from 'html2canvas';
import jsPDF from 'jspdf';

const A4_WIDTH_MM = 210;
const A4_HEIGHT_MM = 297;
const PAGE_MARGIN_MM = 12;

export function buildReportExportFileName(
  competitorName: string,
  generatedAt?: string
): string {
  const safeCompetitor = (competitorName || '竞品分析报告')
    .replace(/[\\/:*?"<>|]/g, '-')
    .trim();
  const safeTime = (generatedAt || '')
    .replace(/[:\s]/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');

  return safeTime
    ? `${safeCompetitor}-综述报告-${safeTime}.pdf`
    : `${safeCompetitor}-综述报告.pdf`;
}

export function calculatePdfPageCount(
  canvasWidth: number,
  canvasHeight: number
): number {
  if (canvasWidth <= 0 || canvasHeight <= 0) return 1;

  const contentWidth = A4_WIDTH_MM - PAGE_MARGIN_MM * 2;
  const singlePageHeight = A4_HEIGHT_MM - PAGE_MARGIN_MM * 2;
  const scaledHeight = (canvasHeight * contentWidth) / canvasWidth;

  return Math.max(1, Math.ceil(scaledHeight / singlePageHeight));
}

export async function exportReportElementToPdf(
  element: HTMLElement,
  competitorName: string,
  generatedAt?: string
) {
  const canvas = await html2canvas(element, {
    scale: 2,
    useCORS: true,
    backgroundColor: '#ffffff',
    logging: false,
  });

  const imgData = canvas.toDataURL('image/png');
  const PdfDocument = jsPDF;
  const pdf = new PdfDocument('p', 'mm', 'a4');
  const contentWidth = A4_WIDTH_MM - PAGE_MARGIN_MM * 2;
  const singlePageHeight = A4_HEIGHT_MM - PAGE_MARGIN_MM * 2;
  const scaledHeight = (canvas.height * contentWidth) / canvas.width;
  let remainingHeight = scaledHeight;
  let positionY = PAGE_MARGIN_MM;

  pdf.addImage(
    imgData,
    'PNG',
    PAGE_MARGIN_MM,
    positionY,
    contentWidth,
    scaledHeight
  );
  remainingHeight -= singlePageHeight;

  while (remainingHeight > 0) {
    positionY = remainingHeight - scaledHeight + PAGE_MARGIN_MM;
    pdf.addPage();
    pdf.addImage(
      imgData,
      'PNG',
      PAGE_MARGIN_MM,
      positionY,
      contentWidth,
      scaledHeight
    );
    remainingHeight -= singlePageHeight;
  }

  pdf.save(buildReportExportFileName(competitorName, generatedAt));
}
