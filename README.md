# 职能沟通翻译助手

产品 ↔ 开发 沟通翻译 Web 应用，支持多轮对话、身份切换、模型与参数设置、Agent 评测。后端 Go + Gin，**Agent 使用 [Eino](https://github.com/cloudwego/eino) 框架**（CloudWeGo 的 LLM 应用框架），前端 HTML + Tailwind CSS + FontAwesome，数据落盘到本地文件。

对话页助手回复支持 **Markdown 渲染**（标题、列表、表格、代码块等），样式见 `static/css/common.css` 中 `.md-body`。

### 会话上下文（模型可见历史）

同一 `session_id` 下，服务端会注入该会话近期「用户原文 + 助手回复」作为模型上下文。轮数与截断见 `translate.go` / `eino_agent.go`。**内置工作指引仅在服务端加载，API 与前端只返回翻译正文，不暴露指引原文。**

## 快速开始

### 环境

- Go 1.21+
- 大模型 API Key（如 OpenAI）

### 配置文件（按环境区分）

API Key 与运行参数放在配置文件中，通过 **环境名** 区分：

1. 复制示例配置并填写 API Key：
   ```bash
   cp config/config.example.json config/config.json
   # 编辑 config/config.json，在 dev 或 prod 中填写 openai_api_key
   ```

2. 配置文件结构（`config/config.json`）支持两种方式：
   - **单供应商**：仅填 `openai_api_key`、`llm_base_url`，所有模型走该端点。
   - **多供应商**：填 `providers`，按模型 ID 自动选用对应供应商的 Key 与 Base URL。
   ```json
   {
     "dev": {
       "openai_api_key": "sk-xxx",
       "llm_base_url": "https://api.openai.com/v1",
       "addr": ":8080",
       "data_dir": ".",
       "providers": {
         "openai": {
           "api_key": "sk-xxx",
           "base_url": "https://api.openai.com/v1",
           "models": ["gpt-4o-mini", "gpt-4o"]
         },
         "qwen": {
           "api_key": "sk-xxx",
           "base_url": "https://dashscope.aliyuncs.com/compatible-mode/v1",
           "models": ["qwen-plus", "qwen-turbo"]
         }
       }
     }
   }
   ```
   前端选择某模型（如 `qwen-plus`）时，后端会根据 `providers` 中该模型所属供应商使用对应的 `api_key` 和 `base_url`。`/api/models` 会返回各 provider 下 `models` 的并集。

3. 通过环境变量选择使用哪套配置：
   - `APP_ENV=dev`（默认）：使用 `dev` 段
   - `APP_ENV=prod`：使用 `prod` 段

**注意**：`config/config.json` 已加入 `.gitignore`，请勿提交真实 Key；仅提交 `config.example.json` 作为模板。

### 安装与运行

```bash
cd /path/to/agents
go mod tidy
cp config/config.example.json config/config.json
# 编辑 config.json 填写 openai_api_key（或保留空，用环境变量兜底）
APP_ENV=dev go run ./cmd/server
```

浏览器访问：**http://localhost:8080**

若提示 **端口已被占用**，可任选其一：
- 修改 `config/config.json` 中当前环境的 `addr`（如 `":8081"`）；
- 或启动时设置环境变量：`ADDR=:8081 go run ./cmd/server`。

- 入口：`/`（会跳转到对话页）
- 对话：`/home.html`（含模型选择与 Temperature、Max Tokens 参数）
- Agent 评测：`/evaluate.html`

### 配置项与兜底

| 配置项（config.json） | 说明 |
|-----------------------|------|
| `openai_api_key` | 单供应商时的 API Key；多供应商时可作为未匹配模型时的兜底 |
| `llm_base_url` | 单供应商时的 API 根地址 |
| `addr` | 服务监听地址，默认 `:8080` |
| `data_dir` | 数据目录根路径，历史与配置写入其下 `data/` |
| `providers` | 多供应商：`"供应商名": { "api_key", "base_url", "models": ["模型id"] }`，请求时按模型 ID 自动选供应商 |

若未建 `config/config.json` 或某模型未在任一 provider 的 `models` 中配置 Key，程序会从环境变量 `OPENAI_API_KEY` / `API_KEY` 读取 Key 作为兜底。

### 依赖（Eino）

Agent 层依赖 [Eino](https://github.com/cloudwego/eino) 与 [eino-ext OpenAI ChatModel](https://github.com/cloudwego/eino-ext/tree/main/components/model/openai)。首次构建请执行：

```bash
go mod tidy
```

若 Eino 版本升级导致 API 变化，请参照 [Eino 文档](https://www.cloudwego.io/docs/eino/) 与 eino-ext 的 ChatModel 说明调整 `internal/agent/eino_agent.go`。

## 功能说明

- **多轮对话**：在对话页连续输入多轮，历史在当页展示。
- **身份/翻译方向切换**：工具栏切换「产品→开发」「开发→产品」；**切换方向会自动新建会话**（避免同一对话混用两种内置策略），原会话仍在左侧列表可继续查看。
- **第三方向（运营→产品）**：支持将运营活动/投放/策略描述，整理为产品可评审的需求表达。
- **模型与参数**：对话页工具栏内选择模型，并直接设置 Temperature、Max Tokens，无需单独设置页。

## 前端结构

- `static/index.html`：跳转到对话页（保留用于兼容旧链接）。
- `static/home.html`：对话页（多轮、方向、模型选择、Temperature / Max Tokens 参数）。
- `static/settings.html`：保留为跳转到对话页（兼容旧链接）。
- `static/css/common.css`：公共样式（无滚动条、圆角等）。
- `static/js/api.js`：与后端 API 封装（流式翻译等）。
- `static/js/chat.js`：对话页逻辑（发送、流式展示、方向/模型）。

可滚动区域已隐藏滚动条；界面最大宽度 1280px，圆角卡片，真实图片（Unsplash）用于首页。

## 测试用例

下列均为**示例**，用于手工或自测时观察「是否抓住该场景下各自关心的点」；**不要求**每题输出相同结构或篇幅。**题目必做的 2 条**为「产品→开发」「开发→产品」两表中的 **#1**；其余在对话页切换方向后粘贴试用（含「运营→产品」）。

### 产品 → 开发

| # | 输入（摘要） | 预期关注点（示例） |
|---|--------------|-------------------|
| 1 | 我们需要一个智能推荐功能，提升用户停留时长。 | 推荐路径、数据、实时性、工作量等（**题目原示例**） |
| 2 | 想在详情页加「看了又看」，先上一版规则推荐即可。 | MVP 范围、规则 vs 模型、数据与后续迭代 |
| 3 | 用户注册时要手机号 + 验证码，要防刷。 | 频控、风控、存储与合规、异常流程 |
| 4 | 运营要在后台批量改价，要权限和操作日志。 | 权限模型、审计、回滚与误操作防护 |
| 5 | App 启动太慢，希望首屏 2 秒内出来。 | 性能指标、拆分（包体/接口/渲染）、如何度量 |
| 6 | 你好。 | 短答引导，勿长篇技术规格（输入判别） |

### 开发 → 产品

| # | 输入（摘要） | 预期关注点（示例） |
|---|--------------|-------------------|
| 1 | 我们优化了数据库查询，QPS 提升了 30%。 | 体验/容量/成本等业务语言（**题目原示例**） |
| 2 | 订单列表接口加了 Redis 缓存，P95 从 800ms 降到 200ms。 | 用户可感知的快慢、峰值容量、缓存副作用说明 |
| 3 | 本周上了限流，高峰期少量用户会排队 2～3 秒。 | 对体验的影响、是否需文案/开关、建议观察指标 |
| 4 | 推荐从离线批处理改成 5 分钟延迟的近实时。 | 「多实时算够」、对转化/体验的潜在影响 |
| 5 | 修了一个会导致偶发白屏的前端 Bug，影响约 0.1% 会话。 | 影响面、是否需公告、如何向产品汇报 |
| 6 | 容器缩容后每月云费用大概降了 15%。 | 成本结论、对稳定性/扩容余量、产品是否要感知 |

### 运营 → 产品

| # | 输入（摘要） | 预期关注点（示例） |
|---|--------------|-------------------|
| 1 | 双 11 想做满 200 减 30，全站可用，预算 50 万，希望拉新 + 提升客单。 | 目标指标、人群与范围、预算与风控、产品入口与规则边界 |
| 2 | 下周 Push 发 3 波优惠券提醒，怕打扰用户，需要频控和 AB。 | 触达策略、频控与对照组、建议产品看的指标（打开率/转化/退订） |
| 3 | 新用户首单 9 元包邮，老用户不参与，要防薅羊毛。 | 人群规则、黑白名单、风控依赖、待澄清（券叠加/退款规则） |
| 4 | 和抖音达人合作带货，要给专属落地页和归因参数。 | 入口、埋点与归因、活动页需求边界、上线与灰度 |
| 5 | 社群裂变：邀请 3 人得积分，积分可兑周边，库存 500。 | 玩法规则、库存与成本上限、合规与客诉风险 |
| 6 | 你好。 | 短答引导，勿套长篇活动方案模板（输入判别） |

## 提示词设计说明

翻译策略与输出约定由 **Eino ADK 内置能力**按翻译方向在服务端加载并注入模型；**不向用户、HTTP 响应或前端暴露**内置文档原文。Agent 负责编排、流式输出与错误透传；产品→开发、开发→产品、运营→产品等方向对应不同内置策略，并约定因题制宜、避免八股模板。

流式接口 `/api/translate/stream` 在失败时通过 SSE 返回 `error`（含错误链）及 `source: agent`。

（阅卷/答辩如需说明 Skill 文件位置与结构，见仓库内开发文档，README 不展开路径。）
