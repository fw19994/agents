package main

import (
	"log"
	"net/http"
	"os"

	"translate-agent/internal/config"
	"translate-agent/internal/handler"
	"translate-agent/internal/server"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config load: ", err)
	}
	addr := cfg.Addr
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}
	dataDir := cfg.DataDir
	prefix := cfg.HTTPRoutePrefix()
	log.Printf("env=%s addr=%s data_dir=%s providers=%d project_path=%q", config.Env, addr, dataDir, len(cfg.Providers), prefix)

	r := gin.Default()
	server.CORS(r)
	if prefix != "" {
		r.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, prefix+"/home.html")
		})
	}
	server.Static(r, prefix)
	handler.New(&cfg, dataDir).Register(r)

	if err := server.Listen(addr, r); err != nil {
		log.Fatal("listen: ", err)
	}
}
