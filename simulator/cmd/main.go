package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"simulator/internal/simulator"
)

func main() {
	// 定义命令行参数
	configPath := flag.String("config", "configs/simulator.json", "配置文件路径")
	backendURL := flag.String("backend", "", "后端API地址")
	flag.Parse()

	// 创建并初始化模拟器管理器
	manager, err := simulator.NewManager(*configPath)
	if err != nil {
		fmt.Printf("初始化模拟器管理器失败: %v\n", err)
		os.Exit(1)
	}

	// 如果提供了后端URL，则覆盖配置中的URL
	if *backendURL != "" {
		// 注意：这个功能现在需要通过manager的方法提供
	}

	// 启动模拟器
	if err := manager.Start(); err != nil {
		fmt.Printf("启动模拟器失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=================================================")
	fmt.Println("              充电桩模拟器已启动")
	fmt.Println("=================================================")
	fmt.Println("输入 'help' 获取命令列表")

	// 等待中断信号以优雅地关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n正在关闭模拟器...")
	manager.Stop()
	fmt.Println("模拟器已关闭")
}
