package main

import (
	"fmt"
	"log"
)

// TestCar 测试车辆结构体
type TestCar struct {
	Speed          int
	MaxSpeed       int
	SpeedIncrement int
	SpeedUp        bool
	SpeedDown      bool
}

// DistanceRecord 距离记录结构体
type DistanceRecord struct {
	TotalDistance int
	LastSpeed     int
}

func main() {
	// 可以选择运行不同的示例
	// carRuleExecutor()
	tidbRuleExecutor()
}

func carRuleExecutor() {
	fmt.Println("=== Grule Rule Engine 示例 ===")

	// 1. 初始化规则执行器
	ruleFile := "rules.grl"
	ruleExecutor, err := NewCarRuleExecutor(ruleFile, "Tutorial", "0.1.1")
	if err != nil {
		log.Fatalf("初始化规则执行器失败: %v", err)
	}
	fmt.Println("✓ 规则文件加载成功")

	// 2. 创建测试数据
	testCar := &TestCar{
		Speed:          50,
		MaxSpeed:       100,
		SpeedIncrement: 10,
		SpeedUp:        true,
		SpeedDown:      false,
	}
	distanceRecord := &DistanceRecord{
		TotalDistance: 0,
		LastSpeed:     0,
	}

	fmt.Printf("\n初始状态:\n")
	fmt.Printf("  车辆速度: %d km/h\n", testCar.Speed)
	fmt.Printf("  最大速度: %d km/h\n", testCar.MaxSpeed)
	fmt.Printf("  总距离: %d km\n", distanceRecord.TotalDistance)
	fmt.Printf("  加速标志: %v\n", testCar.SpeedUp)

	// 3. 执行规则
	err = ruleExecutor.ExecuteWithLog(testCar, distanceRecord)
	if err != nil {
		log.Fatalf("执行规则失败: %v", err)
	}

	// 4. 输出结果
	fmt.Println("\n执行后状态:")
	fmt.Printf("  车辆速度: %d km/h\n", testCar.Speed)
	fmt.Printf("  总距离: %d km\n", distanceRecord.TotalDistance)
	fmt.Printf("  上次速度: %d km/h\n", distanceRecord.LastSpeed)

	// 5. 演示多次执行
	fmt.Println("\n=== 继续执行规则引擎 ===")
	testCar.SpeedUp = true
	testCar.Speed = 80
	fmt.Printf("当前速度: %d km/h\n", testCar.Speed)

	err = ruleExecutor.ExecuteWithLog(testCar, distanceRecord)
	if err != nil {
		log.Fatalf("执行规则失败: %v", err)
	}

	fmt.Printf("执行后速度: %d km/h\n", testCar.Speed)
	fmt.Printf("总距离: %d km\n", distanceRecord.TotalDistance)
}

func tidbRuleExecutor() {
	fmt.Println("=== TiDB 热点检测规则引擎示例 ===")

	// 1. 初始化规则执行器
	ruleFile := "tidb.grl"
	ruleExecutor, err := NewTiDBRuleExecutor(ruleFile, "TiDBHotspot", "1.0.0")
	if err != nil {
		log.Fatalf("初始化规则执行器失败: %v", err)
	}
	fmt.Println("✓ 规则文件加载成功")

	// 2. 创建测试数据 - 模拟多个 TiKV 节点
	tikvNodes := []*TiKVNode{
		{NodeID: "tikv-1", RaftstoreCPU: 30.5, CoprocessorCPU: 25.3},
		{NodeID: "tikv-2", RaftstoreCPU: 32.1, CoprocessorCPU: 28.7},
		{NodeID: "tikv-3", RaftstoreCPU: 85.2, CoprocessorCPU: 22.1}, // 写热点节点
		{NodeID: "tikv-4", RaftstoreCPU: 29.8, CoprocessorCPU: 26.5},
		{NodeID: "tikv-5", RaftstoreCPU: 31.2, CoprocessorCPU: 90.4}, // 读热点节点
	}

	monitor := &TiDBMonitor{
		CheckWriteHotspot: true,
		CheckReadHotspot:  true,
		TiKVNodes:         tikvNodes,
	}

	// 3. 计算统计信息
	monitor.CalculateStatistics()

	fmt.Printf("\n初始监控数据:\n")
	fmt.Printf("  TiKV 节点数量: %d\n", len(monitor.TiKVNodes))
	for _, node := range monitor.TiKVNodes {
		fmt.Printf("  节点 %s: Raftstore CPU=%.2f%%, Coprocessor CPU=%.2f%%\n",
			node.NodeID, node.RaftstoreCPU, node.CoprocessorCPU)
	}
	fmt.Printf("\n统计信息:\n")
	fmt.Printf("  写热点 - 最大值: %.2f%%, 平均值: %.2f%%, 最高节点: %s\n",
		monitor.MaxRaftstoreCPU, monitor.AvgRaftstoreCPU, monitor.WriteHotspotNode)
	fmt.Printf("  读热点 - 最大值: %.2f%%, 平均值: %.2f%%, 最高节点: %s\n",
		monitor.MaxCoprocessorCPU, monitor.AvgCoprocessorCPU, monitor.ReadHotspotNode)

	// 4. 执行规则
	err = ruleExecutor.ExecuteWithLog(monitor)
	if err != nil {
		log.Fatalf("执行规则失败: %v", err)
	}

	// 5. 输出结果
	fmt.Println("\n检测结果:")
	if monitor.WriteHotspotDetected {
		fmt.Printf("  ✓ 检测到写热点！\n")
		fmt.Printf("    热点节点: %s\n", monitor.WriteHotspotNode)
		fmt.Printf("    Raftstore CPU: %.2f%% (平均值: %.2f%%)\n",
			monitor.MaxRaftstoreCPU, monitor.AvgRaftstoreCPU)
		fmt.Printf("    热点比例: %.2f 倍\n", monitor.WriteHotspotRatio)
	} else {
		fmt.Printf("  ✗ 未检测到写热点\n")
	}

	if monitor.ReadHotspotDetected {
		fmt.Printf("  ✓ 检测到读热点！\n")
		fmt.Printf("    热点节点: %s\n", monitor.ReadHotspotNode)
		fmt.Printf("    Coprocessor CPU: %.2f%% (平均值: %.2f%%)\n",
			monitor.MaxCoprocessorCPU, monitor.AvgCoprocessorCPU)
		fmt.Printf("    热点比例: %.2f 倍\n", monitor.ReadHotspotRatio)
	} else {
		fmt.Printf("  ✗ 未检测到读热点\n")
	}

	// 6. 演示正常情况
	fmt.Println("\n=== 测试正常情况（无热点）===")
	normalNodes := []*TiKVNode{
		{NodeID: "tikv-1", RaftstoreCPU: 30.5, CoprocessorCPU: 25.3},
		{NodeID: "tikv-2", RaftstoreCPU: 32.1, CoprocessorCPU: 28.7},
		{NodeID: "tikv-3", RaftstoreCPU: 31.8, CoprocessorCPU: 27.1},
		{NodeID: "tikv-4", RaftstoreCPU: 29.8, CoprocessorCPU: 26.5},
		{NodeID: "tikv-5", RaftstoreCPU: 31.2, CoprocessorCPU: 28.4},
	}

	normalMonitor := &TiDBMonitor{
		CheckWriteHotspot: true,
		CheckReadHotspot:  true,
		TiKVNodes:         normalNodes,
	}

	normalMonitor.CalculateStatistics()
	fmt.Printf("统计信息: 写热点最大值=%.2f%%, 平均值=%.2f%%; 读热点最大值=%.2f%%, 平均值=%.2f%%\n",
		normalMonitor.MaxRaftstoreCPU, normalMonitor.AvgRaftstoreCPU,
		normalMonitor.MaxCoprocessorCPU, normalMonitor.AvgCoprocessorCPU)

	err = ruleExecutor.ExecuteWithLog(normalMonitor)
	if err != nil {
		log.Fatalf("执行规则失败: %v", err)
	}

	fmt.Println("\n检测结果:")
	if normalMonitor.WriteHotspotDetected {
		fmt.Printf("  ✓ 检测到写热点\n")
	} else {
		fmt.Printf("  ✗ 未检测到写热点（正常）\n")
	}
	if normalMonitor.ReadHotspotDetected {
		fmt.Printf("  ✓ 检测到读热点\n")
	} else {
		fmt.Printf("  ✗ 未检测到读热点（正常）\n")
	}
}
