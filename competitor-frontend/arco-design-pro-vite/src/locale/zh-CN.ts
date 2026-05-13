import localeMessageBox from '@/components/message-box/locale/zh-CN';
import localeLogin from '@/views/login/locale/zh-CN';

import localeWorkplace from '@/views/dashboard/workplace/locale/zh-CN';
import localeSearchTable from '@/views/list/search-table/locale/zh-CN';
import localeGroupForm from '@/views/form/group/locale/zh-CN';
import localeDataAnalysis from '@/views/visualization/data-analysis/locale/zh-CN';

import localeUserInfo from '@/views/user/info/locale/zh-CN';
import localeUserSetting from '@/views/user/setting/locale/zh-CN';

import localeSettings from './zh-CN/settings';

export default {
  'menu.workbench': '工作台',
  'menu.competitors': '竞品管理',
  'menu.analysis.new': '新建分析',
  'menu.analysis.report': '分析报告',
  'menu.analysis.overview': '分析概览',
  'menu.user': '个人中心',
  'navbar.docs': '文档中心',
  'navbar.action.locale': '切换为中文',
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
