package main

import (
	"fmt"
	"log"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

// CarRuleExecutor 车辆规则执行器
type CarRuleExecutor struct {
	knowledgeLibrary *ast.KnowledgeLibrary
	knowledgeBase    *ast.KnowledgeBase
	ruleEngine       *engine.GruleEngine
	ruleName         string
	ruleVersion      string
}

// NewCarRuleExecutor 创建并初始化车辆规则执行器
func NewCarRuleExecutor(ruleFile, ruleName, ruleVersion string) (*CarRuleExecutor, error) {
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

	return &CarRuleExecutor{
		knowledgeLibrary: knowledgeLibrary,
		knowledgeBase:    knowledgeBase,
		ruleEngine:       ruleEngine,
		ruleName:         ruleName,
		ruleVersion:      ruleVersion,
	}, nil
}

// Execute 执行规则引擎
func (executor *CarRuleExecutor) Execute(car *TestCar, record *DistanceRecord) error {
	// 创建数据上下文
	dataContext := ast.NewDataContext()
	err := dataContext.Add("TestCar", car)
	if err != nil {
		return fmt.Errorf("添加 TestCar 到数据上下文失败: %v", err)
	}
	err = dataContext.Add("DistanceRecord", record)
	if err != nil {
		return fmt.Errorf("添加 DistanceRecord 到数据上下文失败: %v", err)
	}

	// 执行规则
	err = executor.ruleEngine.Execute(dataContext, executor.knowledgeBase)
	if err != nil {
		return fmt.Errorf("执行规则失败: %v", err)
	}

	return nil
}

// ExecuteWithLog 执行规则引擎并输出日志
func (executor *CarRuleExecutor) ExecuteWithLog(car *TestCar, record *DistanceRecord) error {
	fmt.Println("\n执行规则引擎...")
	err := executor.Execute(car, record)
	if err != nil {
		log.Fatalf("执行规则失败: %v", err)
	}
	return err
}
