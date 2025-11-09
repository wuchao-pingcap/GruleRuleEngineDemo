package main

import (
	"fmt"
	"log"
	"testing"
)

// 1. 演示写热点检测
func TestDetectReadHotspot(t *testing.T) {

	// 1. 初始化规则执行器（单个规则文件）
	ruleFile := "tidb.grl"
	ruleExecutor, err := NewTiDBRuleExecutor(ruleFile, "TiDBHotspot", "1.0.0")
	if err != nil {
		t.Fatalf("初始化规则执行器失败: %v", err)
	}
	fmt.Printf("✓ 规则文件加载成功")

	// 示例：使用多个规则文件
	// ruleFiles := []string{"tidb.grl", "tidb-advanced.grl"}
	// ruleExecutor, err := NewTiDBRuleExecutorWithFiles(ruleFiles, "TiDBHotspot", "1.0.0")
	// if err != nil {
	// 	log.Fatalf("初始化规则执行器失败: %v", err)
	// }
	// fmt.Printf("✓ 成功加载 %d 个规则文件\n", len(ruleFiles))

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

		// 检查是否是非聚簇索引热点
		if monitor.IsNonClusteredIndexHotspot && monitor.RecommendShardRowIDBits {
			fmt.Printf("  ⚠ 非聚簇索引写入热点！\n")
			fmt.Printf("    建议设置 SHARD_ROW_ID_BITS=%d 来打散 RowID，缓解写入热点问题\n", monitor.ShardRowIDBits)
			fmt.Printf("    SQL 示例: ALTER TABLE table_name SHARD_ROW_ID_BITS = %d;\n", monitor.ShardRowIDBits)
		}
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
}

// 6. 演示非聚簇索引写入热点
func TestDetectNonClusteredIndexWriteHotspot(t *testing.T) {

	ruleFile := "tidb.grl"
	ruleExecutor, err := NewTiDBRuleExecutor(ruleFile, "TiDBHotspot", "1.0.0")
	if err != nil {
		t.Fatalf("初始化规则执行器失败: %v", err)
	}
	fmt.Printf("✓ 规则文件加载成功")

	// 6. 演示非聚簇索引写入热点
	fmt.Println("\n=== 测试非聚簇索引写入热点 ===")
	nonClusteredNodes := []*TiKVNode{
		{NodeID: "tikv-1", RaftstoreCPU: 25.3, CoprocessorCPU: 22.1},
		{NodeID: "tikv-2", RaftstoreCPU: 28.7, CoprocessorCPU: 24.5},
		{NodeID: "tikv-3", RaftstoreCPU: 95.8, CoprocessorCPU: 23.2}, // 非聚簇索引写热点
		{NodeID: "tikv-4", RaftstoreCPU: 26.2, CoprocessorCPU: 25.1},
		{NodeID: "tikv-5", RaftstoreCPU: 27.5, CoprocessorCPU: 24.8},
	}

	nonClusteredMonitor := &TiDBMonitor{
		CheckWriteHotspot:          true,
		CheckReadHotspot:           true,
		TiKVNodes:                  nonClusteredNodes,
		IsNonClusteredIndexHotspot: true, // 标识为非聚簇索引热点
	}

	nonClusteredMonitor.CalculateStatistics()
	fmt.Printf("统计信息: 写热点最大值=%.2f%%, 平均值=%.2f%%; 读热点最大值=%.2f%%, 平均值=%.2f%%\n",
		nonClusteredMonitor.MaxRaftstoreCPU, nonClusteredMonitor.AvgRaftstoreCPU,
		nonClusteredMonitor.MaxCoprocessorCPU, nonClusteredMonitor.AvgCoprocessorCPU)

	err = ruleExecutor.ExecuteWithLog(nonClusteredMonitor)
	if err != nil {
		log.Fatalf("执行规则失败: %v", err)
	}

	fmt.Println("\n检测结果:")
	if nonClusteredMonitor.WriteHotspotDetected {
		fmt.Printf("  ✓ 检测到写热点\n")
		if nonClusteredMonitor.RecommendShardRowIDBits {
			fmt.Printf("  ⚠ 非聚簇索引写入热点！\n")
			fmt.Printf("    建议设置 SHARD_ROW_ID_BITS=%d 来打散 RowID\n", nonClusteredMonitor.ShardRowIDBits)
			fmt.Printf("    SQL 示例: ALTER TABLE table_name SHARD_ROW_ID_BITS = %d;\n", nonClusteredMonitor.ShardRowIDBits)
		}
	} else {
		fmt.Printf("  ✗ 未检测到写热点\n")
	}
}

// 7. 演示规则不匹配的情况
func TestWriteHotspotNotMatch(t *testing.T) {
	// 1. 初始化规则执行器（单个规则文件）
	ruleFile := "tidb.grl"
	ruleExecutor, err := NewTiDBRuleExecutor(ruleFile, "TiDBHotspot", "1.0.0")
	if err != nil {
		t.Fatalf("初始化规则执行器失败: %v", err)
	}
	fmt.Printf("✓ 规则文件加载成功")

	// 7. 演示规则不匹配的情况
	fmt.Println("\n=== 测试规则不匹配的情况 ===")

	// 场景1：写热点检测规则不匹配 - CPU 差异不够大
	fmt.Println("\n场景1：写热点检测规则不匹配（CPU差异不够大）")
	lowDiffNodes := []*TiKVNode{
		{NodeID: "tikv-1", RaftstoreCPU: 30.0, CoprocessorCPU: 25.0},
		{NodeID: "tikv-2", RaftstoreCPU: 32.0, CoprocessorCPU: 28.0},
		{NodeID: "tikv-3", RaftstoreCPU: 45.0, CoprocessorCPU: 22.0}, // 最高，但只比平均值高1.4倍（不够1.5倍）
		{NodeID: "tikv-4", RaftstoreCPU: 29.0, CoprocessorCPU: 26.0},
		{NodeID: "tikv-5", RaftstoreCPU: 31.0, CoprocessorCPU: 28.0},
	}

	lowDiffMonitor := &TiDBMonitor{
		CheckWriteHotspot:          true,
		CheckReadHotspot:           true,
		TiKVNodes:                  lowDiffNodes,
		IsNonClusteredIndexHotspot: true,
	}

	lowDiffMonitor.CalculateStatistics()
	fmt.Printf("统计信息: 写热点最大值=%.2f%%, 平均值=%.2f%%, 比例=%.2f倍\n",
		lowDiffMonitor.MaxRaftstoreCPU, lowDiffMonitor.AvgRaftstoreCPU,
		lowDiffMonitor.MaxRaftstoreCPU/lowDiffMonitor.AvgRaftstoreCPU)
	fmt.Printf("说明: 热点比例 %.2f 倍 < 1.5 倍，DetectWriteHotspot 规则不会匹配\n",
		lowDiffMonitor.MaxRaftstoreCPU/lowDiffMonitor.AvgRaftstoreCPU)

	err = ruleExecutor.ExecuteWithLog(lowDiffMonitor)
	if err != nil {
		log.Fatalf("执行规则失败: %v", err)
	}

	fmt.Println("\n检测结果:")
	if lowDiffMonitor.WriteHotspotDetected {
		fmt.Printf("  ✓ 检测到写热点\n")
	} else {
		fmt.Printf("  ✗ 未检测到写热点（规则未匹配：CPU差异不够大）\n")
	}
	if lowDiffMonitor.RecommendShardRowIDBits {
		fmt.Printf("  ⚠ 建议设置 SHARD_ROW_ID_BITS\n")
	} else {
		fmt.Printf("  ✗ 未建议设置 SHARD_ROW_ID_BITS（因为未检测到写热点）\n")
	}
}

// 6. 演示非聚簇索引建议规则不匹配 - 检测到写热点但不是非聚簇索引热点
func TestNonClusteredIndexWriteHotspotNotMatch(t *testing.T) {
	// 1. 初始化规则执行器（单个规则文件）
	ruleFile := "tidb.grl"
	ruleExecutor, err := NewTiDBRuleExecutor(ruleFile, "TiDBHotspot", "1.0.0")
	if err != nil {
		t.Fatalf("初始化规则执行器失败: %v", err)
	}
	fmt.Printf("✓ 规则文件加载成功")

	// 场景2：非聚簇索引建议规则不匹配 - 检测到写热点但不是非聚簇索引热点
	fmt.Println("\n场景2：非聚簇索引建议规则不匹配（不是非聚簇索引热点）")
	normalHotspotNodes := []*TiKVNode{
		{NodeID: "tikv-1", RaftstoreCPU: 30.0, CoprocessorCPU: 25.0},
		{NodeID: "tikv-2", RaftstoreCPU: 32.0, CoprocessorCPU: 28.0},
		{NodeID: "tikv-3", RaftstoreCPU: 85.0, CoprocessorCPU: 22.0}, // 写热点，但不是非聚簇索引导致的
		{NodeID: "tikv-4", RaftstoreCPU: 29.0, CoprocessorCPU: 26.0},
		{NodeID: "tikv-5", RaftstoreCPU: 31.0, CoprocessorCPU: 28.0},
	}

	normalHotspotMonitor := &TiDBMonitor{
		CheckWriteHotspot:          true,
		CheckReadHotspot:           true,
		TiKVNodes:                  normalHotspotNodes,
		IsNonClusteredIndexHotspot: false, // 不是非聚簇索引热点
	}

	normalHotspotMonitor.CalculateStatistics()
	fmt.Printf("统计信息: 写热点最大值=%.2f%%, 平均值=%.2f%%, 比例=%.2f倍\n",
		normalHotspotMonitor.MaxRaftstoreCPU, normalHotspotMonitor.AvgRaftstoreCPU,
		normalHotspotMonitor.MaxRaftstoreCPU/normalHotspotMonitor.AvgRaftstoreCPU)
	fmt.Printf("说明: IsNonClusteredIndexHotspot = false，RecommendShardRowIDBits* 规则不会匹配\n")

	err = ruleExecutor.ExecuteWithLog(normalHotspotMonitor)
	if err != nil {
		log.Fatalf("执行规则失败: %v", err)
	}

	fmt.Println("\n检测结果:")
	if normalHotspotMonitor.WriteHotspotDetected {
		fmt.Printf("  ✓ 检测到写热点\n")
	} else {
		fmt.Printf("  ✗ 未检测到写热点\n")
	}
	if normalHotspotMonitor.RecommendShardRowIDBits {
		fmt.Printf("  ⚠ 建议设置 SHARD_ROW_ID_BITS=%d\n", normalHotspotMonitor.ShardRowIDBits)
	} else {
		fmt.Printf("  ✗ 未建议设置 SHARD_ROW_ID_BITS（规则未匹配：不是非聚簇索引热点）\n")
	}
}

// 7. 演示非聚簇索引建议规则不匹配 - 热点比例不在建议范围内
func TestNonClusteredIndexWriteHotspotNotMatchRatio(t *testing.T) {
	// 1. 初始化规则执行器（单个规则文件）
	ruleFile := "tidb.grl"
	ruleExecutor, err := NewTiDBRuleExecutor(ruleFile, "TiDBHotspot", "1.0.0")
	if err != nil {
		t.Fatalf("初始化规则执行器失败: %v", err)
	}
	fmt.Printf("✓ 规则文件加载成功")

	// 场景3：非聚簇索引建议规则不匹配 - 热点比例不在建议范围内
	fmt.Println("\n场景3：非聚簇索引建议规则不匹配（热点比例不在建议范围内）")
	edgeCaseNodes := []*TiKVNode{
		{NodeID: "tikv-1", RaftstoreCPU: 30.0, CoprocessorCPU: 25.0},
		{NodeID: "tikv-2", RaftstoreCPU: 32.0, CoprocessorCPU: 28.0},
		{NodeID: "tikv-3", RaftstoreCPU: 50.0, CoprocessorCPU: 22.0}, // 热点比例约1.4倍（低于1.5倍阈值）
		{NodeID: "tikv-4", RaftstoreCPU: 29.0, CoprocessorCPU: 26.0},
		{NodeID: "tikv-5", RaftstoreCPU: 31.0, CoprocessorCPU: 28.0},
	}

	edgeCaseMonitor := &TiDBMonitor{
		CheckWriteHotspot:          true,
		CheckReadHotspot:           true,
		TiKVNodes:                  edgeCaseNodes,
		IsNonClusteredIndexHotspot: true,
	}

	edgeCaseMonitor.CalculateStatistics()
	ratio := edgeCaseMonitor.MaxRaftstoreCPU / edgeCaseMonitor.AvgRaftstoreCPU
	fmt.Printf("统计信息: 写热点最大值=%.2f%%, 平均值=%.2f%%, 比例=%.2f倍\n",
		edgeCaseMonitor.MaxRaftstoreCPU, edgeCaseMonitor.AvgRaftstoreCPU, ratio)
	fmt.Printf("说明: 热点比例 %.2f 倍 < 1.5 倍，DetectWriteHotspot 规则不会匹配\n", ratio)
	fmt.Printf("      因此 RecommendShardRowIDBits* 规则也不会匹配\n")

	err = ruleExecutor.ExecuteWithLog(edgeCaseMonitor)
	if err != nil {
		log.Fatalf("执行规则失败: %v", err)
	}

	fmt.Println("\n检测结果:")
	if edgeCaseMonitor.WriteHotspotDetected {
		fmt.Printf("  ✓ 检测到写热点\n")
		if edgeCaseMonitor.RecommendShardRowIDBits {
			fmt.Printf("  ⚠ 建议设置 SHARD_ROW_ID_BITS=%d\n", edgeCaseMonitor.ShardRowIDBits)
		} else {
			fmt.Printf("  ✗ 未建议设置 SHARD_ROW_ID_BITS\n")
		}
	} else {
		fmt.Printf("  ✗ 未检测到写热点（规则未匹配：热点比例 %.2f 倍 < 1.5 倍）\n", ratio)
		fmt.Printf("  ✗ 未建议设置 SHARD_ROW_ID_BITS（因为未检测到写热点）\n")
	}
}

// 8. 演示正常情况
func TestNormalCase(t *testing.T) {
	// 1. 初始化规则执行器（单个规则文件）
	ruleFile := "tidb.grl"
	ruleExecutor, err := NewTiDBRuleExecutor(ruleFile, "TiDBHotspot", "1.0.0")
	if err != nil {
		t.Fatalf("初始化规则执行器失败: %v", err)
	}
	fmt.Printf("✓ 规则文件加载成功")

	// 8. 演示正常情况
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
