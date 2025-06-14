package simulator

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"simulator/internal/config"
	"simulator/internal/utils"
)

// Manager 模拟器管理器
type Manager struct {
	simulator    *PileSimulator
	config       *config.Config
	logger       *utils.Logger
	clockManager *utils.ClockManager
	isRunning    bool
	stopCh       chan struct{}
	wg           sync.WaitGroup
	configPath   string
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

	// 创建时钟管理器
	clockManager := utils.NewClockManager()

	// 根据配置决定是否启用模拟时钟
	if cfg.Simulation.UseSimClock {
		var startTime time.Time
		if cfg.Simulation.SimClockStart != "" {
			startTime, err = time.Parse(time.RFC3339, cfg.Simulation.SimClockStart)
			if err != nil {
				logger.Warning("无效的模拟时钟起始时间格式，使用默认时间: %v", err)
				startTime = time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
			}
		} else {
			startTime = time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
		}

		if err := clockManager.EnableSimulatedClock(startTime); err != nil {
			logger.Warning("启用模拟时钟失败，使用系统时钟: %v", err)
		} else {
			logger.Info("已启用模拟时钟: 起始时间=%s",
				startTime.Format(time.RFC3339))
		}
	}
	// 创建模拟器
	simulator := NewPileSimulator(cfg, logger, clockManager)

	return &Manager{
		simulator:    simulator,
		config:       cfg,
		logger:       logger,
		clockManager: clockManager,
		stopCh:       make(chan struct{}),
		configPath:   configPath,
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

	// 停止时钟管理器
	m.clockManager.Stop()

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
		case "clock":
			m.clockCommand(args)
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
	fmt.Println("  fault <pileID> <type> <description>")
	fmt.Println("                          - 触发充电桩故障")
	fmt.Println("                            type: hardware/software/power")
	fmt.Println("  recover <pileID>        - 手动恢复故障")
	fmt.Println("  clock <subcommand>      - 时钟管理命令")
	fmt.Println("    clock status          - 显示当前时钟状态")
	fmt.Println("    clock set <time>      - 设置模拟时间 (本地时间将转换为UTC)")
	fmt.Println("  help                    - 显示帮助信息")
	fmt.Println("  exit                    - 退出程序")
	fmt.Println("====================")
	fmt.Println("注意: 时间输入支持本地时间，系统会自动转换为UTC时间存储")
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

	description := "手动触发故障"
	if len(args) > 4 {
		description = strings.Join(args[4:], " ")
	}

	if err := m.simulator.TriggerFault(pileID, faultType, description); err != nil {
		fmt.Printf("触发故障失败: %v\n", err)
		return
	}

	fmt.Printf("已触发充电桩 %s 的 %s 故障\n", pileID, faultType)
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

// SetBackendURL 设置后端API URL
func (m *Manager) SetBackendURL(url string) error {
	m.config.BackendAPI.BaseURL = url
	m.logger.Info("后端API URL已更新为: %s", url)
	return nil
}

// EnableSimulatedClock 启用模拟时钟
func (m *Manager) EnableSimulatedClock(startTime time.Time) error {
	if err := m.clockManager.EnableSimulatedClock(startTime); err != nil {
		return fmt.Errorf("启用模拟时钟失败: %w", err)
	}

	// 更新配置
	m.config.Simulation.UseSimClock = true
	m.config.Simulation.SimClockStart = startTime.Format(time.RFC3339)

	m.logger.Info("已启用模拟时钟: 起始时间=%s",
		startTime.Format(time.RFC3339))
	return nil
}

// DisableSimulatedClock 禁用模拟时钟
func (m *Manager) DisableSimulatedClock() {
	m.clockManager.DisableSimulatedClock()
	m.config.Simulation.UseSimClock = false
	m.logger.Info("已禁用模拟时钟，切换为系统时钟")
}

// SetSimulatedTime 设置模拟时钟时间
func (m *Manager) SetSimulatedTime(t time.Time) error {
	if err := m.clockManager.SetSimulatedTime(t); err != nil {
		return fmt.Errorf("设置模拟时间失败: %w", err)
	}

	m.logger.Info("模拟时间已设置为: %s", t.Format(time.RFC3339))
	return nil
}

// GetClockInfo 获取时钟信息
func (m *Manager) GetClockInfo() (bool, time.Time) {
	isSimClock := m.clockManager.IsUsingSimulatedClock()

	if isSimClock {
		currentTime, _ := m.clockManager.GetSimulatedTime()
		return true, currentTime
	}

	return false, time.Now()
}

// clockCommand 处理时钟相关命令
func (m *Manager) clockCommand(args []string) {
	if len(args) < 2 {
		fmt.Println("用法: clock <subcommand>")
		fmt.Println("子命令:")
		fmt.Println("  status          - 显示当前时钟状态")
		fmt.Println("  set <time>      - 设置模拟时间")
		return
	}

	subCmd := args[1]
	switch subCmd {
	case "status":
		m.showClockStatus()
	case "set":
		m.setTimeCommand(args)
	default:
		fmt.Printf("未知的时钟子命令: %s\n", subCmd)
	}
}

// showClockStatus 显示时钟状态
func (m *Manager) showClockStatus() {
	isSimClock, currentTime := m.GetClockInfo()

	if isSimClock {
		fmt.Println("时钟状态: 模拟时钟")
		fmt.Printf("当前时间: %s\n", currentTime.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("时钟状态: 系统时钟")
		fmt.Printf("当前时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	}
}

// setTimeCommand 设置时间命令
func (m *Manager) setTimeCommand(args []string) {
	if !m.clockManager.IsUsingSimulatedClock() {
		fmt.Println("当前未使用模拟时钟，无法设置时间")
		return
	}

	if len(args) < 3 {
		fmt.Println("用法: clock set <time>")
		fmt.Println("时间格式: 2006-01-02T15:04:05Z 或 2006-01-02 15:04:05 (本地时间)")
		return
	}

	timeStr := args[2]
	var newTime time.Time
	var err error

	// 尝试不同的时间格式
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"15:04:05",
	}

	for _, format := range formats {
		if format == "15:04:05" {
			// 只有时间，使用当前模拟日期
			currentTime, _ := m.clockManager.GetSimulatedTime()
			today := currentTime.Format("2006-01-02")
			timeStr = today + " " + args[2]
			format = "2006-01-02 15:04:05"
		}

		if format == time.RFC3339 {
			// RFC3339 格式已包含时区信息，直接解析
			newTime, err = time.Parse(format, timeStr)
		} else {
			// 其他格式按本地时间解析，然后转换为UTC
			newTime, err = time.ParseInLocation(format, timeStr, time.Local)
			if err == nil {
				// 转换为UTC时间
				newTime = newTime.UTC()
			}
		}

		if err == nil {
			break
		}
	}

	if err != nil {
		fmt.Printf("无效的时间格式: %s\n", args[2])
		fmt.Println("支持的格式:")
		fmt.Println("  2024-01-01T08:00:00Z (UTC时间)")
		fmt.Println("  2024-01-01T08:00:00 (本地时间，将转换为UTC)")
		fmt.Println("  2024-01-01 08:00:00 (本地时间，将转换为UTC)")
		fmt.Println("  08:00:00 (本地时间，使用当前模拟日期，将转换为UTC)")
		return
	}

	if err := m.SetSimulatedTime(newTime); err != nil {
		fmt.Printf("设置时间失败: %v\n", err)
		return
	}
	fmt.Printf("模拟时间已设置为: %s (UTC)\n", newTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("本地时间: %s\n", newTime.In(time.Local).Format("2006-01-02 15:04:05"))
}
