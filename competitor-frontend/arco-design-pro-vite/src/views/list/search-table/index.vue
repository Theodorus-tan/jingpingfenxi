<template>
  <div class="page">
    <a-space direction="vertical" :size="16" fill>
      <a-card class="card" title="竞品管理">
        <a-space :size="12" wrap>
          <a-input
            v-model="filters.keyword"
            style="width: 240px"
            allow-clear
            placeholder="搜索竞品名称"
          />
          <a-select
            v-model="filters.category"
            style="width: 180px"
            allow-clear
            :options="categoryOptions"
            placeholder="品类"
          />
          <a-button type="primary" @click="openCreate">添加竞品</a-button>
        </a-space>
      </a-card>

      <a-card class="card" :title="`竞品列表（${filteredRows.length}）`">
        <a-table
          :columns="columns"
          :data="filteredRows"
          :pagination="{ pageSize: 10 }"
          row-key="id"
        >
          <template #status="{ record }">
            <a-tag
              :color="record.status === '已分析' ? 'green' : 'gray'"
            >
              {{ record.status }}
            </a-tag>
          </template>
          <template #actions="{ record }">
            <a-space :size="8">
              <a-button
                size="mini"
                type="primary"
                @click="goNewAnalysis(record)"
              >
                分析
              </a-button>
              <a-button
                size="mini"
                :disabled="record.status !== '已分析'"
                @click="goReport(record)"
              >
                报告
              </a-button>
            </a-space>
          </template>
        </a-table>
      </a-card>
    </a-space>

    <a-modal
      v-model:visible="createVisible"
      title="添加竞品"
      :mask-closable="false"
      ok-text="保存"
      @ok="createCompetitor"
    >
      <a-form :model="createForm" layout="vertical">
        <a-form-item
          field="name"
          label="竞品名称"
          :rules="[{ required: true, message: '请输入竞品名称' }]"
        >
          <a-input
            v-model="createForm.name"
            placeholder="例如：DJI Osmo Action 4"
          />
        </a-form-item>
        <a-form-item field="category" label="品类">
          <a-select
            v-model="createForm.category"
            :options="categoryOptions"
            placeholder="选择品类"
            allow-clear
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { computed, reactive, ref } from 'vue';
  import { useRouter } from 'vue-router';
  import type { TableColumnData } from '@arco-design/web-vue';
  import {
    getCompetitorStorageRecords,
    getHistoryRecords,
    saveCompetitorStorageRecords,
    type CompetitorStorageRecord,
    type HistoryRecord,
  } from '@/utils/analysis-storage';

  type CompetitorRow = {
    id: string;
    name: string;
    category: string;
    status: '待分析' | '已分析';
    project: string;
    scenario: string;
    analysisCount: number;
    updatedAt: string;
  };

  const router = useRouter();

  const categoryOptions = [
    { label: '运动相机', value: '运动相机' },
    { label: '无人机', value: '无人机' },
    { label: '手机', value: '手机' },
    { label: '新能源汽车', value: '新能源汽车' },
    { label: '智能硬件', value: '智能硬件' },
    { label: '软件/SAAS', value: '软件/SAAS' },
    { label: '消费品', value: '消费品' },
    { label: '其他', value: '其他' },
  ];

  const filters = reactive({
    keyword: '',
    category: undefined as string | undefined,
  });

  const normalizeName = (value: string) => value.trim().toLowerCase();

  const getScenarioLabel = (scenario?: string) => {
    if (scenario === 'Product_Improvement') return '已有产品求改进';
    if (scenario === 'Market_Entry') return '无产品求入局';
    return '-';
  };

  const loadCompetitors = (): CompetitorRow[] => {
    const history = new Map<
      string,
      {
        name: string;
        time: string;
        project: string;
        scenario: string;
        analysisCount: number;
      }
    >();
    try {
      const raw: HistoryRecord[] = getHistoryRecords();
      raw.forEach((item) => {
        const name = item.competitor?.trim();
        if (!name) return;
        const key = normalizeName(name);
        const existing = history.get(key);
        if (existing) {
          existing.analysisCount += 1;
          if ((item.time || '') > existing.time) {
            existing.time = item.time || '';
            existing.project = item.project || '';
            existing.scenario = item.scenario || '';
            existing.name = name;
          }
          return;
        }
        history.set(key, {
          name,
          time: item.time || '',
          project: item.project || '',
          scenario: item.scenario || '',
          analysisCount: 1,
        });
      });
    } catch {
      /* ignore */
    }

    const manual: CompetitorRow[] = [];
    try {
      const raw: CompetitorStorageRecord[] = getCompetitorStorageRecords();
      raw.forEach((item) => {
        if (item.name?.trim()) {
          manual.push({
            id: item.id || `m_${Math.random().toString(16).slice(2, 10)}`,
            name: item.name.trim(),
            category: item.category || '其他',
            status: '待分析',
            project: item.project || '',
            scenario: item.scenario || '',
            analysisCount: Number(item.analysisCount || 0),
            updatedAt: item.updatedAt || '',
          });
        }
      });
    } catch {
      /* ignore */
    }

    const seen = new Set<string>();
    const result: CompetitorRow[] = [];

    manual.forEach((item) => {
      const key = normalizeName(item.name);
      seen.add(key);
      const historyItem = history.get(key);
      result.push({
        ...item,
        updatedAt: historyItem?.time || item.updatedAt,
        project: historyItem?.project || item.project,
        scenario: historyItem?.scenario || item.scenario,
        analysisCount: historyItem?.analysisCount || item.analysisCount,
        status:
          (historyItem?.analysisCount || item.analysisCount) > 0
            ? '已分析'
            : '待分析',
      });
    });

    history.forEach((info, key) => {
      if (seen.has(key)) return;
      seen.add(key);
      result.push({
        id: `h_${key}`,
        name: info.name,
        category: '',
        status: '已分析',
        project: info.project,
        scenario: info.scenario,
        analysisCount: info.analysisCount,
        updatedAt: info.time,
      });
    });

    return result.sort(
      (a, b) =>
        new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
    );
  };

  const rows = ref<CompetitorRow[]>(loadCompetitors());

  const persistManual = (list: CompetitorRow[]) => {
    const manual = list.filter(
      (r) => r.id.startsWith('m_') || r.id.startsWith('u_')
    );
    saveCompetitorStorageRecords(manual);
  };

  const filteredRows = computed(() => {
    const kw = filters.keyword.trim().toLowerCase();
    return rows.value.filter((r) => {
      const matchKw = !kw || r.name.toLowerCase().includes(kw);
      const matchCat = !filters.category || r.category === filters.category;
      return matchKw && matchCat;
    });
  });

  const columns: TableColumnData[] = [
    { title: '竞品名称', dataIndex: 'name' },
    { title: '品类', dataIndex: 'category', width: 140 },
    {
      title: '分析场景',
      dataIndex: 'scenario',
      width: 160,
      render: ({ record }) =>
        getScenarioLabel((record as unknown as CompetitorRow).scenario),
    },
    { title: '所属项目', dataIndex: 'project', width: 180 },
    { title: '分析次数', dataIndex: 'analysisCount', width: 100 },
    { title: '状态', slotName: 'status', width: 120 },
    { title: '更新时间', dataIndex: 'updatedAt', width: 180 },
    { title: '操作', slotName: 'actions', width: 160 },
  ];

  const createVisible = ref(false);
  const createForm = reactive({
    name: '',
    category: '' as string,
  });

  const openCreate = () => {
    createForm.name = '';
    createForm.category = '';
    createVisible.value = true;
  };

  const createCompetitor = () => {
    if (!createForm.name.trim()) return;
    const now = new Date().toISOString().slice(0, 16).replace('T', ' ');
    const normalizedName = normalizeName(createForm.name);
    const existingIndex = rows.value.findIndex(
      (item) => normalizeName(item.name) === normalizedName
    );

    if (existingIndex >= 0) {
      rows.value[existingIndex] = {
        ...rows.value[existingIndex],
        category: createForm.category || rows.value[existingIndex].category,
        updatedAt: rows.value[existingIndex].updatedAt || now,
      };
    } else {
      const newRow: CompetitorRow = {
        id: `m_${Math.random().toString(16).slice(2, 10)}`,
        name: createForm.name.trim(),
        category: createForm.category || '其他',
        status: '待分析',
        project: '',
        scenario: '',
        analysisCount: 0,
        updatedAt: now,
      };
      rows.value.unshift(newRow);
    }
    rows.value = loadCompetitors();
    persistManual(rows.value);
    createVisible.value = false;
  };

  const goNewAnalysis = (record: CompetitorRow) => {
    router.push({ name: 'analysisNew', query: { competitor: record.name } });
  };

  const goReport = (record: CompetitorRow) => {
    router.push({
      name: 'analysisReportView',
      query: { competitor: record.name },
    });
  };
</script>

<style lang="less" scoped>
  .page {
    background: var(--color-fill-2);
    padding: 16px 20px;
    min-height: calc(100vh - 120px);
  }

  .card {
    border-radius: 8px;
  }
</style>
