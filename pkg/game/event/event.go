package event

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/mir2go/mir2/pkg/protocol"
)

type EventType int

const (
	ET_NONE EventType = iota
	ET_TRAP
	ET_DIGOUTZOMBI
	ET_PILESTONES
	ET_HOLYCURTAIN
	ET_FIRE
	ET_SCULPEICE
	ET_WEB
	ET_FIREFLOWER
	ET_MAGICLOCK
	ET_MAPMAGIC
	ET_SPACEDOOR
)

type BaseEvent struct {
	ID          int32
	EventType   EventType
	MapName     string
	X, Y        int
	StartTime   time.Time
	Duration    time.Duration
	Enabled     bool
	Owner       interface{}
	DisposeTick time.Time
}

type TEventManager struct {
	Events     map[int32]*BaseEvent
	EventCount int32
	Mutex      sync.RWMutex
}

var DefaultEventManager *TEventManager

func init() {
	DefaultEventManager = NewEventManager()
	go DefaultEventManager.ProcessLoop()
}

func NewEventManager() *TEventManager {
	return &TEventManager{
		Events: make(map[int32]*BaseEvent),
	}
}

func (em *TEventManager) GetNextEventID() int32 {
	return atomic.AddInt32(&em.EventCount, 1)
}

func (em *TEventManager) AddEvent(event *BaseEvent) {
	em.Mutex.Lock()
	defer em.Mutex.Unlock()

	event.ID = em.GetNextEventID()
	em.Events[event.ID] = event

	if event.Duration > 0 {
		go em.timerEvent(event)
	}
}

func (em *TEventManager) timerEvent(event *BaseEvent) {
	ticker := time.NewTicker(event.Duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			em.DelEvent(event.ID)
			return
		}
	}
}

func (em *TEventManager) DelEvent(id int32) {
	em.Mutex.Lock()
	defer em.Mutex.Unlock()
	delete(em.Events, id)
}

func (em *TEventManager) GetEvent(id int32) *BaseEvent {
	em.Mutex.RLock()
	defer em.Mutex.RUnlock()
	return em.Events[id]
}

func (em *TEventManager) GetEventsInMap(mapName string) []*BaseEvent {
	em.Mutex.RLock()
	defer em.Mutex.RUnlock()

	var result []*BaseEvent
	for _, event := range em.Events {
		if event.MapName == mapName {
			result = append(result, event)
		}
	}
	return result
}

func (em *TEventManager) GetEventsInRange(mapName string, x, y, range_ int) []*BaseEvent {
	em.Mutex.RLock()
	defer em.Mutex.RUnlock()

	var result []*BaseEvent
	for _, event := range em.Events {
		if event.MapName != mapName {
			continue
		}

		dx := event.X - x
		dy := event.Y - y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}

		if dx <= range_ && dy <= range_ {
			result = append(result, event)
		}
	}
	return result
}

func (em *TEventManager) ProcessLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		em.ProcessEvents()
	}
}

func (em *TEventManager) ProcessEvents() {
	em.Mutex.Lock()
	defer em.Mutex.Unlock()

	now := time.Now()
	for _, event := range em.Events {
		if event.DisposeTick.IsZero() {
			continue
		}

		if now.After(event.DisposeTick) {
			delete(em.Events, event.ID)
		}
	}
}

type TFireEvent struct {
	*BaseEvent
	Damage int
}

func NewFireEvent(mapName string, x, y int, duration time.Duration, damage int) *TFireEvent {
	return &TFireEvent{
		BaseEvent: &BaseEvent{
			EventType: ET_FIRE,
			MapName:   mapName,
			X:         x,
			Y:         y,
			Duration:  duration,
			StartTime: time.Now(),
			Enabled:   true,
		},
		Damage: damage,
	}
}

type TTrapEvent struct {
	*BaseEvent
	Damage int
}

func NewTrapEvent(mapName string, x, y int, duration time.Duration, damage int) *TTrapEvent {
	return &TTrapEvent{
		BaseEvent: &BaseEvent{
			EventType: ET_TRAP,
			MapName:   mapName,
			X:         x,
			Y:         y,
			Duration:  duration,
			StartTime: time.Now(),
			Enabled:   true,
		},
		Damage: damage,
	}
}

type TMagicLockEvent struct {
	*BaseEvent
	TargetID int32
	Damage   int
}

func NewMagicLockEvent(mapName string, x, y int, targetID int32, duration time.Duration) *TMagicLockEvent {
	return &TMagicLockEvent{
		BaseEvent: &BaseEvent{
			EventType: ET_MAGICLOCK,
			MapName:   mapName,
			X:         x,
			Y:         y,
			Duration:  duration,
			StartTime: time.Now(),
			Enabled:   true,
		},
		TargetID: targetID,
	}
}

type TWebEvent struct {
	*BaseEvent
}

func NewWebEvent(mapName string, x, y int, duration time.Duration) *TWebEvent {
	return &TWebEvent{
		BaseEvent: &BaseEvent{
			EventType: ET_WEB,
			MapName:   mapName,
			X:         x,
			Y:         y,
			Duration:  duration,
			StartTime: time.Now(),
			Enabled:   true,
		},
	}
}

func MapEventType(et EventType) uint16 {
	switch et {
	case ET_TRAP:
		return protocol.ET_TRAP
	case ET_DIGOUTZOMBI:
		return protocol.ET_DIGOUTZOMBI
	case ET_FIRE:
		return protocol.ET_FIRE
	case ET_WEB:
		return protocol.ET_WEB
	case ET_MAGICLOCK:
		return protocol.ET_MAGICLOCK
	default:
		return 0
	}
}
