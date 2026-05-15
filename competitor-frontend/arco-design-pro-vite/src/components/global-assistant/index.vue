<template>
  <div>
    <a-button
      v-if="!visible"
      class="assistant-float-trigger"
      type="primary"
      @click="open"
    >
      <template #icon><icon-robot /></template>
      聊天助手
    </a-button>

    <a-drawer
      :visible="visible"
      placement="right"
      :footer="false"
      :width="420"
      :unmount-on-close="false"
      @cancel="close"
    >
      <template #title>聊天助手</template>
      <div class="assistant-shell">
        <div class="assistant-toolbar">
          <a-button type="outline" size="small" @click="triggerMaterialUpload">
            <template #icon><icon-upload /></template>
            上传材料
          </a-button>
          <input
            ref="materialInputRef"
            type="file"
            multiple
            accept=".txt,.md,.csv,.json,text/plain,text/markdown,text/csv,application/json"
            class="material-input"
            @change="handleMaterialUpload"
          />
        </div>

        <div v-if="uploadedMaterials.length" class="material-list">
          <div
            v-for="material in uploadedMaterials"
            :key="material.id"
            class="material-chip"
          >
            <div class="material-chip-main">
              <div class="material-chip-name">{{ material.name }}</div>
              <div class="material-chip-meta">
                {{ formatTime(material.uploadedAt) }}
              </div>
            </div>
            <a-button
              type="text"
              size="mini"
              @click="removeMaterial(material.id)"
            >
              <template #icon><icon-delete /></template>
            </a-button>
          </div>
        </div>

        <div ref="scrollRef" class="assistant-messages">
          <div class="assistant-message ai">
            <div class="assistant-avatar"><icon-robot /></div>
            <div class="assistant-content">
              您可以直接说“分析一下 Mac mini，项目:
              电脑，看看怎么改进”，我会识别参数并带您进入分析。
              如果当前已有报告，我也可以直接基于报告继续聊。
            </div>
          </div>

          <div
            v-for="(message, index) in messages"
            :key="index"
            :class="['assistant-message', message.role]"
          >
            <div class="assistant-avatar">
              <icon-robot v-if="message.role === 'ai'" />
              <icon-user v-else />
            </div>
            <div class="assistant-content" v-html="message.html"></div>
          </div>
        </div>

        <div class="assistant-input">
          <div class="assistant-hint">
            支持两种模式：无报告时自动识别需求并跳转填参；有报告时直接伴读问答。
          </div>
          <a-input-search
            v-model="inputValue"
            placeholder="描述你的需求或直接提问..."
            search-button
            :loading="loading"
            @search="handleSend"
            @press-enter="handleSend"
          >
            <template #button-icon><icon-send /></template>
          </a-input-search>
        </div>
      </div>
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
  import { nextTick, onMounted, ref, watch } from 'vue';
  import { useRouter } from 'vue-router';
  import { Message } from '@arco-design/web-vue';
  import {
    IconDelete,
    IconRobot,
    IconSend,
    IconUpload,
    IconUser,
  } from '@arco-design/web-vue/es/icon';
  import { marked } from 'marked';
  import { chatWithReportStream } from '@/api/analysis';
  import { buildPreAnalysisChatReply } from '@/utils/analysis-assistant-chat';
  import {
    buildUploadedMaterialOverview,
    buildChatContext,
    createUploadedAnalysisMaterial,
    isSupportedTextMaterial,
    isMaterialSummaryQuestion,
    readTextMaterial,
    type UploadedAnalysisMaterial,
  } from '@/utils/analysis-material';
  import {
    buildIntentSummary,
    parseAssistantIntent,
  } from '@/utils/analysis-assistant-intent';
  import {
    getGlobalAssistantVisible,
    getLatestAnalysisRecord,
    getUploadedAnalysisMaterials,
    saveGlobalAssistantVisible,
    saveUploadedAnalysisMaterials,
  } from '@/utils/analysis-storage';

  const router = useRouter();
  const visible = ref(getGlobalAssistantVisible());
  const inputValue = ref('');
  const loading = ref(false);
  const scrollRef = ref<HTMLElement | null>(null);
  const materialInputRef = ref<HTMLInputElement | null>(null);
  const uploadedMaterials = ref<UploadedAnalysisMaterial[]>(
    getUploadedAnalysisMaterials()
  );
  const lastIntentContext = ref<{
    competitor: string;
    project: string;
    scenario: string;
  } | null>(null);
  const messages = ref<
    Array<{ role: 'ai' | 'user'; html: string; content: string }>
  >([]);

  const scrollToBottom = () => {
    nextTick(() => {
      if (scrollRef.value) {
        scrollRef.value.scrollTop = scrollRef.value.scrollHeight;
      }
    });
  };

  const persistMaterials = () => {
    saveUploadedAnalysisMaterials(uploadedMaterials.value);
  };

  const open = () => {
    visible.value = true;
  };

  const close = () => {
    visible.value = false;
  };

  const triggerMaterialUpload = () => {
    materialInputRef.value?.click();
  };

  const removeMaterial = (id: string) => {
    uploadedMaterials.value = uploadedMaterials.value.filter(
      (material) => material.id !== id
    );
    persistMaterials();
  };

  const formatTime = (value?: string) => {
    if (!value) return '--:--';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return '--:--';
    return date.toLocaleTimeString('zh-CN', { hour12: false });
  };

  const handleMaterialUpload = async (event: Event) => {
    const input = event.target as HTMLInputElement;
    const files = Array.from(input.files || []);

    for (let index = 0; index < files.length; index += 1) {
      const file = files[index];
      if (isSupportedTextMaterial(file.name)) {
        try {
          // eslint-disable-next-line no-await-in-loop
          const content = await readTextMaterial(file);
          const uploadedAt = new Date().toISOString();
          const material = createUploadedAnalysisMaterial(
            file,
            content,
            uploadedAt
          );
          uploadedMaterials.value = [
            material,
            ...uploadedMaterials.value.filter(
              (item) => item.id !== material.id
            ),
          ].slice(0, 6);
          persistMaterials();
        } catch {
          Message.error(`读取材料失败: ${file.name}`);
        }
      } else {
        Message.warning(
          `暂不支持读取 ${file.name}，当前仅支持 txt/md/csv/json`
        );
      }
    }

    input.value = '';
  };

  const pushMessage = async (role: 'ai' | 'user', content: string) => {
    messages.value.push({
      role,
      content,
      html: await marked.parse(content),
    });
    scrollToBottom();
  };

  const handleIntentFlow = async (userMessage: string) => {
    const intent = parseAssistantIntent(userMessage);
    if (!intent.shouldStartAnalysis) {
      await pushMessage(
        'ai',
        '我先没识别到明确竞品。您可以直接说：`分析一下 Mac mini，项目: 电脑，看看怎么改进`。'
      );
      return true;
    }

    lastIntentContext.value = {
      competitor: intent.competitor,
      project: intent.project,
      scenario: intent.scenario,
    };

    await pushMessage(
      'ai',
      `已识别需求，我先帮您填好分析参数并跳转过去。\n\n${buildIntentSummary(
        intent
      )}`
    );

    router.push({
      name: 'analysisNew',
      query: {
        competitor: intent.competitor,
        project: intent.project,
        scenario: intent.scenario,
      },
    });
    visible.value = false;
    return true;
  };

  const handleReportChat = async (userMessage: string) => {
    const latestReport = getLatestAnalysisRecord().report;
    if (!latestReport) return false;

    loading.value = true;
    const context = buildChatContext(latestReport, uploadedMaterials.value);
    let aiContent = '';
    const targetIndex = messages.value.length;
    messages.value.push({ role: 'ai', content: '', html: '' });

    await chatWithReportStream(
      {
        report: context,
        message: `请使用简明模式回答：先结论，再列 3 条要点，最后给动作建议。\n\n用户问题：${userMessage}`,
      },
      async (chunk) => {
        aiContent += chunk;
        messages.value[targetIndex].content = aiContent;
        messages.value[targetIndex].html = await marked.parse(aiContent);
        scrollToBottom();
      },
      () => {
        loading.value = false;
      },
      async (err) => {
        loading.value = false;
        messages.value.splice(targetIndex, 1);
        await pushMessage('ai', `当前报告对话失败：${err}`);
      }
    );

    return true;
  };

  const handleMaterialSummaryQuestion = async (userMessage: string) => {
    if (!isMaterialSummaryQuestion(userMessage) || !uploadedMaterials.value.length) {
      return false;
    }

    await pushMessage(
      'ai',
      `这次我先直接基于已上传材料给你一个快速总结：\n\n${buildUploadedMaterialOverview(
        uploadedMaterials.value
      )}\n\n如果你要，我下一句可以继续帮你提炼成「一句话结论 / 3 个重点 / 可执行动作」。`
    );
    return true;
  };

  const handleNoReportFallback = async (userMessage: string) => {
    const latestRecord = getLatestAnalysisRecord();
    const reply = buildPreAnalysisChatReply(userMessage, {
      competitor:
        lastIntentContext.value?.competitor || latestRecord.competitor || '',
      project: lastIntentContext.value?.project || latestRecord.project || '',
      scenario: lastIntentContext.value?.scenario || latestRecord.scenario || '',
    });

    await pushMessage(
      'ai',
      reply
    );
  };

  const handleSend = async () => {
    const userMessage = inputValue.value.trim();
    if (!userMessage || loading.value) return;
    inputValue.value = '';
    await pushMessage('user', userMessage);

    const latestReport = getLatestAnalysisRecord().report;
    const intent = parseAssistantIntent(userMessage);

    if (await handleMaterialSummaryQuestion(userMessage)) {
      return;
    }

    if (intent.shouldStartAnalysis) {
      await handleIntentFlow(userMessage);
      return;
    }

    if (latestReport) {
      await handleReportChat(userMessage);
      return;
    }

    await handleNoReportFallback(userMessage);
  };

  watch(visible, (next) => {
    saveGlobalAssistantVisible(next);
  });

  onMounted(() => {
    if (visible.value) {
      scrollToBottom();
    }
  });

  defineExpose({
    open,
    close,
  });
</script>

<style scoped lang="less">
  .assistant-float-trigger {
    position: fixed;
    right: 24px;
    bottom: 28px;
    z-index: 120;
    height: 44px;
    padding: 0 18px;
    border-radius: 999px;
    box-shadow: 0 18px 36px rgba(60, 121, 255, 0.24);
  }

  .assistant-shell {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  .assistant-toolbar {
    padding-bottom: 12px;
  }

  .material-input {
    display: none;
  }

  .material-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding-bottom: 12px;
  }

  .material-chip {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 10px 12px;
    border: 1px solid rgba(173, 181, 193, 0.2);
    border-radius: 12px;
    background: rgba(248, 250, 252, 0.92);
  }

  .material-chip-main {
    min-width: 0;
  }

  .material-chip-name {
    color: var(--color-text-1);
    font-size: 13px;
    font-weight: 600;
    word-break: break-all;
  }

  .material-chip-meta {
    margin-top: 4px;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .assistant-messages {
    flex: 1;
    min-height: 0;
    overflow: auto;
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 6px 2px 12px;
  }

  .assistant-message {
    display: flex;
    gap: 10px;
    font-size: 14px;
    line-height: 1.6;
  }

  .assistant-message.user {
    flex-direction: row-reverse;
  }

  .assistant-avatar {
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .assistant-message.ai .assistant-avatar {
    background: var(--color-primary-light-1);
    color: rgb(var(--arcoblue-6));
  }

  .assistant-message.user .assistant-avatar {
    background: rgb(var(--arcoblue-6));
    color: #fff;
  }

  .assistant-content {
    max-width: calc(100% - 42px);
    padding: 10px 12px;
    border: 1px solid rgba(173, 181, 193, 0.2);
    border-radius: 12px;
    background: #fff;
    color: var(--color-text-1);
    word-break: break-word;
  }

  .assistant-input {
    padding-top: 12px;
    border-top: 1px solid rgba(173, 181, 193, 0.2);
  }

  .assistant-hint {
    margin-bottom: 8px;
    color: var(--color-text-3);
    font-size: 12px;
    line-height: 1.5;
  }
</style>
