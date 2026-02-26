<template>
  <div class="page-container" :class="{ 'embedded': embedded }">
    <div class="content-wrapper animate-fade-in">
      <!-- Page Header / 页面头部（嵌入模式下简化） -->
      <PageHeader
        v-if="!embedded"
        :title="$t('aiConfig.title')"
        :subtitle="$t('aiConfig.subtitle') || '管理 AI 服务配置'"
        :show-back="true"
        :back-text="$t('common.back')"
      >
        <template #actions>
          <el-button type="primary" @click="showCreateDialog">
            <el-icon><Plus /></el-icon>
            <span>{{ $t("aiConfig.addConfig") }}</span>
          </el-button>
        </template>
      </PageHeader>
      <div v-else class="embedded-header">
        <h3>{{ $t('aiConfig.title') }}</h3>
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          <span>{{ $t("aiConfig.addConfig") }}</span>
        </el-button>
      </div>

      <!-- 模型开通说明 -->
      <el-alert type="info" :closable="false" class="guide-alert">
        <template #title>
          <span class="guide-title">需开通的模型</span>
        </template>
        <p class="guide-desc">使用前请先在 <a href="https://console.volcengine.com/ark/region:ark+cn-beijing/apiKey" target="_blank">API Key 管理</a> 创建 Key，在开通管理开通以下模型，并填写 API Key。配置完成后可点击「测试连接」验证连通度。</p>
        <ul class="guide-list">
          <li><strong>文本</strong>：<a href="https://console.volcengine.com/ark/region:ark+cn-beijing/openManagement" target="_blank">火山方舟 Doubao</a> <code>doubao-1-5-pro-32k-250115</code></li>
          <li><strong>图片</strong>：<a href="https://console.volcengine.com/ark/region:ark+cn-beijing/openManagement?tab=ComputerVision" target="_blank">火山方舟 Seedream 4.0</a> <code>doubao-seedream-4-0-250828</code></li>
          <li><strong>视频</strong>：<a href="https://console.volcengine.com/ark/region:ark+cn-beijing/openManagement?tab=ComputerVision" target="_blank">火山方舟 Seedance 1.5 Pro</a> <code>doubao-seedance-1-5-pro-251215</code></li>
        </ul>
        <p class="guide-configured">
          <strong>当前已配置：</strong>
          <template v-if="configSummary.text || configSummary.image || configSummary.video">
            <span v-if="configSummary.text">文本：{{ configSummary.text }}</span>
            <span v-if="configSummary.image"> · 图片：{{ configSummary.image }}</span>
            <span v-if="configSummary.video"> · 视频：{{ configSummary.video }}</span>
          </template>
          <span v-else class="guide-configured-empty">暂无，配置并保存后将显示在此</span>
        </p>
        <p class="guide-note">
          同一 API Key 可同时用于文本、图片、视频。Base URL 固定为 <code>https://ark.cn-beijing.volces.com/api/v3</code>。推荐将 API Key 配置在环境变量中，避免泄露。
        </p>
      </el-alert>

      <!-- Tabs / 标签页 -->
      <div class="tabs-wrapper">
        <el-tabs
          v-model="activeTab"
          @tab-change="handleTabChange"
          class="config-tabs"
        >
          <el-tab-pane :label="$t('aiConfig.tabs.text')" name="text">
            <ConfigList
              :configs="configs"
              :loading="loading"
              :show-actions="true"
              :show-test-button="true"
              @edit="handleEdit"
              @delete="handleDelete"
              @toggle-active="handleToggleActive"
              @test="handleTest"
            />
          </el-tab-pane>

          <el-tab-pane :label="$t('aiConfig.tabs.image')" name="image">
            <ConfigList
              :configs="configs"
              :loading="loading"
              :show-actions="true"
              :show-test-button="true"
              @edit="handleEdit"
              @delete="handleDelete"
              @toggle-active="handleToggleActive"
              @test="handleTest"
            />
          </el-tab-pane>

          <el-tab-pane :label="$t('aiConfig.tabs.video')" name="video">
            <ConfigList
              :configs="configs"
              :loading="loading"
              :show-actions="true"
              :show-test-button="true"
              @edit="handleEdit"
              @delete="handleDelete"
              @toggle-active="handleToggleActive"
              @test="handleTest"
            />
          </el-tab-pane>
        </el-tabs>
      </div>

      <!-- Edit/Create Dialog / 编辑创建弹窗 -->
      <el-dialog
        v-model="dialogVisible"
        :title="isEdit ? $t('aiConfig.editConfig') : $t('aiConfig.addConfig')"
        width="520px"
        :close-on-click-modal="false"
        class="ai-config-dialog"
      >
        <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
          <!-- 火山引擎：简洁模式，只需 API Key -->
          <template v-if="isVolcSimpleMode">
            <el-alert type="info" :closable="false" class="simple-mode-tip">
              火山引擎同一 API Key 可用于文本、图片、视频。保存将一次性创建三条配置（Doubao、Seedream、Seedance）。
            </el-alert>
            <el-form-item label="API Key" prop="api_key">
              <el-input
                v-model="form.api_key"
                type="password"
                show-password
                placeholder="在 API Key 管理创建"
                clearable
              />
              <div class="form-tip">
                在 <a href="https://console.volcengine.com/ark/region:ark+cn-beijing/apiKey" target="_blank">API Key 管理</a> 创建。<br />
                API Key 在 API Key 管理中创建。<br />
                Base URL 固定为 <code>https://ark.cn-beijing.volces.com/api/v3</code>，已自动配置。
              </div>
            </el-form-item>
          </template>

          <!-- 非火山引擎 或 编辑模式：完整表单 -->
          <template v-else>
            <el-form-item :label="$t('aiConfig.form.name')" prop="name">
              <el-input
                v-model="form.name"
                :placeholder="$t('aiConfig.form.namePlaceholder')"
              />
            </el-form-item>
            <el-form-item :label="$t('aiConfig.form.provider')" prop="provider">
              <el-select
                v-model="form.provider"
                :placeholder="$t('aiConfig.form.providerPlaceholder')"
                @change="handleProviderChange"
                style="width: 100%"
              >
                <el-option
                  v-for="provider in availableProviders"
                  :key="provider.id"
                  :label="provider.name"
                  :value="provider.id"
                  :disabled="provider.disabled"
                />
              </el-select>
              <div class="form-tip">{{ $t("aiConfig.form.providerTip") }}</div>
            </el-form-item>
            <el-form-item :label="$t('aiConfig.form.priority')" prop="priority">
              <el-input-number
                v-model="form.priority"
                :min="0"
                :max="100"
                :step="1"
                style="width: 100%"
              />
              <div class="form-tip">{{ $t("aiConfig.form.priorityTip") }}</div>
            </el-form-item>
            <el-form-item :label="$t('aiConfig.form.model')" prop="model">
              <el-select
                v-model="form.model"
                :placeholder="form.provider === 'custom' ? '手动输入 Model ID' : $t('aiConfig.form.modelPlaceholder')"
                multiple
                filterable
                allow-create
                default-first-option
                collapse-tags
                collapse-tags-tooltip
                style="width: 100%"
              >
                <el-option
                  v-for="model in availableModels"
                  :key="model"
                  :label="model"
                  :value="model"
                />
              </el-select>
              <div class="form-tip">{{ $t("aiConfig.form.modelTip") }}</div>
            </el-form-item>
            <el-form-item :label="$t('aiConfig.form.baseUrl')" prop="base_url">
              <el-input
                v-model="form.base_url"
                :placeholder="$t('aiConfig.form.baseUrlPlaceholder')"
              />
              <div class="form-tip">
                {{ $t("aiConfig.form.baseUrlTip") }}
                <br />
                {{ $t("aiConfig.form.fullEndpoint") }}: {{ fullEndpointExample }}
              </div>
            </el-form-item>
            <el-form-item :label="$t('aiConfig.form.apiKey')" prop="api_key">
              <el-input
                v-model="form.api_key"
                type="password"
                show-password
                :placeholder="apiKeyPlaceholder"
                clearable
              />
              <div class="form-tip">{{ apiKeyTip }}</div>
            </el-form-item>
            <el-form-item v-if="isEdit" :label="$t('aiConfig.form.isActive')">
              <el-switch v-model="form.is_active" />
            </el-form-item>
          </template>
        </el-form>

        <template #footer>
          <el-button @click="dialogVisible = false">{{
            $t("common.cancel")
          }}</el-button>
          <el-button
            @click="testConnection"
            :loading="testing"
            >{{ $t("aiConfig.actions.test") }}</el-button
          >
          <el-tooltip :content="!testPassed ? '请先点击「测试连接」并全部通过后再保存' : ''" placement="top">
            <span>
              <el-button
                type="primary"
                @click="handleSubmit"
                :loading="submitting"
                :disabled="!testPassed"
              >
                {{ $t("common.save") }}
              </el-button>
            </span>
          </el-tooltip>
        </template>
      </el-dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from "vue";
import { useRouter } from "vue-router";

/** 嵌入模式：用于在 AI 日志页内嵌展示，隐藏返回按钮 */
defineProps<{ embedded?: boolean }>();
import {
  ElMessage,
  ElMessageBox,
  type FormInstance,
  type FormRules,
} from "element-plus";
import { Plus, ArrowLeft } from "@element-plus/icons-vue";
import { aiAPI } from "@/api/ai";
import { PageHeader } from "@/components/common";
import type {
  AIServiceConfig,
  AIServiceType,
  CreateAIConfigRequest,
  UpdateAIConfigRequest,
} from "@/types/ai";
import ConfigList from "./components/ConfigList.vue";

const router = useRouter();

const activeTab = ref<AIServiceType>("text");
const loading = ref(false);
const configsAll = ref<AIServiceConfig[]>([]);
const configs = computed(() =>
  configsAll.value.filter((c) => c.service_type === activeTab.value),
);
const dialogVisible = ref(false);
const isEdit = ref(false);
const editingId = ref<number>();
const savedApiKeyMask = ref(""); // 编辑时已保存 Key 的掩码（前4位+****）
const originalApiKeyForTest = ref(""); // 编辑时用于测试连接的原始 Key（不展示）
const formRef = ref<FormInstance>();
const submitting = ref(false);
const testing = ref(false);
const testPassed = ref(false); // 测试通过后才能保存

const form = reactive<
  CreateAIConfigRequest & { is_active?: boolean; provider?: string }
>({
  service_type: "text",
  provider: "",
  name: "",
  base_url: "",
  api_key: "",
  model: [], // 改为数组支持多选
  priority: 0, // 默认优先级为0
  is_active: true,
});

// 厂商和模型配置
interface ProviderConfig {
  id: string;
  name: string;
  models: string[];
  disabled?: boolean;
}

const providerConfigs: Record<AIServiceType, ProviderConfig[]> = {
  text: [
    { id: "doubao", name: "火山引擎/字节 Doubao", models: ["doubao-1-5-pro-32k-250115", "doubao-1-5-lite-32k-250115", "doubao-pro-32k"] },
    { id: "openai", name: "OpenAI", models: ["gpt-4o", "gpt-4o-mini", "gpt-4"] },
    { id: "gemini", name: "Google Gemini", models: ["gemini-2.0-flash", "gemini-1.5-pro"] },
    { id: "custom", name: "自定义（手动输入）", models: [] },
  ],
  image: [
    { id: "volcengine", name: "火山引擎/字节 Seedream", models: ["doubao-seedream-4-0-250828", "doubao-seedream-4-5-251128"] },
    { id: "openai", name: "OpenAI DALL-E", models: ["dall-e-3"] },
    { id: "custom", name: "自定义（手动输入）", models: [] },
  ],
  video: [
    { id: "volces", name: "火山引擎/字节 Seedance", models: ["doubao-seedance-1-5-pro-251215"] },
    { id: "openai", name: "OpenAI Sora", models: ["sora-1.0"] },
    { id: "custom", name: "自定义（手动输入）", models: [] },
  ],
};

// 火山引擎简洁模式：新建时只需填 API Key，其余自动配置
const isVolcSimpleMode = computed(() => {
  if (isEdit.value) return false;
  const p = form.provider;
  return ["doubao", "volcengine", "volces"].includes(p);
});

// 当前可用的厂商列表（新建时显示全部，编辑时显示全部）
const availableProviders = computed(() => {
  return providerConfigs[form.service_type] || [];
});

// 当前可用的模型列表（优先从已激活配置获取，否则用厂商预设）
const availableModels = computed(() => {
  if (!form.provider) return [];

  const activeConfigsForProvider = configs.value.filter(
    (c) =>
      c.provider === form.provider &&
      c.service_type === form.service_type &&
      c.is_active,
  );

  const models = new Set<string>();
  activeConfigsForProvider.forEach((config) => {
    (Array.isArray(config.model) ? config.model : [config.model]).forEach((m) => models.add(m));
  });

  if (models.size > 0) return Array.from(models);

  const provider = providerConfigs[form.service_type]?.find((p) => p.id === form.provider);
  return provider?.models || [];
});

// API Key 占位符：编辑时显示已保存的掩码，新建时按厂商区分
const apiKeyPlaceholder = computed(() => {
  if (isEdit.value && savedApiKeyMask.value) {
    return `${savedApiKeyMask.value}（已保存，输入新 Key 可修改，留空则不修改）`;
  }
  const p = form.provider;
  if (p === "doubao" || p === "volcengine" || p === "volces") {
    return "在 API Key 管理创建";
  }
  if (p === "openai") return "手动输入 API Key，格式 sk-xxxxxxxx";
  if (p === "gemini" || p === "google") return "手动输入 API Key，格式 AIzaSyxxxxxxxx";
  if (p === "custom") return "手动输入 API Key";
  return "手动输入 API Key";
});

// API Key 提示文案
const apiKeyTip = computed(() => {
  const p = form.provider;
  if (p === "doubao" || p === "volcengine" || p === "volces") {
    return "字节跳动火山方舟 API Key，在 ark.volcengine.com 创建。同一 Key 可同时用于文本、图片、视频。";
  }
  if (p === "custom") return "自定义厂商：请手动输入 Base URL、API Key 和 Model ID。";
  return "请在对应平台申请后，手动复制粘贴 API Key 到此框。";
});

// 完整端点示例
const fullEndpointExample = computed(() => {
  const baseUrl = form.base_url || "https://api.example.com";
  const provider = form.provider;
  const serviceType = form.service_type;

  let endpoint = "";

  if (serviceType === "text") {
    if (provider === "gemini" || provider === "google") {
      endpoint = "/v1beta/models/{model}:generateContent";
    } else {
      endpoint = "/chat/completions";
    }
  } else if (serviceType === "image") {
    if (provider === "gemini" || provider === "google") {
      endpoint = "/v1beta/models/{model}:generateContent";
    } else {
      endpoint = "/images/generations";
    }
  } else if (serviceType === "video") {
    if (provider === "doubao" || provider === "volcengine" || provider === "volces") {
      endpoint = "/contents/generations/tasks";
    } else if (provider === "openai") {
      endpoint = "/videos";
    } else {
      endpoint = "/video/generations"; // 自定义等默认 OpenAI 兼容格式
    }
  }

  return baseUrl + endpoint;
});

const rules: FormRules = {
  name: [{ required: true, message: "请输入配置名称", trigger: "blur" }],
  provider: [{ required: true, message: "请选择厂商", trigger: "change" }],
  base_url: [
    { required: true, message: "请输入 Base URL", trigger: "blur" },
    { type: "url", message: "请输入正确的 URL 格式", trigger: "blur" },
  ],
  api_key: [
    {
      required: true,
      message: "请输入 API Key",
      trigger: "blur",
      validator: (_rule: unknown, value: string, callback: (err?: Error) => void) => {
        if (isEdit.value && !value?.trim() && originalApiKeyForTest.value) {
          callback(); // 编辑时留空表示不改，已有 Key 则通过
          return;
        }
        if (value?.trim()) {
          callback();
          return;
        }
        callback(new Error("请输入 API Key"));
      },
    },
  ],
  model: [
    {
      required: true,
      message: "请至少选择一个模型",
      trigger: "change",
      validator: (rule: any, value: any, callback: any) => {
        if (Array.isArray(value) && value.length > 0) {
          callback();
        } else if (typeof value === "string" && value.length > 0) {
          callback();
        } else {
          callback(new Error("请至少选择一个模型"));
        }
      },
    },
  ],
};

const loadConfigs = async () => {
  loading.value = true;
  try {
    configsAll.value = await aiAPI.list();
  } catch (error: any) {
    ElMessage.error(error.message || "加载失败");
  } finally {
    loading.value = false;
  }
};

// 当前已配置摘要（用于引导区展示，显示各类型第一个已激活配置）
const configSummary = computed(() => {
  const all = configsAll.value.filter((c) => c.is_active);
  const textCfg = all.find((c) => c.service_type === "text");
  const imgCfg = all.find((c) => c.service_type === "image");
  const vidCfg = all.find((c) => c.service_type === "video");
  const fmt = (c: AIServiceConfig) => {
    const m = Array.isArray(c.model) ? c.model[0] : c.model;
    return `${c.name}${m ? ` (${m})` : ""}`;
  };
  return {
    text: textCfg ? fmt(textCfg) : "",
    image: imgCfg ? fmt(imgCfg) : "",
    video: vidCfg ? fmt(vidCfg) : "",
  };
});

// 生成随机配置名称
const generateConfigName = (
  provider: string,
  serviceType: AIServiceType,
): string => {
  const providerNames: Record<string, string> = {
    openai: "OpenAI",
    gemini: "Gemini",
    google: "Google",
    doubao: "火山引擎",
    volcengine: "火山引擎",
    volces: "火山引擎",
    custom: "自定义",
  };

  const serviceNames: Record<AIServiceType, string> = {
    text: "文本",
    image: "图片",
    video: "视频",
  };

  const randomNum = Math.floor(Math.random() * 10000)
    .toString()
    .padStart(4, "0");
  const providerName = providerNames[provider] || provider;
  const serviceName = serviceNames[serviceType] || serviceType;

  return `${providerName}-${serviceName}-${randomNum}`;
};

const showCreateDialog = () => {
  isEdit.value = false;
  editingId.value = undefined;
  savedApiKeyMask.value = "";
  originalApiKeyForTest.value = "";
  testPassed.value = false;
  resetForm();
  form.service_type = activeTab.value;
  const providers = providerConfigs[activeTab.value] || [];
  const defaultProvider = providers[0]?.id || "openai";
  form.provider = defaultProvider;
  handleProviderChange();
  form.name = generateConfigName(defaultProvider, activeTab.value);
  dialogVisible.value = true;
};

const handleEdit = (config: AIServiceConfig) => {
  isEdit.value = true;
  editingId.value = config.id;
  testPassed.value = false; // 编辑时需重新测试
  // 已保存的 Key 仅显示前4位，留空则不修改；原始 Key 用于测试连接
  const raw = config.api_key || "";
  savedApiKeyMask.value = raw.length >= 4 ? raw.slice(0, 4) + "****" : "****";
  originalApiKeyForTest.value = raw;

  Object.assign(form, {
    service_type: config.service_type,
    provider: config.provider || "openai",
    name: config.name,
    base_url: config.base_url,
    api_key: "", // 编辑时留空，留空保存则不改；输入新 Key 则更新
    model: Array.isArray(config.model) ? config.model : [config.model],
    priority: config.priority || 0,
    is_active: config.is_active,
  });
  dialogVisible.value = true;
};

const handleDelete = async (config: AIServiceConfig) => {
  try {
    await ElMessageBox.confirm("确定要删除该配置吗？", "警告", {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    });

    await aiAPI.delete(config.id);
    ElMessage.success("删除成功");
    loadConfigs();
  } catch (error: any) {
    if (error !== "cancel") {
      ElMessage.error(error.message || "删除失败");
    }
  }
};

const handleToggleActive = async (config: AIServiceConfig) => {
  try {
    const newActiveState = !config.is_active;
    await aiAPI.update(config.id, { is_active: newActiveState });
    ElMessage.success(newActiveState ? "已启用配置" : "已禁用配置");
    await loadConfigs();
  } catch (error: any) {
    ElMessage.error(error.message || "操作失败");
  }
};

// 用于测试的 API Key：新建用表单值，编辑且未输入新 Key 时用已保存的
const apiKeyForTest = computed(() =>
  form.api_key?.trim() || (isEdit.value ? originalApiKeyForTest.value : ""),
);

const testConnection = async () => {
  if (!formRef.value) return;
  const keyToTest = apiKeyForTest.value;
  if (!keyToTest) {
    ElMessage.warning("请输入 API Key");
    return;
  }

  const valid = await formRef.value.validate().catch(() => false);
  if (!valid) return;

  testing.value = true;
  testPassed.value = false;
  try {
    const isVolc = ["doubao", "volcengine", "volces"].includes(form.provider);
    if (isVolc) {
      await aiAPI.testConnectionAll({
        base_url: form.base_url,
        api_key: keyToTest,
        provider: form.provider,
      });
      ElMessage.success("文本、图片、视频三个模型测试通过，可保存！");
    } else {
      await aiAPI.testConnection({
        base_url: form.base_url,
        api_key: keyToTest,
        model: form.model,
        provider: form.provider,
        service_type: form.service_type,
      });
      ElMessage.success("连接测试成功！");
    }
    testPassed.value = true;
  } catch (error: any) {
    const msg = error?.response?.data?.error?.message || error?.message || "连接测试失败";
    ElMessage.error(msg);
  } finally {
    testing.value = false;
  }
};

const handleTest = async (config: AIServiceConfig) => {
  testing.value = true;
  try {
    const provider = config.provider || "";
    const isVolc = ["doubao", "volcengine", "volces"].includes(provider);
    if (isVolc) {
      await aiAPI.testConnectionAll({
        base_url: config.base_url,
        api_key: config.api_key,
        provider,
      });
      ElMessage.success("文本、图片、视频三个模型测试通过！");
    } else {
      await aiAPI.testConnection({
        base_url: config.base_url,
        api_key: config.api_key,
        model: config.model,
        provider,
        service_type: config.service_type,
      });
      ElMessage.success("连接测试成功！");
    }
  } catch (error: any) {
    const msg = error?.response?.data?.error?.message || error?.message || "连接测试失败";
    ElMessage.error(msg);
  } finally {
    testing.value = false;
  }
};

const VOLC_BASE_URL = "https://ark.cn-beijing.volces.com/api/v3";
const volcSimpleConfigs: Array<{
  service_type: AIServiceType;
  provider: string;
  model: string[];
  namePrefix: string;
}> = [
  { service_type: "text", provider: "doubao", model: ["doubao-1-5-pro-32k-250115"], namePrefix: "火山方舟 Doubao" },
  { service_type: "image", provider: "volcengine", model: ["doubao-seedream-4-0-250828"], namePrefix: "火山方舟 Seedream 4.0" },
  { service_type: "video", provider: "volces", model: ["doubao-seedance-1-5-pro-251215"], namePrefix: "火山方舟 Seedance 1.5 Pro" },
];

const handleSubmit = async () => {
  if (!formRef.value) return;

  await formRef.value.validate(async (valid) => {
    if (!valid) return;

    submitting.value = true;
    try {
      if (isEdit.value && editingId.value) {
        const updateData: UpdateAIConfigRequest = {
          name: form.name,
          provider: form.provider,
          base_url: form.base_url,
          model: form.model,
          priority: form.priority,
          is_active: form.is_active,
        };
        if (form.api_key?.trim()) {
          updateData.api_key = form.api_key.trim();
        }
        await aiAPI.update(editingId.value, updateData);
        ElMessage.success("更新成功");
      } else if (isVolcSimpleMode.value) {
        const apiKey = form.api_key?.trim() || (isEdit.value ? originalApiKeyForTest.value : "") || "";
        for (const spec of volcSimpleConfigs) {
          const name = `${spec.namePrefix}-${Math.floor(Math.random() * 10000).toString().padStart(4, "0")}`;
          await aiAPI.create({
            service_type: spec.service_type,
            provider: spec.provider,
            name,
            base_url: VOLC_BASE_URL,
            api_key: apiKey,
            model: spec.model,
            priority: 0,
          });
        }
        ElMessage.success("文本、图片、视频三条配置已保存");
      } else {
        await aiAPI.create(form);
        ElMessage.success("保存成功");
      }

      dialogVisible.value = false;
      loadConfigs();
    } catch (error: any) {
      ElMessage.error(error.message || "操作失败");
    } finally {
      submitting.value = false;
    }
  });
};

const handleTabChange = (tabName: string | number) => {
  // 标签页切换时重新加载对应服务类型的配置
  activeTab.value = tabName as AIServiceType;
  loadConfigs();
};

watch(
  () => [form.api_key, form.base_url, form.model, form.provider],
  () => { testPassed.value = false; },
  { deep: true }
);

const handleProviderChange = () => {
  testPassed.value = false;

  if (form.provider === "custom") {
    form.model = [];
    form.base_url = "";
  } else {
    const p = providerConfigs[form.service_type]?.find((x) => x.id === form.provider);
    form.model = p?.models?.length ? [p.models[0]] : [];
  }

  if (form.provider === "gemini" || form.provider === "google") {
    form.base_url = "https://generativelanguage.googleapis.com";
  } else if (form.provider === "doubao" || form.provider === "volcengine" || form.provider === "volces") {
    form.base_url = "https://ark.cn-beijing.volces.com/api/v3";
  } else {
    form.base_url = "https://api.openai.com/v1";
  }

  if (!isEdit.value) {
    form.name = generateConfigName(form.provider, form.service_type);
  }
};

// getDefaultEndpoint 已移除，端点由后端根据 provider 自动设置
// 保留该函数定义以避免编译错误
const getDefaultEndpoint = (serviceType: AIServiceType): string => {
  switch (serviceType) {
    case "text":
      return "";
    case "image":
      return "/v1/images/generations";
    case "video":
      return "/v1/video/generations";
    default:
      return "/v1/chat/completions";
  }
};

const resetForm = () => {
  const serviceType = form.service_type || "text";
  Object.assign(form, {
    service_type: serviceType,
    provider: "",
    name: "",
    base_url: "",
    api_key: "",
    model: [], // 改为空数组
    priority: 0,
    is_active: true,
  });
  formRef.value?.resetFields();
};

const goBack = () => {
  router.back();
};

onMounted(() => {
  loadConfigs();
});
</script>

<style scoped>
/* ========================================
   Page Layout / 页面布局 - 紧凑边距
   ======================================== */
.page-container {
  min-height: 100vh;
  background: var(--bg-primary);
  padding: var(--space-2) var(--space-3);
  transition: background var(--transition-normal);
}

@media (min-width: 768px) {
  .page-container {
    padding: var(--space-3) var(--space-4);
  }
}

@media (min-width: 1024px) {
  .page-container {
    padding: var(--space-4) var(--space-5);
  }
}

.content-wrapper {
  max-width: 1200px;
  margin: 0 auto;
}

.page-container.embedded {
  min-height: auto;
  padding: 0;
}

.embedded-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-4);
}

.embedded-header h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.simple-mode-tip {
  margin-bottom: var(--space-4);
}

/* ========================================
   Tabs / 标签页 - 紧凑内边距
   ======================================== */
.guide-alert {
  margin-bottom: var(--space-4);
}

.guide-title {
  font-weight: 600;
  font-size: 0.9375rem;
}

.guide-desc {
  margin: 0.5rem 0 0.75rem 0;
  font-size: 0.8125rem;
  color: var(--text-secondary);
}

.guide-list {
  margin: 0;
  padding-left: 1.25rem;
  font-size: 0.8125rem;
  color: var(--text-secondary);
}

.guide-list li {
  margin-bottom: 0.25rem;
}

.guide-list a {
  color: var(--el-color-primary);
  text-decoration: none;
}

.guide-list a:hover {
  text-decoration: underline;
}

.guide-note {
  margin: 0.75rem 0 0;
  font-size: 0.75rem;
  color: var(--text-muted);
}

.guide-note code {
  background: var(--bg-secondary);
  padding: 0.125rem 0.375rem;
  border-radius: 4px;
  font-size: 0.75rem;
}

.guide-configured {
  margin: 0.75rem 0 0;
  font-size: 0.8125rem;
  color: var(--text-secondary);
}

.guide-configured-empty {
  color: var(--text-muted);
  font-style: italic;
}

.tabs-wrapper {
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  padding: var(--space-3);
  box-shadow: var(--shadow-card);
}

@media (min-width: 768px) {
  .tabs-wrapper {
    padding: var(--space-4);
  }
}

/* ========================================
   Form Tips / 表单提示
   ======================================== */
.form-tip {
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-top: var(--space-1);
}

/* ========================================
   Dialog / 弹窗 - glass
   ======================================== */
:deep(.el-dialog) {
  background: var(--glass-bg-heavy);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-2xl);
}

:deep(.el-dialog__header) {
  padding: var(--space-5) var(--space-6);
  border-bottom: 1px solid var(--border-primary);
  margin-right: 0;
}

:deep(.el-dialog__title) {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

:deep(.el-dialog__body) {
  padding: var(--space-6);
}

:deep(.el-dialog__footer) {
  padding: var(--space-4) var(--space-6);
  border-top: 1px solid var(--border-primary);
}

/* ========================================
   Dark Mode / 深色模式
   ======================================== */
.dark .tabs-wrapper {
  background: var(--bg-card);
}

.dark :deep(.el-dialog) {
  background: var(--glass-bg-heavy);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border: 1px solid var(--glass-border);
}

.dark :deep(.el-dialog__header) {
  background: transparent;
}

.dark :deep(.el-dialog__body) {
  background: transparent;
}

.dark :deep(.el-form-item__label) {
  color: var(--text-primary);
}

.dark :deep(.el-input__wrapper) {
  background: var(--bg-secondary);
  box-shadow: 0 0 0 1px var(--border-primary) inset;
}

.dark :deep(.el-input__inner) {
  color: var(--text-primary);
}

.dark :deep(.el-input__inner::placeholder) {
  color: var(--text-muted);
}

.dark :deep(.el-select .el-input__wrapper) {
  background: var(--bg-secondary);
}

.dark :deep(.el-textarea__inner) {
  background: var(--bg-secondary);
  color: var(--text-primary);
  box-shadow: 0 0 0 1px var(--border-primary) inset;
}

.dark :deep(.el-input-number) {
  background: var(--bg-secondary);
}

.dark :deep(.el-switch__core) {
  background: var(--bg-secondary);
  border-color: var(--border-primary);
}

.dark :deep(.el-button--default) {
  background: var(--bg-secondary);
  border-color: var(--border-primary);
  color: var(--text-primary);
}

.dark :deep(.el-button--default:hover) {
  background: var(--bg-card-hover);
  border-color: var(--border-secondary);
}
</style>
