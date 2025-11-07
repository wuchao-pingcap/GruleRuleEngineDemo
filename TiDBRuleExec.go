package main

import (
	"fmt"
	"log"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

// TiKVNode TiKV 节点信息
type TiKVNode struct {
	NodeID         string
	RaftstoreCPU   float64
	CoprocessorCPU float64
}

// TiDBMonitor TiDB 监控数据结构
type TiDBMonitor struct {
	// 控制标志
	CheckWriteHotspot bool
	CheckReadHotspot  bool

	// TiKV 节点列表
	TiKVNodes []*TiKVNode

	// 写热点相关统计信息（需要在 Go 代码中计算）
	MaxRaftstoreCPU  float64
	AvgRaftstoreCPU  float64
	WriteHotspotNode string

	// 读热点相关统计信息（需要在 Go 代码中计算）
	MaxCoprocessorCPU float64
	AvgCoprocessorCPU float64
	ReadHotspotNode   string

	// 检测结果
	WriteHotspotDetected bool
	WriteHotspotRatio    float64
	ReadHotspotDetected  bool
	ReadHotspotRatio     float64

	// 非聚簇索引热点相关
	IsNonClusteredIndexHotspot bool // 是否是非聚簇索引导致的写热点
	ShardRowIDBits             int  // 建议的 SHARD_ROW_ID_BITS 值（0-15）
	RecommendShardRowIDBits    bool // 是否建议设置 SHARD_ROW_ID_BITS
}

// CalculateStatistics 计算统计信息（平均值、最大值等）
func (monitor *TiDBMonitor) CalculateStatistics() {
	// 计算写热点统计信息
	if len(monitor.TiKVNodes) > 0 {
		var totalRaftstoreCPU float64
		var totalCoprocessorCPU float64
		var maxRaftstoreCPU float64
		var maxCoprocessorCPU float64
		var writeHotspotNode string
		var readHotspotNode string

		for _, node := range monitor.TiKVNodes {
			// 写热点统计
			totalRaftstoreCPU += node.RaftstoreCPU
			if node.RaftstoreCPU > maxRaftstoreCPU {
				maxRaftstoreCPU = node.RaftstoreCPU
				writeHotspotNode = node.NodeID
			}

			// 读热点统计
			totalCoprocessorCPU += node.CoprocessorCPU
			if node.CoprocessorCPU > maxCoprocessorCPU {
				maxCoprocessorCPU = node.CoprocessorCPU
				readHotspotNode = node.NodeID
			}
		}

		monitor.MaxRaftstoreCPU = maxRaftstoreCPU
		monitor.AvgRaftstoreCPU = totalRaftstoreCPU / float64(len(monitor.TiKVNodes))
		monitor.WriteHotspotNode = writeHotspotNode

		monitor.MaxCoprocessorCPU = maxCoprocessorCPU
		monitor.AvgCoprocessorCPU = totalCoprocessorCPU / float64(len(monitor.TiKVNodes))
		monitor.ReadHotspotNode = readHotspotNode
	}
}

// TiDBRuleExecutor TiDB 规则执行器
type TiDBRuleExecutor struct {
	knowledgeLibrary *ast.KnowledgeLibrary
	knowledgeBase    *ast.KnowledgeBase
	ruleEngine       *engine.GruleEngine
	ruleName         string
	ruleVersion      string
}

// NewTiDBRuleExecutor 创建并初始化 TiDB 规则执行器（支持单个规则文件）
func NewTiDBRuleExecutor(ruleFile, ruleName, ruleVersion string) (*TiDBRuleExecutor, error) {
	return NewTiDBRuleExecutorWithFiles([]string{ruleFile}, ruleName, ruleVersion)
}

// NewTiDBRuleExecutorWithFiles 创建并初始化 TiDB 规则执行器（支持多个规则文件）
// 可以传入多个规则文件，它们会被加载到同一个知识库中
func NewTiDBRuleExecutorWithFiles(ruleFiles []string, ruleName, ruleVersion string) (*TiDBRuleExecutor, error) {
	if len(ruleFiles) == 0 {
		return nil, fmt.Errorf("至少需要提供一个规则文件")
	}

	// 1. 创建知识库
	knowledgeLibrary := ast.NewKnowledgeLibrary()
	ruleBuilder := builder.NewRuleBuilder(knowledgeLibrary)

	// 2. 加载所有规则文件到同一个知识库
	for i, ruleFile := range ruleFiles {
		err := ruleBuilder.BuildRuleFromResource(ruleName, ruleVersion, pkg.NewFileResource(ruleFile))
		if err != nil {
			return nil, fmt.Errorf("加载规则文件 [%d/%d] %s 失败: %v", i+1, len(ruleFiles), ruleFile, err)
		}
	}

	// 3. 获取知识库实例
	knowledgeBase := knowledgeLibrary.NewKnowledgeBaseInstance(ruleName, ruleVersion)

	// 4. 创建规则引擎
	ruleEngine := engine.NewGruleEngine()

	return &TiDBRuleExecutor{
		knowledgeLibrary: knowledgeLibrary,
		knowledgeBase:    knowledgeBase,
		ruleEngine:       ruleEngine,
		ruleName:         ruleName,
		ruleVersion:      ruleVersion,
	}, nil
}

// Execute 执行规则引擎
func (executor *TiDBRuleExecutor) Execute(monitor *TiDBMonitor) error {
	// 创建数据上下文
	dataContext := ast.NewDataContext()
	err := dataContext.Add("TiDBMonitor", monitor)
	if err != nil {
		return fmt.Errorf("添加 TiDBMonitor 到数据上下文失败: %v", err)
	}

	// 执行规则
	err = executor.ruleEngine.Execute(dataContext, executor.knowledgeBase)
	if err != nil {
		return fmt.Errorf("执行规则失败: %v", err)
	}

	return nil
}

// ExecuteWithLog 执行规则引擎并输出日志
func (executor *TiDBRuleExecutor) ExecuteWithLog(monitor *TiDBMonitor) error {
	fmt.Println("\n执行规则引擎...")
	err := executor.Execute(monitor)
	if err != nil {
		log.Fatalf("执行规则失败: %v", err)
	}
	return err
}
