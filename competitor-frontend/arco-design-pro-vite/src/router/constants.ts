export const WHITE_LIST = [
  { name: 'notFound', children: [] },
  { name: 'login', children: [] },
  { name: 'dashboard', children: [{ name: 'workbench' }] },
  { name: 'analysisReport', children: [{ name: 'analysisReportView' }] },
  { name: 'analysisNew', children: [] },
  { name: 'competitorsList', children: [] },
];

export const NOT_FOUND = {
  name: 'notFound',
};

export const REDIRECT_ROUTE_NAME = 'Redirect';

export const DEFAULT_ROUTE_NAME = 'workbench';

export const DEFAULT_ROUTE = {
  title: 'menu.workbench',
  name: DEFAULT_ROUTE_NAME,
  fullPath: '/dashboard/workbench',
};
