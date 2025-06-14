package utils

import (
	"sync"
	"time"
)

// Clock 时钟接口，提供统一的时间获取方式
type Clock interface {
	Now() time.Time
	Since(t time.Time) time.Duration
	Until(t time.Time) time.Duration
	Sleep(d time.Duration)
	NewTicker(d time.Duration) Ticker
	NewTimer(d time.Duration) Timer
	After(d time.Duration) <-chan time.Time
}

// Ticker 计时器接口
type Ticker interface {
	C() <-chan time.Time
	Stop()
}

// Timer 定时器接口
type Timer interface {
	C() <-chan time.Time
	Stop() bool
	Reset(d time.Duration) bool
}

// SystemClock 系统时钟实现
type SystemClock struct{}

// NewSystemClock 创建系统时钟
func NewSystemClock() *SystemClock {
	return &SystemClock{}
}

func (c *SystemClock) Now() time.Time {
	return time.Now().UTC() // 返回UTC时间
}

func (c *SystemClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}

func (c *SystemClock) Until(t time.Time) time.Duration {
	return time.Until(t)
}

func (c *SystemClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

func (c *SystemClock) NewTicker(d time.Duration) Ticker {
	return &systemTicker{time.NewTicker(d)}
}

func (c *SystemClock) NewTimer(d time.Duration) Timer {
	return &systemTimer{time.NewTimer(d)}
}

func (c *SystemClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// systemTicker 系统计时器包装
type systemTicker struct {
	*time.Ticker
}

func (t *systemTicker) C() <-chan time.Time {
	return t.Ticker.C
}

// systemTimer 系统定时器包装
type systemTimer struct {
	*time.Timer
}

func (t *systemTimer) C() <-chan time.Time {
	return t.Timer.C
}

// SimulatedClock 模拟时钟实现
type SimulatedClock struct {
	mu           sync.RWMutex
	currentTime  time.Time
	isRunning    bool
	stopCh       chan struct{}
	tickers      map[*simulatedTicker]bool
	timers       map[*simulatedTimer]bool
	timeChangeCh chan time.Time // 时间变化通知通道
}

// NewSimulatedClock 创建模拟时钟
func NewSimulatedClock(startTime time.Time) *SimulatedClock {
	clock := &SimulatedClock{
		currentTime:  startTime,
		stopCh:       make(chan struct{}),
		tickers:      make(map[*simulatedTicker]bool),
		timers:       make(map[*simulatedTimer]bool),
		timeChangeCh: make(chan time.Time, 10), // 缓冲通道
	}

	return clock
}

// Start 启动模拟时钟
func (c *SimulatedClock) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isRunning {
		return
	}

	c.isRunning = true
	go c.runTimeSimulation()
}

// Stop 停止模拟时钟
func (c *SimulatedClock) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isRunning {
		return
	}

	c.isRunning = false
	close(c.stopCh)

	// 停止所有计时器
	for ticker := range c.tickers {
		ticker.stop()
	}
	for timer := range c.timers {
		timer.stop()
	}
}

// SetTime 设置当前模拟时间
func (c *SimulatedClock) SetTime(t time.Time) {
	c.mu.Lock()
	oldTime := c.currentTime
	c.currentTime = t
	c.mu.Unlock()

	// 通知所有计时器时间已改变
	if !oldTime.Equal(t) {
		select {
		case c.timeChangeCh <- t:
		default:
			// 如果通道已满，则跳过通知
		}
	}
}

func (c *SimulatedClock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentTime
}

func (c *SimulatedClock) Since(t time.Time) time.Duration {
	return c.Now().Sub(t)
}

func (c *SimulatedClock) Until(t time.Time) time.Duration {
	return t.Sub(c.Now())
}

func (c *SimulatedClock) Sleep(d time.Duration) {
	timer := c.NewTimer(d)
	<-timer.C()
}

func (c *SimulatedClock) NewTicker(d time.Duration) Ticker {
	ticker := &simulatedTicker{
		clock:    c,
		interval: d,
		ch:       make(chan time.Time, 1),
		stopCh:   make(chan struct{}),
		stopped:  false,
	}

	c.mu.Lock()
	c.tickers[ticker] = true
	c.mu.Unlock()

	go ticker.run()
	return ticker
}

func (c *SimulatedClock) NewTimer(d time.Duration) Timer {
	timer := &simulatedTimer{
		clock:   c,
		ch:      make(chan time.Time, 1),
		stopCh:  make(chan struct{}),
		stopped: false,
	}

	timer.targetTime = c.Now().Add(d)

	c.mu.Lock()
	c.timers[timer] = true
	c.mu.Unlock()

	go timer.run()
	return timer
}

func (c *SimulatedClock) After(d time.Duration) <-chan time.Time {
	return c.NewTimer(d).C()
}

// runTimeSimulation 运行时间模拟
// 在模拟时钟模式下，时间不会自动流逝，只能通过SetTime手动设置
func (c *SimulatedClock) runTimeSimulation() {
	// 只等待停止信号，不自动推进时间
	<-c.stopCh
}

// simulatedTicker 模拟计时器
type simulatedTicker struct {
	clock    *SimulatedClock
	interval time.Duration
	ch       chan time.Time
	stopCh   chan struct{}
	stopped  bool
	mu       sync.Mutex
}

func (t *simulatedTicker) C() <-chan time.Time {
	return t.ch
}

func (t *simulatedTicker) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.stopped {
		return
	}

	t.stopped = true
	t.stop()

	// 从时钟中移除
	t.clock.mu.Lock()
	delete(t.clock.tickers, t)
	t.clock.mu.Unlock()
}

func (t *simulatedTicker) stop() {
	close(t.stopCh)
}

func (t *simulatedTicker) run() {
	nextTick := t.clock.Now().Add(t.interval)

	for {
		select {
		case <-t.stopCh:
			return
		case currentTime := <-t.clock.timeChangeCh:
			// 当时间变化时，检查是否应该触发
			for currentTime.After(nextTick) || currentTime.Equal(nextTick) {
				select {
				case t.ch <- currentTime:
				default:
				}
				nextTick = nextTick.Add(t.interval)
			}
		}
	}
}

// simulatedTimer 模拟定时器
type simulatedTimer struct {
	clock      *SimulatedClock
	targetTime time.Time
	ch         chan time.Time
	stopCh     chan struct{}
	stopped    bool
	mu         sync.Mutex
}

func (t *simulatedTimer) C() <-chan time.Time {
	return t.ch
}

func (t *simulatedTimer) Stop() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.stopped {
		return false
	}

	t.stopped = true
	t.stop()

	// 从时钟中移除
	t.clock.mu.Lock()
	delete(t.clock.timers, t)
	t.clock.mu.Unlock()

	return true
}

func (t *simulatedTimer) Reset(d time.Duration) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	active := !t.stopped
	t.targetTime = t.clock.Now().Add(d)

	if t.stopped {
		t.stopped = false
		t.stopCh = make(chan struct{})

		t.clock.mu.Lock()
		t.clock.timers[t] = true
		t.clock.mu.Unlock()

		go t.run()
	}

	return active
}

func (t *simulatedTimer) stop() {
	close(t.stopCh)
}

func (t *simulatedTimer) run() {
	for {
		select {
		case <-t.stopCh:
			return
		case currentTime := <-t.clock.timeChangeCh:
			// 当时间变化时，检查是否应该触发
			if currentTime.After(t.targetTime) || currentTime.Equal(t.targetTime) {
				select {
				case t.ch <- currentTime:
				default:
				}

				// 定时器只触发一次，然后停止
				t.mu.Lock()
				t.stopped = true
				t.clock.mu.Lock()
				delete(t.clock.timers, t)
				t.clock.mu.Unlock()
				t.mu.Unlock()
				return
			}
		}
	}
}
