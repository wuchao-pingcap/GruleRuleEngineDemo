package main

import (
	"fmt"
	"log"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
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
	fmt.Println("=== Grule Rule Engine 示例 ===\n")

	// 1. 创建知识库
	knowledgeLibrary := ast.NewKnowledgeLibrary()
	ruleBuilder := builder.NewRuleBuilder(knowledgeLibrary)

	// 2. 加载规则文件
	ruleFile := "rules.grl"
	err := ruleBuilder.BuildRuleFromResource("Tutorial", "0.1.1", pkg.NewFileResource(ruleFile))
	if err != nil {
		log.Fatalf("加载规则文件失败: %v", err)
	}
	fmt.Println("✓ 规则文件加载成功")

	// 3. 创建测试数据
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

	// 4. 创建数据上下文
	dataContext := ast.NewDataContext()
	err = dataContext.Add("TestCar", testCar)
	if err != nil {
		log.Fatalf("添加 TestCar 到数据上下文失败: %v", err)
	}
	err = dataContext.Add("DistanceRecord", distanceRecord)
	if err != nil {
		log.Fatalf("添加 DistanceRecord 到数据上下文失败: %v", err)
	}

	// 5. 获取知识库实例
	knowledgeBase := knowledgeLibrary.NewKnowledgeBaseInstance("Tutorial", "0.1.1")

	// 6. 创建规则引擎
	ruleEngine := engine.NewGruleEngine()

	// 7. 执行规则
	fmt.Println("\n执行规则引擎...")
	err = ruleEngine.Execute(dataContext, knowledgeBase)
	if err != nil {
		log.Fatalf("执行规则失败: %v", err)
	}

	// 8. 输出结果
	fmt.Println("\n执行后状态:")
	fmt.Printf("  车辆速度: %d km/h\n", testCar.Speed)
	fmt.Printf("  总距离: %d km\n", distanceRecord.TotalDistance)
	fmt.Printf("  上次速度: %d km/h\n", distanceRecord.LastSpeed)

	// 9. 演示多次执行
	fmt.Println("\n=== 继续执行规则引擎 ===")
	testCar.SpeedUp = true
	testCar.Speed = 80
	fmt.Printf("当前速度: %d km/h\n", testCar.Speed)

	err = ruleEngine.Execute(dataContext, knowledgeBase)
	if err != nil {
		log.Fatalf("执行规则失败: %v", err)
	}

	fmt.Printf("执行后速度: %d km/h\n", testCar.Speed)
	fmt.Printf("总距离: %d km\n", distanceRecord.TotalDistance)
}
