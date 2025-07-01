# goctl-validate

一个专为 go-zero 框架设计的验证插件，自动生成符合 go-zero 规范的验证代码。

## ✨ 特性

- 🚀 **符合 go-zero 规范** - 生成 `func (r *Req) Validate() error` 方法
- 🔧 **零侵入设计** - 不修改现有文件，生成独立的验证文件
- 🌐 **完整的 import 支持** - 支持单行和块级 import 语法
- 🌍 **国际化支持** - 基于官方翻译库的中文错误信息
- 🎨 **自定义翻译** - 支持自定义翻译规则且不会被覆盖
- ⚡ **高性能** - 共享 validator 实例优化性能
- 📁 **模块化设计** - 翻译功能独立分离，代码结构清晰

## 🏗️ 架构设计

### 核心原则

1. **简单优于复杂** - 使用模板生成替代复杂的 AST 操作
2. **职责分离** - goctl 处理类型生成，插件专注验证逻辑
3. **标准化** - 完全符合 go-zero 开发习惯和规范

### 工作流程

```
API文件 (带validate标签) → goctl生成types → 插件生成验证代码
```

## 📖 使用指南

### 1. 在 API 文件中添加 validate 标签

支持多种 API 文件组织方式：

#### 单文件方式

```go
// user.api
syntax = "v1"

type (
    UserRegisterReq {
        Username string `json:"username" validate:"required,min=3,max=20,alphanum"`
        Password string `json:"password" validate:"required,min=6,max=20"`
        Email    string `json:"email" validate:"required,email"`
        Phone    string `json:"phone" validate:"required,len=11,numeric"`
    }
)

service user-api {
    @handler UserRegister
    post /api/user/register (UserRegisterReq) returns (CommonResp)
}
```

#### 多文件方式（推荐）

```go
// main.api
syntax = "v1"

import (
    "types/user_types.api"
    "types/common_types.api"
)

service user-api {
    @handler UserRegister
    post /api/user/register (UserRegisterReq) returns (CommonResp)
}
```

```go
// types/user_types.api
syntax = "v1"

type (
    UserRegisterReq {
        Username string `json:"username" validate:"required,min=3,max=20,alphanum"`
        Password string `json:"password" validate:"required,min=6,max=20"`
        Email    string `json:"email" validate:"required,email"`
        Phone    string `json:"phone" validate:"required,len=11,numeric"`
    }
)
```

### 2. 生成验证代码

```bash
# 基本使用（仅生成验证方法）
goctl api plugin -plugin goctl-validate -api user.api -dir .

# 启用翻译器（生成中文错误信息支持）,以下两种方式二选一
GOCTL_VALIDATE_TRANSLATOR=true goctl api plugin -plugin "goctl-validate" -api user.api -dir .
goctl api plugin -plugin "goctl-validate -translator" -api user.api -dir .
```

### 3. 在业务代码中使用

```go
import "your-project/internal/types"

func (l *UserRegisterLogic) UserRegister(req *types.UserRegisterReq) error {
    // 验证请求参数
    if err := req.Validate(); err != nil {
        return err
    }

    // 业务逻辑...
    return nil
}
```

### 4. 使用翻译功能（可选）

```go
func (l *UserRegisterLogic) UserRegister(req *types.UserRegisterReq) error {
    if err := req.Validate(); err != nil {
        // 翻译为中文错误信息
        return types.Translate(err)
    }

    // 业务逻辑...
    return nil
}
```

## 📁 生成的文件结构

启用翻译器后，会生成以下文件：

```
internal/types/
├── validate.go           # 验证方法（会被重新生成）
├── translator.go         # 翻译器主文件（会被重新生成）
├── translator_custom.go  # 自定义翻译（受保护，不会被覆盖）
└── types.go              # goctl生成的类型文件
```

### validate.go - 验证方法文件

```go
package types

import (
    "github.com/go-playground/validator/v10"
)

// 共享的validator实例
var validate = validator.New()

// Validate 验证UserRegisterReq结构体
func (r *UserRegisterReq) Validate() error {
    return validate.Struct(r)
}

// Validate 验证UserLoginReq结构体
func (r *UserLoginReq) Validate() error {
    return validate.Struct(r)
}
```

### translator.go - 翻译器文件（启用翻译器时生成）

```go
package types

import (
    "errors"
    "fmt"
    "github.com/go-playground/validator/v10"
    "github.com/go-playground/validator/v10/translations/zh"
    "github.com/go-playground/universal-translator"
    zhongwen "github.com/go-playground/locales/zh"
)

// 翻译器实例
var translator ut.Translator

func init() {
    // 初始化翻译器
    zw := zhongwen.New()
    uni := ut.New(zw, zw)
    trans, _ := uni.GetTranslator("zh")
    translator = trans

    // 注册官方默认翻译
    zh.RegisterDefaultTranslations(validate, translator)

    // 注册自定义翻译（如果存在）
    registerCustomTranslations()
}

// TranslateError 翻译验证错误信息
func Translate(err error) error {
    // 翻译逻辑...
}

// TranslateErrors 翻译所有验证错误信息
func TranslateErrors(err error) []string {
    // 翻译逻辑...
}
```

### translator_custom.go - 自定义翻译文件（受保护）

```go
package types

import (
    "github.com/go-playground/validator/v10"
    "github.com/go-playground/universal-translator"
)

// registerCustomTranslationsImpl 注册自定义翻译规则的实现
// 在这里添加您的自定义验证规则翻译
// 此文件不会被 goctl-validate 重新生成覆盖
func registerCustomTranslationsImpl(validate *validator.Validate, translator ut.Translator) {
    // 示例：自定义翻译
    validate.RegisterTranslation("alphanum", translator, func(ut ut.Translator) error {
        return ut.Add("alphanum", "{0}只能包含字母和数字，不允许特殊字符", true)
    }, func(ut ut.Translator, fe validator.FieldError) string {
        t, _ := ut.T("alphanum", fe.Field())
        return t
    })
}
```

## 🎯 支持的验证规则

插件支持所有 `github.com/go-playground/validator/v10` 的验证规则：

| 规则       | 说明        | 示例                        |
| ---------- | ----------- | --------------------------- |
| `required` | 必填字段    | `validate:"required"`       |
| `email`    | 邮箱格式    | `validate:"required,email"` |
| `min`      | 最小长度/值 | `validate:"min=3"`          |
| `max`      | 最大长度/值 | `validate:"max=20"`         |
| `len`      | 固定长度    | `validate:"len=11"`         |
| `numeric`  | 数字字符    | `validate:"numeric"`        |
| `alphanum` | 字母数字    | `validate:"alphanum"`       |
| `oneof`    | 枚举值      | `validate:"oneof=0 1 2"`    |
| `url`      | URL 格式    | `validate:"url"`            |

## 🌍 翻译功能

### 官方翻译 vs 自定义翻译

| 验证规则   | 官方翻译                  | 自定义翻译示例                       |
| ---------- | ------------------------- | ------------------------------------ |
| `required` | "为必填字段"              | 可自定义                             |
| `email`    | "必须是一个有效的邮箱"    | 可自定义                             |
| `min`      | "长度必须至少为 3 个字符" | 可自定义                             |
| `alphanum` | "只能包含字母和数字"      | "只能包含字母和数字，不允许特殊字符" |

### 翻译使用示例

```go
// 基本翻译
if err := req.Validate(); err != nil {
    return types.Translate(err)  // 返回第一个错误的翻译
}

// 获取所有错误的翻译
if err := req.Validate(); err != nil {
    errors := types.TranslateErrors(err)  // 返回所有错误的翻译列表
    for _, errMsg := range errors {
        fmt.Println(errMsg)
    }
}
```

## 🚀 构建和安装

```bash
# 克隆项目
git clone <repository-url>
cd goctl-validate

# 构建
go build -o goctl-validate .

# 测试
goctl api plugin -plugin "goctl-validate" -api example/mixed_import.api -dir example
```

## 📂 项目结构

```
goctl-validate/
├── main.go                     # 主程序
├── generator/
│   ├── simple_generator.go     # 核心生成器
│   └── simple_parser.go        # API解析器
├── example/                    # 示例项目
│   ├── mixed_import.api        # 主API文件
│   ├── types/                  # 类型定义文件
│   │   ├── user_types.api      # 用户相关类型
│   │   ├── admin_types.api     # 管理员相关类型
│   │   └── common_types.api    # 通用类型
│   └── internal/types/         # 生成的代码
│       ├── validate.go         # 验证方法
│       ├── translator.go       # 翻译器（可选）
│       └── translator_custom.go # 自定义翻译（可选）
└── README.md                   # 本文档
```

## 💡 最佳实践

1. **API 文件组织**

   - 推荐使用多文件方式组织 API
   - 按功能模块分离类型定义
   - 使用 import 块导入相关类型

2. **验证规则设计**

   - 在 API 文件中直接添加 validate 标签
   - 合理使用验证规则组合
   - 考虑前端验证的一致性

3. **错误处理**

   - 在 handler 或 logic 层入口处验证
   - 使用翻译器提供友好的错误信息
   - 记录详细的验证错误日志

4. **性能优化**
   - 默认使用共享 validator 实例
   - 避免在热路径中重复验证
   - 合理使用 `omitempty` 标签
