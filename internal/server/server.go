package server

import (
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS 注册 CORS 中间件
func CORS(r *gin.Engine) {
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
}

// Static 注册静态资源与页面路由。prefix 非空时资源挂在 /{prefix}/ 下（与 API 一致）。
func Static(r *gin.Engine, prefix string) {
	if prefix == "" {
		r.StaticFile("/", "./static/index.html")
		r.StaticFile("/index.html", "./static/index.html")
		r.Static("/css", "./static/css")
		r.Static("/js", "./static/js")
		r.StaticFile("/home.html", "./static/home.html")
		r.StaticFile("/settings.html", "./static/settings.html")
		r.StaticFile("/evaluate.html", "./static/evaluate.html")
		return
	}
	g := r.Group(prefix)
	g.StaticFile("/", "./static/index.html")
	g.StaticFile("/index.html", "./static/index.html")
	g.Static("/css", "./static/css")
	g.Static("/js", "./static/js")
	g.StaticFile("/home.html", "./static/home.html")
	g.StaticFile("/settings.html", "./static/settings.html")
	g.StaticFile("/evaluate.html", "./static/evaluate.html")
}

// Listen 在 addr 上监听并服务，绑定失败时输出友好提示
func Listen(addr string, r *gin.Engine) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		if strings.Contains(err.Error(), "address already in use") || strings.Contains(err.Error(), "bind") {
			log.Printf("[ERROR] 端口 %s 已被占用。请修改 config/config.json 中的 addr（如 :9002）或设置环境变量 ADDR=:9002，也可关闭占用该端口的进程。", addr)
		}
		return err
	}
	log.Println("listen", addr)
	return http.Serve(ln, r)
}
