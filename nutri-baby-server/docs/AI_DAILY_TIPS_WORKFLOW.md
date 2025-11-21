# AI 每日建议生成工作流详解

本文档详细介绍了 `AnalysisChainBuilder` 中每日建议（Daily Tips）生成的工作原理，重点解释了循环机制、早停策略（Early Stop）以及上下文管理（Context Management）的设计思考。

## 1. 核心工作流程 (The Loop Mechanism)

AI 分析链采用**迭代式对话循环（Iterative Conversational Loop）**来处理复杂的分析任务。

### 工作原理
1.  **初始化**：构建 `System Prompt`（角色与规则）和 `User Prompt`（具体任务），并预加载必要的历史数据（Batch Data）。
2.  **进入循环**：系统进入一个最大为 `maxIterations` (10次) 的循环。
3.  **LLM 生成**：在每一轮中，将当前的 `messages` 历史发送给 LLM。
4.  **决策分支**：
    *   **情况 A：工具调用 (Tool Calls)**
        *   如果 LLM 决定需要更多数据，它会返回 `ToolCalls`。
        *   系统执行这些工具（如 `get_feeding_data`），获取数据。
        *   将工具执行结果作为 `Tool Message` 追加到 `messages` 历史中。
        *   **继续下一轮循环**，让 LLM 基于新数据继续思考。
    *   **情况 B：最终响应 (Final Response)**
        *   如果 LLM 认为数据足够，或者直接根据预加载数据生成了结果，它会返回纯文本内容（JSON）。
        *   此时 `ToolCalls` 为空。
        *   系统触发**早停机制**。

## 2. 为什么 earlyStopThreshold 设置为 1？

### 背景
在之前的实现中，`earlyStopThreshold` 被设置为 2，意味着系统需要连续两次“无工具调用”才会停止。这导致了以下问题：
1.  **第 1 轮**：LLM 返回了完美的 JSON 建议（无工具调用）。
2.  **第 2 轮**：系统因为阈值未到，将这个 JSON 建议作为历史又发回给了 LLM。
3.  **结果**：LLM 看到任务已完成，往往会返回空字符串、"已完成" 或其他非 JSON 内容。
4.  **报错**：系统尝试解析这第 2 轮的无效响应，导致 `unexpected end of JSON input` 错误。

### 优化方案
我们将 `earlyStopThreshold` 设置为 **1**。

### 原因
*   **明确的结束信号**：我们的 Prompt 强制要求 LLM **只返回 JSON**。当 LLM 返回非 ToolCall 的内容时，它必然就是我们要的最终 JSON 结果。
*   **避免冗余**：一旦获得 JSON，任务即告完成。没有任何理由再进行下一轮对话。
*   **提高性能**：减少了一次无意义的 LLM 调用，降低了延迟和 Token 消耗。

## 3. 上下文管理：为什么 "System + User + Latest" 就够了？

为了防止 Token 超出限制并保持 LLM 专注，我们使用了 `trimMessageHistory` 方法来修剪消息历史。

### 保留策略
我们始终保留：
1.  **System Prompt (第 1 条)**：包含“你是谁”、“输出 JSON 格式”等核心**规则**。如果丢失，LLM 会忘记格式要求，导致解析失败。
2.  **User Prompt (第 2 条)**：包含“分析宝宝 ID xxx”等核心**任务**。如果丢失，LLM 会忘记它在做什么。
3.  **Latest Messages (最近 N 条)**：包含最近的工具调用和数据。这是 LLM 当前推理所需的**短期记忆**。

### 为什么中间的历史可以丢弃？
*   **过时信息**：早期的工具调用（比如第 1 轮查的“基本信息”）通常已经被 LLM 消化并整合到了后续的推理中，或者 LLM 已经基于这些信息决定了下一步动作。
*   **上下文窗口限制**：无限保留历史会导致 Token 溢出。
*   **抗干扰**：过长的历史包含大量中间步骤的噪音，反而可能干扰 LLM 对当前最终目标的判断。

通过“**首尾保留法**”，我们确保了 LLM 既不忘“初心”（System/User），又能掌握“当下”（Latest），从而保证最终输出的正确性。
