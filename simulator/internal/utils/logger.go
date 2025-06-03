package utils

import (
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

var (
	// 随机数生成器
	rnd *rand.Rand

	// 初始化函数
	once sync.Once

	// 日志级别
	logLevel string = "info" // 默认日志级别
)

func init() {
	once.Do(func() {
		// 初始化日志
		log.New(os.Stdout, "[模拟器] ", log.LstdFlags)

		// 初始化随机数生成器
		src := rand.NewSource(time.Now().UTC().UnixNano())
		rnd = rand.New(src)
	})
}

// SetLogLevel 设置日志级别
func SetLogLevel(level string) {
	logLevel = level
}

// GetLogLevel 获取日志级别
func GetLogLevel() string {
	return logLevel
}

// ShouldLogDetail 是否应该记录详细日志
func ShouldLogDetail() bool {
	return logLevel == "debug" || logLevel == "trace"
}

// RandomInt 生成指定范围内的随机整数 [min, max]
func RandomInt(min, max int) int {
	if min >= max {
		return min
	}
	return min + rnd.Intn(max-min+1)
}

// RandomFloat 生成指定范围内的随机浮点数 [min, max)
func RandomFloat(min, max float64) float64 {
	if min >= max {
		return min
	}
	return min + rnd.Float64()*(max-min)
}

// RandomElement 从切片中随机选择一个元素
func RandomElement[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	return slice[rnd.Intn(len(slice))]
}

// FormatDuration 格式化时间间隔为易读的形式
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return time.Duration(h).String() + " " + time.Duration(m).String() + " " + time.Duration(s).String()
	} else if m > 0 {
		return time.Duration(m).String() + " " + time.Duration(s).String()
	}
	return time.Duration(s).String()
}
