package models

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Admin struct {
	Id       string `form:"id" json:"id"`
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

// Login 管理员登录
func (admin *Admin) Login(username, password string) error {
	// 对密码进行MD5加密
	h := md5.New()
	h.Write([]byte(password))
	encryptedPassword := hex.EncodeToString(h.Sum(nil))

	fmt.Printf("加密后的密码: %s\n", encryptedPassword)

	// 查询数据库
	result := DB.Where("username = ? AND password = ?", username, encryptedPassword).First(admin)

	// 如果出错，先尝试只通过用户名查询，看看是否存在该用户
	if result.Error != nil {
		var tempAdmin Admin
		userResult := DB.Where("username = ?", username).First(&tempAdmin)
		if userResult.Error != nil {
			if errors.Is(userResult.Error, gorm.ErrRecordNotFound) {
				fmt.Printf("用户名不存在: %s\n", username)
				return errors.New("用户名不存在")
			}
			return userResult.Error
		}
		fmt.Printf("密码不匹配: 期望=%s, 实际=%s\n", tempAdmin.Password, encryptedPassword)
		return errors.New("密码错误")
	}

	return nil
}
