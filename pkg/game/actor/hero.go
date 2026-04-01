package actor

import (
	"sync"
	"time"

	"github.com/mir2go/mir2/pkg/protocol"
)

type HeroManager struct {
	Heroes    map[int32]*Hero
	HeroesByOwner map[int32]*Hero
	Mutex     sync.RWMutex
	NextID    int32
}

var DefaultHeroManager *HeroManager

func init() {
	DefaultHeroManager = NewHeroManager()
}

func NewHeroManager() *HeroManager {
	return &HeroManager{
		Heroes:       make(map[int32]*Hero),
		HeroesByOwner: make(map[int32]*Hero),
		NextID:       20000,
	}
}

func (hm *HeroManager) GetNextID() int32 {
	id := hm.NextID
	hm.NextID++
	return id
}

func (hm *HeroManager) CreateHero(owner *Player, name string, job Job, gender Gender) *Hero {
	hm.Mutex.Lock()
	defer hm.Mutex.Unlock()
	
	if owner.MyHero != nil {
		return nil
	}
	
	hero := NewHero(owner, name, job, gender)
	hero.ID = hm.GetNextID()
	
	hm.Heroes[hero.ID] = hero
	hm.HeroesByOwner[owner.ID] = hero
	owner.MyHero = hero
	
	return hero
}

func (hm *HeroManager) DelHero(heroID int32) bool {
	hm.Mutex.Lock()
	defer hm.Mutex.Unlock()
	
	hero, ok := hm.Heroes[heroID]
	if !ok {
		return false
	}
	
	if hero.Owner != nil {
		hero.Owner.MyHero = nil
		delete(hm.HeroesByOwner, hero.Owner.ID)
	}
	
	delete(hm.Heroes, heroID)
	return true
}

func (hm *HeroManager) GetHero(heroID int32) *Hero {
	hm.Mutex.RLock()
	defer hm.Mutex.RUnlock()
	return hm.Heroes[heroID]
}

func (hm *HeroManager) GetHeroByOwner(ownerID int32) *Hero {
	hm.Mutex.RLock()
	defer hm.Mutex.RUnlock()
	return hm.HeroesByOwner[ownerID]
}

func (hm *HeroManager) GetAllHeroes() []*Hero {
	hm.Mutex.RLock()
	defer hm.Mutex.RUnlock()
	
	result := make([]*Hero, 0, len(hm.Heroes))
	for _, h := range hm.Heroes {
		result = append(result, h)
	}
	return result
}

func (hm *HeroManager) ProcessAll() {
	hm.Mutex.RLock()
	heroes := make([]*Hero, 0, len(hm.Heroes))
	for _, h := range hm.Heroes {
		heroes = append(heroes, h)
	}
	hm.Mutex.RUnlock()
	
	for _, h := range heroes {
		h.Process()
	}
}

func GetHeroManager() *HeroManager {
	return DefaultHeroManager
}

func (h *Hero) Process() {
	if !h.Alive {
		return
	}
	
	h.ProcessStatus()
	h.ProcessAI()
}

func (h *Hero) ProcessStatus() {
	for i := 0; i < protocol.MAX_STATUS_ATTRIBUTE; i++ {
		if h.StatusTimeArr[i] > 0 {
			h.StatusTimeArr[i]--
			if h.StatusTimeArr[i] == 0 {
				h.RemoveStatus(i)
			}
		}
	}
}

func (h *Hero) ProcessAI() {
}

func (h *Hero) RemoveStatus(status int) {
	h.Status &= ^uint32(1 << status)
}

func (h *Hero) AddAngryValue(value int32) {
	h.AngryValue += value
	if h.AngryValue > 1000 {
		h.AngryValue = 1000
	}
	if h.AngryValue < 0 {
		h.AngryValue = 0
	}
}

func (h *Hero) GetFeature() int32 {
	feature := int32(h.HeroGender)
	feature |= int32(h.Hair) << 8
	feature |= int32(h.HeroJob) << 16
	feature |= 0x02000000
	return feature
}

func (h *Hero) GetAbility() *protocol.TAbility {
	level := h.Level
	baseHP := int32(50 + level*20)
	baseMP := int32(30 + level*15)
	
	switch h.HeroJob {
	case JobWarr:
		baseHP += int32(level * 30)
	case JobWizard:
		baseMP += int32(level * 25)
	case JobTaos:
		baseHP += int32(level * 15)
		baseMP += int32(level * 20)
	}
	
	return &protocol.TAbility{
		Level:    h.Level,
		HP:       h.HP,
		MP:       h.MP,
		MaxHP:    h.MaxHP,
		MaxMP:    h.MaxMP,
		Exp:      h.Exp,
		MaxExp:   uint32(100 * level * level),
		AC:       h.AddAbility.AC,
		MAC:      h.AddAbility.MAC,
		DC:       h.AddAbility.DC,
		MC:       h.AddAbility.MC,
		SC:       h.AddAbility.SC,
	}
}

func (h *Hero) AddExp(amount int64) bool {
	h.Exp += uint32(amount)
	
	needExp := uint32(100 * int64(h.Level) * int64(h.Level))
	if h.Exp >= needExp {
		h.Level++
		h.Exp = 0
		h.MaxHP += 20
		h.MaxMP += 10
		h.HP = h.MaxHP
		h.MP = h.MaxMP
		return true
	}
	return false
}

func (h *Hero) RecalculateAbility() {
	level := int(h.Level)
	
	h.MaxHP = int32(50 + level*20)
	h.MaxMP = int32(30 + level*15)
	
	switch h.Job {
	case JobWarr:
		h.MaxHP += int32(level * 30)
		h.AddAbility.DC = int32(1 + level/5)
		h.AddAbility.AC = int32(1 + level/7)
	case JobWizard:
		h.MaxMP += int32(level * 25)
		h.AddAbility.MC = int32(1 + level/5)
	case JobTaos:
		h.MaxHP += int32(level * 15)
		h.MaxMP += int32(level * 20)
		h.AddAbility.SC = int32(1 + level/5)
		h.AddAbility.AC = int32(1 + level/10)
	}
	
	if h.HP > h.MaxHP {
		h.HP = h.MaxHP
	}
	if h.MP > h.MaxMP {
		h.MP = h.MaxMP
	}
}