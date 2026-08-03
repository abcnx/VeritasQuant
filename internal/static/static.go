// Package static 提供内嵌前端静态资源服务（SPA fallback）。
package static

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/acanx/finvquant/internal/webui"
)

// Handler 返回前端静态文件处理器；未知路径回退 index.html（SPA 路由）。
func Handler() http.Handler {
	sub, err := fs.Sub(webui.Dist, "dist")
	if err != nil {
		panic("内嵌前端资源缺失: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(sub, path); err != nil {
			// SPA 回退：非静态资源路径一律返回 index.html
			r.URL.Path = "/index.html"
		}
		fileServer.ServeHTTP(w, r)
	})
}
