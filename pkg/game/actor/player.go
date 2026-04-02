package actor

import (
	"sync"
	"time"

	"github.com/mir2go/mir2/pkg/game/item"
	"github.com/mir2go/mir2/pkg/protocol"
)

type Gender byte
type Job byte

const (
	GenderMan   Gender = 0
	GenderWoman Gender = 1
)

const (
	JobWarr  Job = 0
	JobWizard Job = 1
	JobTaos  Job = 2
	JobRogue Job = 3
)

type TAbility struct {
	Level      int32
	AC, MAC    int32
	DC, MC, SC int32
	CC         int32
	HP, MP     int32
	MaxHP, MaxMP int32
	Exp, MaxExp uint32
	Weight     uint16
	MaxWeight  uint16
	WearWeight uint16
	MaxWearWeight uint16
	HandWeight uint16
	MaxHandWeight uint16
}

type TAddAbility struct {
	WHP, WMP       int32
	HitPoint       uint16
	SpeedPoint     uint16
	AC, MAC        int32
	DC, MC, SC     int32
	CC             int32
	AntiPoison     uint16
	PoisonRecover  uint16
	HealthRecover  uint16
	SpellRecover   uint16
	AntiMagic      uint16
	Luck, UnLuck   int8
	HitSpeed       int32
}

type TCharState struct {
	X, Y        int
	Direction   byte
	Feature     int32
	Status      int32
	Level       int32
	HP, MaxHP   int32
}

type Actor interface {
	GetID() int32
	GetName() string
	GetX() int
	GetY() int
	SetX(int)
	SetY(int)
	GetDirection() byte
	SetDirection(byte)
	GetRaceServer() byte
	GetMapName() string
	IsAlive() bool
	GetHP() int32
	AddHP(value int32)
}

type BaseObject struct {
	ID            int32
	Name          string
	MapName       string
	X, Y          int
	Direction     byte
	Hair          byte
	Gender        Gender
	Job           Job
	Gold          int32
	Level         int32
	Ability       TAbility
	AddAbility    TAddAbility
	HP, MP        int32
	MaxHP, MaxMP  int32
	PKPoint       int32
	Feature       int32
	Status        int32
	RaceServer    byte
	ViewRange     int
	WalkSpeed     int
	HitSpeed      int
	OnHorse       bool
	HorseType     byte
	
	GroupID       int32
	GuildID       int32
	GuildRank     int
	
	Alive         bool
	LoginTime     time.Time
	LastActiveTime time.Time
	
	lock          sync.RWMutex
}

func (o *BaseObject) GetID() int32 { return o.ID }
func (o *BaseObject) GetName() string { return o.Name }
func (o *BaseObject) GetX() int { return o.X }
func (o *BaseObject) GetY() int { return o.Y }
func (o *BaseObject) SetX(x int) { o.X = x }
func (o *BaseObject) SetY(y int) { o.Y = y }
func (o *BaseObject) GetDirection() byte { return o.Direction }
func (o *BaseObject) SetDirection(dir byte) { o.Direction = dir }
func (o *BaseObject) GetRaceServer() byte { return o.RaceServer }
func (o *BaseObject) GetMapName() string { return o.MapName }
func (o *BaseObject) IsAlive() bool { return o.Alive }
func (o *BaseObject) GetHP() int32 { return o.HP }
func (o *BaseObject) AddHP(value int32) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.HP += value
	if o.HP > o.MaxHP {
		o.HP = o.MaxHP
	}
	if o.HP < 0 {
		o.HP = 0
	}
}

func (o *BaseObject) AddMP(value int32) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.MP += value
	if o.MP > o.MaxMP {
		o.MP = o.MaxMP
	}
	if o.MP < 0 {
		o.MP = 0
	}
}

type TItem struct {
	MakeIndex int32
	Index     uint16
	Dura      uint16
	DuraMax   uint16
	Value     [14]byte
	AddValue  [14]byte
	AddPoint  [14]byte
	MaxDate   int64
}

type TUseItems [protocol.U_CHARM + 1]TItem
type TBagItems []*TItem

type Player struct {
	*BaseObject
	CharID        int32
	Account       string
	UserAddr      string
	SessionID     int32
	
	UseItems      TUseItems
	BagItems      TBagItems
	MagicList     []*protocol.THumMagic
	
	BagGold       int32
	StorageGold   int32
	
	Master         *Player
	SlaveList      []*Player
	
	MyHero         *Hero
	
	QuestFlag      protocol.TQuestFlag
	BonusAbil      protocol.TNakedAbility
	BunusPoint     int32
	
	CharStatus     int32
	StatusTimeArr  [protocol.MAX_STATUS_ATTRIBUTE]uint32
	StatusArrValue [protocol.MAX_STATUS_ATTRIBUTE]uint16

	Poisoned       bool
	PoisonDamage   int32
	
	DeathTime      time.Time
	ReclaimTick    time.Time
	
	ConnectionID   int32
	Session        interface{}
	
	Permission     int
	GameMaster     bool
	AllowMsg       bool
	
	TradeInfo     interface{}
	
	Items         []*item.TUserItem
}

type Hero struct {
	*BaseObject
	Owner          *Player
	AngryValue     int32
	Loyality       int32
	Exp            uint32
	HeroGender     Gender
	HeroJob        Job
	Hair           byte
	Weapon          uint16
	Clothes        uint16
	Armor           uint16
	Helmet          uint16
	Necklace        uint16
	Bracelet        uint16
	Ring            uint16
	Belt            uint16
	Shoes           uint16
	Bag             TBagItems
	MagicList      []*protocol.THumMagic
	Items           []*TItem
	StatusTimeArr  [protocol.MAX_STATUS_ATTRIBUTE]uint32
}

func NewHero(owner *Player, name string, job Job, gender Gender) *Hero {
	return &Hero{
		BaseObject: &BaseObject{
			ID:         0,
			Name:       name,
			Job:        job,
			Gender:     gender,
			Level:      1,
			Alive:      true,
			WalkSpeed:  350,
			HitSpeed:   1000,
			RaceServer: protocol.RC_HEROOBJECT,
			HP:         100,
			MP:         100,
			MaxHP:      100,
			MaxMP:      100,
		},
		Owner:     owner,
		HeroJob:   job,
		HeroGender: gender,
		Hair:     1,
		Bag:      make(TBagItems, 0, protocol.MAXHEROBAGITEM),
		MagicList: make([]*protocol.THumMagic, 0, protocol.MAXMAGIC),
		Items:    make([]*TItem, 0),
	}
}

func NewPlayer(id int32, name string) *Player {
	return &Player{
		BaseObject: &BaseObject{
			ID:        id,
			Name:      name,
			Alive:     true,
			WalkSpeed: 300,
			HitSpeed:  1000,
			Level:     1,
			RaceServer: protocol.RC_PLAYOBJECT,
			HP:        100,
			MP:        100,
			MaxHP:     100,
			MaxMP:     100,
			Ability: TAbility{
				Level:    1,
				MaxHP:    100,
				MaxMP:    100,
				MaxWeight: 50,
			},
		},
		BagItems: make(TBagItems, 0, protocol.MAXBAGITEM),
		MagicList: make([]*protocol.THumMagic, 0, protocol.MAXMAGIC),
	}
}

func (p *Player) GetFeature() int32 {
	feature := int32(p.Gender)
	feature |= int32(p.Hair) << 8
	feature |= int32(p.BaseObject.Job) << 16
	feature |= int32(0) << 24
	return feature
}

func (p *Player) GetCharDesc() *protocol.TCharDesc {
	return &protocol.TCharDesc{
		Feature:   p.GetFeature(),
		Status:    p.Status,
		Level:     p.Level,
		HP:        p.HP,
		MaxHP:     p.MaxHP,
		AddStatus: 0,
	}
}

func (p *Player) GetAbility() *protocol.TAbility {
	return &protocol.TAbility{
		Level:         p.Level,
		AC:            int32(p.AddAbility.AC),
		MAC:           int32(p.AddAbility.MAC),
		DC:            int32(p.AddAbility.DC),
		MC:            int32(p.AddAbility.MC),
		SC:            int32(p.AddAbility.SC),
		CC:            int32(p.AddAbility.CC),
		HP:            p.HP,
		MP:            p.MP,
		MaxHP:         p.MaxHP,
		MaxMP:         p.MaxMP,
		Exp:           p.Ability.Exp,
		MaxExp:        p.Ability.MaxExp,
		Weight:        p.Ability.Weight,
		MaxWeight:     p.Ability.MaxWeight,
		WearWeight:    p.Ability.WearWeight,
		MaxWearWeight: p.Ability.MaxWearWeight,
		HandWeight:    p.Ability.HandWeight,
		MaxHandWeight: p.Ability.MaxHandWeight,
	}
}

func (p *Player) AddItem(item *TItem) bool {
	if len(p.BagItems) >= protocol.MAXBAGITEM {
		return false
	}
	p.BagItems = append(p.BagItems, item)
	return true
}

func (p *Player) DelItem(makeIndex int32) bool {
	for i, item := range p.BagItems {
		if item != nil && item.MakeIndex == makeIndex {
			p.BagItems = append(p.BagItems[:i], p.BagItems[i+1:]...)
			return true
		}
	}
	return false
}

func (p *Player) AddMagic(magic *protocol.THumMagic) bool {
	if len(p.MagicList) >= protocol.MAXMAGIC {
		return false
	}
	for _, m := range p.MagicList {
		if m.MagIdx == magic.MagIdx {
			return false
		}
	}
	p.MagicList = append(p.MagicList, magic)
	return true
}

func (p *Player) AddExp(amount int64) bool {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.Ability.Exp += uint32(amount)
	if p.Ability.Exp >= p.Ability.MaxExp {
		p.Level++
		p.Ability.Exp = 0
		p.Ability.MaxExp = uint32(float64(p.Ability.MaxExp) * 1.5)
		return true
	}
	return false
}

type PlayerManager struct {
	Players      map[int32]*Player
	PlayersByName map[string]*Player
	Mutex        sync.RWMutex
	NextID       int32
}

var DefaultPlayerManager *PlayerManager

func init() {
	DefaultPlayerManager = NewPlayerManager()
}

func NewPlayerManager() *PlayerManager {
	return &PlayerManager{
		Players:      make(map[int32]*Player),
		PlayersByName: make(map[string]*Player),
		NextID:       1,
	}
}

func (pm *PlayerManager) GetNextID() int32 {
	id := pm.NextID
	pm.NextID++
	return id
}

func (pm *PlayerManager) AddPlayer(player *Player) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	
	if player.ID == 0 {
		player.ID = pm.GetNextID()
	}
	pm.Players[player.ID] = player
	pm.PlayersByName[player.Name] = player
}

func (pm *PlayerManager) DelPlayer(id int32) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	
	if player, ok := pm.Players[id]; ok {
		delete(pm.PlayersByName, player.Name)
	}
	delete(pm.Players, id)
}

func (pm *PlayerManager) GetPlayer(id int32) *Player {
	pm.Mutex.RLock()
	defer pm.Mutex.RUnlock()
	return pm.Players[id]
}

func (pm *PlayerManager) GetPlayerByName(name string) *Player {
	pm.Mutex.RLock()
	defer pm.Mutex.RUnlock()
	return pm.PlayersByName[name]
}

func (pm *PlayerManager) GetAllPlayers() []*Player {
	pm.Mutex.RLock()
	defer pm.Mutex.RUnlock()
	
	result := make([]*Player, 0, len(pm.Players))
	for _, p := range pm.Players {
		result = append(result, p)
	}
	return result
}

func GetPlayerManager() *PlayerManager {
	return DefaultPlayerManager
}

func (p *Player) SendMessage(msg string) {
}

func (p *Player) Kick() {
	p.Alive = false
}

func FindPlayerByName(name string) *Player {
	return GetPlayerManager().GetPlayerByName(name)
}
