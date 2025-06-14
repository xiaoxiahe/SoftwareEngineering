package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"simulator/internal/simulator"
)

func main() { // 定义命令行参数
	configPath := flag.String("config", "configs/simulator.json", "配置文件路径")
	backendURL := flag.String("backend", "", "后端API地址")
	useSimClock := flag.Bool("sim-clock", false, "使用模拟时钟")
	simClockTime := flag.String("sim-time", "", "设置模拟时钟时间 (RFC3339格式，如: 2024-01-01T08:00:00Z)")
	flag.Parse()

	// 创建并初始化模拟器管理器
	manager, err := simulator.NewManager(*configPath)
	if err != nil {
		fmt.Printf("初始化模拟器管理器失败: %v\n", err)
		os.Exit(1)
	}

	// 如果提供了后端URL，则覆盖配置中的URL
	if *backendURL != "" {
		if err := manager.SetBackendURL(*backendURL); err != nil {
			fmt.Printf("设置后端URL失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 处理模拟时钟设置
	if *useSimClock {
		var startTime time.Time
		if *simClockTime != "" {
			startTime, err = time.Parse(time.RFC3339, *simClockTime)
			if err != nil {
				fmt.Printf("无效的时间格式: %v\n", err)
				os.Exit(1)
			}
		} else {
			startTime = time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
		}
		if err := manager.EnableSimulatedClock(startTime); err != nil {
			fmt.Printf("启用模拟时钟失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("已启用模拟时钟: 起始时间=%s\n",
			startTime.Format(time.RFC3339))
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
