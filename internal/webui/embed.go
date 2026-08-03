// Package webui 内嵌前端构建产物（All-in-One 镜像：单进程双端口）。
//
// 构建流程：Web/ npm run build → dist 由本包 embed，服务端在 16002 端口
// 直接提供前端静态资源（SPA fallback），16001 端口提供 API。
package webui

import "embed"

// Dist 内嵌的前端静态文件（Web/dist 构建产物）。
//
//go:embed dist
var Dist embed.FS
