package types

import (
	"fmt"
	"github.com/go-playground/locales/zh_Hans_CN"
	"github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/translations/zh"
)

// 翻译器实例
var translator ut.Translator

func init() {
	// 初始化翻译器
	zhCN := zh_Hans_CN.New()
	uni := ut.New(zhCN, zhCN)
	trans, _ := uni.GetTranslator("zh_Hans_CN")
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

// TranslateError 翻译验证错误信息
// 使用方法:
//
//	if err := req.Validate(); err != nil {
//	    return TranslateError(err)
//	}
func TranslateError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
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

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
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
