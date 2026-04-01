package script

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ScriptCommand string

const (
	CMD_GIVE        = "GIVE"
	CMD_TAKE        = "TAKE"
	CMD_GOLD        = "GOLD"
	CMD_SET         = "SET"
	CMD_RESET       = "RESET"
	CMD_BREAK       = "BREAK"
	CMD_CLOSE       = "CLOSE"
	CMD_GOTO        = "GOTO"
	CMD_SAY         = "SAY"
	CMD_NOSAY       = "NOSAY"
	CMD_IF          = "IF"
	CMD_ELSE        = "ELSE"
	CMD_ENDIF       = "ENDIF"
	CMD_SLEEP       = "SLEEP"
	CMD_CALL        = "CALL"
	CMD_RETURN      = "RETURN"
	CMD_LOADVAR     = "LOADVAR"
	CMD_SAVEVAR     = "SAVEVAR"
	CMD_EQUAL       = "EQUAL"
	CMD_NOTEQUAL    = "NOTEQUAL"
	CMD_LARGE       = "LARGE"
	CMD_SMALL       = "SMALL"
	CMD_RANDOM      = "RANDOM"
	CMD_DELAYGOTO   = "DELAYGOTO"
	CMD_TIMER       = "TIMER"
	CMD_CHECK       = "CHECK"
	CMD_CLEAR       = "CLEAR"
	CMD_ADD         = "ADD"
	CMD_SUB         = "SUB"
	CMD_MUL         = "MUL"
	CMD_DIV         = "DIV"
	CMD_MOV         = "MOV"
	CMD_INC         = "INC"
	CMD_DEC         = "DEC"
	CMD_POW         = "POW"
	CMD_PERFORM     = "PERFORM"
	CMD_CREATEGROUP = "CREATEGROUP"
	CMD_ADDGROUP    = "ADDGROUP"
	CMD_DELGROUP    = "DELGROUP"
	CMD_GROUPLIST   = "GROUPLIST"
	CMD_MAPMOVE     = "MAPMOVE"
	CMD_MAP         = "MAP"
	CMD_MONSTER     = "MONSTER"
	CMD_ITEM        = "ITEM"
	CMD_GIVEITEM    = "GIVEITEM"
	CMD_TAKEITEM    = "TAKEITEM"
	CMD_ADDGOLD     = "ADDGOLD"
	CMD_TAKEGOLD    = "TAKEGOLD"
	CMD_SETPK       = "SETPK"
	CMD_SETLEVEL    = "SETLEVEL"
	CMD_SETSKILL    = "SETSKILL"
	CMD_GIVEHERO    = "GIVEHERO"
	CMD_GIVEHORSE   = "GIVEHORSE"
	CMD_TAKEHORSE   = "TAKEHORSE"
	CMD_SETSTATUS   = "SETSTATUS"
	CMD_CLEARSTATUS = "CLEARSTATUS"
	CMD_SENDMSG     = "SENDMSG"
	CMD_SENDTIMERMSG = "SENDTIMERMSG"
	CMD_SENDUPGRADEmsg = "SENDUPGRADEmsg"
	CMD_SENDDEALMENU = "SENDDEALMENU"
	CMD_OPENSHOP    = "OPENSHOP"
	CMD_RECALLGROUP = "RECALLGROUP"
	CMD_CLEARLIST   = "CLEARLIST"
	CMD_FORMATMSG  = "FORMATMSG"
	CMD_CHANGEMODE  = "CHANGEMODE"
	CMD_GIVEEX     = "GIVEEX"
	CMD_TAKEEX     = "TAKEEX"
	CMD_CHECKEX    = "CHECKEX"
	CMD_REVIVE     = "REVIVE"
	CMD_MAKE       = "MAKE"
	CMD_UPGRADE    = "UPGRADE"
	CMD_REPAIR     = "REPAIR"
	CMD_DESTROY    = "DESTROY"
	CMD_CLEARPET   = "CLEARPET"
	CMD_CLEARHERO  = "CLEARHERO"
	CMD_CLEARSKILL = "CLEARSKILL"
	CMD_TEST       = "TEST"
	CMD_TESTEX     = "TESTEX"
	CMD_RESETVAR   = "RESETVAR"
	CMD_GETRANDOM  = "GETRANDOM"
	CMD_SETVAR     = "SETVAR"
	CMD_GETVAR     = "GETVAR"
	CMD_CMPVAR     = "CMPVAR"
	CMD_DELAY      = "DELAY"
	CMD_GETITEM    = "GETITEM"
	CMD_GM         = "GM"
	CMD_KICK       = "KICK"
	CMD_CLOSEALL   = "CLOSEALL"
)

type Condition struct {
	Type    string
	Params  []string
	Negated bool
}

type Action struct {
	Command string
	Params  []string
}

type QuestInfo struct {
	boQuest    bool
	QuestList  []QuestCondition
	RecordList []string
	nQuest     int
}

type QuestCondition struct {
	Flag  uint16
	Value byte
}

type ScriptLine struct {
	LineNum    int
	Label      string
	Command    string
	Params     []string
	Conditions []Condition
	Actions    []Action
	IsJump     bool
	JumpLabel  string
}

type Script struct {
	Name      string
	FileName  string
	Lines     []ScriptLine
	Labels    map[string]int
	Variables map[string]interface{}
}

type ScriptEngine struct {
	Scripts   map[string]*Script
	CurScript *Script
	CurLine   int
	Variables map[string]interface{}
}

func NewScriptEngine() *ScriptEngine {
	return &ScriptEngine{
		Scripts:   make(map[string]*Script),
		Variables: make(map[string]interface{}),
	}
}

func (se *ScriptEngine) LoadScript(filename string) (*Script, error) {
	script := &Script{
		Name:      filename,
		FileName:  filename,
		Lines:     make([]ScriptLine, 0),
		Labels:    make(map[string]int),
		Variables: make(map[string]interface{}),
	}
	
	se.Scripts[filename] = script
	return script, nil
}

func (se *ScriptEngine) ParseScript(filename, content string) (*Script, error) {
	script := &Script{
		Name:      filename,
		FileName:  filename,
		Lines:     make([]ScriptLine, 0),
		Labels:    make(map[string]int),
		Variables: make(map[string]interface{}),
	}
	
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}
		
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			label := strings.Trim(line, "[]")
			script.Labels[label] = i
			continue
		}
		
		parsed, err := se.parseLine(line, i)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		if parsed != nil {
			script.Lines = append(script.Lines, *parsed)
		}
	}
	
	se.Scripts[filename] = script
	return script, nil
}

func (se *ScriptEngine) parseLine(line string, lineNum int) (*ScriptLine, error) {
	parsed := &ScriptLine{
		LineNum: lineNum,
	}
	
	if strings.HasPrefix(line, "@") {
		parts := strings.SplitN(line, " ", 2)
		parsed.Command = strings.TrimPrefix(parts[0], "@")
		if len(parts) > 1 {
			parsed.Params = se.parseParams(parts[1])
		}
		return parsed, nil
	}
	
	if strings.HasPrefix(line, "#") {
		if strings.HasPrefix(line, "#IF") {
			parsed.Command = "IF"
			rest := strings.TrimPrefix(line, "#IF")
			rest = strings.TrimSpace(rest)
			parsed.Conditions = se.parseConditions(rest)
			return parsed, nil
		}
		
		if strings.HasPrefix(line, "#ACT") || strings.HasPrefix(line, "#ELSEACT") || 
		   strings.HasPrefix(line, "#ELSE") || strings.HasPrefix(line, "#ENDIF") {
			parsed.Command = strings.TrimPrefix(strings.TrimSpace(line), "#")
			return parsed, nil
		}
		
		if strings.HasPrefix(line, "#SAY") {
			parsed.Command = "SAY"
			parsed.Params = []string{strings.TrimPrefix(line, "#SAY ")}
			return parsed, nil
		}
		
		if strings.HasPrefix(line, "#ELSE") {
			parsed.Command = "ELSE"
			return parsed, nil
		}
	}
	
	parts := strings.SplitN(line, " ", 2)
	parsed.Command = strings.ToUpper(parts[0])
	if len(parts) > 1 {
		parsed.Params = se.parseParams(parts[1])
	}
	
	return parsed, nil
}

func (se *ScriptEngine) parseParams(params string) []string {
	var result []string
	var current strings.Builder
	inQuote := false
	
	for _, ch := range params {
		if ch == '"' {
			inQuote = !inQuote
			continue
		}
		if ch == ' ' && !inQuote {
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteRune(ch)
	}
	
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	
	return result
}

func (se *ScriptEngine) parseConditions(condStr string) []Condition {
	var conditions []Condition
	
	condStr = strings.TrimSpace(condStr)
	if condStr == "" {
		return conditions
	}
	
	parts := strings.Fields(condStr)
	
	for i := 0; i < len(parts); i++ {
		cond := Condition{Type: parts[i]}
		if strings.HasPrefix(parts[i], "!") {
			cond.Negated = true
			cond.Type = strings.TrimPrefix(parts[i], "!")
		}
		
		i++
		for i < len(parts) && !strings.HasPrefix(parts[i], "CHECK") && 
			parts[i] != "AND" && parts[i] != "OR" {
			cond.Params = append(cond.Params, parts[i])
			i++
		}
		i--
		
		conditions = append(conditions, cond)
	}
	
	return conditions
}

func (se *ScriptEngine) ExecuteScript(scriptName string) error {
	script, ok := se.Scripts[scriptName]
	if !ok {
		return fmt.Errorf("script not found: %s", scriptName)
	}
	
	se.CurScript = script
	se.CurLine = 0
	
	return se.runScript(script)
}

func (se *ScriptEngine) runScript(script *Script) error {
	for se.CurLine < len(script.Lines) {
		line := script.Lines[se.CurLine]
		
		switch line.Command {
		case "IF":
			if !se.checkConditions(line.Conditions) {
				se.skipToEndif()
			}
		case "ELSE":
			se.skipToEndif()
		case "ENDIF":
		case "SAY":
			fmt.Println(strings.Join(line.Params, " "))
		case "GOTO":
			if len(line.Params) > 0 {
				if target, ok := script.Labels[line.Params[0]]; ok {
					se.CurLine = target
					continue
				}
			}
		case "GIVE":
			fmt.Printf("Give: %v\n", line.Params)
		case "TAKE":
			fmt.Printf("Take: %v\n", line.Params)
		case "CLOSE":
			return nil
		case "BREAK":
			return nil
		}
		
		se.CurLine++
	}
	
	return nil
}

func (se *ScriptEngine) checkConditions(conditions []Condition) bool {
	for _, cond := range conditions {
		result := se.evaluateCondition(cond)
		if cond.Negated {
			result = !result
		}
		if !result {
			return false
		}
	}
	return true
}

func (se *ScriptEngine) evaluateCondition(cond Condition) bool {
	switch cond.Type {
	case "CHECKLEVELEX":
		if len(cond.Params) < 2 {
			return false
		}
		op := cond.Params[0]
		val, err := strconv.Atoi(cond.Params[1])
		if err != nil {
			return false
		}
		return se.compareInt(10, val, op)
		
	case "CHECKGOLD":
		if len(cond.Params) < 2 {
			return false
		}
		op := cond.Params[0]
		val, err := strconv.ParseInt(cond.Params[1], 10, 64)
		if err != nil {
			return false
		}
		return se.compareInt(1000, int(val), op)
		
	case "CHECKNAMELIST":
		return false
		
	case "RANDOM":
		if len(cond.Params) < 1 {
			return false
		}
		rate, err := strconv.Atoi(cond.Params[0])
		if err != nil {
			return false
		}
		return rand.Intn(100) < rate
		
	case "CHECKITEM":
		return se.checkPlayerItem(cond.Params)
		
	case "CHECKGENDER":
		if len(cond.Params) < 1 {
			return false
		}
		return se.checkGender(cond.Params[0])
		
	case "CHECKJOB":
		if len(cond.Params) < 1 {
			return false
		}
		return se.checkJob(cond.Params[0])
		
	case "CHECKDAY":
		if len(cond.Params) < 1 {
			return false
		}
		return se.checkDay(cond.Params[0])
		
	case "CHECKHOUR":
		if len(cond.Params) < 1 {
			return false
		}
		return se.checkHour(cond.Params[0])
		
	case "CHECKVAR":
		return se.checkVar(cond.Params)
		
	case "CHECKGROUP":
		return se.checkGroup()
		
	case "CHECKINMAP":
		return se.checkInMap(cond.Params)
		
	case "CHECKONLINE":
		return se.checkOnline(cond.Params)
		
	case "CHECKPKPOINT":
		return se.checkPKPoint(cond.Params)
		
	case "CHECKDEAD":
		return se.checkDead(cond.Params)
		
	case "CHECKCASTLE":
		return se.checkCastle(cond.Params)
		
	case "CHECKQUEST":
		return se.checkQuest(cond.Params)
		
	case "CHECKRANGE":
		return se.checkRange(cond.Params)
		
	case "CHECKSKILL":
		return se.checkSkill(cond.Params)
		
	case "CHECKMONSTER":
		return se.checkMonster(cond.Params)
		
	case "CHECKGUILDMEMBER":
		return se.checkGuildMember(cond.Params)
		
	case "CHECKRIDING":
		return se.checkRiding()
		
	case "CHECKSTORE":
		return se.checkStore(cond.Params)
		
	case "CHECKREPAIR":
		return se.checkRepair(cond.Params)
		
	default:
		return false
	}
}

func (se *ScriptEngine) compareInt(a, b int, op string) bool {
	switch op {
	case "=":
		return a == b
	case ">":
		return a > b
	case "<":
		return a < b
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	default:
		return false
	}
}

func (se *ScriptEngine) checkPlayerItem(params []string) bool {
	return false
}

func (se *ScriptEngine) checkGender(gender string) bool {
	return false
}

func (se *ScriptEngine) checkJob(job string) bool {
	return false
}

func (se *ScriptEngine) checkDay(day string) bool {
	return false
}

func (se *ScriptEngine) checkHour(hour string) bool {
	return false
}

func (se *ScriptEngine) checkVar(params []string) bool {
	return false
}

func (se *ScriptEngine) checkGroup() bool {
	return false
}

func (se *ScriptEngine) checkInMap(params []string) bool {
	if len(params) < 1 {
		return false
	}
	return false
}

func (se *ScriptEngine) checkOnline(params []string) bool {
	if len(params) < 1 {
		return false
	}
	return false
}

func (se *ScriptEngine) checkPKPoint(params []string) bool {
	if len(params) < 2 {
		return false
	}
	return false
}

func (se *ScriptEngine) checkDead(params []string) bool {
	return false
}

func (se *ScriptEngine) checkCastle(params []string) bool {
	return false
}

func (se *ScriptEngine) checkQuest(params []string) bool {
	return false
}

func (se *ScriptEngine) checkRange(params []string) bool {
	if len(params) < 4 {
		return false
	}
	return false
}

func (se *ScriptEngine) checkSkill(params []string) bool {
	if len(params) < 1 {
		return false
	}
	return false
}

func (se *ScriptEngine) checkMonster(params []string) bool {
	if len(params) < 1 {
		return false
	}
	return false
}

func (se *ScriptEngine) checkGuildMember(params []string) bool {
	return false
}

func (se *ScriptEngine) checkRiding() bool {
	return false
}

func (se *ScriptEngine) checkStore(params []string) bool {
	return false
}

func (se *ScriptEngine) checkRepair(params []string) bool {
	return false
}

func (se *ScriptEngine) doGiveItem(itemName string, count int) bool {
	return true
}

func (se *ScriptEngine) doTakeItem(itemName string, count int) bool {
	return true
}

func (se *ScriptEngine) doAddGold(amount int) bool {
	return true
}

func (se *ScriptEngine) doTakeGold(amount int) bool {
	return true
}

func (se *ScriptEngine) doMapMove(mapName string, x, y int) bool {
	return true
}

func (se *ScriptEngine) doSetLevel(level int) bool {
	return true
}

func (se *ScriptEngine) doSetPK(pkPoint int) bool {
	return true
}

func (se *ScriptEngine) doSetSkill(skillName string, level int) bool {
	return true
}

func (se *ScriptEngine) doAddSkill(skillName string) bool {
	return true
}

func (se *ScriptEngine) doSendMessage(msgType int, message string) bool {
	return true
}

func (se *ScriptEngine) skipToEndif() {
	if se.CurScript == nil {
		return
	}
	
	depth := 1
	for se.CurLine < len(se.CurScript.Lines) {
		se.CurLine++
		if se.CurLine >= len(se.CurScript.Lines) {
			break
		}
		
		line := se.CurScript.Lines[se.CurLine]
		if line.Command == "IF" {
			depth++
		} else if line.Command == "ENDIF" {
			depth--
			if depth == 0 {
				break
			}
		}
	}
}

func (se *ScriptEngine) GetVariable(name string) interface{} {
	if val, ok := se.Variables[name]; ok {
		return val
	}
	return nil
}

func (se *ScriptEngine) SetVariable(name string, value interface{}) {
	se.Variables[name] = value
}

type NpcScript struct {
	Script   *Script
	NpcName  string
	MapName  string
	X, Y     int
}

func ParseNpcScript(content string) (*NpcScript, error) {
	engine := NewScriptEngine()
	script, err := engine.ParseScript("npc", content)
	if err != nil {
		return nil, err
	}
	
	return &NpcScript{
		Script: script,
	}, nil
}

func ParseQuestScript(content string) (*QuestInfo, error) {
	qi := &QuestInfo{
		boQuest:    true,
		QuestList:  make([]QuestCondition, 0),
		RecordList: make([]string, 0),
	}
	
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, ";") {
			continue
		}
		
		re := regexp.MustCompile(`(\d+)\s*=\s*(\d+)`)
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			flag, _ := strconv.ParseUint(matches[1], 10, 16)
			val, _ := strconv.ParseUint(matches[2], 10, 8)
			qi.QuestList = append(qi.QuestList, QuestCondition{
				Flag:  uint16(flag),
				Value: byte(val),
			})
		}
	}
	
	return qi, nil
}
