package llm

// SystemPrompt defines the LLM's persona and role
const SystemPrompt = `你是一个贴心的天气助手，专门为用户提供实用的天气建议。你的任务是：
1. 用简洁、友好的中文语言与用户交流
2. 根据天气数据提供具体、可操作的建议
3. 关注用户的日常出行和生活需求
4. 在天气变化显著时提醒用户注意事项
5. 保持语气温暖、自然，避免过于正式或机械

请用2-3句话提供建议，重点突出最重要的信息。`

// Note: The prompt building functions have been moved to internal/notifications/handlers.go
// where they are implemented using WeatherResponse instead of a Weather type.
// See buildMorningPromptFromResponse, buildEveningPromptFromResponse, and
// buildChangeAlertPromptFromResponse in the notifications package.
