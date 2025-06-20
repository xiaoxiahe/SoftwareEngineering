# 测试脚本使用说明

## 环境准备

1. 安装 Python 依赖

2. 确保后端服务器运行在 `http://localhost:8080`

3. 确保模拟器 `simulator.exe` 在当前目录或指定路径

## 运行测试

```bash
python main.py test_case_1.txt
```

## 测试用例格式

测试用例文件格式如下：

```text
6:00:00
(A,V1,T,7) (A,V2,F,30)

6:30:00
(A,V3,T,28) (A,V4,F,120)

7:00:00
(B,T2,O,0)
```
