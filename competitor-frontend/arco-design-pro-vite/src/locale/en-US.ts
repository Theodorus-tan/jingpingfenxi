import localeMessageBox from '@/components/message-box/locale/en-US';
import localeLogin from '@/views/login/locale/en-US';

import localeWorkplace from '@/views/dashboard/workplace/locale/en-US';
import localeSearchTable from '@/views/list/search-table/locale/en-US';
import localeGroupForm from '@/views/form/group/locale/en-US';
import localeDataAnalysis from '@/views/visualization/data-analysis/locale/en-US';

import localeUserInfo from '@/views/user/info/locale/en-US';
import localeUserSetting from '@/views/user/setting/locale/en-US';

import localeSettings from './en-US/settings';

export default {
  'menu.workbench': 'Workbench',
  'menu.competitors': 'Competitors',
  'menu.analysis.new': 'New Analysis',
  'menu.analysis.report': 'Analysis Report',
  'menu.analysis.overview': 'Analysis Overview',
  'menu.user': 'User Center',
  'navbar.docs': 'Docs',
  'navbar.action.locale': 'Switch to English',
  ...localeSettings,
  ...localeMessageBox,
  ...localeLogin,
  ...localeWorkplace,
  ...localeSearchTable,
  ...localeGroupForm,
  ...localeDataAnalysis,
  ...localeUserInfo,
  ...localeUserSetting,
};
