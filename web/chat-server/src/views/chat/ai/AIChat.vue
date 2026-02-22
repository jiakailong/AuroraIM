<template>
  <div class="chat-wrap">
    <div
      class="chat-window"
      :style="{
        boxShadow: `var(--el-box-shadow-dark)`,
      }"
    >
      <el-container class="chat-window-container">
        <el-aside class="aside-container">
          <NavigationModal></NavigationModal>
          <div class="aichat-panel">
            <div class="aichat-panel-header">
              <h3 class="aichat-panel-title">AI聊天</h3>
            </div>
            <div class="aichat-panel-body">
              <el-button class="aichat-action-btn" @click="handleNewChat">
                新建聊天
              </el-button>
              <el-input
                v-model="chatSearch"
                class="aichat-search-input"
                placeholder="搜索聊天"
                size="small"
                suffix-icon="Search"
              />
              <div class="aichat-knowledge-wrap">
                <span class="aichat-knowledge-title">知识库</span>
                <el-input
                  v-model="knowledgeSearch"
                  class="aichat-search-input"
                  placeholder="搜索知识库"
                  size="small"
                  suffix-icon="Search"
                />
                <ul class="aichat-knowledge-list">
                  <li
                    v-for="knowledge in filteredKnowledgeList"
                    :key="knowledge"
                    class="aichat-knowledge-item"
                  >
                    {{ knowledge }}
                  </li>
                </ul>
              </div>
            </div>
          </div>
        </el-aside>

        <el-container class="chat-container">
          <el-header class="aichat-header">
            <div class="aichat-header-left">
              <h2 class="chat-name">AI助手</h2>
            </div>
            <div class="aichat-header-right">
              <el-select
                v-model="selectedModel"
                placeholder="选择模型"
                size="small"
                class="aichat-model-select"
              >
                <el-option
                  v-for="model in modelOptions"
                  :key="model.value"
                  :label="model.label"
                  :value="model.value"
                />
              </el-select>
            </div>
          </el-header>

          <el-main class="main-container">
            <div class="aichat-screen">
              <div
                v-for="message in messages"
                :key="message.id"
                :class="[
                  'aichat-message-row',
                  message.role === 'user'
                    ? 'aichat-message-row-user'
                    : 'aichat-message-row-ai',
                ]"
              >
                <div
                  :class="[
                    'aichat-message-bubble',
                    message.role === 'user'
                      ? 'aichat-message-bubble-user'
                      : 'aichat-message-bubble-ai',
                  ]"
                >
                  {{ message.content }}
                </div>
              </div>
            </div>
          </el-main>

          <el-footer>
            <div class="chat-input">
              <el-input
                v-model="inputMessage"
                type="textarea"
                resize="none"
                :rows="4"
                placeholder="请输入内容，与AI对话"
                @keyup.enter.exact="handleSendMessage"
              />
            </div>
            <div class="chat-send">
              <button class="send-btn" @click="handleSendMessage">发送</button>
            </div>
          </el-footer>
        </el-container>
      </el-container>
    </div>
  </div>
</template>

<script>
import { computed, reactive, toRefs } from "vue";
import NavigationModal from "@/components/NavigationModal.vue";

export default {
  name: "AIChat",
  components: {
    NavigationModal,
  },
  setup() {
    const data = reactive({
      chatSearch: "",
      knowledgeSearch: "",
      inputMessage: "",
      selectedModel: "model-kamachat-default",
      modelOptions: [
        { label: "KamaChat-Default", value: "model-kamachat-default" },
        { label: "KamaChat-Pro", value: "model-kamachat-pro" },
      ],
      knowledgeList: ["产品使用说明", "常见问题", "项目规范"],
      messages: [
        {
          id: 1,
          role: "ai",
          content: "你好，我是AI助手。请选择模型后开始对话。",
        },
      ],
    });

    const filteredKnowledgeList = computed(() => {
      if (!data.knowledgeSearch) {
        return data.knowledgeList;
      }
      return data.knowledgeList.filter((item) =>
        item.includes(data.knowledgeSearch)
      );
    });

    const handleNewChat = () => {
      data.inputMessage = "";
      data.messages = [
        {
          id: Date.now(),
          role: "ai",
          content: "新的会话已创建，请输入你的问题。",
        },
      ];
    };

    const handleSendMessage = () => {
      const content = data.inputMessage.trim();
      if (!content) {
        return;
      }

      data.messages.push({
        id: Date.now(),
        role: "user",
        content,
      });

      data.messages.push({
        id: Date.now() + 1,
        role: "ai",
        content: "已收到你的消息。后续将接入模型API返回真实回答。",
      });

      data.inputMessage = "";
    };

    return {
      ...toRefs(data),
      filteredKnowledgeList,
      handleNewChat,
      handleSendMessage,
    };
  },
};
</script>

<style scoped>
.aichat-panel {
  height: 100%;
  width: 74%;
}

.aichat-panel-header {
  height: 8%;
  border-bottom: 3px solid #ccc;
  display: flex;
  align-items: center;
  padding: 0 12px;
}

.aichat-panel-title {
  margin: 0;
  color: #4e4e4e;
  font-size: 16px;
}

.aichat-panel-body {
  height: 92%;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.aichat-action-btn {
  width: 100%;
}

.aichat-search-input {
  width: 100%;
}

.aichat-knowledge-wrap {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  flex: 1;
  min-height: 0;
}

.aichat-knowledge-title {
  color: #4e4e4e;
  font-size: 14px;
  font-weight: 600;
}

.aichat-knowledge-list {
  list-style: none;
  padding: 0;
  margin: 0;
  overflow-y: auto;
}

.aichat-knowledge-item {
  padding: 8px;
  border-radius: 6px;
  color: #606266;
}

.aichat-knowledge-item:hover {
  background: #f5f7fa;
}

.aichat-header {
  border-bottom: 3px solid #ccc;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.aichat-header-left,
.aichat-header-right {
  display: flex;
  align-items: center;
}

.aichat-model-select {
  width: 180px;
}

.aichat-screen {
  height: 100%;
  overflow-y: auto;
  padding: 12px;
  box-sizing: border-box;
}

.aichat-message-row {
  display: flex;
  margin-bottom: 10px;
}

.aichat-message-row-user {
  justify-content: flex-end;
}

.aichat-message-row-ai {
  justify-content: flex-start;
}

.aichat-message-bubble {
  max-width: 70%;
  padding: 10px 12px;
  border-radius: 10px;
  word-break: break-word;
  line-height: 1.5;
}

.aichat-message-bubble-user {
  background: rgb(252, 210.9, 210.9);
  color: #303133;
}

.aichat-message-bubble-ai {
  background: #f4f4f5;
  color: #303133;
}
</style>
