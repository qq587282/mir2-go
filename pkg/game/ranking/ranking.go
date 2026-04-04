package ranking

import (
	"sort"
	"sync"
	"time"
)

type RankingType int

const (
	RankingTypeLevel RankingType = iota
	RankingTypePK
	RankingTypeRich
	RankingTypeGuild
	RankingTypeHero
	RankingTypeArena
)

type RankingItem struct {
	Rank       int
	PlayerID   int32
	Name       string
	Value      int64
	Job        byte
	Level      int
	GuildName  string
	UpdateTime time.Time
}

type Ranking struct {
	RankingType RankingType
	Title       string
	Items       []*RankingItem
	UpdateTime  time.Time
	MaxCount    int
}

type RankingManager struct {
	Rankings map[RankingType]*Ranking
	Mutex    sync.RWMutex
}

var DefaultRankingManager *RankingManager

func init() {
	DefaultRankingManager = NewRankingManager()
	DefaultRankingManager.InitRankings()
}

func NewRankingManager() *RankingManager {
	return &RankingManager{
		Rankings: make(map[RankingType]*Ranking),
	}
}

func (rm *RankingManager) InitRankings() {
	rm.Rankings[RankingTypeLevel] = &Ranking{
		RankingType: RankingTypeLevel,
		Title:       "等级排行榜",
		MaxCount:    100,
	}

	rm.Rankings[RankingTypePK] = &Ranking{
		RankingType: RankingTypePK,
		Title:       "PK排行榜",
		MaxCount:    100,
	}

	rm.Rankings[RankingTypeRich] = &Ranking{
		RankingType: RankingTypeRich,
		Title:       "财富排行榜",
		MaxCount:    100,
	}

	rm.Rankings[RankingTypeGuild] = &Ranking{
		RankingType: RankingTypeGuild,
		Title:       "行会排行榜",
		MaxCount:    50,
	}

	rm.Rankings[RankingTypeHero] = &Ranking{
		RankingType: RankingTypeHero,
		Title:       "英雄排行榜",
		MaxCount:    100,
	}

	rm.Rankings[RankingTypeArena] = &Ranking{
		RankingType: RankingTypeArena,
		Title:       "竞技场排行榜",
		MaxCount:    100,
	}
}

func (rm *RankingManager) GetRanking(rankingType RankingType) *Ranking {
	rm.Mutex.RLock()
	defer rm.Mutex.RUnlock()
	return rm.Rankings[rankingType]
}

func (rm *RankingManager) UpdateLevelRanking(playerID int32, name string, level int, job byte) {
	rm.updateRanking(RankingTypeLevel, playerID, name, int64(level), job, level, "")
}

func (rm *RankingManager) UpdatePKRanking(playerID int32, name string, pkPoint int) {
	rm.updateRanking(RankingTypePK, playerID, name, int64(pkPoint), 0, 0, "")
}

func (rm *RankingManager) UpdateRichRanking(playerID int32, name string, gold int32, job byte) {
	rm.updateRanking(RankingTypeRich, playerID, name, int64(gold), job, 0, "")
}

func (rm *RankingManager) UpdateGuildRanking(guildID int32, guildName string, level int, exp uint64) {
	rm.updateRanking(RankingTypeGuild, guildID, guildName, int64(exp), 0, level, "")
}

func (rm *RankingManager) UpdateHeroRanking(playerID int32, name string, heroName string, level int) {
	rm.updateRanking(RankingTypeHero, playerID, heroName, int64(level), 0, level, "")
}

func (rm *RankingManager) updateRanking(rankingType RankingType, id int32, name string, value int64, job byte, level int, guildName string) {
	rm.Mutex.Lock()
	defer rm.Mutex.Unlock()

	ranking, ok := rm.Rankings[rankingType]
	if !ok {
		return
	}

	found := false
	for i, item := range ranking.Items {
		if item.PlayerID == id {
			item.Value = value
			item.Level = level
			item.GuildName = guildName
			item.UpdateTime = time.Now()
			ranking.Items[i] = item
			found = true
			break
		}
	}

	if !found {
		ranking.Items = append(ranking.Items, &RankingItem{
			PlayerID:   id,
			Name:       name,
			Value:      value,
			Job:        job,
			Level:      level,
			GuildName:  guildName,
			UpdateTime: time.Now(),
		})
	}

	ranking.UpdateTime = time.Now()
	rm.sortRanking(ranking)
}

func (rm *RankingManager) sortRanking(ranking *Ranking) {
	switch ranking.RankingType {
	case RankingTypeLevel:
		sort.Slice(ranking.Items, func(i, j int) bool {
			if ranking.Items[i].Level != ranking.Items[j].Level {
				return ranking.Items[i].Level > ranking.Items[j].Level
			}
			return ranking.Items[i].Value > ranking.Items[j].Value
		})
	case RankingTypePK:
		sort.Slice(ranking.Items, func(i, j int) bool {
			return ranking.Items[i].Value > ranking.Items[j].Value
		})
	case RankingTypeRich:
		sort.Slice(ranking.Items, func(i, j int) bool {
			return ranking.Items[i].Value > ranking.Items[j].Value
		})
	case RankingTypeGuild:
		sort.Slice(ranking.Items, func(i, j int) bool {
			return ranking.Items[i].Value > ranking.Items[j].Value
		})
	case RankingTypeHero:
		sort.Slice(ranking.Items, func(i, j int) bool {
			if ranking.Items[i].Level != ranking.Items[j].Level {
				return ranking.Items[i].Level > ranking.Items[j].Level
			}
			return ranking.Items[i].Value > ranking.Items[j].Value
		})
	case RankingTypeArena:
		sort.Slice(ranking.Items, func(i, j int) bool {
			return ranking.Items[i].Value > ranking.Items[j].Value
		})
	}

	if len(ranking.Items) > ranking.MaxCount {
		ranking.Items = ranking.Items[:ranking.MaxCount]
	}

	for i := range ranking.Items {
		ranking.Items[i].Rank = i + 1
	}
}

func (rm *RankingManager) GetTopPlayer(rankingType RankingType, count int) []*RankingItem {
	rm.Mutex.RLock()
	defer rm.Mutex.RUnlock()

	ranking, ok := rm.Rankings[rankingType]
	if !ok {
		return nil
	}

	if count > len(ranking.Items) {
		count = len(ranking.Items)
	}

	result := make([]*RankingItem, count)
	copy(result, ranking.Items[:count])
	return result
}

func (rm *RankingManager) GetPlayerRank(rankingType RankingType, playerID int32) int {
	rm.Mutex.RLock()
	defer rm.Mutex.RUnlock()

	ranking, ok := rm.Rankings[rankingType]
	if !ok {
		return -1
	}

	for _, item := range ranking.Items {
		if item.PlayerID == playerID {
			return item.Rank
		}
	}

	return -1
}

func (rm *RankingManager) RefreshAllRankings() {
	rm.Mutex.Lock()
	defer rm.Mutex.Unlock()

	for _, ranking := range rm.Rankings {
		rm.sortRanking(ranking)
		ranking.UpdateTime = time.Now()
	}
}

func (rm *RankingManager) GetAllRankings() map[RankingType]*Ranking {
	rm.Mutex.RLock()
	defer rm.Mutex.RUnlock()

	result := make(map[RankingType]*Ranking)
	for k, v := range rm.Rankings {
		result[k] = v
	}
	return result
}

func GetRankingManager() *RankingManager {
	return DefaultRankingManager
}

type DailyRanking struct {
	Date     string
	Items    []*RankingItem
	MaxCount int
}

type DailyRankingManager struct {
	DailyRankings map[string]*DailyRanking
	Mutex         sync.RWMutex
}

var DefaultDailyRankingManager *DailyRankingManager

func init() {
	DefaultDailyRankingManager = NewDailyRankingManager()
}

func NewDailyRankingManager() *DailyRankingManager {
	return &DailyRankingManager{
		DailyRankings: make(map[string]*DailyRanking),
	}
}

func (drm *DailyRankingManager) GetTodayDate() string {
	return time.Now().Format("2006-01-02")
}

func (drm *DailyRankingManager) GetTodayRanking(rankingType RankingType) *DailyRanking {
	drm.Mutex.RLock()
	defer drm.Mutex.RUnlock()

	date := drm.GetTodayDate()
	key := date

	return drm.DailyRankings[key]
}

func (drm *DailyRankingManager) UpdateDailyRanking(rankingType RankingType, playerID int32, name string, value int64) {
	drm.Mutex.Lock()
	defer drm.Mutex.Unlock()

	date := drm.GetTodayDate()
	key := date

	if _, ok := drm.DailyRankings[key]; !ok {
		drm.DailyRankings[key] = &DailyRanking{
			Date:     date,
			Items:    make([]*RankingItem, 0),
			MaxCount: 100,
		}
	}

	ranking := drm.DailyRankings[key]

	found := false
	for _, item := range ranking.Items {
		if item.PlayerID == playerID {
			item.Value += value
			found = true
			break
		}
	}

	if !found {
		ranking.Items = append(ranking.Items, &RankingItem{
			PlayerID:   playerID,
			Name:       name,
			Value:      value,
			UpdateTime: time.Now(),
		})
	}

	sort.Slice(ranking.Items, func(i, j int) bool {
		return ranking.Items[i].Value > ranking.Items[j].Value
	})

	if len(ranking.Items) > ranking.MaxCount {
		ranking.Items = ranking.Items[:ranking.MaxCount]
	}
}

func GetDailyRankingManager() *DailyRankingManager {
	return DefaultDailyRankingManager
}
