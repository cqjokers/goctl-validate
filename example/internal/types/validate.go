package types

import (
	"github.com/go-playground/validator/v10"
)

// 共享的validator实例
var validate = validator.New()

// Validate 验证AdminLoginReq结构体
func (r *AdminLoginReq) Validate() error {
	return validate.Struct(r)
}

// Validate 验证AssignRoleReq结构体
func (r *AssignRoleReq) Validate() error {
	return validate.Struct(r)
}

// Validate 验证UserRegisterReq结构体
func (r *UserRegisterReq) Validate() error {
	return validate.Struct(r)
}

// Validate 验证UserLoginReq结构体
func (r *UserLoginReq) Validate() error {
	return validate.Struct(r)
}

// Validate 验证UserUpdateReq结构体
func (r *UserUpdateReq) Validate() error {
	return validate.Struct(r)
}

// Validate 验证UserQueryReq结构体
func (r *UserQueryReq) Validate() error {
	return validate.Struct(r)
}

// Validate 验证PasswordChangeReq结构体
func (r *PasswordChangeReq) Validate() error {
	return validate.Struct(r)
}
