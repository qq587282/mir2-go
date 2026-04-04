package quest

import (
	"sync"
	"time"
)

type QuestState int

const (
	QuestStateNone QuestState = iota
	QuestStateAccepted
	QuestStateCompleted
	QuestStateFailed
)

type QuestType int

const (
	QuestTypeMain QuestType = iota
	QuestTypeBranch
	QuestTypeDaily
	QuestTypeEvent
)

type QuestTarget struct {
	TargetType string
	TargetID   int32
	TargetName string
	Count      int
	Current    int
}

type TQuest struct {
	QuestID   int32
	Name      string
	Desc      string
	QuestType QuestType
	MinLevel  int
	MaxLevel  int
	JobLimit  byte
	PrevQuest int32
	NextQuest int32

	GoldReward  int32
	ExpReward   uint64
	ItemRewards []string

	Targets        []QuestTarget
	TimeLimit      time.Duration
	Repeatable     bool
	RepeatInterval time.Duration
}

type TPlayerQuest struct {
	QuestID      int32
	State        QuestState
	AcceptTime   time.Time
	CompleteTime time.Time
	Targets      []QuestTarget
}

type QuestManager struct {
	Quests map[int32]*TQuest
	Mutex  sync.RWMutex
	NextID int32
}

var DefaultQuestManager *QuestManager

func init() {
	DefaultQuestManager = NewQuestManager()
	DefaultQuestManager.LoadDefaultQuests()
}

func NewQuestManager() *QuestManager {
	return &QuestManager{
		Quests: make(map[int32]*TQuest),
		NextID: 1,
	}
}

func (qm *QuestManager) LoadDefaultQuests() {
	qm.Quests[1] = &TQuest{
		QuestID:    1,
		Name:       "新手任务",
		Desc:       "与比奇省的新手引导员对话，了解游戏基础",
		QuestType:  QuestTypeMain,
		MinLevel:   1,
		MaxLevel:   5,
		GoldReward: 100,
		ExpReward:  50,
	}

	qm.Quests[2] = &TQuest{
		QuestID:    2,
		Name:       "初级装备",
		Desc:       "击败鸡怪物，获得掉落的新手装备",
		QuestType:  QuestTypeMain,
		MinLevel:   1,
		MaxLevel:   7,
		GoldReward: 200,
		ExpReward:  100,
		Targets: []QuestTarget{
			{TargetType: "kill", TargetName: "鸡", Count: 5, Current: 0},
		},
	}

	qm.Quests[3] = &TQuest{
		QuestID:    3,
		Name:       "森林探险",
		Desc:       "探索比奇森林，找到并击败森林里的怪物",
		QuestType:  QuestTypeBranch,
		MinLevel:   5,
		MaxLevel:   15,
		GoldReward: 500,
		ExpReward:  300,
		Targets: []QuestTarget{
			{TargetType: "kill", TargetName: "森林雪人", Count: 10, Current: 0},
		},
	}

	qm.Quests[4] = &TQuest{
		QuestID:        4,
		Name:           "收集材料",
		Desc:           "为铁匠收集特定的矿石材料",
		QuestType:      QuestTypeDaily,
		MinLevel:       10,
		MaxLevel:       50,
		GoldReward:     1000,
		ExpReward:      500,
		ItemRewards:    []string{"铁矿"},
		Repeatable:     true,
		RepeatInterval: time.Hour * 24,
	}

	qm.Quests[5] = &TQuest{
		QuestID:    5,
		Name:       "沃玛寺庙",
		Desc:       "进入沃玛寺庙，击败沃玛教主",
		QuestType:  QuestTypeMain,
		MinLevel:   20,
		MaxLevel:   30,
		GoldReward: 5000,
		ExpReward:  2000,
		Targets: []QuestTarget{
			{TargetType: "kill", TargetName: "沃玛教主", Count: 1, Current: 0},
		},
	}

	qm.Quests[6] = &TQuest{
		QuestID:    6,
		Name:       "每日挑战",
		Desc:       "完成每日击杀任务",
		QuestType:  QuestTypeDaily,
		MinLevel:   1,
		MaxLevel:   60,
		GoldReward: 2000,
		ExpReward:  1000,
		Targets: []QuestTarget{
			{TargetType: "kill", TargetName: "怪物", Count: 50, Current: 0},
		},
		Repeatable:     true,
		RepeatInterval: time.Hour * 24,
	}
}

func (qm *QuestManager) GetQuest(questID int32) *TQuest {
	qm.Mutex.RLock()
	defer qm.Mutex.RUnlock()
	return qm.Quests[questID]
}

func (qm *QuestManager) GetQuestsByLevel(level int) []*TQuest {
	qm.Mutex.RLock()
	defer qm.Mutex.RUnlock()

	var result []*TQuest
	for _, q := range qm.Quests {
		if level >= q.MinLevel && level <= q.MaxLevel {
			result = append(result, q)
		}
	}
	return result
}

func (qm *QuestManager) GetQuestsByType(questType QuestType) []*TQuest {
	qm.Mutex.RLock()
	defer qm.Mutex.RUnlock()

	var result []*TQuest
	for _, q := range qm.Quests {
		if q.QuestType == questType {
			result = append(result, q)
		}
	}
	return result
}

func (qm *QuestManager) AddQuest(quest *TQuest) {
	qm.Mutex.Lock()
	defer qm.Mutex.Unlock()
	quest.QuestID = qm.NextID
	qm.NextID++
	qm.Quests[quest.QuestID] = quest
}

func GetQuestManager() *QuestManager {
	return DefaultQuestManager
}

type PlayerQuestManager struct {
	PlayerQuests map[int32]map[int32]*TPlayerQuest
	Mutex        sync.RWMutex
}

var DefaultPlayerQuestManager *PlayerQuestManager

func init() {
	DefaultPlayerQuestManager = NewPlayerQuestManager()
}

func NewPlayerQuestManager() *PlayerQuestManager {
	return &PlayerQuestManager{
		PlayerQuests: make(map[int32]map[int32]*TPlayerQuest),
	}
}

func (pqm *PlayerQuestManager) AcceptQuest(playerID, questID int32) *TPlayerQuest {
	pqm.Mutex.Lock()
	defer pqm.Mutex.Unlock()

	if _, ok := pqm.PlayerQuests[playerID]; !ok {
		pqm.PlayerQuests[playerID] = make(map[int32]*TPlayerQuest)
	}

	quest := GetQuestManager().GetQuest(questID)
	if quest == nil {
		return nil
	}

	if _, ok := pqm.PlayerQuests[playerID][questID]; ok {
		return nil
	}

	targets := make([]QuestTarget, len(quest.Targets))
	copy(targets, quest.Targets)

	pq := &TPlayerQuest{
		QuestID:    questID,
		State:      QuestStateAccepted,
		AcceptTime: time.Now(),
		Targets:    targets,
	}

	pqm.PlayerQuests[playerID][questID] = pq
	return pq
}

func (pqm *PlayerQuestManager) CompleteQuest(playerID, questID int32) bool {
	pqm.Mutex.Lock()
	defer pqm.Mutex.Unlock()

	playerQuests, ok := pqm.PlayerQuests[playerID]
	if !ok {
		return false
	}

	pq, ok := playerQuests[questID]
	if !ok || pq.State != QuestStateAccepted {
		return false
	}

	for _, target := range pq.Targets {
		if target.Current < target.Count {
			return false
		}
	}

	pq.State = QuestStateCompleted
	pq.CompleteTime = time.Now()
	return true
}

func (pqm *PlayerQuestManager) UpdateTarget(playerID, questID int32, targetType, targetName string) {
	pqm.Mutex.Lock()
	defer pqm.Mutex.Unlock()

	playerQuests, ok := pqm.PlayerQuests[playerID]
	if !ok {
		return
	}

	pq, ok := playerQuests[questID]
	if !ok || pq.State != QuestStateAccepted {
		return
	}

	for i := range pq.Targets {
		if pq.Targets[i].TargetType == targetType && pq.Targets[i].TargetName == targetName {
			pq.Targets[i].Current++
			break
		}
	}
}

func (pqm *PlayerQuestManager) GetPlayerQuest(playerID, questID int32) *TPlayerQuest {
	pqm.Mutex.RLock()
	defer pqm.Mutex.RUnlock()

	if playerQuests, ok := pqm.PlayerQuests[playerID]; ok {
		return playerQuests[questID]
	}
	return nil
}

func (pqm *PlayerQuestManager) GetPlayerQuests(playerID int32) []*TPlayerQuest {
	pqm.Mutex.RLock()
	defer pqm.Mutex.RUnlock()

	var result []*TPlayerQuest
	if playerQuests, ok := pqm.PlayerQuests[playerID]; ok {
		for _, pq := range playerQuests {
			result = append(result, pq)
		}
	}
	return result
}

func (pqm *PlayerQuestManager) AbandonQuest(playerID, questID int32) bool {
	pqm.Mutex.Lock()
	defer pqm.Mutex.Unlock()

	playerQuests, ok := pqm.PlayerQuests[playerID]
	if !ok {
		return false
	}

	if _, ok := playerQuests[questID]; !ok {
		return false
	}

	delete(playerQuests, questID)
	return true
}

func GetPlayerQuestManager() *PlayerQuestManager {
	return DefaultPlayerQuestManager
}
