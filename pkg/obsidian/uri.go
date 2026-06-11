package obsidian

import (
	"errors"
	"net/url"
	"strings"

	"github.com/skratchdot/open-golang/open"
)

// Uri 是 UriManager 的具体实现，负责构造和执行 Obsidian URI。
type Uri struct{}

// UriManager 定义了 URI 构造与执行的接口。
// 通过接口解耦，方便在测试中注入 Mock 对象。
type UriManager interface {
	Construct(baseUri string, params map[string]string) string // 构造带查询参数的 URI
	Execute(uri string) error                                  // 执行 URI（唤起系统默认程序）
}

// Construct 将基础 URI 与参数 map 拼接成完整的查询字符串。
// 空值会被忽略；空格会被编码为 %20 而非 +，以兼容 Obsidian。
func (u *Uri) Construct(baseUri string, params map[string]string) string {
	uri := baseUri
	for key, value := range params {
		if value != "" {
			// url.QueryEscape 会将空格编码为 +，Obsidian 需要 %20
			encoded := strings.ReplaceAll(url.QueryEscape(value), "+", "%20")
			if uri == baseUri {
				uri += "?" + key + "=" + encoded
			} else {
				uri += "&" + key + "=" + encoded
			}
		}
	}
	return uri
}

// Run 是一个包级变量，指向 open.Run 函数。
// 在测试中可以通过替换此变量来避免真正唤起外部程序。
var Run = open.Run

// Execute 调用系统默认程序打开给定的 URI。
func (u *Uri) Execute(uri string) error {
	err := Run(uri)
	if err != nil {
		return errors.New(ExecuteUriError)
	}
	return nil
}
