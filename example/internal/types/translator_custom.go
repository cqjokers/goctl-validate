package types

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
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
	// 自定义翻译：覆盖默认的alphanum翻译
	validate.RegisterTranslation("alphanum", translator, func(ut ut.Translator) error {
		return ut.Add("alphanum", "{0}只能包含字母和数字，不允许特殊字符", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("alphanum", fe.Field())
		return t
	})

	// 自定义翻译：覆盖默认的oneof翻译，使其更友好
	validate.RegisterTranslation("oneof", translator, func(ut ut.Translator) error {
		return ut.Add("oneof", "{0}的值无效，请选择正确的选项", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("oneof", fe.Field())
		return t
	})

	// 您可以在这里添加更多自定义翻译规则...
}
