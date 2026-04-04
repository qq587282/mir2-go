package game

import (
	"strings"
	"sync"

	"github.com/mir2go/mir2/pkg/game/actor"
	gamemap "github.com/mir2go/mir2/pkg/game/map"
)

type CommandPermission int

const (
	PermNone   CommandPermission = 0
	PermPlayer CommandPermission = 3
	PermGM     CommandPermission = 10
	PermAdmin  CommandPermission = 10
)

type GameCommand struct {
	Name        string
	Permission  CommandPermission
	MinPerm     int
	MaxPerm     int
	Description string
	Handler     func(player *actor.Player, params []string) bool
}

type CommandManager struct {
	Commands map[string]*GameCommand
	Mutex    sync.RWMutex
}

var DefaultCommandManager *CommandManager

func init() {
	DefaultCommandManager = NewCommandManager()
	DefaultCommandManager.RegisterCommands()
}

func NewCommandManager() *CommandManager {
	return &CommandManager{
		Commands: make(map[string]*GameCommand),
	}
}

func (cm *CommandManager) Register(cmd *GameCommand) {
	cm.Mutex.Lock()
	defer cm.Mutex.Unlock()
	cmd.Name = strings.ToUpper(cmd.Name)
	cm.Commands[cmd.Name] = cmd
}

func (cm *CommandManager) GetCommand(name string) *GameCommand {
	cm.Mutex.RLock()
	defer cm.Mutex.RUnlock()
	name = strings.ToUpper(name)
	return cm.Commands[name]
}

func (cm *CommandManager) Execute(player *actor.Player, cmdStr string) bool {
	if player == nil {
		return false
	}

	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return false
	}

	cmdName := strings.ToUpper(parts[0])
	params := parts[1:]

	cmd := cm.GetCommand(cmdName)
	if cmd == nil {
		player.SendMessage("未知命令: " + cmdName)
		return false
	}

	if player.Permission < cmd.MinPerm || player.Permission > cmd.MaxPerm {
		player.SendMessage("权限不足")
		return false
	}

	return cmd.Handler(player, params)
}

func (cm *CommandManager) RegisterCommands() {
	cm.Register(&GameCommand{
		Name:        "DATE",
		Permission:  PermNone,
		MinPerm:     0,
		MaxPerm:     10,
		Description: "显示系统日期",
		Handler:     cmdDate,
	})

	cm.Register(&GameCommand{
		Name:        "PRVMSG",
		Permission:  PermNone,
		MinPerm:     0,
		MaxPerm:     10,
		Description: "禁止私聊",
		Handler:     cmdPrvMsg,
	})

	cm.Register(&GameCommand{
		Name:        "ALLOWMSG",
		Permission:  PermNone,
		MinPerm:     0,
		MaxPerm:     10,
		Description: "允许接收消息",
		Handler:     cmdAllowMsg,
	})

	cm.Register(&GameCommand{
		Name:        "LETSHOUT",
		Permission:  PermNone,
		MinPerm:     0,
		MaxPerm:     10,
		Description: "允许喊话",
		Handler:     cmdLetShout,
	})

	cm.Register(&GameCommand{
		Name:        "LETTRADE",
		Permission:  PermNone,
		MinPerm:     0,
		MaxPerm:     10,
		Description: "允许交易",
		Handler:     cmdLetTrade,
	})

	cm.Register(&GameCommand{
		Name:        "LETGUILD",
		Permission:  PermNone,
		MinPerm:     0,
		MaxPerm:     10,
		Description: "允许行会",
		Handler:     cmdLetGuild,
	})

	cm.Register(&GameCommand{
		Name:        "ENDGUILD",
		Permission:  PermNone,
		MinPerm:     0,
		MaxPerm:     10,
		Description: "解散行会",
		Handler:     cmdEndGuild,
	})

	cm.Register(&GameCommand{
		Name:        "MOVE",
		Permission:  PermPlayer,
		MinPerm:     3,
		MaxPerm:     6,
		Description: "移动到指定地图",
		Handler:     cmdMove,
	})

	cm.Register(&GameCommand{
		Name:        "POSITIONMOVE",
		Permission:  PermPlayer,
		MinPerm:     3,
		MaxPerm:     6,
		Description: "移动到指定坐标",
		Handler:     cmdPositionMove,
	})

	cm.Register(&GameCommand{
		Name:        "INFO",
		Permission:  PermPlayer,
		MinPerm:     3,
		MaxPerm:     10,
		Description: "查看玩家信息",
		Handler:     cmdInfo,
	})

	cm.Register(&GameCommand{
		Name:        "MAP",
		Permission:  PermPlayer,
		MinPerm:     3,
		MaxPerm:     10,
		Description: "查看当前地图信息",
		Handler:     cmdMap,
	})

	cm.Register(&GameCommand{
		Name:        "KICK",
		Permission:  PermGM,
		MinPerm:     10,
		MaxPerm:     10,
		Description: "踢出玩家",
		Handler:     cmdKick,
	})

	cm.Register(&GameCommand{
		Name:        "MAPMOVE",
		Permission:  PermGM,
		MinPerm:     10,
		MaxPerm:     10,
		Description: "GM移动命令",
		Handler:     cmdMapMove,
	})

	cm.Register(&GameCommand{
		Name:        "RECALL",
		Permission:  PermGM,
		MinPerm:     10,
		MaxPerm:     10,
		Description: "召唤玩家",
		Handler:     cmdRecall,
	})

	cm.Register(&GameCommand{
		Name:        "GAMEMASTER",
		Permission:  PermGM,
		MinPerm:     10,
		MaxPerm:     10,
		Description: "进入游戏管理模式",
		Handler:     cmdGameMaster,
	})

	cm.Register(&GameCommand{
		Name:        "LEVEL",
		Permission:  PermGM,
		MinPerm:     10,
		MaxPerm:     10,
		Description: "调整玩家等级",
		Handler:     cmdLevel,
	})

	cm.Register(&GameCommand{
		Name:        "MOB",
		Permission:  PermGM,
		MinPerm:     10,
		MaxPerm:     10,
		Description: "在当前位置生成怪物",
		Handler:     cmdMob,
	})

	cm.Register(&GameCommand{
		Name:        "DELETEITEM",
		Permission:  PermGM,
		MinPerm:     10,
		MaxPerm:     10,
		Description: "删除玩家物品",
		Handler:     cmdDeleteItem,
	})

	cm.Register(&GameCommand{
		Name:        "CHANGEJOB",
		Permission:  PermGM,
		MinPerm:     10,
		MaxPerm:     10,
		Description: "改变玩家职业",
		Handler:     cmdChangeJob,
	})

	cm.Register(&GameCommand{
		Name:        "CHANGELUCK",
		Permission:  PermGM,
		MinPerm:     10,
		MaxPerm:     10,
		Description: "改变玩家幸运值",
		Handler:     cmdChangeLuck,
	})
}

func cmdDate(p *actor.Player, params []string) bool {
	return true
}

func cmdPrvMsg(p *actor.Player, params []string) bool {
	return true
}

func cmdAllowMsg(p *actor.Player, params []string) bool {
	if len(params) > 0 {
		p.AllowMsg = params[0] == "1"
	}
	return true
}

func cmdLetShout(p *actor.Player, params []string) bool {
	return true
}

func cmdLetTrade(p *actor.Player, params []string) bool {
	return true
}

func cmdLetGuild(p *actor.Player, params []string) bool {
	return true
}

func cmdEndGuild(p *actor.Player, params []string) bool {
	return true
}

func cmdMove(p *actor.Player, params []string) bool {
	if len(params) < 1 {
		p.SendMessage("用法: @Move 地图名")
		return false
	}
	mapName := params[0]
	gm := gamemap.GetMapManager()
	gameMap := gm.GetMap(mapName)
	if gameMap == nil {
		p.SendMessage("地图不存在")
		return false
	}
	p.MapName = mapName
	p.SendMessage("已移动到地图: " + mapName)
	return true
}

func cmdPositionMove(p *actor.Player, params []string) bool {
	if len(params) < 3 {
		p.SendMessage("用法: @PositionMove 地图名 X Y")
		return false
	}
	mapName := params[0]
	x := parseInt(params[1])
	y := parseInt(params[2])
	gm := gamemap.GetMapManager()
	gameMap := gm.GetMap(mapName)
	if gameMap == nil {
		p.SendMessage("地图不存在")
		return false
	}
	if !gameMap.CanWalk(x, y) {
		p.SendMessage("该位置无法行走")
		return false
	}
	p.MapName = mapName
	p.X = x
	p.Y = y
	p.SendMessage("已移动到: " + mapName + "(" + params[1] + "," + params[2] + ")")
	return true
}

func cmdInfo(p *actor.Player, params []string) bool {
	if len(params) < 1 {
		p.SendMessage("用法: @Info 玩家名")
		return false
	}
	targetName := params[0]
	target := actor.FindPlayerByName(targetName)
	if target == nil {
		p.SendMessage("玩家不在线")
		return false
	}
	p.SendMessage("玩家: " + target.Name + " 等级: " + itoa(int(target.Level)))
	return true
}

func cmdMap(p *actor.Player, params []string) bool {
	p.SendMessage("当前地图: " + p.MapName + " 坐标: (" + itoa(p.X) + "," + itoa(p.Y) + ")")
	return true
}

func cmdKick(p *actor.Player, params []string) bool {
	if len(params) < 1 {
		p.SendMessage("用法: @Kick 玩家名")
		return false
	}
	targetName := params[0]
	target := actor.FindPlayerByName(targetName)
	if target == nil {
		p.SendMessage("玩家不在线")
		return false
	}
	target.Kick()
	p.SendMessage("已踢出玩家: " + targetName)
	return true
}

func cmdMapMove(p *actor.Player, params []string) bool {
	return cmdPositionMove(p, params)
}

func cmdRecall(p *actor.Player, params []string) bool {
	if len(params) < 1 {
		p.SendMessage("用法: @Recall 玩家名")
		return false
	}
	targetName := params[0]
	target := actor.FindPlayerByName(targetName)
	if target == nil {
		p.SendMessage("玩家不在线")
		return false
	}
	target.MapName = p.MapName
	target.X = p.X
	target.Y = p.Y
	target.SendMessage("你被召唤到: " + p.Name)
	return true
}

func cmdGameMaster(p *actor.Player, params []string) bool {
	p.GameMaster = !p.GameMaster
	if p.GameMaster {
		p.SendMessage("已进入游戏管理模式")
	} else {
		p.SendMessage("已退出游戏管理模式")
	}
	return true
}

func cmdLevel(p *actor.Player, params []string) bool {
	if len(params) < 1 {
		p.SendMessage("用法: @Level 等级")
		return false
	}
	level := parseInt(params[0])
	if level < 1 || level > 255 {
		p.SendMessage("等级范围: 1-255")
		return false
	}
	p.Level = int32(level)
	p.SendMessage("等级已设置为: " + params[0])
	return true
}

func cmdMob(p *actor.Player, params []string) bool {
	if len(params) < 1 {
		p.SendMessage("用法: @Mob 怪物名称")
		return false
	}
	monsterName := params[0]
	monster := actor.NewMonster(0, monsterName, 0)
	if monster == nil {
		p.SendMessage("怪物不存在")
		return false
	}
	monster.MapName = p.MapName
	monster.X = p.X
	monster.Y = p.Y
	gm := gamemap.GetMapManager()
	gameMap := gm.GetMap(p.MapName)
	if gameMap != nil {
		gameMap.AddMonster(monster)
	}
	p.SendMessage("已生成怪物: " + monsterName)
	return true
}

func cmdDeleteItem(p *actor.Player, params []string) bool {
	if len(params) < 1 {
		p.SendMessage("用法: @DeleteItem 物品名称")
		return false
	}
	itemName := params[0]
	removed := false
	for i := len(p.Items) - 1; i >= 0; i-- {
		if p.Items[i] != nil && p.Items[i].Name == itemName {
			p.Items = append(p.Items[:i], p.Items[i+1:]...)
			removed = true
			break
		}
	}
	if removed {
		p.SendMessage("已删除物品: " + itemName)
	} else {
		p.SendMessage("物品不存在或背包中未找到")
	}
	return true
}

func cmdChangeJob(p *actor.Player, params []string) bool {
	if len(params) < 1 {
		p.SendMessage("用法: @ChangeJob 职业(Warrior/Wizard/Taoist)")
		return false
	}
	jobName := strings.ToLower(params[0])
	switch jobName {
	case "warrior":
		p.Job = actor.JobWarr
	case "wizard":
		p.Job = actor.JobWizard
	case "taoist":
		p.Job = actor.JobTaos
	default:
		p.SendMessage("未知职业: " + params[0])
		return false
	}
	p.SendMessage("职业已改为: " + params[0])
	return true
}

func cmdChangeLuck(p *actor.Player, params []string) bool {
	if len(params) < 1 {
		p.SendMessage("用法: @ChangeLuck 幸运值")
		return false
	}
	luck := int8(parseInt(params[0]))
	p.AddAbility.Luck = luck
	p.SendMessage("幸运值已设置为: " + params[0])
	return true
}

func GetCommandManager() *CommandManager {
	return DefaultCommandManager
}

func parseInt(s string) int {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	var result []byte
	for n > 0 {
		result = append([]byte{'0' + byte(n%10)}, result...)
		n /= 10
	}
	if negative {
		result = append([]byte{'-'}, result...)
	}
	return string(result)
}
