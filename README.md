# Grule Rule Engine Demo

这是一个使用 Grule-Rule-Engine 的 Go 语言示例项目。

## 项目简介

Grule 是一个基于 Go 语言的规则引擎库，受 JBOSS Drools 的启发，提供了自己的领域特定语言（DSL）来定义业务规则。

## 项目结构

```
.
├── go.mod          # Go 模块定义
├── main.go         # 主程序文件
├── rules.grl       # 规则定义文件
└── README.md      # 项目说明文档
```

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

或者直接运行（会自动下载依赖）：

```bash
go run main.go
```

### 2. 运行示例

```bash
go run main.go
```

## 代码说明

### main.go

主程序文件包含以下内容：

- **TestCar**: 测试车辆结构体，包含速度、最大速度、加速度等属性
- **DistanceRecord**: 距离记录结构体，记录总距离和上次速度
- **规则引擎初始化**: 创建知识库、规则构建器、加载规则文件
- **规则执行**: 创建数据上下文，执行规则引擎

### rules.grl

规则文件定义了三个业务规则：

1. **SpeedUp**: 当车辆加速且速度小于最大速度时，增加速度并更新距离记录
2. **SpeedDown**: 当车辆减速且速度大于0时，减少速度
3. **MaxSpeedReached**: 当车辆达到最大速度时，停止加速

## 规则语法说明

Grule 规则文件使用 `.grl` 扩展名，基本语法如下：

```grl
rule "规则名称" "规则描述" salience 优先级 {
    when
        条件表达式
    then
        执行动作;
}
```

- **rule**: 规则关键字
- **规则名称**: 规则的唯一标识符
- **规则描述**: 规则的描述信息
- **salience**: 规则优先级，数字越大优先级越高
- **when**: 条件部分，定义规则触发的条件
- **then**: 执行部分，定义规则触发后执行的动作

## 扩展示例

### 添加新的规则

在 `rules.grl` 文件中添加新规则：

```grl
rule "NewRule" "新规则描述" salience 10 {
    when
        条件表达式
    then
        执行动作;
}
```

### 添加新的数据结构

在 `main.go` 中定义新的结构体，并在数据上下文中添加：

```go
type NewStruct struct {
    Field1 string
    Field2 int
}

newInstance := &NewStruct{
    Field1: "value",
    Field2: 100,
}

dataContext.Add("NewStruct", newInstance)
```

## 参考资料

- [Grule-Rule-Engine GitHub](https://github.com/hyperjumptech/grule-rule-engine)
- [Grule 文档](https://github.com/hyperjumptech/grule-rule-engine/blob/master/README.md)

## License

MIT

