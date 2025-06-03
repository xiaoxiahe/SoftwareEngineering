package simulator

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"simulator/internal/config"
	"simulator/internal/utils"
)

// Manager 模拟器管理器
type Manager struct {
	simulator  *PileSimulator
	config     *config.Config
	logger     *utils.Logger
	isRunning  bool
	stopCh     chan struct{}
	wg         sync.WaitGroup
	configPath string
}

// NewManager 创建模拟器管理器
func NewManager(configPath string) (*Manager, error) {
	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 创建日志工具
	logger := utils.NewLogger(cfg.Simulation.LogLevel)

	// 创建模拟器
	simulator := NewPileSimulator(cfg, logger)

	return &Manager{
		simulator:  simulator,
		config:     cfg,
		logger:     logger,
		stopCh:     make(chan struct{}),
		configPath: configPath,
	}, nil
}

// Start 启动管理器
func (m *Manager) Start() error {
	if m.isRunning {
		return nil
	}

	m.logger.Info("启动模拟器管理器")
	m.isRunning = true

	// 启动模拟器
	if err := m.simulator.Start(); err != nil {
		return fmt.Errorf("启动模拟器失败: %w", err)
	}

	// 启动命令行界面
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.runCLI()
	}()

	return nil
}

// Stop 停止管理器
func (m *Manager) Stop() error {
	if !m.isRunning {
		return nil
	}

	m.logger.Info("停止模拟器管理器")
	close(m.stopCh)

	// 停止模拟器
	if err := m.simulator.Stop(); err != nil {
		m.logger.Error("停止模拟器失败: %v", err)
	}

	// 等待所有协程退出
	m.wg.Wait()
	m.isRunning = false
	return nil
}

// runCLI 运行命令行界面
func (m *Manager) runCLI() {
	m.printHelp()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")

	for scanner.Scan() && m.isRunning {
		line := scanner.Text()
		args := strings.Fields(line)

		if len(args) == 0 {
			fmt.Print("> ")
			continue
		}

		cmd := args[0]
		switch cmd {
		case "help":
			m.printHelp()
		case "status":
			m.showStatus(args)
		case "fault":
			m.triggerFault(args)
		case "recover":
			m.recoverFault(args)
		case "sim":
			m.simulateRequest(args)
		case "reload":
			m.reloadConfig()
		case "exit", "quit", "stop":
			m.Stop()
			return
		default:
			fmt.Println("未知命令，输入 'help' 获取帮助")
		}

		fmt.Print("> ")
	}
}

// printHelp 打印帮助信息
func (m *Manager) printHelp() {
	fmt.Println("\n充电桩模拟器命令行界面")
	fmt.Println("====================")
	fmt.Println("可用命令:")
	fmt.Println("  status [pileID]         - 查看充电桩状态，不指定ID则显示所有")
	fmt.Println("  fault <pileID> <type> <minutes> <desc>")
	fmt.Println("                          - 触发充电桩故障")
	fmt.Println("                            type: hardware/software/power")
	fmt.Println("  recover <pileID>        - 手动恢复故障")
	fmt.Println("  sim <userID> <amount> <mode>")
	fmt.Println("                          - 模拟充电请求")
	fmt.Println("                            mode: fast/trickle")
	fmt.Println("  reload                  - 重新加载配置")
	fmt.Println("  help                    - 显示帮助信息")
	fmt.Println("  exit                    - 退出程序")
	fmt.Println("====================")
}

// showStatus 显示充电桩状态
func (m *Manager) showStatus(args []string) {
	if len(args) > 1 {
		// 显示指定充电桩状态
		pileID := args[1]
		status, vehicle, err := m.simulator.GetPileStatus(pileID)

		if err != nil {
			fmt.Printf("获取充电桩状态失败: %v\n", err)
			return
		}

		fmt.Printf("充电桩 %s 状态:\n", pileID)
		fmt.Printf("  状态: %s\n", status)

		if vehicle != nil {
			fmt.Printf("  当前充电车辆:\n")
			fmt.Printf("    用户ID: %s\n", vehicle.UserID)
			fmt.Printf("    开始时间: %s\n", vehicle.StartTime.Format("15:04:05"))
			fmt.Printf("    请求电量: %.1f kWh\n", vehicle.RequestedCapacity)
			fmt.Printf("    当前电量: %.1f kWh (%.1f%%)\n",
				vehicle.CurrentCapacity,
				vehicle.CurrentCapacity/vehicle.RequestedCapacity*100)
		}
	} else {
		// 显示所有充电桩状态
		piles := m.simulator.GetAllPiles()
		fmt.Printf("充电桩状态 (总数: %d)\n", len(piles))
		fmt.Println("-------------------------------------")

		for _, pile := range piles {
			status, vehicle := pile.GetStatus()

			fmt.Printf("充电桩 %s (%s):\n", pile.ID, pile.Type)
			fmt.Printf("  状态: %s\n", status)

			if vehicle != nil {
				fmt.Printf("  当前: %s (%.1f/%.1f kWh)\n",
					vehicle.UserID, vehicle.CurrentCapacity, vehicle.RequestedCapacity)
			}
			fmt.Println("-------------------------------------")
		}
	}
}

// triggerFault 触发故障
func (m *Manager) triggerFault(args []string) {
	if len(args) < 4 {
		fmt.Println("用法: fault <pileID> <type> <minutes> [description]")
		fmt.Println("类型: hardware, software, power")
		return
	}

	pileID := args[1]
	faultType := args[2]

	minutes, err := strconv.Atoi(args[3])
	if err != nil {
		fmt.Printf("无效的时间格式: %s\n", args[3])
		return
	}

	description := "手动触发故障"
	if len(args) > 4 {
		description = strings.Join(args[4:], " ")
	}

	if err := m.simulator.TriggerFault(pileID, faultType, description, minutes); err != nil {
		fmt.Printf("触发故障失败: %v\n", err)
		return
	}

	fmt.Printf("已触发充电桩 %s 的 %s 故障，持续时间: %d分钟\n", pileID, faultType, minutes)
}

// recoverFault 恢复故障
func (m *Manager) recoverFault(args []string) {
	if len(args) < 2 {
		fmt.Println("用法: recover <pileID>")
		return
	}

	pileID := args[1]

	if err := m.simulator.RecoverFault(pileID); err != nil {
		fmt.Printf("恢复故障失败: %v\n", err)
		return
	}

	fmt.Printf("已恢复充电桩 %s 的故障\n", pileID)
}

// simulateRequest 模拟充电请求
func (m *Manager) simulateRequest(args []string) {
	if len(args) < 4 {
		fmt.Println("用法: sim <userID> <amount> <mode>")
		fmt.Println("模式: fast, trickle")
		return
	}

	userID := args[1]

	amount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		fmt.Printf("无效的电量格式: %s\n", args[2])
		return
	}

	mode := args[3]
	if mode != "fast" && mode != "trickle" {
		fmt.Printf("无效的充电模式: %s, 必须是 'fast' 或 'trickle'\n", mode)
		return
	}

	m.simulator.SimulateChargingRequest(userID, amount, mode)
	fmt.Printf("已模拟用户 %s 的充电请求: %.1f kWh, 模式: %s\n", userID, amount, mode)
}

// reloadConfig 重新加载配置
func (m *Manager) reloadConfig() {
	// 停止当前模拟器
	if err := m.simulator.Stop(); err != nil {
		fmt.Printf("停止模拟器失败: %v\n", err)
		return
	}

	// 重新加载配置
	cfg, err := config.LoadConfig(m.configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	// 更新配置
	m.config = cfg

	// 创建新的模拟器
	m.simulator = NewPileSimulator(cfg, m.logger)

	// 启动新的模拟器
	if err := m.simulator.Start(); err != nil {
		fmt.Printf("启动模拟器失败: %v\n", err)
		return
	}

	fmt.Println("配置已重新加载，模拟器已重启")
}
