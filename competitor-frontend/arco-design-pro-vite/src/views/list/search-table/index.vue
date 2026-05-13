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
              :color="
                record.status === '已分析'
                  ? 'green'
                  : record.status === '分析中'
                  ? 'orange'
                  : 'gray'
              "
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

  type CompetitorRow = {
    id: string;
    name: string;
    category: string;
    status: '待分析' | '分析中' | '已分析';
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

  const STORAGE_KEY = 'competitor_list';

  const loadCompetitors = (): CompetitorRow[] => {
    const history = new Map<string, { time: string }>();
    try {
      const raw = JSON.parse(localStorage.getItem('analysis_history') || '[]');
      raw.forEach((item: any) => {
        const name = item.competitor?.trim();
        if (name && !history.has(name)) {
          history.set(name, { time: item.time || '' });
        }
      });
    } catch {
      /* ignore */
    }

    const latestCompetitor =
      localStorage.getItem('latest_analysis_competitor') || '';

    const manual: CompetitorRow[] = [];
    try {
      const raw = JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]');
      raw.forEach((item: any) => {
        if (item.name?.trim()) {
          manual.push(item);
        }
      });
    } catch {
      /* ignore */
    }

    const seen = new Set<string>();
    const result: CompetitorRow[] = [];

    manual.forEach((item) => {
      seen.add(item.name);
      const historyItem = history.get(item.name);
      result.push({
        ...item,
        updatedAt: historyItem?.time || item.updatedAt,
        status: item.name === latestCompetitor ? '已分析' : (item.status as any),
      });
    });

    history.forEach((info, name) => {
      if (seen.has(name)) return;
      seen.add(name);
      result.push({
        id: `h_${name}`,
        name,
        category: '',
        status: name === latestCompetitor ? '已分析' : '待分析',
        updatedAt: info.time,
      });
    });

    return result.sort(
      (a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
    );
  };

  const rows = ref<CompetitorRow[]>(loadCompetitors());

  const persistManual = (list: CompetitorRow[]) => {
    const manual = list.filter(
      (r) => r.id.startsWith('m_') || r.id.startsWith('u_')
    );
    localStorage.setItem(STORAGE_KEY, JSON.stringify(manual));
  };

  const filteredRows = computed(() => {
    const kw = filters.keyword.trim().toLowerCase();
    return rows.value.filter((r) => {
      const matchKw = !kw || r.name.toLowerCase().includes(kw);
      const matchCat = !filters.category || r.category === filters.category;
      return matchKw && matchCat;
    });
  });

  const columns = [
    { title: '竞品名称', dataIndex: 'name' },
    { title: '品类', dataIndex: 'category', width: 140 },
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
    const newRow: CompetitorRow = {
      id: `m_${Math.random().toString(16).slice(2, 10)}`,
      name: createForm.name.trim(),
      category: createForm.category || '其他',
      status: '待分析',
      updatedAt: now,
    };
    rows.value.unshift(newRow);
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
