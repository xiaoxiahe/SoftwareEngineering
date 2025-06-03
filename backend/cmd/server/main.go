package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/api"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/service"
)

func main() {

	// 加载配置
	cfg, err := config.Load("./configs/config.json")
	if err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}

	// 初始化数据库连接
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 运行数据库迁移
	if err := database.RunMigrations(cfg.Database); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	// 初始化服务
	services := service.NewServices(db, cfg)

	// 初始化系统配置和数据
	if err := services.Bootstrap.InitializeSystem(); err != nil {
		log.Printf("系统初始化警告: %v", err)
	} else {
		log.Println("系统配置和初始数据加载成功")
	}

	// 初始化路由
	router := api.SetupRouter(services, cfg)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// 启动HTTP服务器
	go func() {
		log.Printf("后端服务启动在 http://localhost:%d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP服务器启动失败: %v", err)
		}
	}()

	// 等待信号以优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	// 创建一个带有超时的上下文来等待关闭
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	// 优雅关闭服务器
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("服务器强制关闭: %v", err)
	}

	log.Println("服务器成功关闭")
}
