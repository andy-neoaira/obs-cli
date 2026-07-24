package actions

// Obsidian URI 协议相关常量。
// Obsidian 支持通过自定义 URI 协议打开笔记，格式为 obsidian://open?vault=...&file=...
const (
	obsBaseUrl = "obsidian://" // Obsidian URI 协议前缀
	openAction = "open"        // 打开动作

	ObsOpenUrl = obsBaseUrl + openAction // 完整的打开 URI 基础地址
)
