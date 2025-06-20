#!/usr/bin/env python3
"""充电桩管理系统测试脚本
根据测试用例文件自动执行测试流程
"""

import subprocess
import sys
import time
from typing import List, Optional, Tuple

import requests


class TestClient:
    """测试客户端"""

    def __init__(
        self,
        base_url: str = "http://localhost:8080",
        simulator_path: str = "./simulator.exe",
    ):
        self.base_url = base_url
        self.simulator_path = simulator_path
        self.session = requests.Session()
        self.tokens = {}  # 存储用户token
        self.simulator_process = None  # 模拟器进程

    def __enter__(self):
        """进入上下文管理器"""
        if self.start_simulator():
            return self
        raise RuntimeError("无法启动模拟器")

    def __exit__(self, exc_type, exc_val, exc_tb):
        """退出上下文管理器"""
        self.stop_simulator()

    def start_simulator(self) -> bool:
        """启动模拟器进程"""
        try:
            print("正在启动模拟器...")
            self.simulator_process = subprocess.Popen(
                [self.simulator_path],
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
                bufsize=1,  # 行缓冲
            )

            # 等待一下确保进程启动
            time.sleep(2)

            if self.simulator_process.poll() is None:
                print("✓ 模拟器启动成功")
                return True
            print("✗ 模拟器启动失败")
            return False

        except Exception as e:
            print(f"✗ 模拟器启动异常: {e}")
            return False

    def stop_simulator(self):
        """停止模拟器进程"""
        if self.simulator_process:
            try:
                print("正在停止模拟器...")
                self.simulator_process.stdin.close()
                self.simulator_process.terminate()
                self.simulator_process.wait(timeout=5)
                print("✓ 模拟器已停止")
            except subprocess.TimeoutExpired:
                print("模拟器未响应，强制终止...")
                self.simulator_process.kill()
                self.simulator_process.wait()
                print("✓ 模拟器已强制终止")
            except Exception as e:
                print(f"停止模拟器时出错: {e}")
            finally:
                self.simulator_process = None

    def send_simulator_command(self, command: str) -> bool:
        """向模拟器发送命令"""
        if not self.simulator_process or self.simulator_process.poll() is not None:
            print("✗ 模拟器进程未运行")
            return False

        try:
            self.simulator_process.stdin.write(command + "\n")
            self.simulator_process.stdin.flush()

            # 等待一下让命令处理完成
            time.sleep(1)

            return True

        except Exception as e:
            print(f"✗ 发送模拟器命令失败: {command}, 错误: {e}")
            return False

    def run_simulator_command(self, command: str) -> bool:
        """执行模拟器命令(旧版本兼容，实际调用send_simulator_command)"""
        return self.send_simulator_command(command)

    def set_clock(self, time_str: str) -> bool:
        """设置模拟器时钟"""
        command = f"clock set {time_str}"
        return self.run_simulator_command(command)

    def login_user(self, username: str, password: str = "3") -> Optional[str]:
        """用户登录"""
        if username in self.tokens:
            return self.tokens[username]

        url = f"{self.base_url}/api/v1/auth/login"
        data = {"Username": username, "Password": password}

        try:
            response = self.session.post(url, json=data)
            if response.status_code == 200:
                result = response.json()
                token = result.get("data").get("token")
                if token:
                    self.tokens[username] = token
                    return token
                print(f"✗ 用户 {username} 登录失败: 未获取到token")
                return None
            print(f"✗ 用户 {username} 登录失败: HTTP {response.status_code}")
            return None
        except Exception as e:
            print(f"✗ 用户 {username} 登录异常: {e}")
            return None

    def create_charging_request(
        self, username: str, charging_mode: str, requested_capacity: float
    ) -> bool:
        """创建充电请求"""
        token = self.login_user(username)
        if not token:
            return False

        url = f"{self.base_url}/api/v1/charging/requests"
        data = {"ChargingMode": charging_mode, "RequestedCapacity": requested_capacity}
        headers = {
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json",
        }

        try:
            response = self.session.post(url, json=data, headers=headers)
            if response.status_code in [200, 201]:
                return True
            print(f"✗ 用户 {username} 充电请求创建失败: HTTP {response.status_code}")
            return False
        except Exception as e:
            print(f"✗ 用户 {username} 充电请求创建异常: {e}")
            return False

    def set_pile_fault(self, pile_id: str) -> bool:
        """设置充电桩故障"""
        command = f"fault {pile_id} power desc"
        return self.run_simulator_command(command)

    def recover_pile(self, pile_id: str) -> bool:
        """恢复充电桩"""
        command = f"recover {pile_id}"
        return self.run_simulator_command(command)

    def get_queue_status(self) -> dict:
        """获取排队状态"""
        url = f"{self.base_url}/api/v1/admin/charging-piles/queue-vehicles"

        try:
            response = self.session.get(url)
            if response.status_code == 200:
                data = response.json()
                print("=== 排队状态 ===")
                self._print_queue_status_formatted(data)
                return data
            print(f"✗ 获取排队状态失败: HTTP {response.status_code}")
            if response.text:
                print(f"  响应: {response.text}")
            return {}
        except Exception as e:
            print(f"✗ 获取排队状态异常: {e}")
            return {}

    def _print_queue_status_formatted(self, data: dict):
        """格式化打印排队状态"""
        if not data or "data" not in data or "piles" not in data["data"]:
            print("无排队数据")
            return

        piles = data["data"]["piles"]
        for pile in piles:
            pile_id = pile.get("pileId", "Unknown")
            queue_vehicles = pile.get("queueVehicles", [])

            print(f"{pile_id}：")

            if not queue_vehicles:
                print("(无排队车辆)")
            else:
                for vehicle in queue_vehicles:
                    vehicle_id = vehicle.get("vehicleId", "Unknown")
                    current_charged = vehicle.get("currentChargedCapacity", 0)
                    current_fee = vehicle.get("currentFee", 0)
                    print(f"({vehicle_id},{current_charged},{current_fee})")

            print()  # 空行分隔不同充电桩

    def get_waiting_vehicles(self) -> dict:
        """获取等候区车辆信息"""
        url = f"{self.base_url}/api/v1/queue/waiting-vehicles"

        try:
            response = self.session.get(url)
            if response.status_code == 200:
                data = response.json()
                print("=== 等候区车辆信息 ===")

                # 检查是否有等候区车辆信息
                waiting_vehicles = data.get("waitingVehicles", [])
                if not waiting_vehicles:
                    print("(无等候车辆)")
                else:
                    for vehicle in waiting_vehicles:
                        license_plate = vehicle.get("licensePlate", "Unknown")
                        request_type = vehicle.get("requestType", "Unknown")
                        requested_capacity = vehicle.get("requestedCapacity", 0)
                        print(f"({license_plate},{request_type},{requested_capacity})")

                return data
            print(f"✗ 获取等候区车辆信息失败: HTTP {response.status_code}")
            return {}
        except Exception as e:
            print(f"✗ 获取等候区车辆信息异常: {e}")
            return {}


def parse_test_case(file_path: str) -> List[Tuple[str, List[str]]]:
    """解析测试用例文件"""
    test_steps = []

    try:
        with open(file_path, encoding="utf-8") as f:
            lines = [line.strip() for line in f if line.strip()]

        i = 0
        while i < len(lines):
            time_line = lines[i]
            commands = []

            # 获取下一行的命令（如果存在）
            i += 1
            if i < len(lines) and not lines[i].endswith(":00"):
                # 这行包含命令
                command_line = lines[i]
                # 解析命令，支持多个命令在同一行
                import re

                command_pattern = r"\([^)]+\)"
                commands = re.findall(command_pattern, command_line)
                i += 1

            test_steps.append((time_line, commands))

        return test_steps
    except Exception as e:
        print(f"✗ 解析测试用例文件失败: {e}")
        return []


def parse_command(command: str) -> Tuple[str, str, str, str]:
    """解析单个命令"""
    # 移除括号并分割
    command = command.strip("()")
    parts = command.split(",")

    if len(parts) != 4:
        raise ValueError(f"命令格式错误: {command}")

    return parts[0], parts[1], parts[2], parts[3]


def execute_test_case(test_case_file: str):
    """执行测试用例"""
    print(f"开始执行测试用例: {test_case_file}")
    print("=" * 60)

    try:
        with TestClient() as client:
            test_steps = parse_test_case(test_case_file)

            if not test_steps:
                print("✗ 没有找到有效的测试步骤")
                return

            for step_num, (time_str, commands) in enumerate(test_steps, 1):
                print(f"\n【步骤 {step_num}】时间: {time_str}")
                print("-" * 40)

                # 设置时钟
                if not client.set_clock(time_str):
                    print("✗ 设置时钟失败，继续执行...")

                time.sleep(12)  # 等待时钟设置生效

                # 执行命令
                for command in commands:
                    try:
                        cmd_type, param1, param2, param3 = parse_command(command)

                        if cmd_type == "A":
                            # 充电请求命令: (A,V1,T,7)
                            username = param1
                            charging_mode = "slow" if param2 == "T" else "fast"
                            requested_capacity = float(param3)

                            print(
                                f"执行充电请求: 用户={username}, 模式={charging_mode}, 容量={requested_capacity}"
                            )
                            client.create_charging_request(
                                username, charging_mode, requested_capacity
                            )

                        elif cmd_type == "B":
                            # 充电桩控制命令: (B,T2,O,0) 或 (B,T2,O,1)
                            pile_id = param1
                            operation = param3

                            if operation == "0":
                                print(f"执行充电桩故障: {pile_id}")
                                client.set_pile_fault(pile_id)
                            elif operation == "1":
                                print(f"执行充电桩恢复: {pile_id}")
                                client.recover_pile(pile_id)
                            else:
                                print(f"✗ 未知的充电桩操作: {operation}")

                        else:
                            print(f"✗ 未知的命令类型: {cmd_type}")

                    except Exception as e:
                        print(f"✗ 命令执行异常: {command}, 错误: {e}")

                # 等待一下，让系统处理完成
                time.sleep(5)

                # 获取并打印状态信息
                client.get_queue_status()
                client.get_waiting_vehicles()

                print("\n" + "=" * 60)

            print("测试用例执行完成!")

    except RuntimeError as e:
        print(f"✗ 初始化失败: {e}")
    except Exception as e:
        print(f"✗ 测试执行过程中发生异常: {e}")


def main():
    """主函数"""
    if len(sys.argv) != 2:
        print("使用方法: python main.py <测试用例文件>")
        print("示例: python main.py test_case1.txt")
        sys.exit(1)

    test_case_file = sys.argv[1]
    execute_test_case(test_case_file)


if __name__ == "__main__":
    main()
