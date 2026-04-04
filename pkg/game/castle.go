package game

import (
	"sync"
	"time"

	"github.com/mir2go/mir2/pkg/game/guild"
)

type CastleState int

const (
	CASTLE_STATE_NONE    CastleState = 0
	CASTLE_STATE_WAR     CastleState = 1
	CASTLE_STATE_WAR_END CastleState = 2
	CASTLE_STATE_OPEN    CastleState = 3
)

type MapPoint struct {
	MapName string
	X, Y    int
}

type TCastle struct {
	CastleID   int32
	CastleName string
	Guild      *guild.TGuild
	GuildName  string

	MapInfo interface{}
	MapName string

	Palace     *MapPoint
	BkWall     *MapPoint
	LeftWall   *MapPoint
	RightWall  *MapPoint
	MainDoor   *MapPoint
	SecretGate *MapPoint

	State        CastleState
	WarDate      time.Time
	WarStartTime time.Time
	WarEndTime   time.Time
	WarGold      int32
	Tax          int32

	RegUsers        map[int32]bool
	AttackGuild     *guild.TGuild
	AttackGuildName string

	TodayDate   time.Time
	IncomeToday int32

	StateChangedTick time.Time
	OwnerChangedTick time.Time

	SaveDate time.Time

	Mutex sync.RWMutex
}

type TCastleManager struct {
	Castles map[int32]*TCastle
	Mutex   sync.RWMutex
}

var DefaultCastleManager *TCastleManager

func init() {
	DefaultCastleManager = NewCastleManager()
}

func NewCastleManager() *TCastleManager {
	return &TCastleManager{
		Castles: make(map[int32]*TCastle),
	}
}

func (cm *TCastleManager) AddCastle(castle *TCastle) {
	cm.Mutex.Lock()
	defer cm.Mutex.Unlock()
	cm.Castles[castle.CastleID] = castle
}

func (cm *TCastleManager) GetCastle(castleID int32) *TCastle {
	cm.Mutex.RLock()
	defer cm.Mutex.RUnlock()
	return cm.Castles[castleID]
}

func (cm *TCastleManager) GetCastleByName(name string) *TCastle {
	cm.Mutex.RLock()
	defer cm.Mutex.RUnlock()

	for _, castle := range cm.Castles {
		if castle.CastleName == name {
			return castle
		}
	}
	return nil
}

func (cm *TCastleManager) GetCastleByGuild(guildID int32) *TCastle {
	cm.Mutex.RLock()
	defer cm.Mutex.RUnlock()

	for _, castle := range cm.Castles {
		if castle.Guild != nil && castle.Guild.ID == guildID {
			return castle
		}
	}
	return nil
}

func (cm *TCastleManager) GetAllCastles() []*TCastle {
	cm.Mutex.RLock()
	defer cm.Mutex.RUnlock()

	result := make([]*TCastle, 0, len(cm.Castles))
	for _, castle := range cm.Castles {
		result = append(result, castle)
	}
	return result
}

func (c *TCastle) GetOwnerGuild() *guild.TGuild {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.Guild
}

func (c *TCastle) SetOwner(guild *guild.TGuild) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	c.Guild = guild
	if guild != nil {
		c.GuildName = guild.Name
	}
	c.OwnerChangedTick = time.Now()
}

func (c *TCastle) IsOwnedGuild(guildID int32) bool {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	if c.Guild == nil {
		return false
	}
	return c.Guild.ID == guildID
}

func (c *TCastle) IsWarTime() bool {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	if c.State != CASTLE_STATE_WAR {
		return false
	}

	return time.Now().Before(c.WarEndTime)
}

func (c *TCastle) StartWar(attackGuild *guild.TGuild) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	c.State = CASTLE_STATE_WAR
	c.AttackGuild = attackGuild
	if attackGuild != nil {
		c.AttackGuildName = attackGuild.Name
	}
	c.WarStartTime = time.Now()
	c.WarEndTime = time.Now().Add(time.Hour * 2)
	c.StateChangedTick = time.Now()
}

func (c *TCastle) EndWar() {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	c.State = CASTLE_STATE_WAR_END
	c.WarEndTime = time.Now()
	c.StateChangedTick = time.Now()

	c.AttackGuild = nil
	c.AttackGuildName = ""
	c.RegUsers = make(map[int32]bool)
}

func (c *TCastle) CheckWarTime() {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if c.State == CASTLE_STATE_WAR {
		if time.Now().After(c.WarEndTime) {
			c.State = CASTLE_STATE_WAR_END
			c.WarEndTime = time.Now()
			c.StateChangedTick = time.Now()
		}
	}
}

func (c *TCastle) AddTax(gold int32) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	c.Tax += gold

	if time.Now().Day() != c.TodayDate.Day() {
		c.IncomeToday = 0
		c.TodayDate = time.Now()
	}

	c.IncomeToday += gold
}

func (c *TCastle) GetTax() int32 {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.Tax
}

func (c *TCastle) CollectTax(amount int32) bool {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if c.Tax < amount {
		return false
	}

	c.Tax -= amount
	return true
}

func (c *TCastle) AddWarGold(gold int32) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.WarGold += gold
}

func (c *TCastle) RegisterAttackGuild(guildID int32) bool {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if c.State != CASTLE_STATE_WAR_END {
		return false
	}

	if c.Guild != nil && c.Guild.ID == guildID {
		return false
	}

	c.RegUsers[guildID] = true
	return true
}

func (c *TCastle) IsRegisteredGuild(guildID int32) bool {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.RegUsers[guildID]
}

func (c *TCastle) GetNextWarDate() time.Time {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.WarDate
}

func (c *TCastle) SetNextWarDate(date time.Time) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.WarDate = date
}

func (c *TCastle) IsUnderAttack() bool {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	if c.State != CASTLE_STATE_WAR {
		return false
	}

	if c.AttackGuild == nil {
		return false
	}

	return true
}

func (c *TCastle) GetAttackGuild() *guild.TGuild {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.AttackGuild
}

func (c *TCastle) Save() error {
	return nil
}

func (c *TCastle) Load() error {
	return nil
}

type TCastleSiege struct {
	Castle    *TCastle
	StartTime time.Time
	EndTime   time.Time
	Status    int
}

func (cm *TCastleManager) LoadCastleConfig(filename string) error {
	return nil
}

func (cm *TCastleManager) Initialize() {
	castle := &TCastle{
		CastleID:   0,
		CastleName: "沙巴克城",
		MapName:    "0",
		State:      CASTLE_STATE_OPEN,
		WarDate:    time.Now(),
		TodayDate:  time.Now(),
	}

	castle.Palace = &MapPoint{MapName: "0", X: 617, Y: 274}
	castle.MainDoor = &MapPoint{MapName: "0", X: 679, Y: 330}
	castle.BkWall = &MapPoint{MapName: "0", X: 655, Y: 274}

	cm.AddCastle(castle)
}

func GetCastleManager() *TCastleManager {
	return DefaultCastleManager
}

func GetCastle(castleID int32) *TCastle {
	return DefaultCastleManager.GetCastle(castleID)
}

func GetCastleByName(name string) *TCastle {
	return DefaultCastleManager.GetCastleByName(name)
}

func IsCastleMap(mapName string) bool {
	for _, castle := range DefaultCastleManager.Castles {
		if castle.MapName == mapName {
			return true
		}
	}
	return false
}

func IsCastleWarTime(castleID int32) bool {
	castle := GetCastle(castleID)
	if castle == nil {
		return false
	}
	return castle.IsWarTime()
}

func IsAttackGuild(castleID int32, guildID int32) bool {
	castle := GetCastle(castleID)
	if castle == nil {
		return false
	}
	return castle.AttackGuild != nil && castle.AttackGuild.ID == guildID
}

func IsOwnerGuild(castleID int32, guildID int32) bool {
	castle := GetCastle(castleID)
	if castle == nil {
		return false
	}
	return castle.IsOwnedGuild(guildID)
}
