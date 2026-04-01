package actor

import (
	"math/rand"
	"sync"
	"time"

	"github.com/mir2go/mir2/pkg/game/event"
	gamemap "github.com/mir2go/mir2/pkg/game/map"
	"github.com/mir2go/mir2/pkg/protocol"
)

type MonsterRace int

const (
	RACE_NONE         MonsterRace = 0
	RACE_PLAYER       MonsterRace = 1
	RACE_NPC          MonsterRace = 2
	RACE_MONSTER      MonsterRace = 3
	RACE_ANIMAL       MonsterRace = 4
	RACEGuard         MonsterRace = 5
	RACE_HUMAN        MonsterRace = 6
	RACE_MERC         MonsterRace = 7
	RACE_ESSENCE      MonsterRace = 8
	RACE_NOMAL_NPC    MonsterRace = 10
	RACE_ANGEL        MonsterRace = 11
	RACE_ARCHERGUARD  MonsterRace = 12
	RACE_HERO         MonsterRace = 66
)

type AIState int

const (
	AI_STATE_IDLE    AIState = iota
	AI_STATE_MOVE
	AI_STATE_TRACE
	AI_STATE_ATTACK
	AI_STATE_RETURN
	AI_STATE_FLEE
	AI_STATE_DEAD
)

type AIMode int

const (
	AI_MODE_NORMAL    AIMode = iota
	AI_MODE_ATTACK
	AI_MODE_GUARD
	AI_MODE_FOLLOW
	AI_MODE_PATROL
)

type TMonster struct {
	*BaseObject
	Master         *Player
	
	MonsterType    MonsterRace
	Appr           uint16
	MonsterName    string
	
	ViewRange      int
	AttackRange    int
	WalkSpeed      int
	RunSpeed       int
	
	HP             int32
	MaxHP          int32
	MP             int32
	MaxMP          int32
	AC             int32
	MAC            int32
	DC             int32
	MaxDC          int32
	MC             int32
	SC             int32
	Speed          int
	HitRate        int
	
	Exp            uint32
	Gold           int32
	
	AIState        AIState
	AIMode         AIMode
	
	WalkCount      int
	WalkTick       time.Time
	AttackTick     time.Time
	AIThinkTick    time.Time
	SearchTick     time.Time
	
	TargetID       int32
	TargetX, TargetY int
	HomeX, HomeY   int
	
	DropItemList   []*DropItem
	DropGold       int32
	
	Undead         bool
	HolySeize      bool
	HolySeizeTick  time.Time
	
	Poisoned       bool
	PoisonDamage   int32
	
	Event          *event.BaseEvent
	
	SkillList      []*MonsterSkill
	
	NoDropItem     bool
	NoGold         bool
	Castle         bool
	
	OnMouseMove    bool
	OnMouseReturn  bool
	OnMouseStruck  bool
	
	WalkWait       int
	WalkStep       int
	
	DeathTime      time.Time
}

type DropItem struct {
	ItemName   string
	DropRate   int
	DropCount  int
	ItemType   string
	NeedLevel  int
	JobLimit   Job
	MaxDura    uint16
	AC         int32
	MAC        int32
	DC         int32
	MC         int32
	SC         int32
	HP         int32
	MP         int32
	Speed      int32
	HIT        int32
}

type ItemDrops struct {
	Items      []DropItem
	Mutex      sync.RWMutex
}

type MonsterSkill struct {
	SkillID    uint16
	SkillName  string
	Power      int
	Range      int
	Delay      time.Duration
	LastUse    time.Time
}

func NewMonster(id int32, name string, appr uint16) *TMonster {
	return &TMonster{
		BaseObject: &BaseObject{
			ID:         id,
			Name:       name,
			RaceServer: protocol.RC_MONSTER,
			Alive:      true,
			WalkSpeed:  500,
			ViewRange:  6,
		},
		MonsterName:    name,
		Appr:           appr,
		RunSpeed:       350,
		AttackRange:    1,
		AIState:        AI_STATE_IDLE,
		AIMode:         AI_MODE_NORMAL,
		WalkTick:       time.Now(),
		AIThinkTick:    time.Now(),
		SearchTick:     time.Now(),
		WalkStep:       1,
		WalkWait:       600,
		SkillList:      make([]*MonsterSkill, 0),
		DropItemList:   make([]*DropItem, 0),
	}
}

func (m *TMonster) Initialize() {
	m.RecalculateAbility()
}

func (m *TMonster) RecalculateAbility() {
	m.MaxHP = int32(100 + m.Level*10)
	m.MaxMP = int32(50 + m.Level*5)
	
	if m.HP == 0 {
		m.HP = m.MaxHP
	}
	if m.MP == 0 {
		m.MP = m.MaxMP
	}
}

func (m *TMonster) Think() {
	if !m.Alive {
		return
	}
	
	if time.Since(m.AIThinkTick) < time.Millisecond*500 {
		return
	}
	m.AIThinkTick = time.Now()
	
	if m.HolySeize && time.Since(m.HolySeizeTick) > time.Second*10 {
		m.HolySeize = false
	}
	
	if m.HolySeize {
		m.HolySeizeTick = time.Now()
		return
	}
	
	switch m.AIState {
	case AI_STATE_IDLE:
		m.AI_Idle()
	case AI_STATE_MOVE:
		m.AI_Move()
	case AI_STATE_TRACE:
		m.AI_Trace()
	case AI_STATE_ATTACK:
		m.AI_Attack()
	case AI_STATE_RETURN:
		m.AI_Return()
	}
}

func (m *TMonster) AI_Idle() {
	if time.Since(m.SearchTick) < time.Second*2 {
		return
	}
	m.SearchTick = time.Now()
	
	target := m.SearchTarget()
	if target != nil {
		m.TargetID = target.GetID()
		m.AIState = AI_STATE_TRACE
		return
	}
	
	if rand.Intn(100) < 30 {
		m.AIState = AI_STATE_MOVE
	}
}

func (m *TMonster) AI_Move() {
	if time.Since(m.WalkTick) < time.Duration(m.WalkSpeed)*time.Millisecond {
		return
	}
	
	if m.WalkCount >= m.WalkStep {
		m.WalkCount = 0
		m.AIState = AI_STATE_IDLE
		return
	}
	
	dir := rand.Intn(8)
nx, ny := m.GetNextPos(dir, 1)
	
	gm := gamemap.GetMapManager()
	gameMap := gm.GetMap(m.MapName)
	if gameMap == nil {
		return
	}
	
	if gameMap.CanWalk(nx, ny) {
		m.X = nx
		m.Y = ny
		m.Direction = byte(dir)
		m.WalkCount++
	}
	
	m.WalkTick = time.Now()
}

func (m *TMonster) AI_Trace() {
	if m.TargetID == 0 {
		m.AIState = AI_STATE_IDLE
		return
	}
	
	target := GetActorByID(m.TargetID)
	if target == nil || !target.IsAlive() {
		m.TargetID = 0
		m.AIState = AI_STATE_IDLE
		return
	}
	
	dist := Distance(m.X, m.Y, target.GetX(), target.GetY())
	
	if dist <= m.AttackRange {
		m.AIState = AI_STATE_ATTACK
		return
	}
	
	if dist > m.ViewRange*2 {
		m.TargetID = 0
		m.AIState = AI_STATE_IDLE
		return
	}
	
	m.TraceTarget(target)
}

func (m *TMonster) AI_Attack() {
	if time.Since(m.AttackTick) < time.Duration(m.WalkSpeed)*time.Millisecond {
		return
	}
	
	if m.TargetID == 0 {
		m.AIState = AI_STATE_IDLE
		return
	}
	
	target := GetActorByID(m.TargetID)
	if target == nil || !target.IsAlive() {
		m.TargetID = 0
		m.AIState = AI_STATE_IDLE
		return
	}
	
	dist := Distance(m.X, m.Y, target.GetX(), target.GetY())
	if dist > m.AttackRange {
		m.AIState = AI_STATE_TRACE
		return
	}
	
	m.Attack(target)
	m.AttackTick = time.Now()
	
	if rand.Intn(100) < 50 {
		m.AIState = AI_STATE_IDLE
	}
}

func (m *TMonster) AI_Return() {
	if m.X == m.HomeX && m.Y == m.HomeY {
		m.AIState = AI_STATE_IDLE
		return
	}
	
	m.TraceToHome()
	
	if m.X == m.HomeX && m.Y == m.HomeY {
		m.AIState = AI_STATE_IDLE
	}
}

func (m *TMonster) SearchTarget() Actor {
	gm := gamemap.GetMapManager()
	gameMap := gm.GetMap(m.MapName)
	if gameMap == nil {
		return nil
	}
	
	players := gameMap.Players
	var nearest Actor
	minDist := m.ViewRange + 1
	
	for _, p := range players {
		if player, ok := p.(*Player); ok {
			if !player.IsAlive() {
				continue
			}
			
			if m.Master != nil && player.GetID() == m.Master.GetID() {
				continue
			}
			
			dist := Distance(m.X, m.Y, player.GetX(), player.GetY())
			if dist <= m.ViewRange && dist < minDist {
				minDist = dist
				nearest = player
			}
		}
	}
	
	return nearest
}

func (m *TMonster) TraceTarget(target Actor) {
	if time.Since(m.WalkTick) < time.Duration(m.RunSpeed)*time.Millisecond {
		return
	}
	
	dx := target.GetX() - m.X
	dy := target.GetY() - m.Y
	
	var dir int
	if dx == 0 && dy < 0 {
		dir = 0
	} else if dx > 0 && dy < 0 {
		dir = 1
	} else if dx > 0 && dy == 0 {
		dir = 2
	} else if dx > 0 && dy > 0 {
		dir = 3
	} else if dx == 0 && dy > 0 {
		dir = 4
	} else if dx < 0 && dy > 0 {
		dir = 5
	} else if dx < 0 && dy == 0 {
		dir = 6
	} else {
		dir = 7
	}
	
	nx, ny := m.GetNextPos(dir, 1)
	
	gm := gamemap.GetMapManager()
	gameMap := gm.GetMap(m.MapName)
	if gameMap == nil {
		return
	}
	
	if gameMap.CanWalk(nx, ny) {
		m.X = nx
		m.Y = ny
		m.Direction = byte(dir)
	}
	
	m.WalkTick = time.Now()
}

func (m *TMonster) TraceToHome() {
	if time.Since(m.WalkTick) < time.Duration(m.RunSpeed)*time.Millisecond {
		return
	}
	
	dx := m.HomeX - m.X
	dy := m.HomeY - m.Y
	
	var dir int
	if dx == 0 && dy < 0 {
		dir = 0
	} else if dx > 0 && dy < 0 {
		dir = 1
	} else if dx > 0 && dy == 0 {
		dir = 2
	} else if dx > 0 && dy > 0 {
		dir = 3
	} else if dx == 0 && dy > 0 {
		dir = 4
	} else if dx < 0 && dy > 0 {
		dir = 5
	} else if dx < 0 && dy == 0 {
		dir = 6
	} else {
		dir = 7
	}
	
	nx, ny := m.GetNextPos(dir, 1)
	
	gm := gamemap.GetMapManager()
	gameMap := gm.GetMap(m.MapName)
	if gameMap == nil {
		return
	}
	
	if gameMap.CanWalk(nx, ny) {
		m.X = nx
		m.Y = ny
	}
	
	m.WalkTick = time.Now()
}

func (m *TMonster) Attack(target Actor) {
	damage := m.CalcAttackPower(m.DC, m.MaxDC)
	
	crit := false
	if rand.Intn(100) < 10 {
		damage *= 2
		crit = true
	}
	
	target.AddHP(-damage)
	
	m.SendAttackMsg(target.GetID(), damage, crit)
	
	if target.GetHP() <= 0 {
		m.OnKill(target)
	}
}

func (m *TMonster) CalcAttackPower(dc, maxDC int32) int32 {
	base := dc
	if maxDC > dc {
		base += int32(rand.Intn(int(maxDC - dc + 1)))
	}
	return base
}

func (m *TMonster) SendAttackMsg(targetID int32, damage int32, crit bool) {
}

func (m *TMonster) OnKill(target Actor) {
	if _, ok := target.(*Player); ok {
		if m.Master != nil {
			m.Master.AddExp(int64(m.Exp))
		}
	}
	
	if !m.NoDropItem {
		m.DropItems()
	}
	
	if !m.NoGold && m.DropGold > 0 {
		m.DropGoldItems()
	}
	
	m.Die()
}

func (m *TMonster) DropItems() {
	if len(m.DropItemList) == 0 {
		m.DropDefaultItems()
		return
	}
	
	for _, drop := range m.DropItemList {
		if rand.Intn(10000) < drop.DropRate {
			player := GetActorByID(m.TargetID)
			if player != nil {
				createDroppedItem(player, &drop)
			}
			break
		}
	}
}

func (m *TMonster) DropDefaultItems() {
	dropTable := GetDefaultDropTable()
	
	for _, item := range dropTable.Items {
		if rand.Intn(10000) < item.DropRate {
			player := GetActorByID(m.TargetID)
			if player != nil {
				createDroppedItem(player, &item)
			}
			break
		}
	}
}

func createDroppedItem(target Actor, drop *DropItem) {
	if p, ok := target.(*Player); ok {
		item := &TItem{
			MakeIndex: 0,
			Index:     0,
			Dura:      drop.MaxDura,
			DuraMax:   drop.MaxDura,
		}
		
		p.AddItem(item)
	}
}

func (m *TMonster) DropGoldItems() {
}

func (m *TMonster) Die() {
	m.Alive = false
	m.AIState = AI_STATE_DEAD
	
	m.DeathTime = time.Now()
}

func (m *TMonster) AddHP(value int32) {
	m.HP += value
	if m.HP > m.MaxHP {
		m.HP = m.MaxHP
	}
	if m.HP < 0 {
		m.HP = 0
	}
}

func (m *TMonster) AddMP(value int32) {
	m.MP += value
	if m.MP > m.MaxMP {
		m.MP = m.MaxMP
	}
	if m.MP < 0 {
		m.MP = 0
	}
}

func (m *TMonster) Struck(damage int32) {
	m.AddHP(-damage)
	
	if m.HP <= 0 {
		m.Die()
	}
}

func (m *TMonster) GetNextPos(dir, step int) (int, int) {
	nx := m.X
	ny := m.Y
	
	switch dir {
	case 0:
		ny -= step
	case 1:
		nx += step
		ny -= step
	case 2:
		nx += step
	case 3:
		nx += step
		ny += step
	case 4:
		ny += step
	case 5:
		nx -= step
		ny += step
	case 6:
		nx -= step
	case 7:
		nx -= step
		ny -= step
	}
	
	return nx, ny
}

type MonsterManager struct {
	Monsters    map[int32]*TMonster
	MonstersByName map[string][]*TMonster
	Mutex      sync.RWMutex
	NextID     int32
}

var DefaultMonsterManager *MonsterManager

func init() {
	DefaultMonsterManager = NewMonsterManager()
}

func NewMonsterManager() *MonsterManager {
	return &MonsterManager{
		Monsters:       make(map[int32]*TMonster),
		MonstersByName: make(map[string][]*TMonster),
		NextID:         10000,
	}
}

func (mm *MonsterManager) GetNextID() int32 {
	id := mm.NextID
	mm.NextID++
	return id
}

func (mm *MonsterManager) AddMonster(mon *TMonster) {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()
	
	mon.ID = mm.GetNextID()
	mm.Monsters[mon.ID] = mon
	
	if _, ok := mm.MonstersByName[mon.MonsterName]; !ok {
		mm.MonstersByName[mon.MonsterName] = make([]*TMonster, 0)
	}
	mm.MonstersByName[mon.MonsterName] = append(mm.MonstersByName[mon.MonsterName], mon)
}

func (mm *MonsterManager) DelMonster(id int32) {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()
	
	if mon, ok := mm.Monsters[id]; ok {
		name := mon.MonsterName
		if list, ok := mm.MonstersByName[name]; ok {
			for i, m := range list {
				if m.ID == id {
					mm.MonstersByName[name] = append(list[:i], list[i+1:]...)
					break
				}
			}
		}
	}
	delete(mm.Monsters, id)
}

func (mm *MonsterManager) GetMonster(id int32) *TMonster {
	mm.Mutex.RLock()
	defer mm.Mutex.RUnlock()
	return mm.Monsters[id]
}

func (mm *MonsterManager) GetMonstersByName(name string) []*TMonster {
	mm.Mutex.RLock()
	defer mm.Mutex.RUnlock()
	return mm.MonstersByName[name]
}

func (mm *MonsterManager) GetAllMonsters() []*TMonster {
	mm.Mutex.RLock()
	defer mm.Mutex.RUnlock()
	
	result := make([]*TMonster, 0, len(mm.Monsters))
	for _, m := range mm.Monsters {
		result = append(result, m)
	}
	return result
}

func (mm *MonsterManager) ProcessAll() {
	mm.Mutex.RLock()
	monsters := make([]*TMonster, 0, len(mm.Monsters))
	for _, m := range mm.Monsters {
		monsters = append(monsters, m)
	}
	mm.Mutex.RUnlock()
	
	for _, m := range monsters {
		m.Think()
	}
}

func GetMonsterManager() *MonsterManager {
	return DefaultMonsterManager
}

func GetActorByID(id int32) Actor {
	if player := GetPlayerManager().GetPlayer(id); player != nil {
		return player
	}
	return GetMonsterManager().GetMonster(id)
}

func Distance(x1, y1, x2, y2 int) int {
	dx := x1 - x2
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y2
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

type DropTable struct {
	Items []DropItem
	Mutex sync.RWMutex
}

var DefaultDropTable *DropTable

func init() {
	DefaultDropTable = NewDropTable()
}

func NewDropTable() *DropTable {
	dt := &DropTable{
		Items: make([]DropItem, 0),
	}
	dt.initDefaultItems()
	return dt
}

func (dt *DropTable) initDefaultItems() {
	dt.Items = []DropItem{
		{ItemName: "金创药(小)", DropRate: 800, DropCount: 1, MaxDura: 10},
		{ItemName: "金创药(中)", DropRate: 500, DropCount: 1, MaxDura: 20},
		{ItemName: "金创药(大)", DropRate: 300, DropCount: 1, MaxDura: 50},
		{ItemName: "魔法药(小)", DropRate: 800, DropCount: 1, MaxDura: 10},
		{ItemName: "魔法药(中)", DropRate: 500, DropCount: 1, MaxDura: 20},
		{ItemName: "魔法药(大)", DropRate: 300, DropCount: 1, MaxDura: 50},
		{ItemName: "强效金创药", DropRate: 200, DropCount: 1, MaxDura: 100},
		{ItemName: "强效魔法药", DropRate: 200, DropCount: 1, MaxDura: 100},
		{ItemName: "回城卷", DropRate: 500, DropCount: 1, MaxDura: 1},
		{ItemName: "随机卷", DropRate: 400, DropCount: 1, MaxDura: 1},
		{ItemName: "地牢逃脱卷", DropRate: 300, DropCount: 1, MaxDura: 1},
		{ItemName: "祝福油", DropRate: 100, DropCount: 1, MaxDura: 1},
		{ItemName: "战神油", DropRate: 50, DropCount: 1, MaxDura: 1},
		{ItemName: "修复油", DropRate: 80, DropCount: 1, MaxDura: 1},
		{ItemName: "任务卷轴", DropRate: 150, DropCount: 1, MaxDura: 1},
		{ItemName: "回城石", DropRate: 100, DropCount: 1, MaxDura: 1},
		{ItemName: "行会回城卷", DropRate: 50, DropCount: 1, MaxDura: 1},
		{ItemName: "随机传送石", DropRate: 80, DropCount: 1, MaxDura: 1},
		{ItemName: "回城卷(包)", DropRate: 100, DropCount: 1, MaxDura: 1},
		{ItemName: "疗伤药", DropRate: 250, DropCount: 1, MaxDura: 30},
	}
}

func (dt *DropTable) AddItem(item DropItem) {
	dt.Mutex.Lock()
	defer dt.Mutex.Unlock()
	dt.Items = append(dt.Items, item)
}

func (dt *DropTable) GetDropItems() []DropItem {
	dt.Mutex.RLock()
	defer dt.Mutex.RUnlock()
	result := make([]DropItem, len(dt.Items))
	copy(result, dt.Items)
	return result
}

func GetDefaultDropTable() *DropTable {
	return DefaultDropTable
}
