package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/zeromicro/go-zero/tools/goctl/plugin"
)

// ValidateGenerator 简化的验证代码生成器
type ValidateGenerator struct {
	plugin  *plugin.Plugin
	options *Options
}

// NewValidateGenerator 创建简化的验证代码生成器
func NewValidateGenerator(p *plugin.Plugin, opts *Options) *ValidateGenerator {
	if opts == nil {
		opts = &Options{
			EnableTranslator: false,
		}
	}
	return &ValidateGenerator{
		plugin:  p,
		options: opts,
	}
}

// Generate 生成验证代码
func (g *ValidateGenerator) Generate() error {
	// 解析API文件获取带有validate标签的结构体
	validateStructs, err := g.parseAPIFileForValidateTags()
	if err != nil {
		return fmt.Errorf("failed to parse API file: %v", err)
	}

	if len(validateStructs) == 0 {
		fmt.Println("goctl-validate: no structures with validate tags found")
		return nil
	}

	// 查找types目录
	typesDir := filepath.Join(g.plugin.Dir, "internal", "types")
	if !dirExists(typesDir) {
		return fmt.Errorf("types directory not found: %s", typesDir)
	}

	// 生成验证文件
	validateFile := filepath.Join(typesDir, "validate.go")
	if err := g.generateValidateFile(validateFile, validateStructs); err != nil {
		return fmt.Errorf("failed to generate validate file: %v", err)
	}

	fmt.Printf("goctl-validate: generated validation code for %d structures in %s\n",
		len(validateStructs), validateFile)

	// 如果启用翻译器，生成翻译器文件
	if g.options.EnableTranslator {
		translatorFile := filepath.Join(typesDir, "translator.go")
		if err := g.generateTranslatorFile(translatorFile); err != nil {
			return fmt.Errorf("failed to generate translator file: %v", err)
		}

		fmt.Printf("goctl-validate: generated translator code in %s\n", translatorFile)

		// 生成自定义翻译模板文件（如果不存在）
		customTranslatorFile := filepath.Join(typesDir, "translator_custom.go")
		if !fileExists(customTranslatorFile) {
			if err := g.generateCustomTranslatorTemplate(customTranslatorFile); err != nil {
				return fmt.Errorf("failed to generate custom translator template: %v", err)
			}
			fmt.Printf("goctl-validate: generated custom translator template in %s\n", customTranslatorFile)
		} else {
			fmt.Printf("goctl-validate: custom translator file already exists, skipped: %s\n", customTranslatorFile)
		}
	}

	return nil
}

// generateValidateFile 生成验证文件
func (g *ValidateGenerator) generateValidateFile(filename string, validateStructs []ValidateStruct) error {
	// 准备模板数据
	data := struct {
		Package          string
		EnableTranslator bool
		Structs          []ValidateStruct
	}{
		Package:          "types",
		EnableTranslator: g.options.EnableTranslator,
		Structs:          validateStructs,
	}

	// 生成代码
	content, err := g.renderTemplate(data)
	if err != nil {
		return fmt.Errorf("failed to render template: %v", err)
	}

	// 写入文件
	return os.WriteFile(filename, []byte(content), 0644)
}

// renderTemplate 渲染验证代码模板
func (g *ValidateGenerator) renderTemplate(data interface{}) (string, error) {
	tmpl := `package {{.Package}}

import (
	"github.com/go-playground/validator/v10"
)

// 共享的validator实例
var validate = validator.New()

{{- range .Structs}}
// Validate 验证{{.Name}}结构体
func (r *{{.Name}}) Validate() error {
	return validate.Struct(r)
}
{{- end}}
`

	t, err := template.New("validate").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateTranslatorFile 生成翻译器文件
func (g *ValidateGenerator) generateTranslatorFile(filename string) error {
	content, err := g.renderTranslatorTemplate()
	if err != nil {
		return fmt.Errorf("failed to render translator template: %v", err)
	}

	return os.WriteFile(filename, []byte(content), 0644)
}

// renderTranslatorTemplate 渲染翻译器模板
func (g *ValidateGenerator) renderTranslatorTemplate() (string, error) {
	tmpl := `package types

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

// registerCustomTranslations 注册自定义翻译规则
// 此方法为预留方法，用于注册自定义的验证规则翻译
// 自定义翻译应该在 translator_custom.go 文件中实现
func registerCustomTranslations() {
	// 检查是否存在自定义翻译注册函数
	if customRegister := getCustomTranslationRegister(); customRegister != nil {
		customRegister(validate, translator)
	}
}

// getCustomTranslationRegister 获取自定义翻译注册函数
// 这是一个弱引用，如果 translator_custom.go 文件存在，则会被重写
var getCustomTranslationRegister = func() func(*validator.Validate, ut.Translator) {
	return nil
}

// Translate 翻译验证错误信息
// 使用方法:
//   if err := req.Validate(); err != nil {
//       return Translate(err)
//   }
func Translate(err error) error {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		var translatedErrors []string
		for _, fieldError := range validationErrors {
			translatedMsg := fieldError.Translate(translator)
			translatedErrors = append(translatedErrors, translatedMsg)
		}

		// 返回第一个翻译后的错误信息
		if len(translatedErrors) > 0 {
			return fmt.Errorf(translatedErrors[0])
		}
	}

	// 如果不是验证错误，返回原始错误
	return err
}

// TranslateErrors 翻译所有验证错误信息
// 返回所有翻译后的错误信息列表
func TranslateErrors(err error) []string {
	var translatedErrors []string
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, fieldError := range validationErrors {
			translatedMsg := fieldError.Translate(translator)
			translatedErrors = append(translatedErrors, translatedMsg)
		}
	} else {
		// 如果不是验证错误，返回原始错误信息
		translatedErrors = append(translatedErrors, err.Error())
	}

	return translatedErrors
}
`

	t, err := template.New("translator").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse translator template: %v", err)
	}

	var buf strings.Builder
	if err := t.Execute(&buf, nil); err != nil {
		return "", fmt.Errorf("failed to execute translator template: %v", err)
	}

	return buf.String(), nil
}

// generateCustomTranslatorTemplate 生成自定义翻译器模板文件
func (g *ValidateGenerator) generateCustomTranslatorTemplate(filename string) error {
	content := g.renderCustomTranslatorTemplate()
	return os.WriteFile(filename, []byte(content), 0644)
}

// renderCustomTranslatorTemplate 渲染自定义翻译器模板
func (g *ValidateGenerator) renderCustomTranslatorTemplate() string {
	return `package types

import (
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/universal-translator"
)

// 重写默认的自定义翻译注册函数
func init() {
	getCustomTranslationRegister = func() func(*validator.Validate, ut.Translator) {
		return registerCustomTranslationsImpl
	}
}

// registerCustomTranslationsImpl 注册自定义翻译规则的实现
// 在这里添加您的自定义验证规则翻译
// 此文件不会被 goctl-validate 重新生成覆盖
func registerCustomTranslationsImpl(validate *validator.Validate, translator ut.Translator) {
	// 示例：注册自定义验证规则翻译
	// validate.RegisterTranslation("custom_rule", translator, func(ut ut.Translator) error {
	//     return ut.Add("custom_rule", "{0}不符合自定义规则", true)
	// }, func(ut ut.Translator, fe validator.FieldError) string {
	//     t, _ := ut.T("custom_rule", fe.Field())
	//     return t
	// })

	// 您可以在这里添加更多自定义翻译规则...
}
`
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// parseAPIFileForValidateTags 解析API文件获取validate标签
func (g *ValidateGenerator) parseAPIFileForValidateTags() ([]ValidateStruct, error) {
	return parseAPIFileForValidateStructs(g.plugin.ApiFilePath)
}
