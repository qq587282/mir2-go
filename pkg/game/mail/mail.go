package mail

import (
	"sync"
	"time"
)

type MailType int

const (
	MailTypeSystem MailType = iota
	MailTypePlayer
	MailTypeGuild
	MailTypeGM
	MailTypeAuction
)

type MailItem struct {
	ItemName   string
	Count      int
	Dura       uint16
	Index      uint16
}

type TMail struct {
	MailID       int32
	MailType     MailType
	FromName     string
	ToName       string
	Title        string
	Content      string
	Items        []MailItem
	Gold         int32
	Readed       bool
	Deleted      bool
	KeepTime     time.Duration
	CreateTime   time.Time
	ExpireTime   time.Time
}

type MailManager struct {
	Mails        map[int32]*TMail
	MailsByOwner map[string]map[int32]*TMail
	Mutex        sync.RWMutex
	NextID       int32
}

var DefaultMailManager *MailManager

func init() {
	DefaultMailManager = NewMailManager()
}

func NewMailManager() *MailManager {
	return &MailManager{
		Mails:        make(map[int32]*TMail),
		MailsByOwner: make(map[string]map[int32]*TMail),
		NextID:       1,
	}
}

func (mm *MailManager) GetNextID() int32 {
	id := mm.NextID
	mm.NextID++
	return id
}

func (mm *MailManager) SendMail(mail *TMail) int32 {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()
	
	mail.MailID = mm.GetNextID()
	mail.CreateTime = time.Now()
	mail.ExpireTime = mail.CreateTime.Add(mail.KeepTime)
	
	mm.Mails[mail.MailID] = mail
	
	if _, ok := mm.MailsByOwner[mail.ToName]; !ok {
		mm.MailsByOwner[mail.ToName] = make(map[int32]*TMail)
	}
	mm.MailsByOwner[mail.ToName][mail.MailID] = mail
	
	return mail.MailID
}

func (mm *MailManager) SendSystemMail(toName, title, content string, gold int32, items []MailItem) int32 {
	mail := &TMail{
		MailType:   MailTypeSystem,
		ToName:     toName,
		Title:      title,
		Content:    content,
		Gold:       gold,
		Items:      items,
		KeepTime:   time.Hour * 72,
	}
	return mm.SendMail(mail)
}

func (mm *MailManager) SendPlayerMail(fromName, toName, title, content string, gold int32, items []MailItem) int32 {
	mail := &TMail{
		MailType:   MailTypePlayer,
		FromName:   fromName,
		ToName:     toName,
		Title:      title,
		Content:    content,
		Gold:       gold,
		Items:      items,
		KeepTime:   time.Hour * 168,
	}
	return mm.SendMail(mail)
}

func (mm *MailManager) SendGMMail(toName, title, content string, gold int32, items []MailItem) int32 {
	mail := &TMail{
		MailType:   MailTypeGM,
		FromName:   "系统管理员",
		ToName:     toName,
		Title:      title,
		Content:    content,
		Gold:       gold,
		Items:      items,
		KeepTime:   time.Hour * 720,
	}
	return mm.SendMail(mail)
}

func (mm *MailManager) GetMail(mailID int32) *TMail {
	mm.Mutex.RLock()
	defer mm.Mutex.RUnlock()
	return mm.Mails[mailID]
}

func (mm *MailManager) GetPlayerMails(playerName string) []*TMail {
	mm.Mutex.RLock()
	defer mm.Mutex.RUnlock()
	
	var result []*TMail
	if mailMap, ok := mm.MailsByOwner[playerName]; ok {
		for _, mail := range mailMap {
			if !mail.Deleted {
				result = append(result, mail)
			}
		}
	}
	return result
}

func (mm *MailManager) GetUnreadCount(playerName string) int {
	mm.Mutex.RLock()
	defer mm.Mutex.RUnlock()
	
	count := 0
	if mailMap, ok := mm.MailsByOwner[playerName]; ok {
		for _, mail := range mailMap {
			if !mail.Readed && !mail.Deleted {
				count++
			}
		}
	}
	return count
}

func (mm *MailManager) ReadMail(mailID int32) bool {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()
	
	mail, ok := mm.Mails[mailID]
	if !ok {
		return false
	}
	
	mail.Readed = true
	return true
}

func (mm *MailManager) DeleteMail(mailID int32) bool {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()
	
	mail, ok := mm.Mails[mailID]
	if !ok {
		return false
	}
	
	mail.Deleted = true
	return true
}

func (mm *MailManager) CollectItems(mailID int32) ([]MailItem, int32, bool) {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()
	
	mail, ok := mm.Mails[mailID]
	if !ok || mail.Deleted {
		return nil, 0, false
	}
	
	if len(mail.Items) == 0 && mail.Gold == 0 {
		return nil, 0, false
	}
	
	items := make([]MailItem, len(mail.Items))
	copy(items, mail.Items)
	gold := mail.Gold
	
	mail.Items = nil
	mail.Gold = 0
	
	return items, gold, true
}

func (mm *MailManager) CleanExpiredMails() {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()
	
	now := time.Now()
	for mailID, mail := range mm.Mails {
		if now.After(mail.ExpireTime) {
			mail.Deleted = true
			delete(mm.Mails, mailID)
		}
	}
	
	for name, mailMap := range mm.MailsByOwner {
		for mailID := range mailMap {
			if _, ok := mm.Mails[mailID]; !ok {
				delete(mailMap, mailID)
			}
		}
		if len(mailMap) == 0 {
			delete(mm.MailsByOwner, name)
		}
	}
}

func GetMailManager() *MailManager {
	return DefaultMailManager
}

type MailDB struct {
	db map[int32]*TMail
	mu sync.RWMutex
}

func (mdb *MailDB) Save() {
}

func (mdb *MailDB) Load() {
}