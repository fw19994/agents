package main

import (
	"log"
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
	log.Printf("env=%s addr=%s data_dir=%s providers=%d", config.Env, addr, dataDir, len(cfg.Providers))

	r := gin.Default()
	server.CORS(r)
	server.Static(r)
	handler.New(&cfg, dataDir).Register(r)

	if err := server.Listen(addr, r); err != nil {
		log.Fatal("listen: ", err)
	}
}
