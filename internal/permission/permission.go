package permission

import (
	"context"
	"sync"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JiayuXu0/MiniCode/internal/pubsub"
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

// RequestInfo is a lightweight permission request payload for events.
type RequestInfo struct {
	ToolName    string
	Action      string
	Description string
}

// PermissionEvent is the payload for permission events.
type PermissionEvent struct {
	Request  RequestInfo
	Response Response
}

// Service 权限服务
type Service struct {
	mu         sync.RWMutex
	persistent map[string]bool // 持久授权存储: "toolName:action" -> true
	skipAll    bool            // YOLO 模式

	program *tea.Program // 用于发送消息到 TUI

	broker *pubsub.Broker[PermissionEvent]
}

// NewService 创建权限服务
func NewService() *Service {
	return &Service{
		persistent: make(map[string]bool),
		broker:     pubsub.NewBroker[PermissionEvent](),
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

// Subscribe 订阅权限事件
func (s *Service) Subscribe(ctx context.Context) <-chan pubsub.Event[PermissionEvent] {
	return s.broker.Subscribe(ctx)
}

// Shutdown 关闭权限事件 Broker
func (s *Service) Shutdown() {
	if s.broker != nil {
		s.broker.Shutdown()
	}
}

// Request 请求权限（阻塞直到用户响应）
func (s *Service) Request(toolName, action, description string) bool {
	reqInfo := RequestInfo{
		ToolName:    toolName,
		Action:      action,
		Description: description,
	}

	// 1. 检查 skipAll
	s.mu.RLock()
	if s.skipAll {
		s.mu.RUnlock()
		s.publishEvent(reqInfo, Granted)
		return true
	}
	s.mu.RUnlock()

	// 2. 检查持久授权
	key := s.permissionKey(toolName, action)
	s.mu.RLock()
	if s.persistent[key] {
		s.mu.RUnlock()
		s.publishEvent(reqInfo, Persistent)
		return true
	}
	program := s.program
	s.mu.RUnlock()

	// 3. 没有 program 引用，无法请求权限，拒绝
	if program == nil {
		s.publishEvent(reqInfo, Denied)
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
		s.publishEvent(reqInfo, Granted)
		return true
	case Persistent:
		s.mu.Lock()
		s.persistent[key] = true
		s.mu.Unlock()
		s.publishEvent(reqInfo, Persistent)
		return true
	default:
		s.publishEvent(reqInfo, Denied)
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

func (s *Service) publishEvent(req RequestInfo, resp Response) {
	if s.broker == nil {
		return
	}

	eventType := pubsub.CreatedEvent
	if resp == Denied {
		eventType = pubsub.DeletedEvent
	}

	s.broker.Publish(eventType, PermissionEvent{
		Request:  req,
		Response: resp,
	})
}
