package cmd

import (
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

// resolveUseEditor 判断当前命令是否应该使用编辑器打开笔记。
// 优先级：
//  1. 如果用户在命令行显式传了 --editor/-e，直接使用该值；
//  2. 否则读取 CLI 配置中的 default_open_type，若配置为 "editor" 则返回 true。
func resolveUseEditor(cmd *cobra.Command, vault obsidian.VaultManager) bool {
	useEditor, err := cmd.Flags().GetBool("editor")
	if err != nil {
		return false
	}
	// 用户没有在命令行显式指定 --editor 时，查询配置文件中的默认值
	if !cmd.Flags().Changed("editor") {
		defaultOpenType, configErr := vault.DefaultOpenType()
		if configErr == nil && defaultOpenType == "editor" {
			useEditor = true
		}
	}
	return useEditor
}
