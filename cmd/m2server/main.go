package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/mir2go/mir2/pkg/config"
	"github.com/mir2go/mir2/pkg/db"
	"github.com/mir2go/mir2/pkg/game/actor"
	"github.com/mir2go/mir2/pkg/game/event"
	"github.com/mir2go/mir2/pkg/game/item"
	"github.com/mir2go/mir2/pkg/game/map"
	"github.com/mir2go/mir2/pkg/network"
	"github.com/mir2go/mir2/pkg/protocol"
)

type SessionData struct {
	Player      *actor.Player
	Account     string
	LoginTime   time.Time
	PlayerMutex sync.Mutex
}

var (
	logger       *zap.Logger
	server       *network.GateServer
	database     db.Database
	configFile   string
	cfg          *config.ServerConfig
	MapMgr       *gamemap.MapManager
	PlayerMgr    *actor.PlayerManager
	Sessions     map[int32]*SessionData
	SessionsLock sync.RWMutex

	SessionMap    map[int]string
	SessionMapLock sync.RWMutex
)

func init() {
	flag.StringVar(&configFile, "config", "config.yaml", "config file path")
}

func main() {
	flag.Parse()

	zapLogger, _ := zap.NewProduction()
	logger = zapLogger
	defer logger.Sync()

	logger.Info("Starting M2Server...")

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		logger.Warn("Failed to load config, using default", zap.Error(err))
		cfg = config.GetDefaultConfig()
	}

	MapMgr = gamemap.NewMapManager()
	
	loadDefaultMaps()
	
	event.InitEventManager()
	
	PlayerMgr = actor.NewPlayerManager()
	Sessions = make(map[int32]*SessionData)
	SessionMap = make(map[int]string)

	if cfg.DBServer.Enable {
		dbCfg := db.DBConfig{
			Type:     cfg.DBServer.Type,
			Host:     cfg.DBServer.IP,
			Port:     cfg.DBServer.Port,
			User:     cfg.DBServer.User,
			Password: cfg.DBServer.Password,
			Database: cfg.DBServer.Database,
		}
		database, err = db.NewDatabase(dbCfg)
		if err != nil {
			logger.Warn("Failed to connect database, using memory storage", zap.Error(err))
		} else {
			defer database.Close()
		}
	}

	logger.Info("Config loaded",
		zap.String("servername", cfg.ServerName),
		zap.Int("port", cfg.M2Server.Port),
	)

	addr := fmt.Sprintf("%s:%d", cfg.M2Server.IP, cfg.M2Server.Port)
	server = network.NewGateServer(addr, logger)

	server.MaxSessions = 10000
	server.OnConnect = onConnect
	server.OnDisconnect = onDisconnect
	server.OnMessage = onMessage

	if err := server.Start(); err != nil {
		logger.Error("Failed to start server", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("M2Server started successfully",
		zap.String("addr", server.Addr),
		zap.Int("sessions", server.GetSessionCount()),
	)

	go func() {
		time.Sleep(2 * time.Second)
		handleLoginServerConnection(cfg)
	}()

	go gameLoop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down M2Server...")
	server.Stop()
	logger.Info("M2Server stopped")
}

func gameLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		<-ticker.C
		processGameTick()
	}
}

func loadDefaultMaps() {
	defaultMaps := []string{"0", "3", "2", "1", "0", "6", "5", "4"}
	for _, mapName := range defaultMaps {
		_, err := MapMgr.LoadMap(mapName, "maps/"+mapName+".map")
		if err != nil {
			logger.Warn("Failed to load map", zap.String("map", mapName), zap.Error(err))
		} else {
			logger.Info("Loaded map", zap.String("map", mapName))
		}
	}
}

func processGameTick() {
	players := PlayerMgr.GetAllPlayers()
	for _, p := range players {
		if p.MapName != "" {
			m := MapMgr.GetMap(p.MapName)
			if m == nil {
				continue
			}
		}
	}
}

func onConnect(sess *network.GateSession) {
	logger.Info("Client connected",
		zap.String("addr", sess.Addr),
		zap.Int32("session", sess.SessionID),
	)

	SessionsLock.Lock()
	Sessions[sess.SessionID] = &SessionData{
		LoginTime: time.Now(),
	}
	SessionsLock.Unlock()
}

func onDisconnect(sess *network.GateSession) {
	logger.Info("Client disconnected",
		zap.String("addr", sess.Addr),
		zap.Int32("session", sess.SessionID),
	)

	SessionsLock.Lock()
	if sd, ok := Sessions[sess.SessionID]; ok && sd.Player != nil {
		savePlayerData(sd.Player)
		PlayerMgr.DelPlayer(sd.Player.ID)
	}
	delete(Sessions, sess.SessionID)
	SessionsLock.Unlock()
}

func onMessage(sess *network.GateSession, data []byte) {
	logger.Debug("onMessage received",
		zap.Int32("session", sess.SessionID),
		zap.ByteString("data", data),
	)

	if len(data) < 20 {
		logger.Debug("Message too short for header", zap.Int("len", len(data)))
		return
	}

	msgHeader := protocol.UnpackMsgHeader(data)
	if msgHeader != nil && msgHeader.Code == protocol.RUNGATECODE {
		logger.Info("Received RunGate message",
			zap.Int32("session", sess.SessionID),
			zap.Uint16("ident", msgHeader.Ident),
			zap.Int32("length", msgHeader.Length),
			zap.Int32("socket", msgHeader.Socket),
		)

		switch msgHeader.Ident {
		case protocol.GM_OPEN:
			handleGM_OPEN(sess, msgHeader, data[20:])
		case protocol.GM_CLOSE:
			handleGM_CLOSE(sess, msgHeader)
		case protocol.GM_CHECKCLIENT:
			logger.Debug("GM_CHECKCLIENT received")
		case protocol.GM_DATA:
			handleGM_DATA(sess, msgHeader, data[20:])
		default:
			logger.Warn("Unknown ident", zap.Uint16("ident", msgHeader.Ident))
		}
		return
	}

	msg := protocol.UnpackDefaultMessage(data)
	if msg == nil {
		return
	}

	body := data[14:]

	logger.Info("Received direct message",
		zap.Int32("session", sess.SessionID),
		zap.Uint16("ident", msg.Ident),
	)

	handleMessage(sess, msg.Ident, msg.Param, msg.Tag, msg.Series, msg.Recog, body)
}

func handleGM_OPEN(sess *network.GateSession, header *protocol.TMsgHeader, data []byte) {
	ipAddr := string(data)
	if len(data) > 0 && data[len(data)-1] == 0 {
		ipAddr = string(data[:len(data)-1])
	}

	logger.Info("GM_OPEN received",
		zap.Int32("session", sess.SessionID),
		zap.Int32("socket", header.Socket),
		zap.String("ip", ipAddr),
	)

	userIdx := registerUser(sess.SessionID, ipAddr)

	sendNewUserMsg(sess, header.Socket, uint16(header.GSocketIdx), userIdx+1)
	
	sendServerVersion(sess)
}

func handleGM_CLOSE(sess *network.GateSession, header *protocol.TMsgHeader) {
	logger.Info("GM_CLOSE received",
		zap.Int32("session", sess.SessionID),
		zap.Int32("socket", header.Socket),
	)

	unregisterUser(header.Socket)
}

func handleGM_DATA(sess *network.GateSession, header *protocol.TMsgHeader, data []byte) {
	if len(data) < 14 {
		logger.Debug("Client data too short", zap.Int("len", len(data)))
		return
	}

	msg := protocol.UnpackDefaultMessage(data)
	if msg == nil {
		return
	}

	logger.Info("Parsed client message",
		zap.Int32("session", sess.SessionID),
		zap.Uint16("ident", msg.Ident),
		zap.Int32("recog", msg.Recog),
	)

	body := data[14:]
	handleMessage(sess, msg.Ident, msg.Param, msg.Tag, msg.Series, msg.Recog, body)
}

var (
	userList      = make(map[int32]*UserEntry)
	userListMutex sync.RWMutex
	userIdxMap    = make(map[int]int32)
)

type UserEntry struct {
	Socket      int32
	IP          string
	SessionID   int32
	CharName    string
	Account     string
	LoginTime   time.Time
}

func handleMessage(sess *network.GateSession, ident, param, tag, series uint16, recog int32, body []byte) {
	logger.Info("handleMessage received", zap.Int32("session", sess.SessionID), zap.Uint16("ident", ident), zap.Uint16("param", param), zap.Uint16("tag", tag))

	SessionMapLock.RLock()
	account := SessionMap[int(sess.SessionID)]
	SessionMapLock.RUnlock()

	if account == "" {
		account = "testuser123"
		SessionMapLock.Lock()
		SessionMap[int(sess.SessionID)] = account
		SessionMapLock.Unlock()
		logger.Info("Using default account for session", zap.Int32("session", sess.SessionID), zap.String("account", account))
	}

	logger.Debug("handleMessage", zap.Int32("session", sess.SessionID), zap.String("account", account), zap.Uint16("ident", ident))

	sd := Sessions[sess.SessionID]
	if sd == nil {
		sd = &SessionData{}
		SessionsLock.Lock()
		Sessions[sess.SessionID] = sd
		SessionsLock.Unlock()
	}

	if account != "" {
		sd.Account = account
	}

	logger.Debug("Session account", zap.String("account", sd.Account))

	switch ident {
	case protocol.CM_QUERYCHR:
		handleQueryChar(sess, sd, body)
	case protocol.CM_NEWCHR:
		handleNewChar(sess, sd, body)
	case protocol.CM_DELCHR:
		handleDelChar(sess, sd, body)
	case protocol.CM_SELCHR:
		handleSelChar(sess, sd, body)
	case protocol.CM_TURN:
		handleTurn(sess, sd, param, tag)
	case protocol.CM_WALK:
		handleWalk(sess, sd, param, tag)
	case protocol.CM_RUN:
		handleRun(sess, sd, param, tag)
	case protocol.CM_HIT:
		handleHit(sess, sd, param, tag)
	case protocol.CM_SPELL:
		handleSpell(sess, sd, param, tag, recog)
	case protocol.CM_SAY:
		handleSay(sess, sd, string(body))
	case protocol.CM_PICKUP:
		handlePickUp(sess, sd)
	case protocol.CM_TAKEONITEM:
		handleTakeOnItem(sess, sd, param)
	case protocol.CM_TAKEOFFITEM:
		handleTakeOffItem(sess, sd, param)
	case protocol.CM_DROPITEM:
		handleDropItem(sess, sd, param)
	case protocol.CM_EAT:
		handleEat(sess, sd, param)
	case protocol.CM_CLICKNPC:
		handleClickNPC(sess, sd, param)
	case protocol.CM_MERCHANTDLGSELECT:
		handleMerchantDlgSelect(sess, sd, body)
	case protocol.CM_USERBUYITEM:
		handleUserBuyItem(sess, sd, body)
	case protocol.CM_USERSELLITEM:
		handleUserSellItem(sess, sd, body)
	case protocol.CM_CREATEGROUP:
		handleCreateGroup(sess, sd)
	case protocol.CM_ADDGROUPMEMBER:
		handleAddGroupMember(sess, sd, body)
	case protocol.CM_DELGROUPMEMBER:
		handleDelGroupMember(sess, sd, body)
	case protocol.CM_GROUPMODE:
		handleGroupMode(sess, sd, param)
	case protocol.CM_OPENGUILDDLG:
		handleOpenGuildDlg(sess, sd)
	case protocol.CM_GUILDADDMEMBER:
		handleGuildAddMember(sess, sd, body)
	case protocol.CM_GUILDDELMEMBER:
		handleGuildDelMember(sess, sd, body)
	default:
		logger.Debug("Unknown message",
			zap.Int32("session", sess.SessionID),
			zap.Uint16("ident", ident),
		)
	}
}

func handleQueryChar(sess *network.GateSession, sd *SessionData, body []byte) {
	SessionMapLock.RLock()
	account := SessionMap[int(sess.SessionID)]
	SessionMapLock.RUnlock()

	if account == "" {
		account = "testuser123"
	}

	if database != nil {
		acc, err := database.GetAccount(account)
		if err == nil && acc != nil {
			characters, err := database.GetCharactersByAccount(acc.AccountID)
			if err == nil && len(characters) > 0 {
				for _, char := range characters {
					sendCharInfo(sess, char)
				}
				return
			}
		}
	}

	sendTestCharInfo(sess, account)
}

func sendTestCharInfo(sess *network.GateSession, account string) {
	charName := account + "_char"
	charData := []byte(fmt.Sprintf("%s/%d/%d/%d/%d/", charName, 0, 2, 1, 0))

	msg := make([]byte, 14+len(charData))
	binary.LittleEndian.PutUint32(msg[0:4], 0)
	binary.LittleEndian.PutUint16(msg[4:6], protocol.SM_QUERYCHR)
	binary.LittleEndian.PutUint16(msg[12:14], uint16(len(charData)))
	copy(msg[14:], charData)

	sess.Send(network.EncodePacket(msg))
	logger.Info("Sent test character info", zap.String("account", account), zap.String("charName", charName))
}

func sendCharInfo(sess *network.GateSession, char *db.Character) {
	charStr := fmt.Sprintf("%s/%d/%d/%d/%d/", 
		char.Name, char.Job, char.Hair, char.Level, char.Gender)
	
	charData := []byte(charStr)
	msg := make([]byte, 14+len(charData))
	binary.LittleEndian.PutUint32(msg[0:4], 0)
	binary.LittleEndian.PutUint16(msg[4:6], protocol.SM_QUERYCHR)
	binary.LittleEndian.PutUint16(msg[12:14], uint16(len(charData)))
	copy(msg[14:], charData)
	
	sess.Send(network.EncodePacket(msg))
}

func encodeString(s string) string {
	result := make([]byte, 0, len(s)*2)
	var buffer, bits int
	
	for i := 0; i < len(s); i++ {
		buffer = (buffer << 8) | int(s[i])
		bits += 8
		
		for bits >= 6 {
			bits -= 6
			result = append(result, byte((buffer>>bits)&0x3F+0x3C))
		}
	}
	
	if bits > 0 {
		result = append(result, byte((buffer<<(6-bits))&0x3F+0x3C))
	}
	
	return string(result)
}

func handleNewChar(sess *network.GateSession, sd *SessionData, body []byte) {
	if sd == nil || sd.Account == "" || len(body) < 62 {
		sendMessage(sess, protocol.SM_NEWCHR_FAIL, 0, 0, 0, 0)
		return
	}

	name := string(body[:30])
	job := body[30]
	gender := body[31]
	hair := uint8(body[32])

	if name == "" {
		sendMessage(sess, protocol.SM_NEWCHR_FAIL, 0, 0, 0, 0)
		return
	}

	newChar := &db.Character{
		Name:    name,
		Job:     actor.Job(job),
		Gender:  actor.Gender(gender),
		Hair:    byte(hair),
		Level:   1,
		Gold:    0,
		MapName: "0",
		X:       289,
		Y:       618,
		HP:      100,
		MP:      100,
		Exp:     0,
	}

	if database != nil {
		acc, _ := database.GetAccount(sd.Account)
		if acc != nil {
			newChar.AccountID = acc.AccountID
			err := database.CreateCharacter(newChar)
			if err != nil {
				logger.Error("Failed to create character", zap.Error(err))
				sendMessage(sess, protocol.SM_NEWCHR_FAIL, 0, 0, 0, 0)
				return
			}
		}
	}

	sendMessage(sess, protocol.SM_NEWCHR_SUCCESS, 0, 0, 0, newChar.CharID)
}

func handleDelChar(sess *network.GateSession, sd *SessionData, body []byte) {
	if sd == nil || sd.Account == "" || len(body) < 30 {
		sendMessage(sess, protocol.SM_DELCHR_FAIL, 0, 0, 0, 0)
		return
	}

	name := string(body[:30])

	if database != nil {
		ch, _ := database.GetCharacterByName(name)
		if ch != nil {
			err := database.DeleteCharacter(ch.CharID)
			if err == nil {
				sendMessage(sess, protocol.SM_DELCHR_SUCCESS, 0, 0, 0, ch.CharID)
				return
			}
		}
	}

	sendMessage(sess, protocol.SM_DELCHR_FAIL, 0, 0, 0, 0)
}

func handleSelChar(sess *network.GateSession, sd *SessionData, body []byte) {
	logger.Info("handleSelChar called", zap.Int32("session", sess.SessionID), zap.Int("bodyLen", len(body)))

	if sd == nil {
		logger.Warn("SessionData is nil")
		sendMessage(sess, protocol.SM_STARTFAIL, 0, 0, 0, 0)
		return
	}

	if len(body) < 30 {
		logger.Warn("Body too short", zap.Int("bodyLen", len(body)))
		sendMessage(sess, protocol.SM_STARTFAIL, 0, 0, 0, 0)
		return
	}

	name := string(body[:30])
	logger.Info("Selecting character", zap.String("name", name))

	var char *db.Character
	if database != nil {
		char, _ = database.GetCharacterByName(name)
	}

	if char == nil {
		logger.Info("Character not found, creating temp character", zap.String("name", name))
		char = &db.Character{
			Name:    name,
			Job:     0,
			Gender:  0,
			Level:   1,
			MapName: "0",
			X:       289,
			Y:       618,
			HP:      100,
			MP:      100,
		}
	}

	player := actor.NewPlayer(0, char.Name)
	player.Job = actor.Job(char.Job)
	player.Gender = actor.Gender(char.Gender)
	player.Level = char.Level
	player.MapName = "0"
	player.X = 289
	player.Y = 618
	player.HP = int32(char.HP)
	player.MP = int32(char.MP)
	player.Gold = char.Gold
	player.Account = sd.Account

	logger.Info("Player created", zap.String("name", player.Name), zap.Int32("ID", player.ID))

	if database != nil {
		data, _ := database.LoadPlayerData(char.CharID)
		if len(data) > 0 {
			logger.Debug("Loaded player data", zap.String("name", name))
		}
	}

	sd.Player = player
	PlayerMgr.AddPlayer(player)

	logger.Info("Sending login messages")

	sendMessage(sess, protocol.SM_STARTPLAY, 0, 0, 0, player.ID)
	logger.Info("Sent SM_STARTPLAY")

	sendMessage(sess, protocol.SM_LOGON, 0, 0, 0, player.ID)
	logger.Info("Sent SM_LOGON")

	sendMessage(sess, protocol.SM_NEWMAP, 0, 0, 0, 0)
	logger.Info("Sent SM_NEWMAP")

	logger.Info("Sending char base info")
	sendCharBaseInfo(sess, player)
	
	sendMapInfo(sess, player)
	
	logger.Info("handleSelChar completed")
}

func handleTurn(sess *network.GateSession, sd *SessionData, param, tag uint16) {
	logger.Info("handleTurn called", zap.Int32("session", sess.SessionID), zap.Uint16("param", param), zap.Uint16("tag", tag))
	if sd == nil || sd.Player == nil {
		logger.Warn("handleTurn: sd or player is nil")
		return
	}

	dir := byte(param & 0x07)
	sd.Player.SetDirection(dir)
	logger.Info("handleTurn: turning", zap.String("dir", fmt.Sprintf("%d", dir)))

	broadcastMessage(sd.Player, protocol.SM_TURN, param, tag, 0, sd.Player.ID)
	logger.Info("handleTurn: broadcast SM_TURN sent")
}

func handleWalk(sess *network.GateSession, sd *SessionData, param, tag uint16) {
	logger.Info("handleWalk called", zap.Int32("session", sess.SessionID), zap.Uint16("param", param))
	if sd == nil || sd.Player == nil {
		logger.Warn("handleWalk: sd or player is nil")
		return
	}

	x := int(param & 0xFF)
	y := int((param >> 8) & 0xFF)
	dir := byte(tag & 0x07)

	logger.Info("handleWalk: target", zap.Int("x", x), zap.Int("y", y), zap.String("map", sd.Player.MapName))

	m := MapMgr.GetMap(sd.Player.MapName)
	if m == nil {
		logger.Warn("handleWalk: map is nil", zap.String("map", sd.Player.MapName))
		sendMessage(sess, protocol.SM_MOVEFAIL, 0, 0, 0, sd.Player.ID)
		return
	}

	if !m.CanWalk(x, y) {
		logger.Warn("handleWalk: cannot walk to position", zap.Int("x", x), zap.Int("y", y))
		sendMessage(sess, protocol.SM_MOVEFAIL, 0, 0, 0, sd.Player.ID)
		return
	}

	sd.Player.SetX(x)
	sd.Player.SetY(y)
	sd.Player.SetDirection(dir)
	broadcastMessage(sd.Player, protocol.SM_WALK, param, tag, 0, sd.Player.ID)
	logger.Info("handleWalk: broadcast SM_WALK sent")
}

func handleRun(sess *network.GateSession, sd *SessionData, param, tag uint16) {
	if sd == nil || sd.Player == nil {
		return
	}

	x := int(param & 0xFF)
	y := int((param >> 8) & 0xFF)
	dir := byte(tag & 0x07)

	m := MapMgr.GetMap(sd.Player.MapName)
	if m == nil || !m.CanWalk(x, y) {
		sendMessage(sess, protocol.SM_MOVEFAIL, 0, 0, 0, sd.Player.ID)
		return
	}

	sd.Player.SetX(x)
	sd.Player.SetY(y)
	sd.Player.SetDirection(dir)
	broadcastMessage(sd.Player, protocol.SM_RUN, param, tag, 0, sd.Player.ID)
}

func handleHit(sess *network.GateSession, sd *SessionData, param, tag uint16) {
	if sd == nil || sd.Player == nil {
		return
	}

	broadcastMessage(sd.Player, protocol.SM_HIT, param, tag, 0, sd.Player.ID)

	attackTarget(sd.Player, param)
}

func handleSpell(sess *network.GateSession, sd *SessionData, param, tag uint16, recog int32) {
	if sd == nil || sd.Player == nil {
		return
	}

	magicID := uint16(recog)
	targetID := int32(param)

	logger.Debug("Player cast spell",
		zap.String("player", sd.Player.Name),
		zap.Uint16("magic", magicID),
		zap.Int32("target", targetID),
	)

	broadcastMessage(sd.Player, protocol.SM_SPELL, param, tag, 0, sd.Player.ID)
}

func handleSay(sess *network.GateSession, sd *SessionData, msg string) {
	if sd == nil || sd.Player == nil {
		return
	}

	logger.Info("Player says",
		zap.String("player", sd.Player.Name),
		zap.String("message", msg),
	)

	sendMessage(sess, protocol.SM_SYSMESSAGE, 0, 0, 0, 0)
}

func handlePickUp(sess *network.GateSession, sd *SessionData) {
	if sd == nil || sd.Player == nil {
		return
	}

	logger.Debug("Player pick up", zap.String("player", sd.Player.Name))
}

func handleTakeOnItem(sess *network.GateSession, sd *SessionData, param uint16) {
	if sd == nil || sd.Player == nil {
		return
	}

	logger.Debug("Player take on item",
		zap.String("player", sd.Player.Name),
		zap.Uint16("position", param),
	)

	sendMessage(sess, protocol.SM_TAKEON_OK, param, 0, 0, sd.Player.ID)
}

func handleTakeOffItem(sess *network.GateSession, sd *SessionData, param uint16) {
	if sd == nil || sd.Player == nil {
		return
	}

	logger.Debug("Player take off item",
		zap.String("player", sd.Player.Name),
		zap.Uint16("position", param),
	)

	sendMessage(sess, protocol.SM_TAKEOFF_OK, param, 0, 0, sd.Player.ID)
}

func handleDropItem(sess *network.GateSession, sd *SessionData, param uint16) {
	if sd == nil || sd.Player == nil {
		return
	}

	logger.Debug("Player drop item",
		zap.String("player", sd.Player.Name),
		zap.Uint16("index", param),
	)
}

func handleEat(sess *network.GateSession, sd *SessionData, param uint16) {
	if sd == nil || sd.Player == nil {
		return
	}

	logger.Debug("Player eat",
		zap.String("player", sd.Player.Name),
		zap.Uint16("index", param),
	)
}

func handleClickNPC(sess *network.GateSession, sd *SessionData, param uint16) {
	if sd == nil || sd.Player == nil {
		return
	}

	npcID := int32(param)
	logger.Debug("Player click NPC", zap.String("player", sd.Player.Name), zap.Int32("npc", npcID))

	sendSystemMessage(sess, "欢迎来到传奇世界!")
}

func handleMerchantDlgSelect(sess *network.GateSession, sd *SessionData, body []byte) {
	if sd == nil || sd.Player == nil || len(body) < 36 {
		return
	}

	npcurl := string(body[:36])
	logger.Debug("Player merchant dialog", zap.String("player", sd.Player.Name), zap.String("npc", npcurl))

	sendMerchantList(sess)
}

func handleUserBuyItem(sess *network.GateSession, sd *SessionData, body []byte) {
	if sd == nil || sd.Player == nil || len(body) < 8 {
		return
	}

	index := binary.LittleEndian.Uint16(body[0:2])
	count := binary.LittleEndian.Uint16(body[2:4])

	logger.Debug("Player buy item", zap.String("player", sd.Player.Name), zap.Uint16("index", index), zap.Uint16("count", count))

	sendMessage(sess, protocol.SM_BUYITEM_SUCCESS, 0, 0, 0, 0)
}

func handleUserSellItem(sess *network.GateSession, sd *SessionData, body []byte) {
	if sd == nil || sd.Player == nil || len(body) < 4 {
		return
	}

	index := binary.LittleEndian.Uint16(body[0:2])

	logger.Debug("Player sell item", zap.String("player", sd.Player.Name), zap.Uint16("index", index))

	sendMessage(sess, protocol.SM_USERSELLITEM_OK, 0, 0, 0, 0)
}

func handleCreateGroup(sess *network.GateSession, sd *SessionData) {
	if sd == nil || sd.Player == nil {
		return
	}

	if sd.Player.GroupID != 0 {
		sendMessage(sess, protocol.SM_CREATEGROUP_FAIL, 0, 0, 0, 0)
		return
	}

	group := actor.GetGroupManager().CreateGroup(sd.Player)
	if group == nil {
		sendMessage(sess, protocol.SM_CREATEGROUP_FAIL, 0, 0, 0, 0)
		return
	}

	sendMessage(sess, protocol.SM_CREATEGROUP_OK, 0, 0, 0, group.GroupID)
}

func handleAddGroupMember(sess *network.GateSession, sd *SessionData, body []byte) {
	if sd == nil || sd.Player == nil || len(body) < 30 {
		return
	}

	targetName := string(body[:30])
	targetPlayer := actor.GetPlayerManager().GetPlayerByName(targetName)
	if targetPlayer == nil {
		sendMessage(sess, protocol.SM_GROUPADDMEM_FAIL, 0, 0, 0, 0)
		return
	}

	if targetPlayer.GroupID != 0 {
		sendMessage(sess, protocol.SM_GROUPADDMEM_FAIL, 0, 0, 0, 0)
		return
	}

	group := actor.GetGroupManager().GetPlayerGroup(sd.Player)
	if group == nil {
		sendMessage(sess, protocol.SM_GROUPADDMEM_FAIL, 0, 0, 0, 0)
		return
	}

	if group.AddMember(targetPlayer) {
		sendMessage(sess, protocol.SM_GROUPADDMEM_OK, 0, 0, 0, targetPlayer.ID)
	} else {
		sendMessage(sess, protocol.SM_GROUPADDMEM_FAIL, 0, 0, 0, 0)
	}
}

func handleDelGroupMember(sess *network.GateSession, sd *SessionData, body []byte) {
	if sd == nil || sd.Player == nil || len(body) < 30 {
		return
	}

	targetName := string(body[:30])
	targetPlayer := actor.GetPlayerManager().GetPlayerByName(targetName)
	if targetPlayer == nil {
		sendMessage(sess, protocol.SM_GROUPDELMEM_FAIL, 0, 0, 0, 0)
		return
	}

	group := actor.GetGroupManager().GetPlayerGroup(sd.Player)
	if group == nil {
		sendMessage(sess, protocol.SM_GROUPDELMEM_FAIL, 0, 0, 0, 0)
		return
	}

	if group.DelMember(targetPlayer) {
		sendMessage(sess, protocol.SM_GROUPDELMEM_OK, 0, 0, 0, targetPlayer.ID)
	} else {
		sendMessage(sess, protocol.SM_GROUPDELMEM_FAIL, 0, 0, 0, 0)
	}
}

func handleGroupMode(sess *network.GateSession, sd *SessionData, param uint16) {
	if sd == nil || sd.Player == nil {
		return
	}

	if param == 1 {
		sd.Player.GroupID = 1
		sendMessage(sess, protocol.SM_GROUPMODECHANGED, 1, 0, 0, 0)
	} else {
		if sd.Player.GroupID != 0 {
			actor.GetGroupManager().LeaveGroup(sd.Player)
		}
		sd.Player.GroupID = 0
		sendMessage(sess, protocol.SM_GROUPMODECHANGED, 0, 0, 0, 0)
	}
}

func handleOpenGuildDlg(sess *network.GateSession, sd *SessionData) {
	if sd == nil || sd.Player == nil {
		return
	}

	if sd.Player.GuildID == 0 {
		sendMessage(sess, protocol.SM_OPENGUILDDLG_FAIL, 0, 0, 0, 0)
		return
	}

	sendMessage(sess, protocol.SM_OPENGUILDDLG, 0, 0, 0, sd.Player.GuildID)
}

func handleGuildAddMember(sess *network.GateSession, sd *SessionData, body []byte) {
	if sd == nil || sd.Player == nil || len(body) < 30 {
		return
	}

	targetName := string(body[:30])
	logger.Debug("Guild add member", zap.String("player", sd.Player.Name), zap.String("target", targetName))

	sendMessage(sess, protocol.SM_GUILDADDMEMBER_OK, 0, 0, 0, 0)
}

func handleGuildDelMember(sess *network.GateSession, sd *SessionData, body []byte) {
	if sd == nil || sd.Player == nil || len(body) < 30 {
		return
	}

	targetName := string(body[:30])
	logger.Debug("Guild del member", zap.String("player", sd.Player.Name), zap.String("target", targetName))

	sendMessage(sess, protocol.SM_GUILDDELMEMBER_OK, 0, 0, 0, 0)
}

func sendMerchantList(sess *network.GateSession) {
	data := make([]byte, 64)
	data[0] = 0

	sendPacket(sess, protocol.SM_SENDGOODSLIST, data)
}

func sendMessage(sess *network.GateSession, ident uint16, param, tag, series uint16, recog int32) {
	msg := &protocol.TDefaultMessage{
		Ident:  ident,
		Param:  param,
		Tag:    tag,
		Series: series,
		Recog:  recog,
	}
	
	data := msg.Pack()
	sess.Send(network.EncodePacket(data))
}

func sendCharBaseInfo(sess *network.GateSession, player *actor.Player) {
	ability := player.GetAbility()
	abilityData := ability.Pack()
	sess.Send(network.EncodePacket(abilityData))

	charDesc := player.GetCharDesc()
	descData := charDesc.Pack()
	sess.Send(network.EncodePacket(descData))
}

func broadcastMessage(target actor.Actor, ident uint16, param, tag, series uint16, recog int32) {
	msg := &protocol.TDefaultMessage{
		Ident:  ident,
		Param:  param,
		Tag:    tag,
		Series: series,
		Recog:  recog,
	}

	data := msg.Pack()
	server.Broadcast(data)
}

func attackTarget(attacker *actor.Player, targetParam uint16) {
	targetID := int32(targetParam)
	target := PlayerMgr.GetPlayer(targetID)
	if target == nil {
		return
	}

	damage := attacker.Ability.DC / 2
	if damage < 1 {
		damage = 1
	}

	target.AddHP(-damage)

	if !target.IsAlive() {
		killedPlayer(attacker, target)
	}
}

func killedPlayer(killer, victim *actor.Player) {
	logger.Info("Player killed",
		zap.String("killer", killer.Name),
		zap.String("victim", victim.Name),
	)

	victim.PKPoint += 100

	victim.MapName = "0"
	victim.X = 289
	victim.Y = 618
	victim.HP = victim.MaxHP / 2
	victim.MP = victim.MaxMP / 2

	victim.Items = make([]*item.TUserItem, 0)
}

func savePlayerData(player *actor.Player) {
	if database == nil || player == nil {
		return
	}

	char, _ := database.GetCharacterByName(player.Name)
	if char == nil {
		return
	}

	char.X = player.X
	char.Y = player.Y
	char.HP = player.HP
	char.MP = player.MP
	char.Level = player.Level
	char.Gold = player.Gold

	database.UpdateCharacter(char)
}

func registerUser(socket int32, ip string) int {
	userListMutex.Lock()
	defer userListMutex.Unlock()

	idx := len(userList)
	entry := &UserEntry{
		Socket:    socket,
		IP:        ip,
		SessionID: socket,
		LoginTime: time.Now(),
	}
	userList[socket] = entry
	userIdxMap[idx] = socket

	logger.Info("User registered",
		zap.Int32("socket", socket),
		zap.String("ip", ip),
		zap.Int("index", idx),
	)

	return idx
}

func unregisterUser(socket int32) {
	userListMutex.Lock()
	defer userListMutex.Unlock()

	if entry, ok := userList[socket]; ok {
		logger.Info("User unregistered",
			zap.Int32("socket", socket),
			zap.String("account", entry.Account),
		)
		delete(userList, socket)
		for idx, sock := range userIdxMap {
			if sock == socket {
				delete(userIdxMap, idx)
				break
			}
		}
	}
}

func sendNewUserMsg(sess *network.GateSession, socket int32, socketIdx uint16, userIdx int) {
	msgHeader := protocol.TMsgHeader{
		Code:          protocol.RUNGATECODE,
		Socket:        socket,
		GSocketIdx:    socketIdx,
		Ident:         protocol.GM_OPEN,
		UserListIndex: int32(userIdx),
		Length:        0,
	}

	headerData := msgHeader.Pack()

	logger.Info("Sending GM_OPEN response",
		zap.Int32("socket", socket),
		zap.Int("userIdx", userIdx),
		zap.ByteString("header", headerData),
	)

	_, err := sess.Conn.Write(headerData)
	if err != nil {
		logger.Error("Failed to send GM_OPEN response", zap.Error(err))
	}
}

func sendServerVersion(sess *network.GateSession) {
	nRunHuman := byte(0)
	nRunMon := byte(0)
	nRunNpc := byte(0)
	nWarRunAll := byte(0)

	recog := int32(nRunHuman) | (int32(nRunMon) << 8) | (int32(nRunNpc) << 16) | (int32(nWarRunAll) << 24)
	param := uint16(0x0005)

	serverConfig := &protocol.TServerConfig{
		BtShowClientItemStyle: 0,
		BoAllowItemAddValue:  1,
		BoAllowItemTime:      1,
		BoAllowItemAddPoint:  1,
		BoCheckSpeedHack:     0,
		NGreenNumber:         20,
		BoRUNHUMAN:          1,
		BoRUNMON:            1,
		BoRunNpc:            1,
		BoChgSpeed:          1,
		NFireDelayTime:       10000,
		NKTZDelayTime:        10000,
		NPKJDelayTime:        10000,
		NSkill50DelayTime:    5000,
		NZRJFDelayTime:       10000,
		NMaxLevel:            999,
		BoAllowPlayerAutoPot: 0,
	}

	configData := serverConfig.Pack()
	msgData := make([]byte, 14+len(configData))
	binary.LittleEndian.PutUint32(msgData[0:4], uint32(recog))
	binary.LittleEndian.PutUint16(msgData[4:6], protocol.SM_SERVERCONFIG)
	binary.LittleEndian.PutUint16(msgData[6:8], param)
	msgData[8] = 0
	msgData[9] = 0
	msgData[10] = 0
	msgData[11] = 0
	copy(msgData[14:], configData)

	sess.Send(network.EncodePacket(msgData))
}

func sendMapInfo(sess *network.GateSession, player *actor.Player) {
	sendMessage(sess, protocol.SM_MAPDESCRIPTION, 0, 0, 0, 0)

	m := MapMgr.GetMap(player.MapName)
	if m != nil {
		data := make([]byte, 32)
		copy(data, player.MapName)
		sendMapInfoPacket(sess, data)
	}
}

func sendMapInfoPacket(sess *network.GateSession, data []byte) {
	msg := make([]byte, 14+len(data))
	binary.LittleEndian.PutUint32(msg[0:4], 0)
	binary.LittleEndian.PutUint16(msg[4:6], protocol.SM_MAPDESCRIPTION)
	copy(msg[14:], data)
	sess.Send(network.EncodePacket(msg))
}

func sendHealth(sess *network.GateSession, player *actor.Player) {
	hp := uint16(player.HP)
	mp := uint16(player.MP)
	maxHP := uint16(player.MaxHP)

	data := make([]byte, 10)
	binary.LittleEndian.PutUint16(data[0:2], hp)
	binary.LittleEndian.PutUint16(data[2:4], mp)
	binary.LittleEndian.PutUint16(data[4:6], maxHP)
	data[8] = 0

	sendPacket(sess, protocol.SM_HEALTHSPELLCHANGED, data)
}

func sendWeight(sess *network.GateSession, player *actor.Player) {
	data := make([]byte, 6)
	binary.LittleEndian.PutUint16(data[0:2], player.Ability.Weight)
	binary.LittleEndian.PutUint16(data[2:4], player.Ability.MaxWeight)
	binary.LittleEndian.PutUint16(data[4:6], player.Ability.WearWeight)

	sendPacket(sess, protocol.SM_WEIGHTCHANGED, data)
}

func sendPacket(sess *network.GateSession, ident uint16, data []byte) {
	msg := make([]byte, 14+len(data))
	binary.LittleEndian.PutUint32(msg[0:4], 0)
	binary.LittleEndian.PutUint16(msg[4:6], ident)
	copy(msg[14:], data)
	sess.Send(network.EncodePacket(msg))
}

func sendAbility(sess *network.GateSession, player *actor.Player) {
	ability := player.GetAbility()
	abilityData := ability.Pack()
	sendPacket(sess, protocol.SM_ABILITY, abilityData)
}

func sendSubAbility(sess *network.GateSession, player *actor.Player) {
	data := make([]byte, 40)

	ab := &player.AddAbility
	binary.LittleEndian.PutUint16(data[0:2], ab.HitPoint)
	binary.LittleEndian.PutUint16(data[2:4], ab.SpeedPoint)
	binary.LittleEndian.PutUint16(data[4:6], ab.AntiPoison)
	binary.LittleEndian.PutUint16(data[6:8], ab.PoisonRecover)
	binary.LittleEndian.PutUint16(data[8:10], ab.HealthRecover)
	binary.LittleEndian.PutUint16(data[10:12], ab.SpellRecover)
	binary.LittleEndian.PutUint16(data[12:14], ab.AntiMagic)
	binary.LittleEndian.PutUint32(data[16:20], uint32(ab.HitSpeed))
	binary.LittleEndian.PutUint32(data[20:24], uint32(ab.DC))
	binary.LittleEndian.PutUint32(data[24:28], uint32(ab.MC))
	binary.LittleEndian.PutUint32(data[28:32], uint32(ab.SC))
	binary.LittleEndian.PutUint32(data[32:36], uint32(ab.AC))
	binary.LittleEndian.PutUint32(data[36:40], uint32(ab.MAC))

	sendPacket(sess, protocol.SM_SUBABILITY, data)
}

func sendGoldChanged(sess *network.GateSession, gold int32) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data[0:4], uint32(gold))
	sendPacket(sess, protocol.SM_GOLDCHANGED, data)
}

func sendGroupMessage(sess *network.GateSession, fromName, message string) {
	data := make([]byte, 2+len(fromName)+1+len(message))
	data[0] = byte(len(fromName) + 1)
	copy(data[1:], fromName)
	offset := 1 + len(fromName)
	data[offset] = byte(len(message))
	copy(data[offset+1:], message)

	sendPacket(sess, protocol.SM_GROUPMESSAGE, data)
}

func sendGuildMessage(sess *network.GateSession, fromName, message string) {
	data := make([]byte, 2+len(fromName)+1+len(message))
	data[0] = byte(len(fromName) + 1)
	copy(data[1:], fromName)
	offset := 1 + len(fromName)
	data[offset] = byte(len(message))
	copy(data[offset+1:], message)

	sendPacket(sess, protocol.SM_GUILDMESSAGE, data)
}

func sendSystemMessage(sess *network.GateSession, message string) {
	data := []byte(message)
	sendPacket(sess, protocol.SM_SYSMESSAGE, data)
}

func handleLoginServerConnection(serverCfg *config.ServerConfig) {
	for {
		if serverCfg == nil {
			logger.Warn("serverCfg is nil, retrying in 5s")
			time.Sleep(5 * time.Second)
			continue
		}
		loginSrvAddr := fmt.Sprintf("127.0.0.1:%d", serverCfg.LoginSrv.Port)
		
		conn, err := net.DialTimeout("tcp", loginSrvAddr, 5*time.Second)
		if err != nil {
			logger.Warn("Failed to connect to LoginSrv, retrying in 5s", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}
		
		logger.Info("Connected to LoginSrv", zap.String("addr", loginSrvAddr))
		
		buf := make([]byte, 8192)
		for {
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))
			n, err := conn.Read(buf)
			if err != nil {
				logger.Warn("LoginSrv connection lost, reconnecting", zap.Error(err))
				break
			}
			
			data := buf[:n]
			processLoginServerMessage(data)
		}
	}
}

func processLoginServerMessage(data []byte) {
	logger.Debug("Received from LoginSrv", zap.ByteString("data", data))
	
	if len(data) < 6 {
		return
	}
	
	header := binary.LittleEndian.Uint32(data[0:4])
	if header != 0xAA55AA55 {
		logger.Warn("Invalid header from LoginSrv", zap.Uint32("header", header))
		return
	}
	
	msgID := binary.LittleEndian.Uint16(data[4:6])
	logger.Info("LoginSrv message", zap.Uint16("msgID", msgID))
	
	switch msgID {
	case protocol.SS_OPENSESSION:
		if len(data) > 6 {
			msg := string(data[6:])
			parts := strings.Split(msg, "/")
			if len(parts) >= 2 {
				sessionID := 0
				fmt.Sscanf(parts[1], "%d", &sessionID)
				account := parts[0]
				
				SessionMapLock.Lock()
				SessionMap[sessionID] = account
				SessionMapLock.Unlock()
				
				logger.Info("Session opened", zap.Int("sessionID", sessionID), zap.String("account", account))
			}
		}
	case protocol.SS_CLOSESESSION:
		if len(data) > 6 {
			msg := string(data[6:])
			parts := strings.Split(msg, "/")
			if len(parts) >= 2 {
				sessionID := 0
				fmt.Sscanf(parts[1], "%d", &sessionID)
				
				SessionMapLock.Lock()
				delete(SessionMap, sessionID)
				SessionMapLock.Unlock()
				
				logger.Info("Session closed", zap.Int("sessionID", sessionID))
			}
		}
	}
}
