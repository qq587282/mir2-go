package guild

import (
	"sync"
	"time"
)

type GuildRank int

const (
	GUILDRANK_LEADER     GuildRank = 1
	GUILDRANK_VICELEADER GuildRank = 2
	GUILDRANK_ELDER      GuildRank = 3
	GUILDRANK_MEMBER     GuildRank = 4
	GUILDRANK_NEWBIE     GuildRank = 5
)

type TGuildMember struct {
	Name       string
	Rank       GuildRank
	RankName   string
	JoinDate   time.Time
	Contribute int
	Offer      int
}

type TGuild struct {
	ID         int32
	Name       string
	Leader     string
	Members    map[string]*TGuildMember
	RankList   []string
	Notice     string
	GuildWar   map[string]time.Time
	AllyList   []int32
	EnemyList  []int32
	CreateTime time.Time
	TotalCount int
	MaxCount   int
	Level      int
	Exp        uint64
	Gold       int32
	CastleID   int32
}

type TGuildManager struct {
	Guilds       map[int32]*TGuild
	GuildsByName map[string]*TGuild
	Mutex        sync.RWMutex
}

var DefaultGuildManager *TGuildManager

func init() {
	DefaultGuildManager = NewGuildManager()
}

func NewGuildManager() *TGuildManager {
	return &TGuildManager{
		Guilds:       make(map[int32]*TGuild),
		GuildsByName: make(map[string]*TGuild),
	}
}

func (gm *TGuildManager) CreateGuild(leaderName, guildName string) *TGuild {
	gm.Mutex.Lock()
	defer gm.Mutex.Unlock()

	if _, exists := gm.GuildsByName[guildName]; exists {
		return nil
	}

	guild := &TGuild{
		ID:         int32(len(gm.Guilds) + 1),
		Name:       guildName,
		Leader:     leaderName,
		Members:    make(map[string]*TGuildMember),
		GuildWar:   make(map[string]time.Time),
		Notice:     "Welcome to " + guildName,
		CreateTime: time.Now(),
		MaxCount:   50,
		Level:      1,
	}

	guild.Members[leaderName] = &TGuildMember{
		Name:     leaderName,
		Rank:     GUILDRANK_LEADER,
		RankName: "掌门",
		JoinDate: time.Now(),
	}

	guild.RankList = append(guild.RankList, leaderName)

	gm.Guilds[guild.ID] = guild
	gm.GuildsByName[guildName] = guild

	return guild
}

func (gm *TGuildManager) GetGuild(id int32) *TGuild {
	gm.Mutex.RLock()
	defer gm.Mutex.RUnlock()
	return gm.Guilds[id]
}

func (gm *TGuildManager) GetGuildByName(name string) *TGuild {
	gm.Mutex.RLock()
	defer gm.Mutex.RUnlock()
	return gm.GuildsByName[name]
}

func (gm *TGuildManager) AddMember(guildID int32, memberName string, rank GuildRank, rankName string) bool {
	guild := gm.GetGuild(guildID)
	if guild == nil {
		return false
	}

	if _, exists := guild.Members[memberName]; exists {
		return false
	}

	guild.Members[memberName] = &TGuildMember{
		Name:     memberName,
		Rank:     rank,
		RankName: rankName,
		JoinDate: time.Now(),
	}

	guild.RankList = append(guild.RankList, memberName)
	guild.TotalCount++

	return true
}

func (gm *TGuildManager) DelMember(guildID int32, memberName string) bool {
	guild := gm.GetGuild(guildID)
	if guild == nil {
		return false
	}

	if _, exists := guild.Members[memberName]; !exists {
		return false
	}

	delete(guild.Members, memberName)

	for i, name := range guild.RankList {
		if name == memberName {
			guild.RankList = append(guild.RankList[:i], guild.RankList[i+1:]...)
			break
		}
	}

	guild.TotalCount--
	return true
}

func (gm *TGuildManager) SetMemberRank(guildID int32, memberName string, rank GuildRank, rankName string) bool {
	guild := gm.GetGuild(guildID)
	if guild == nil {
		return false
	}

	member, exists := guild.Members[memberName]
	if !exists {
		return false
	}

	member.Rank = rank
	member.RankName = rankName
	return true
}

func (gm *TGuildManager) AddAlly(guildID, allyID int32) bool {
	guild := gm.GetGuild(guildID)
	if guild == nil {
		return false
	}

	for _, id := range guild.AllyList {
		if id == allyID {
			return false
		}
	}

	guild.AllyList = append(guild.AllyList, allyID)
	return true
}

func (gm *TGuildManager) DelAlly(guildID, allyID int32) bool {
	guild := gm.GetGuild(guildID)
	if guild == nil {
		return false
	}

	for i, id := range guild.AllyList {
		if id == allyID {
			guild.AllyList = append(guild.AllyList[:i], guild.AllyList[i+1:]...)
			return true
		}
	}
	return false
}

func (gm *TGuildManager) AddEnemy(guildID, enemyID int32) bool {
	guild := gm.GetGuild(guildID)
	if guild == nil {
		return false
	}

	for _, id := range guild.EnemyList {
		if id == enemyID {
			return false
		}
	}

	guild.EnemyList = append(guild.EnemyList, enemyID)
	return true
}

func (gm *TGuildManager) StartGuildWar(guild1ID, guild2ID int32, duration time.Duration) bool {
	guild1 := gm.GetGuild(guild1ID)
	guild2 := gm.GetGuild(guild2ID)

	if guild1 == nil || guild2 == nil {
		return false
	}

	endTime := time.Now().Add(duration)
	guild1.GuildWar[guild2.Name] = endTime
	guild2.GuildWar[guild1.Name] = endTime

	return true
}

func (gm *TGuildManager) IsGuildWar(guild1ID, guild2ID int32) bool {
	guild1 := gm.GetGuild(guild1ID)
	guild2 := gm.GetGuild(guild2ID)

	if guild1 == nil || guild2 == nil {
		return false
	}

	if endTime, ok := guild1.GuildWar[guild2.Name]; ok {
		return time.Now().Before(endTime)
	}

	return false
}

func (gm *TGuildManager) GetAllGuilds() []*TGuild {
	gm.Mutex.RLock()
	defer gm.Mutex.RUnlock()

	result := make([]*TGuild, 0, len(gm.Guilds))
	for _, guild := range gm.Guilds {
		result = append(result, guild)
	}
	return result
}

func (gm *TGuildManager) GetGuildMemberList(guildID int32) []*TGuildMember {
	guild := gm.GetGuild(guildID)
	if guild == nil {
		return nil
	}

	result := make([]*TGuildMember, 0, len(guild.Members))
	for _, member := range guild.Members {
		result = append(result, member)
	}
	return result
}
