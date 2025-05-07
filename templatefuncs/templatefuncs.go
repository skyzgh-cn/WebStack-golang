// 创建新文件 templatefuncs/templatefuncs.go
package templatefuncs

import (
	"time"
)

// FormatDate 格式化时间
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

// Max 返回较大值
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min 返回较小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Sub 减法运算
func Sub(a, b int) int {
	return a - b
}

// Seq 生成整数序列
func Seq(start, end int) []int {
	if start > end {
		return []int{}
	}
	result := make([]int, end-start+1)
	for i := range result {
		result[i] = start + i
	}
	return result
}

// Add 加法运算
func Add(a, b int) int {
	return a + b
}
