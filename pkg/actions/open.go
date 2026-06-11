package actions

import (
	"fmt"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
)

// OpenParams 定义了 open 命令所需的业务参数。
type OpenParams struct {
	NoteName  string // 笔记名称或路径
	Section   string // 要定位到的标题（heading），可选
	UseEditor bool   // 是否使用编辑器打开而非 Obsidian
}

// OpenNote 是 "open" 命令的业务核心。
// 根据 UseEditor 参数决定用编辑器直接打开文件，还是构造 Obsidian URI 唤起 Obsidian 应用。
func OpenNote(vault obsidian.VaultManager, uri obsidian.UriManager, params OpenParams) error {
	// 获取默认 vault 名称（构造 Obsidian URI 需要）
	vaultName, err := vault.DefaultName()
	if err != nil {
		return err
	}

	// 如果使用编辑器打开，直接操作文件系统，不需要 URI
	if params.UseEditor {
		if params.Section != "" {
			return fmt.Errorf("--section cannot be used with --editor")
		}
		vaultPath, err := vault.Path()
		if err != nil {
			return err
		}
		filePath, err := obsidian.ValidatePath(vaultPath, obsidian.AddMdSuffix(params.NoteName))
		if err != nil {
			return err
		}
		return obsidian.OpenInEditor(filePath)
	}

	// 构造 file 参数：如果指定了 section，追加 #heading 以支持 Obsidian 的定位功能
	fileParam := params.NoteName
	if params.Section != "" {
		fileParam = params.NoteName + "#" + params.Section
	}

	// 构造 Obsidian URI 并执行
	obsidianUri := uri.Construct(ObsOpenUrl, map[string]string{
		"vault": vaultName,
		"file":  fileParam,
	})

	return uri.Execute(obsidianUri)
}
