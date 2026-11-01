package models

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// setupTestDB 在测试中初始化一个独立的临时数据库，
// 避免测试用例污染真实的 db/webstack.db。
func setupTestDB(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	t.Setenv("WEBSTACK_SQLITE_FILE", filepath.Join(dir, "test.db"))
	t.Setenv("WEBSTACK_DB_TYPE", "sqlite")

	// LoadConfig 使用 sync.Once 单例，测试中需要重置以应用新的环境变量
	configInstance = nil
	loadErr = nil
	loadOnce = sync.Once{}

	if err := InitDB(); err != nil {
		t.Fatalf("初始化测试数据库失败: %v", err)
	}

	// 关闭数据库连接，使临时目录可以被清理
	t.Cleanup(func() {
		if DB != nil {
			if sqlDB, err := DB.DB(); err == nil {
				sqlDB.Close()
			}
		}
	})
}

// TestMain 确保测试进程退出时清理临时目录。
func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
