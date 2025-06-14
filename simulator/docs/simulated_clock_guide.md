# 模拟时钟功能使用指南

充电桩模拟器现在支持模拟时钟功能，允许您控制模拟器中的时间流逝，这对于测试和演示非常有用。

## 功能特性

1. **系统时钟**: 使用真实的系统时间（默认模式）
2. **模拟时钟**: 使用可控制的模拟时间，支持：
   - 自定义起始时间
   - 可调节的时间加速因子
   - 实时时间调整

## 配置文件设置

在 `configs/simulator.json` 中添加模拟时钟配置：

```json
{
  "simulation": {
    "speedFactor": 1.0,
    "logLevel": "error",
    "useSimClock": true,
    "simClockStart": "2024-01-01T08:00:00Z"
  }
}
```

配置选项说明：

- `useSimClock`: 是否启用模拟时钟（true/false）
- `simClockStart`: 模拟时钟的起始时间（RFC3339 格式）
- `speedFactor`: 时间加速因子（1.0 表示正常速度，2.0 表示 2 倍速）

## 命令行参数

启动模拟器时可以通过命令行参数控制模拟时钟：

```bash
# 使用模拟时钟，默认起始时间
./main.exe -sim-clock

# 设置自定义起始时间
./main.exe -sim-clock -sim-time "2024-06-14T09:00:00Z"

# 设置时间加速因子（10倍速）
./main.exe -sim-clock -sim-time "2024-06-14T09:00:00Z" -speed 10.0

# 组合使用
./main.exe -config configs/simulator.json -backend http://localhost:8080 -sim-clock -sim-time "2024-01-01T08:00:00Z" -speed 5.0
```

参数说明：

- `-sim-clock`: 启用模拟时钟
- `-sim-time`: 设置模拟时钟起始时间（RFC3339 格式）
- `-speed`: 设置时间加速因子

## 运行时命令

在模拟器运行期间，可以通过 CLI 命令控制时钟：

### 查看时钟状态

```
> clock status
```

### 启用模拟时钟

```
# 基本用法
> clock enable 2024-06-14T09:00:00Z

# 设置加速因子
> clock enable 2024-06-14T09:00:00Z 5.0

# 支持的时间格式
> clock enable 2024-06-14T09:00:00Z
> clock enable 2024-06-14T09:00:00
> clock enable "2024-06-14 09:00:00"
> clock enable 09:00:00  # 使用当前日期
```

### 禁用模拟时钟

```
> clock disable
```

### 设置模拟时间

```
> clock set 2024-06-14T12:00:00Z
> clock set "2024-06-14 12:00:00"
> clock set 12:00:00  # 只设置时间，保持当前日期
```

### 调整时间加速因子

```
> clock speed 2.0   # 2倍速
> clock speed 0.5   # 半速
> clock speed 10.0  # 10倍速
```

## 使用场景

### 1. 快速测试

使用高加速因子来快速模拟长时间的充电过程：

```bash
./main.exe -sim-clock -speed 100.0
```

### 2. 特定时间点测试

测试特定时间的系统行为：

```bash
./main.exe -sim-clock -sim-time "2024-06-14T23:59:50Z"
```

### 3. 慢速调试

使用低加速因子进行详细调试：

```bash
./main.exe -sim-clock -speed 0.1
```

### 4. 跨天测试

模拟跨越多天的场景：

```bash
./main.exe -sim-clock -sim-time "2024-06-13T23:00:00Z" -speed 24.0
```

## 时间格式支持

模拟器支持以下时间格式：

1. **RFC3339 格式**: `2024-06-14T09:00:00Z`
2. **本地时间格式**: `2024-06-14T09:00:00`
3. **日期时间格式**: `2024-06-14 09:00:00`
4. **仅时间格式**: `09:00:00` (使用当前或模拟日期)

## 注意事项

1. **性能影响**: 高加速因子可能会增加 CPU 使用率
2. **精度限制**: 模拟时钟的精度约为 10 毫秒
3. **系统兼容**: 切换时钟模式时，建议重启模拟器以确保状态一致
4. **时间同步**: 模拟时钟与系统时钟独立运行，不会影响系统时间

## 示例使用流程

```bash
# 1. 启动模拟器（使用模拟时钟）
./main.exe -sim-clock -sim-time "2024-06-14T08:00:00Z" -speed 5.0

# 2. 查看当前时钟状态
> clock status

# 3. 模拟充电请求
> sim user001 50.0 fast

# 4. 快进时间到中午
> clock set 12:00:00

# 5. 调整加速因子进行详细观察
> clock speed 1.0

# 6. 查看充电桩状态
> status

# 7. 切换回系统时钟
> clock disable
```

这样您就可以完全控制模拟器中的时间流逝，进行各种时间相关的测试和演示。
