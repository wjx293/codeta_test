// Package idgen 提供业务编号生成功能。
// 编号格式：{前缀} + yyyyMMdd + 10 位序列号（左侧 0 填充）
//   - 订单编号：TO202608070000000001
//   - 支付单号：PW202608070000000001
//   - 回调编号：CB202608070000000001
//   - 模板编号：TP202608070000000001
//
// 并发安全：使用互斥锁保护计数器，同一秒内并发不重复。
package idgen

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Generator 业务编号生成器。
// 每个生成器实例维护独立的日期与计数器，不同前缀可共用同一实例。
type Generator struct {
	mu      sync.Mutex
	date    string // 当前日期 yyyyMMdd
	counter uint64 // 当日序列号计数器
}

// New 创建生成器实例。
func New() *Generator {
	now := time.Now()
	return &Generator{
		date:    formatDate(now),
		counter: 0,
	}
}

// NewSeeded 创建已同步数据库最大序列号的生成器（重启后避免编号冲突）。
func NewSeeded(counter uint64) *Generator {
	now := time.Now()
	return &Generator{
		date:    formatDate(now),
		counter: counter,
	}
}

// ParseSeqFromNo 从业务编号（如 TO202608100000000003）解析当日序列号。
func ParseSeqFromNo(prefix, no string) uint64 {
	today := formatDate(time.Now())
	head := prefix + today
	if !strings.HasPrefix(no, head) || len(no) != len(head)+10 {
		return 0
	}
	seq, err := strconv.ParseUint(no[len(head):], 10, 64)
	if err != nil {
		return 0
	}
	return seq
}

// GenerateOrderNo 生成订单编号：TO + yyyyMMdd + 10位序列。
func (g *Generator) GenerateOrderNo() string {
	return g.generate("TO")
}

// GeneratePaymentNo 生成支付单编号：PW + yyyyMMdd + 10位序列。
func (g *Generator) GeneratePaymentNo() string {
	return g.generate("PW")
}

// GenerateCallbackNo 生成回调编号：CB + yyyyMMdd + 10位序列。
func (g *Generator) GenerateCallbackNo() string {
	return g.generate("CB")
}

// GenerateTemplateNo 生成模板编号：TP + yyyyMMdd + 10位序列。
func (g *Generator) GenerateTemplateNo() string {
	return g.generate("TP")
}

// generate 内部方法：前缀 + 日期 + 10 位序列。
func (g *Generator) generate(prefix string) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	today := formatDate(now)

	// 日期变更时重置计数器
	if g.date != today {
		g.date = today
		g.counter = 0
	}

	g.counter++
	seq := g.counter

	return fmt.Sprintf("%s%s%010d", prefix, today, seq)
}

// formatDate 返回 yyyyMMdd 格式的日期字符串。
func formatDate(t time.Time) string {
	return t.Format("20060102")
}
