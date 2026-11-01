package models

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Admin struct {
	Id       int    `form:"id" json:"id"`
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"-"`
}

// bcryptCost 为密码哈希计算强度。
const bcryptCost = bcrypt.DefaultCost

// md5Pattern 用于识别历史遗留的 MD5 密码（32 位十六进制字符串）。
var md5Pattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

// HashPassword 使用 bcrypt 生成密码哈希。
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// verifyPassword 校验明文密码与存储的哈希是否匹配。
//
// 为兼容旧版本数据，MD5 格式的历史密码仍可登录，但登录成功后会自动升级为 bcrypt。
func verifyPassword(stored, plaintext string) (ok bool, needUpgrade bool) {
	if stored == "" {
		return false, false
	}

	if md5Pattern.MatchString(stored) {
		sum := md5.Sum([]byte(plaintext))
		legacy := hex.EncodeToString(sum[:])
		if subtle.ConstantTimeCompare([]byte(legacy), []byte(stored)) == 1 {
			return true, true
		}
		return false, false
	}

	if err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(plaintext)); err != nil {
		return false, false
	}
	return true, false
}

// Login 管理员登录
func (admin *Admin) Login(username, password string) error {
	var user Admin
	if err := DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户名不存在")
		}
		return err
	}

	ok, needUpgrade := verifyPassword(user.Password, password)
	if !ok {
		return errors.New("密码错误")
	}

	// 历史 MD5 密码登录成功后自动升级为 bcrypt
	if needUpgrade {
		if hashed, err := HashPassword(password); err == nil {
			if err := DB.Model(&Admin{}).Where("id = ?", user.Id).Update("password", hashed).Error; err != nil {
				// 升级失败不影响本次登录
				_ = err
			}
		}
	}

	*admin = user
	return nil
}
