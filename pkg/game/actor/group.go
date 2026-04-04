package actor

import (
	"sync"
	"time"
)

type TGroup struct {
	GroupID    int32
	Leader     *Player
	Members    map[int32]*Player
	MemberLock sync.RWMutex
	CreateTime time.Time
	Exp        uint64
	Gold       int32
}

func NewGroup(leader *Player) *TGroup {
	return &TGroup{
		GroupID:    0,
		Leader:     leader,
		Members:    make(map[int32]*Player),
		CreateTime: time.Now(),
	}
}

func (g *TGroup) AddMember(player *Player) bool {
	g.MemberLock.Lock()
	defer g.MemberLock.Unlock()

	if len(g.Members) >= 9 {
		return false
	}

	if player.GroupID != 0 {
		return false
	}

	g.Members[player.ID] = player
	player.GroupID = g.GroupID

	return true
}

func (g *TGroup) DelMember(player *Player) bool {
	g.MemberLock.Lock()
	defer g.MemberLock.Unlock()

	if _, ok := g.Members[player.ID]; !ok {
		return false
	}

	delete(g.Members, player.ID)
	player.GroupID = 0

	if player == g.Leader && len(g.Members) > 0 {
		for _, m := range g.Members {
			g.Leader = m
			break
		}
	}

	return true
}

func (g *TGroup) GetMembers() []*Player {
	g.MemberLock.RLock()
	defer g.MemberLock.RUnlock()

	result := make([]*Player, 0, len(g.Members))
	for _, m := range g.Members {
		result = append(result, m)
	}
	return result
}

func (g *TGroup) IsLeader(player *Player) bool {
	return player == g.Leader
}

func (g *TGroup) IsMember(player *Player) bool {
	g.MemberLock.RLock()
	defer g.MemberLock.RUnlock()

	_, ok := g.Members[player.ID]
	return ok
}

func (g *TGroup) Broadcast(ident uint16, data []byte) {
	g.MemberLock.RLock()
	defer g.MemberLock.RUnlock()

	for _, m := range g.Members {
		if m.Session != nil {
		}
	}
}

func (g *TGroup) ShareExp(exp uint64) {
	g.MemberLock.RLock()
	members := make([]*Player, 0, len(g.Members))
	for _, m := range g.Members {
		members = append(members, m)
	}
	g.MemberLock.RUnlock()

	sharedExp := exp / uint64(len(members))

	for _, m := range members {
		m.AddExp(int64(sharedExp))
	}
}

type GroupManager struct {
	Groups map[int32]*TGroup
	Mutex  sync.RWMutex
	NextID int32
}

var DefaultGroupManager *GroupManager

func init() {
	DefaultGroupManager = NewGroupManager()
}

func NewGroupManager() *GroupManager {
	return &GroupManager{
		Groups: make(map[int32]*TGroup),
		NextID: 1,
	}
}

func (gm *GroupManager) GetNextID() int32 {
	id := gm.NextID
	gm.NextID++
	return id
}

func (gm *GroupManager) CreateGroup(leader *Player) *TGroup {
	gm.Mutex.Lock()
	defer gm.Mutex.Unlock()

	if leader.GroupID != 0 {
		return nil
	}

	group := NewGroup(leader)
	group.GroupID = gm.GetNextID()
	group.AddMember(leader)

	gm.Groups[group.GroupID] = group

	return group
}

func (gm *GroupManager) DeleteGroup(groupID int32) bool {
	gm.Mutex.Lock()
	defer gm.Mutex.Unlock()

	group, ok := gm.Groups[groupID]
	if !ok {
		return false
	}

	members := group.GetMembers()
	for _, m := range members {
		m.GroupID = 0
	}

	delete(gm.Groups, groupID)
	return true
}

func (gm *GroupManager) GetGroup(groupID int32) *TGroup {
	gm.Mutex.RLock()
	defer gm.Mutex.RUnlock()
	return gm.Groups[groupID]
}

func (gm *GroupManager) GetPlayerGroup(player *Player) *TGroup {
	gm.Mutex.RLock()
	defer gm.Mutex.RUnlock()

	for _, group := range gm.Groups {
		if group.IsMember(player) {
			return group
		}
	}
	return nil
}

func (gm *GroupManager) InviteToGroup(leader, target *Player) bool {
	if leader.GroupID != 0 {
		group := gm.GetGroup(leader.GroupID)
		if group == nil {
			return false
		}
		return group.AddMember(target)
	}

	group := gm.CreateGroup(leader)
	if group == nil {
		return false
	}

	return group.AddMember(target)
}

func (gm *GroupManager) LeaveGroup(player *Player) bool {
	gm.Mutex.RLock()
	var group *TGroup
	for _, g := range gm.Groups {
		if g.IsMember(player) {
			group = g
			break
		}
	}
	gm.Mutex.RUnlock()

	if group == nil {
		return false
	}

	group.DelMember(player)

	if group.Leader == player && len(group.Members) == 0 {
		gm.DeleteGroup(group.GroupID)
	}

	return true
}

func GetGroupManager() *GroupManager {
	return DefaultGroupManager
}
