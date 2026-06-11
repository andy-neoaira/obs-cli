package cmd

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// stubVaultManager 是测试专用的 VaultManager 简易实现，避免引入完整的 Mock 包。
type stubVaultManager struct {
	defaultName    string
	defaultNameErr error
	path           string
	pathErr        error
	openType       string
	openTypeErr    error
}

func (s *stubVaultManager) DefaultName() (string, error) {
	if s.defaultNameErr != nil {
		return "", s.defaultNameErr
	}
	if s.defaultName == "" {
		return "vault", nil
	}
	return s.defaultName, nil
}

func (s *stubVaultManager) SetDefaultName(name string) error {
	s.defaultName = name
	return nil
}

func (s *stubVaultManager) Path() (string, error) {
	if s.pathErr != nil {
		return "", s.pathErr
	}
	if s.path == "" {
		return "path", nil
	}
	return s.path, nil
}

func (s *stubVaultManager) DefaultOpenType() (string, error) {
	if s.openTypeErr != nil {
		return "", s.openTypeErr
	}
	if s.openType == "" {
		return "obsidian", nil
	}
	return s.openType, nil
}

// newSearchContentOptionsTestCmd 创建一个用于测试的 cobra.Command，注册 search-content 相关的 flag。
func newSearchContentOptionsTestCmd() *cobra.Command {
	c := &cobra.Command{Use: "test"}
	c.Flags().BoolP("editor", "e", false, "")
	c.Flags().Bool("no-interactive", false, "")
	c.Flags().String("format", "text", "")
	c.Flags().Int("page", 0, "")
	c.Flags().Int("page-size", 0, "")
	return c
}

// TestSearchContentCommandFlagsWired 验证 search-content 命令的所有 flag 都已正确注册。
func TestSearchContentCommandFlagsWired(t *testing.T) {
	assert.NotNil(t, searchContentCmd.Flags().Lookup("no-interactive"))
	assert.NotNil(t, searchContentCmd.Flags().Lookup("format"))
	assert.NotNil(t, searchContentCmd.Flags().Lookup("editor"))
	assert.NotNil(t, searchContentCmd.Flags().Lookup("vault"))
	assert.NotNil(t, searchContentCmd.Flags().Lookup("page"))
	assert.NotNil(t, searchContentCmd.Flags().Lookup("page-size"))

	assert.Equal(t, "text", searchContentCmd.Flags().Lookup("format").DefValue)
	assert.Equal(t, "0", searchContentCmd.Flags().Lookup("page").DefValue)
	assert.Equal(t, "0", searchContentCmd.Flags().Lookup("page-size").DefValue)
}

// TestBuildSearchContentOptionsParsesExplicitFlags 测试显式传入的 flag 被正确解析。
func TestBuildSearchContentOptionsParsesExplicitFlags(t *testing.T) {
	c := newSearchContentOptionsTestCmd()
	err := c.ParseFlags([]string{"--editor", "--no-interactive", "--format", "json"})
	assert.NoError(t, err)

	vault := &stubVaultManager{openType: "obsidian"}
	options, err := buildSearchContentOptions(c, vault, false)
	assert.NoError(t, err)
	assert.True(t, options.UseEditor)
	assert.True(t, options.EditorFlagExplicit)
	assert.True(t, options.NoInteractive)
	assert.Equal(t, "json", options.Format)
	assert.False(t, options.InteractiveTerminal)
	assert.NotNil(t, options.Output)
}

// TestBuildSearchContentOptionsRespectsDefaultOpenType 测试当用户没有显式传 --editor 时，
// 是否尊重配置文件中的 default_open_type。
func TestBuildSearchContentOptionsRespectsDefaultOpenType(t *testing.T) {
	c := newSearchContentOptionsTestCmd()
	err := c.ParseFlags([]string{})
	assert.NoError(t, err)

	vault := &stubVaultManager{openType: "editor"}
	options, err := buildSearchContentOptions(c, vault, true)
	assert.NoError(t, err)
	assert.True(t, options.UseEditor)
	assert.False(t, options.EditorFlagExplicit)
	assert.False(t, options.NoInteractive)
	assert.Equal(t, "text", options.Format)
	assert.True(t, options.InteractiveTerminal)
}

// TestBuildSearchContentOptionsDefaultOpenTypeErrorFallsBack 测试读取默认打开方式出错时，
// 是否能优雅回退到不使用编辑器。
func TestBuildSearchContentOptionsDefaultOpenTypeErrorFallsBack(t *testing.T) {
	c := newSearchContentOptionsTestCmd()
	err := c.ParseFlags([]string{})
	assert.NoError(t, err)

	vault := &stubVaultManager{openTypeErr: errors.New("config error")}
	options, err := buildSearchContentOptions(c, vault, true)
	assert.NoError(t, err)
	assert.False(t, options.UseEditor)
	assert.False(t, options.EditorFlagExplicit)
}
