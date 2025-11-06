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

// NewTiDBRuleExecutor 创建并初始化 TiDB 规则执行器
func NewTiDBRuleExecutor(ruleFile, ruleName, ruleVersion string) (*TiDBRuleExecutor, error) {
	// 1. 创建知识库
	knowledgeLibrary := ast.NewKnowledgeLibrary()
	ruleBuilder := builder.NewRuleBuilder(knowledgeLibrary)

	// 2. 加载规则文件
	err := ruleBuilder.BuildRuleFromResource(ruleName, ruleVersion, pkg.NewFileResource(ruleFile))
	if err != nil {
		return nil, fmt.Errorf("加载规则文件失败: %v", err)
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
