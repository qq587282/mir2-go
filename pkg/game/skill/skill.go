package skill

import (
	"math/rand"
	"sync"
	"time"

	"github.com/mir2go/mir2/pkg/game/actor"
	"github.com/mir2go/mir2/pkg/protocol"
)

type SkillEffect byte

const (
	EFFECT_NONE SkillEffect = iota
	EFFECT_FIRE
	EFFECT_ICE
	EFFECT_LIGHTNING
	EFFECT_WIND
)

type TMagic struct {
	MagicID     uint16
	MagicName   string
	EffectType  SkillEffect
	Effect      uint8
	Spell       uint16
	Power       uint16
	TrainLevel  [4]uint8
	MaxTrain    [4]int32
	TrainLv     uint8
	Job         uint8
	MagicIdx    uint16
	DelayTime   uint32
	DefSpell    uint8
	DefPower    uint8
	MaxPower    uint16
	DefMaxPower uint8
	Description string
}

type TUserMagic struct {
	*TMagic
	MagIdx    uint16
	Level     uint8
	Key       uint8
	TranPoint int32
	TrainTime time.Time
}

type MagicManager struct {
	MagicList map[uint16]*TMagic
	Mutex     sync.RWMutex
}

var DefaultMagicManager *MagicManager

func init() {
	DefaultMagicManager = NewMagicManager()
	DefaultMagicManager.LoadDefaultMagics()
}

func NewMagicManager() *MagicManager {
	return &MagicManager{
		MagicList: make(map[uint16]*TMagic),
	}
}

func (mm *MagicManager) LoadDefaultMagics() {
	magics := []*TMagic{
		{MagicID: protocol.SKILL_FIREBALL, MagicName: "火球术", EffectType: EFFECT_FIRE, Spell: 15, Power: 20},
		{MagicID: protocol.SKILL_HEALLING, MagicName: "治愈术", EffectType: EFFECT_NONE, Spell: 20, Power: 30},
		{MagicID: protocol.SKILL_FENCING, MagicName: "基本剑术", EffectType: EFFECT_NONE, Spell: 0, Power: 0},
		{MagicID: protocol.SKILL_SPIRITSWORD, MagicName: "精神力战法", EffectType: EFFECT_NONE, Spell: 0, Power: 0},
		{MagicID: protocol.SKILL_GREATFIREBALL, MagicName: "大火球", EffectType: EFFECT_FIRE, Spell: 25, Power: 40},
		{MagicID: protocol.SKILL_POISONING, MagicName: "施毒术", EffectType: EFFECT_NONE, Spell: 30, Power: 20},
		{MagicID: protocol.SKILL_SLAYING, MagicName: "刺杀剑术", EffectType: EFFECT_NONE, Spell: 25, Power: 30},
		{MagicID: protocol.SKILL_REPULSION, MagicName: "抗拒火环", EffectType: EFFECT_FIRE, Spell: 22, Power: 0},
		{MagicID: protocol.SKILL_HELLFIRE, MagicName: "地狱火", EffectType: EFFECT_FIRE, Spell: 35, Power: 50},
		{MagicID: protocol.SKILL_LIGHTNING, MagicName: "雷电术", EffectType: EFFECT_LIGHTNING, Spell: 26, Power: 45},
		{MagicID: protocol.SKILL_THUNDERBOLT, MagicName: "疾光电影", EffectType: EFFECT_LIGHTNING, Spell: 32, Power: 55},
		{MagicID: protocol.SKILL_THRUSTING, MagicName: "野蛮冲撞", EffectType: EFFECT_NONE, Spell: 30, Power: 0},
		{MagicID: protocol.SKILL_SOULFIREBALL, MagicName: "灵魂火符", EffectType: EFFECT_FIRE, Spell: 28, Power: 35},
		{MagicID: protocol.SKILL_SOULSHIELD, MagicName: "幽灵盾", EffectType: EFFECT_NONE, Spell: 32, Power: 0},
		{MagicID: protocol.SKILL_DEJIWONHO, MagicName: "召唤骷髅", EffectType: EFFECT_NONE, Spell: 35, Power: 0},
		{MagicID: protocol.SKILL_HOLYSHIELD, MagicName: "神圣战甲术", EffectType: EFFECT_NONE, Spell: 38, Power: 0},
		{MagicID: protocol.SKILL_TAMMING, MagicName: "诱惑之光", EffectType: EFFECT_NONE, Spell: 20, Power: 0},
		{MagicID: protocol.SKILL_SPACEMOVE, MagicName: "瞬息移动", EffectType: EFFECT_NONE, Spell: 28, Power: 0},
		{MagicID: protocol.SKILL_EARTHFIRE, MagicName: "地狱雷光", EffectType: EFFECT_LIGHTNING, Spell: 40, Power: 60},
		{MagicID: protocol.SKILL_FIREBOOM, MagicName: "爆裂火焰", EffectType: EFFECT_FIRE, Spell: 30, Power: 45},
		{MagicID: protocol.SKILL_LIGHTFLOWER, MagicName: "圣言术", EffectType: EFFECT_LIGHTNING, Spell: 38, Power: 50},
		{MagicID: protocol.SKILL_BANWOL, MagicName: "冰咆哮", EffectType: EFFECT_ICE, Spell: 42, Power: 70},
		{MagicID: protocol.SKILL_FIRESWORD, MagicName: "烈火剑法", EffectType: EFFECT_FIRE, Spell: 45, Power: 80},
		{MagicID: protocol.SKILL_SNOWWIND, MagicName: "寒冰掌", EffectType: EFFECT_ICE, Spell: 33, Power: 40},
		{MagicID: protocol.SKILL_CROSSHALFMOON, MagicName: "半月弯刀", EffectType: EFFECT_NONE, Spell: 30, Power: 25},
		{MagicID: protocol.SKILL_BLIZZARD, MagicName: "冰旋风", EffectType: EFFECT_ICE, Spell: 48, Power: 85},
		{MagicID: protocol.SKILL_TWINDRAKEBLADE, MagicName: "双龙斩", EffectType: EFFECT_NONE, Spell: 46, Power: 75},
		{MagicID: protocol.SKILL_FROSTCRUNCH, MagicName: "霜冰冻", EffectType: EFFECT_ICE, Spell: 40, Power: 65},
	}
	
	for _, m := range magics {
		mm.MagicList[m.MagicID] = m
	}
}

func (mm *MagicManager) GetMagic(id uint16) *TMagic {
	mm.Mutex.RLock()
	defer mm.Mutex.RUnlock()
	return mm.MagicList[id]
}

func (mm *MagicManager) AddMagic(m *TMagic) {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()
	mm.MagicList[m.MagicID] = m
}

func (mm *MagicManager) GetAllMagics() []*TMagic {
	mm.Mutex.RLock()
	defer mm.Mutex.RUnlock()
	
	result := make([]*TMagic, 0, len(mm.MagicList))
	for _, m := range mm.MagicList {
		result = append(result, m)
	}
	return result
}

type MagicHandler func(caster interface{}, target interface{}, magic *TUserMagic) bool

type MagicSystem struct {
	*MagicManager
	Handlers map[uint16]MagicHandler
}

func NewMagicSystem() *MagicSystem {
	return &MagicSystem{
		MagicManager: DefaultMagicManager,
		Handlers:     make(map[uint16]MagicHandler),
	}
}

func (ms *MagicSystem) RegisterHandler(magicID uint16, handler MagicHandler) {
	ms.Handlers[magicID] = handler
}

func (ms *MagicSystem) UseMagic(caster, target interface{}, magicID uint16) bool {
	magic := ms.GetMagic(magicID)
	if magic == nil {
		return false
	}
	
	if handler, ok := ms.Handlers[magicID]; ok {
		userMagic := &TUserMagic{
			TMagic: magic,
			Level:  1,
		}
		return handler(caster, target, userMagic)
	}
	
	return ms.DefaultMagicHandler(caster, target, magic)
}

func (ms *MagicSystem) DefaultMagicHandler(caster, target interface{}, magic *TMagic) bool {
	return true
}

func (ms *MagicSystem) CheckSpellCondition(caster interface{}, magic *TMagic) bool {
	return true
}

func (ms *MagicSystem) GetMagicPower(magic *TUserMagic, level uint8) int {
	power := int(magic.Power)
	switch level {
	case 1:
		power = int(float32(power) * 1.0)
	case 2:
		power = int(float32(power) * 1.2)
	case 3:
		power = int(float32(power) * 1.5)
	}
	return power
}

type MagicEffectType int

const (
	MagicEffectNone MagicEffectType = iota
	MagicEffectDirectDamage
	MagicEffectAreaDamage
	MagicEffectHeal
	MagicEffectPoison
	MagicEffectBuff
	MagicEffectDebuff
	MagicEffectSummon
	MagicEffectTeleport
	MagicEffectCure
)

type MagicEffectHandler func(caster, target interface{}, magic *TUserMagic, power int) bool

func (ms *MagicSystem) RegisterDefaultHandlers() {
	ms.RegisterHandler(protocol.SKILL_FIREBALL, ms.handleFireBall)
	ms.RegisterHandler(protocol.SKILL_LIGHTNING, ms.handleLightning)
	ms.RegisterHandler(protocol.SKILL_HEALLING, ms.handleHealing)
	ms.RegisterHandler(protocol.SKILL_POISONING, ms.handlePoisoning)
	ms.RegisterHandler(protocol.SKILL_DEJIWONHO, ms.handleSummonSkeleton)
	ms.RegisterHandler(protocol.SKILL_REPULSION, ms.handleRepulsion)
	ms.RegisterHandler(protocol.SKILL_TAMMING, ms.handleTamming)
	ms.RegisterHandler(protocol.SKILL_SPACEMOVE, ms.handleSpaceMove)
}

func (ms *MagicSystem) handleFireBall(caster, target interface{}, magic *TUserMagic) bool {
	power := ms.GetMagicPower(magic, magic.Level)
	
	if target == nil {
		return false
	}
	
	return ms.applyDamage(target, power, EFFECT_FIRE)
}

func (ms *MagicSystem) handleLightning(caster, target interface{}, magic *TUserMagic) bool {
	power := ms.GetMagicPower(magic, magic.Level)
	
	if target == nil {
		return false
	}
	
	return ms.applyDamage(target, power, EFFECT_LIGHTNING)
}

func (ms *MagicSystem) handleHealing(caster, target interface{}, magic *TUserMagic) bool {
	power := ms.GetMagicPower(magic, magic.Level)
	
	if target == nil {
		return false
	}
	
	if p, ok := target.(*actor.Player); ok {
		p.AddHP(int32(power))
		return true
	}
	return false
}

func (ms *MagicSystem) handlePoisoning(caster, target interface{}, magic *TUserMagic) bool {
	power := ms.GetMagicPower(magic, magic.Level)
	
	if target == nil {
		return false
	}
	
	if m, ok := target.(*actor.TMonster); ok {
		m.Poisoned = true
		m.PoisonDamage = int32(power / 5)
		return true
	}
	
	if p, ok := target.(*actor.Player); ok {
		p.Poisoned = true
		p.Status |= protocol.POISON_DECHEALTH
		return true
	}
	
	return false
}

func (ms *MagicSystem) handleSummonSkeleton(caster, target interface{}, magic *TUserMagic) bool {
	c, ok := caster.(*actor.Player)
	if !ok {
		return false
	}
	
	monster := actor.NewMonster(0, "变异骷髅", 47)
	monster.Master = c
	monster.HP = 100
	monster.MaxHP = 100
	monster.DC = 5
	monster.MaxDC = 10
	monster.Speed = 500
	
	actor.GetMonsterManager().AddMonster(monster)
	
	return true
}

func (ms *MagicSystem) handleRepulsion(caster, target interface{}, magic *TUserMagic) bool {
	if target == nil {
		return false
	}
	
	return true
}

func (ms *MagicSystem) handleTamming(caster, target interface{}, magic *TUserMagic) bool {
	if target == nil {
		return false
	}
	
	if m, ok := target.(*actor.TMonster); ok {
		m.AIMode = actor.AI_MODE_FOLLOW
		m.Master = nil
		return true
	}
	
	return false
}

func (ms *MagicSystem) handleSpaceMove(caster, target interface{}, magic *TUserMagic) bool {
	_, ok := caster.(*actor.Player)
	if !ok {
		return false
	}
	
	return true
}

func MPow(magic *TUserMagic) int {
	return int(magic.Power) + rand.Intn(int(magic.MaxPower)-int(magic.Power))
}

func GetPower(nPower int, magic *TUserMagic) int {
	base := nPower / (int(magic.TrainLv) + 1)
	levelBonus := base * int(magic.Level)
	defPower := int(magic.DefPower) + rand.Intn(int(magic.DefMaxPower)-int(magic.DefPower))
	return levelBonus + defPower
}

func GetRPow(wInt uint16) uint16 {
	hi := int(uint16((uint32(wInt) >> 16) & 0xFFFF))
	lo := int(wInt & 0xFFFF)
	if hi > lo {
		return uint16(rand.Intn(hi-lo+1) + lo)
	}
	return wInt
}

func loWord(w uint32) uint16 {
	return uint16(w & 0xFFFF)
}

func hiWord(w uint32) uint16 {
	return uint16((w >> 16) & 0xFFFF)
}

func (ms *MagicSystem) handleBigHealing(caster interface{}, magic *TUserMagic, x, y int, power int) bool {
	return true
}

func (ms *MagicSystem) handlePushArround(caster interface{}, pushLevel int) int {
	return 0
}

func (ms *MagicSystem) handleTurnUndead(caster, target interface{}, magic *TUserMagic) bool {
	if target == nil {
		return false
	}
	
	if m, ok := target.(*actor.TMonster); ok {
		if m.Undead {
			damage := ms.GetMagicPower(magic, magic.Level) * 2
			ms.applyDamage(target, damage, EFFECT_NONE)
			return true
		}
	}
	return false
}

func (ms *MagicSystem) handleHolyCurtain(caster interface{}, damage, duration, x, y int) int {
	return 0
}

func (ms *MagicSystem) handleMakeTrap(caster interface{}, damage, duration, x, y int) int {
	return 0
}

func (ms *MagicSystem) handleCrash(caster interface{}, power int) bool {
	return true
}

func (ms *MagicSystem) handleEntrapment(caster, target interface{}, magic *TUserMagic) bool {
	return true
}

func (ms *MagicSystem) handleEnhancer(caster, target interface{}, power, duration int) bool {
	return true
}

func (ms *MagicSystem) handleLightBody(caster interface{}, power, duration int) bool {
	return true
}

func (ms *MagicSystem) handleHaste(caster interface{}, power, duration int) bool {
	return true
}

func (ms *MagicSystem) handleSwiftFeet(caster interface{}, power, duration int) bool {
	return true
}

func (ms *MagicSystem) handleBladeAvalanche(caster, target interface{}, magic *TUserMagic) bool {
	power := GetPower(int(magic.Power), magic)
	return ms.applyDamage(target, power, EFFECT_FIRE)
}

func (ms *MagicSystem) handleCresentSlash(caster, target interface{}, magic *TUserMagic) bool {
	power := GetPower(int(magic.Power), magic)
	return ms.applyDamage(target, power, EFFECT_NONE)
}

func (ms *MagicSystem) handleProtectionField(caster interface{}, power, sec int) bool {
	return true
}

func (ms *MagicSystem) handleRage(caster interface{}, power, sec int) bool {
	return true
}

func (ms *MagicSystem) handleGroupTransparent(caster interface{}, x, y, duration int) bool {
	return true
}

func (ms *MagicSystem) handleTammingMonster(caster, target interface{}, x, y, magicLevel int) bool {
	if target == nil {
		return false
	}
	
	if m, ok := target.(*actor.TMonster); ok {
		if int(m.Level) <= magicLevel*5+10 {
			m.AIMode = actor.AI_MODE_FOLLOW
			m.Master = nil
			return true
		}
	}
	return false
}

func (ms *MagicSystem) handleSpaceMove2(caster interface{}, level int) bool {
	return true
}

func (ms *MagicSystem) applyDamage(target interface{}, power int, effect SkillEffect) bool {
	switch t := target.(type) {
	case *actor.Player:
		t.AddHP(-int32(power))
		if !t.IsAlive() {
			return true
		}
	case *actor.TMonster:
		t.AddHP(-int32(power))
		if !t.Alive {
			return true
		}
	case *actor.Hero:
		t.AddHP(-int32(power))
	}
	
	return false
}
