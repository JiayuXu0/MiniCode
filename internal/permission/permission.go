package permission

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// Response 权限响应类型
type Response int

const (
	Denied Response = iota
	Granted
	Persistent
)

// Request 权限请求
type Request struct {
	ID          string        // 唯一标识
	ToolName    string        // 工具名称
	Action      string        // 具体操作（命令或文件路径）
	Description string        // 操作描述
	ResponseCh  chan Response // 用于接收响应的 channel
}

// RequestMsg 权限请求消息（发送到 TUI）
type RequestMsg struct {
	Req *Request
}

// Service 权限服务
type Service struct {
	mu         sync.RWMutex
	persistent map[string]bool // 持久授权存储: "toolName:action" -> true
	skipAll    bool            // YOLO 模式

	program *tea.Program // 用于发送消息到 TUI
}

// NewService 创建权限服务
func NewService() *Service {
	return &Service{
		persistent: make(map[string]bool),
	}
}

// SetProgram 设置 tea.Program 引用
func (s *Service) SetProgram(p *tea.Program) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.program = p
}

// SetSkipAll 设置 YOLO 模式（跳过所有权限检查）
func (s *Service) SetSkipAll(skip bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.skipAll = skip
}

// IsSkipAll 检查是否是 YOLO 模式
func (s *Service) IsSkipAll() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.skipAll
}

// Request 请求权限（阻塞直到用户响应）
func (s *Service) Request(toolName, action, description string) bool {
	// 1. 检查 skipAll
	s.mu.RLock()
	if s.skipAll {
		s.mu.RUnlock()
		return true
	}
	s.mu.RUnlock()

	// 2. 检查持久授权
	key := s.permissionKey(toolName, action)
	s.mu.RLock()
	if s.persistent[key] {
		s.mu.RUnlock()
		return true
	}
	program := s.program
	s.mu.RUnlock()

	// 3. 没有 program 引用，无法请求权限，拒绝
	if program == nil {
		return false
	}

	// 4. 创建请求并发送到 TUI
	req := &Request{
		ToolName:    toolName,
		Action:      action,
		Description: description,
		ResponseCh:  make(chan Response, 1),
	}

	// 发送请求到 TUI
	program.Send(RequestMsg{Req: req})

	// 5. 阻塞等待响应
	resp := <-req.ResponseCh

	// 6. 处理响应
	switch resp {
	case Granted:
		return true
	case Persistent:
		s.mu.Lock()
		s.persistent[key] = true
		s.mu.Unlock()
		return true
	default:
		return false
	}
}

// Respond 响应权限请求（由 TUI 调用）
func (s *Service) Respond(req *Request, resp Response) {
	if req != nil && req.ResponseCh != nil {
		req.ResponseCh <- resp
	}
}

// ClearPersistent 清除所有持久授权
func (s *Service) ClearPersistent() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.persistent = make(map[string]bool)
}

// permissionKey 生成权限键
func (s *Service) permissionKey(toolName, action string) string {
	return toolName + ":" + action
}
