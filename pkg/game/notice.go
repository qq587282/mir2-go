package game

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

type NoticeMsg struct {
	Msg   string
	List  []string
	Valid bool
}

type NoticeManager struct {
	NoticeList [100]*NoticeMsg
	Mutex      sync.RWMutex
	NoticeDir  string
}

var DefaultNoticeManager *NoticeManager

func init() {
	DefaultNoticeManager = NewNoticeManager()
}

func NewNoticeManager() *NoticeManager {
	nm := &NoticeManager{
		NoticeDir: "./notice/",
	}
	for i := range nm.NoticeList {
		nm.NoticeList[i] = &NoticeMsg{
			Valid: true,
		}
	}
	return nm
}

func (nm *NoticeManager) SetNoticeDir(dir string) {
	nm.NoticeDir = dir
}

func (nm *NoticeManager) LoadNotices() {
	nm.Mutex.Lock()
	defer nm.Mutex.Unlock()

	os.MkdirAll(nm.NoticeDir, 0755)

	for _, notice := range nm.NoticeList {
		if notice == nil || notice.Msg == "" {
			continue
		}
		filename := filepath.Join(nm.NoticeDir, notice.Msg+".txt")
		data, err := os.ReadFile(filename)
		if err != nil {
			continue
		}
		lines := parseLines(string(data))
		notice.List = lines
	}
}

func (nm *NoticeManager) GetNotice(msg string) []string {
	nm.Mutex.RLock()
	defer nm.Mutex.RUnlock()

	for _, notice := range nm.NoticeList {
		if notice == nil || notice.Msg != msg {
			continue
		}
		if len(notice.List) > 0 {
			return notice.List
		}
	}

	for i, notice := range nm.NoticeList {
		if notice == nil || notice.Msg != "" {
			continue
		}
		filename := filepath.Join(nm.NoticeDir, msg+".txt")
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil
		}
		lines := parseLines(string(data))
		nm.NoticeList[i].Msg = msg
		nm.NoticeList[i].List = lines
		return lines
	}

	return nil
}

func (nm *NoticeManager) SetNotice(index int, msg string) bool {
	if index < 0 || index >= 100 {
		return false
	}
	nm.Mutex.Lock()
	defer nm.Mutex.Unlock()
	nm.NoticeList[index].Msg = msg
	return true
}

func parseLines(s string) []string {
	var lines []string
	var current []byte
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, string(current))
			current = nil
		} else if c != '\r' {
			current = append(current, byte(c))
		}
	}
	if len(current) > 0 {
		lines = append(lines, string(current))
	}
	return lines
}

func GetNoticeManager() *NoticeManager {
	return DefaultNoticeManager
}

type LineNotice struct {
	Notices    []string
	Index      int
	Interval   time.Duration
	LastUpdate time.Time
	Running    bool
	mu         sync.RWMutex
}

var GlobalLineNotice *LineNotice

func init() {
	GlobalLineNotice = &LineNotice{
		Notices:  make([]string, 0),
		Interval: 30 * time.Second,
	}
}

func (ln *LineNotice) AddNotice(notice string) {
	ln.mu.Lock()
	defer ln.mu.Unlock()
	ln.Notices = append(ln.Notices, notice)
}

func (ln *LineNotice) ClearNotices() {
	ln.mu.Lock()
	defer ln.mu.Unlock()
	ln.Notices = nil
	ln.Index = 0
}

func (ln *LineNotice) GetCurrentNotice() string {
	ln.mu.RLock()
	defer ln.mu.RUnlock()
	if len(ln.Notices) == 0 {
		return ""
	}
	return ln.Notices[ln.Index]
}

func (ln *LineNotice) NextNotice() string {
	ln.mu.Lock()
	defer ln.mu.Unlock()
	if len(ln.Notices) == 0 {
		return ""
	}
	notice := ln.Notices[ln.Index]
	ln.Index = (ln.Index + 1) % len(ln.Notices)
	return notice
}

func (ln *LineNotice) SetInterval(interval time.Duration) {
	ln.mu.Lock()
	defer ln.mu.Unlock()
	ln.Interval = interval
}

func GetLineNotice() *LineNotice {
	return GlobalLineNotice
}
