package util

import (
	"fmt"
	"sync"
	"time"
)

type SessionMgr struct {
	mu        sync.Mutex
	sessions  map[uint32]chan interface{}
	sessionID uint32
}

// 创建一个包级别的单例 SessionMgr 实例
var mgr *SessionMgr
var once sync.Once

// 初始化 SessionMgr 实例
func init() {
	once.Do(func() {
		mgr = &SessionMgr{
			sessions: make(map[uint32]chan interface{}),
		}
	})
}

// GetSession 获取一个 session，并返回一个 channel 和一个等待通知的通道
func GetSession() (uint32, chan interface{}) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	mgr.sessionID++

	ch := make(chan interface{}, 2)
	mgr.sessions[mgr.sessionID] = ch

	return mgr.sessionID, ch
}

// CompleteSession 完成 session，通知等待的协程
func CompleteSession(sessionID uint32, data interface{}) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// 获取对应的 channel 和等待通知通道
	ch := mgr.sessions[sessionID]
	if ch == nil {
		return
	}

	ch <- data
	delete(mgr.sessions, sessionID)
}

func WaitForSession(sessionID uint32, timeout time.Duration) (interface{}, error) {
	mgr.mu.Lock()
	ch, exists := mgr.sessions[sessionID]
	mgr.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("session %d does not exist", sessionID)
	}

	if timeout > 0 {
		select {
		case data := <-ch:
			// 数据已收到
			return data, nil
		case <-time.After(timeout):
			// 超时
			return nil, fmt.Errorf("session %d timed out", sessionID)
		}
	} else {
		data := <-ch
		return data, nil
	}
}
