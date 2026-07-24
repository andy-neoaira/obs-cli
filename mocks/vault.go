package mocks

import "os"

// MockVaultOperator 是 VaultManager 接口的 Mock 实现，用于单元测试。
// 通过设置各个字段的值，可以控制 Mock 对象的行为（返回指定值或模拟错误）。
type MockVaultOperator struct {
	DefaultNameErr error  // 模拟 DefaultName() 返回的错误
	PathError      error  // 模拟 Path() 返回的错误
	Name           string // DefaultName() 返回的 vault 名称
	PathValue      string // Path() 返回的路径
	OpenType       string // DefaultOpenType() 返回的打开方式
	OpenTypeErr    error  // 模拟 DefaultOpenType() 返回的错误
}

func (m *MockVaultOperator) DefaultName() (string, error) {
	if m.DefaultNameErr != nil {
		return "", m.DefaultNameErr
	}
	return m.Name, nil
}

func (m *MockVaultOperator) SetDefaultName(_ string) error {
	if m.DefaultNameErr != nil {
		return m.DefaultNameErr
	}
	return nil
}

func (m *MockVaultOperator) Path() (string, error) {
	if m.PathError != nil {
		return "", m.PathError
	}
	if m.PathValue != "" {
		return m.PathValue, nil
	}
	// 安全路径解析要求 Vault 根目录真实存在；未指定时提供只用于 Mock
	// 交互验证的系统临时目录。
	return os.TempDir(), nil
}

func (m *MockVaultOperator) DefaultOpenType() (string, error) {
	if m.OpenTypeErr != nil {
		return "", m.OpenTypeErr
	}
	if m.OpenType != "" {
		return m.OpenType, nil
	}
	return "obsidian", nil
}
