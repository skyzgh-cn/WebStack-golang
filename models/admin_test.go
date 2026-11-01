package models

import (
	"testing"
)

func TestHashPasswordIsBcrypt(t *testing.T) {
	hashed, err := HashPassword("admin")
	if err != nil {
		t.Fatalf("HashPassword 失败: %v", err)
	}
	if hashed == "admin" {
		t.Fatal("密码不应被明文存储")
	}
	if len(hashed) < 50 {
		t.Fatalf("bcrypt 哈希长度异常: %d", len(hashed))
	}
}

func TestVerifyPasswordBcrypt(t *testing.T) {
	hashed, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword 失败: %v", err)
	}

	ok, upgrade := verifyPassword(hashed, "correct-password")
	if !ok {
		t.Fatal("正确密码应校验通过")
	}
	if upgrade {
		t.Fatal("bcrypt 密码不应触发升级")
	}

	ok, _ = verifyPassword(hashed, "wrong-password")
	if ok {
		t.Fatal("错误密码应校验失败")
	}
}

func TestVerifyPasswordLegacyMD5(t *testing.T) {
	// 21232f297a57a5a743894a0e4a801fc3 是 md5("admin")
	legacy := "21232f297a57a5a743894a0e4a801fc3"

	ok, upgrade := verifyPassword(legacy, "admin")
	if !ok {
		t.Fatal("历史 MD5 密码应能通过校验")
	}
	if !upgrade {
		t.Fatal("历史 MD5 密码应标记为需要升级")
	}

	ok, _ = verifyPassword(legacy, "not-admin")
	if ok {
		t.Fatal("错误密码不应通过 MD5 校验")
	}
}

func TestVerifyPasswordEmpty(t *testing.T) {
	ok, _ := verifyPassword("", "anything")
	if ok {
		t.Fatal("空哈希不应通过校验")
	}
}

// TestAdminLoginUsesBcrypt 验证登录流程：默认管理员使用 bcrypt，
// 且历史 MD5 密码登录后会被自动升级为 bcrypt。
func TestAdminLoginUsesBcrypt(t *testing.T) {
	setupTestDB(t)

	var admin Admin
	if err := admin.Login("admin", "admin"); err != nil {
		t.Fatalf("默认管理员应能登录: %v", err)
	}

	var stored Admin
	DB.Where("username = ?", "admin").First(&stored)
	if stored.Password == "admin" {
		t.Fatal("数据库中不应存储明文密码")
	}
	if len(stored.Password) < 50 {
		t.Fatalf("存储的密码应为 bcrypt 哈希，实际长度 %d", len(stored.Password))
	}

	// 错误密码应被拒绝
	var bad Admin
	if err := bad.Login("admin", "wrong-password"); err == nil {
		t.Fatal("错误密码应登录失败")
	}
}

// TestLegacyMD5PasswordAutoUpgrade 验证历史 MD5 密码可登录并自动升级。
func TestLegacyMD5PasswordAutoUpgrade(t *testing.T) {
	setupTestDB(t)

	legacy := "21232f297a57a5a743894a0e4a801fc3"
	DB.Where("username = ?", "admin").Delete(&Admin{})
	if err := DB.Create(&Admin{Username: "legacy", Password: legacy}).Error; err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	var admin Admin
	if err := admin.Login("legacy", "admin"); err != nil {
		t.Fatalf("历史 MD5 密码应能登录: %v", err)
	}

	var stored Admin
	DB.Where("username = ?", "legacy").First(&stored)
	if stored.Password == legacy {
		t.Fatal("登录后 MD5 密码应被升级为 bcrypt")
	}
	if len(stored.Password) < 50 {
		t.Fatal("升级后的密码应为 bcrypt 哈希")
	}
}
