package obsidian

import (
	"errors"

	"github.com/andy-neoaira/obs-cli/pkg/pathpolicy"
)

// ErrPathTraversal 保留旧 API 名称；V2 稳定错误语义为 PATH_OUTSIDE_VAULT。
var ErrPathTraversal = pathpolicy.ErrOutsideVault

// ErrPhysicalPathConflict 表示逻辑路径通过 Vault 内符号链接形成物理别名。
// 读取可以使用该别名；修改必须要求调用方选择无歧义的规范逻辑路径。
var ErrPhysicalPathConflict = errors.New("physical path identity conflict: mutation through symbolic link is forbidden")

// ValidatePath 是统一 Vault path policy 的兼容入口。
//
// 新代码如需显式读取隐藏路径，应直接使用 pathpolicy.Resolver 并传入可审计选项；
// 普通 Note、附件、模板和 Daily Note 不允许隐藏路径。空路径仅供 Vault 根目录
// 查询兼容使用。
func ValidatePath(basePath, relativePath string) (string, error) {
	result, err := ResolvePath(basePath, relativePath)
	if err != nil {
		return "", err
	}
	return result.Path, nil
}

// ValidateWritablePath 在普通边界检查之上拒绝经 Vault 内符号链接的物理别名。
func ValidateWritablePath(basePath, relativePath string) (string, error) {
	result, err := ResolvePath(basePath, relativePath)
	if err != nil {
		return "", err
	}
	if result.ThroughSymlink {
		return "", ErrPhysicalPathConflict
	}
	return result.Path, nil
}

// ResolvePath 返回统一 resolver 的完整身份信息。
func ResolvePath(basePath, relativePath string) (pathpolicy.Result, error) {
	resolver, err := pathpolicy.NewResolver(basePath)
	if err != nil {
		return pathpolicy.Result{}, err
	}
	result, err := resolver.Resolve(relativePath, pathpolicy.ResolveOptions{AllowRoot: true})
	if err != nil {
		return pathpolicy.Result{}, err
	}
	return result, nil
}
