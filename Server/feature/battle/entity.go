package battle

import (
	"cardGame/model/cards"
	"cardGame/model/constants"
)

type (
	// Team 是戰鬥中的陣營。我方（Ally）為進攻方，敵方（Enemy）為守方。
	Team int

	// EventType 是事件日誌的種類。
	EventType int

	// Ref 以陣營與位置指向一個單位，位置為 1~6（前排 1~3、後排 4~6）。
	Ref struct {
		Team     Team `json:"team"`
		Position int  `json:"position"`
	}

	// UnitInput 是進入戰鬥的單位，三圍由呼叫端算好等級後傳入。
	UnitInput struct {
		ID       string
		Position int
		Attack   int
		Defense  int
		Health   int
		Star     int // 判斷天賦是否解鎖
		Level    int // 判斷天賦是否解鎖
		Skills   cards.SkillConfig
	}

	// Input 是一場戰鬥的輸入。同樣的輸入（含 Seed）一定得到同樣的結果。
	Input struct {
		Ally       []UnitInput
		Enemy      []UnitInput
		EnemyFirst bool // 預設我方先手
		Seed       int64
	}

	// Event 是事件日誌的一筆，前端播放戰鬥即依此日誌。
	Event struct {
		Round  int       `json:"round"` // 0 代表戰鬥開始前
		Type   EventType `json:"type"`
		Actor  Ref       `json:"actor"`
		Target Ref       `json:"target"`
		Skill  string    `json:"skill,omitempty"`
		Value  int       `json:"value"`  // 傷害、實際回復量、怒氣變化
		Remain int       `json:"remain"` // 作用後目標的血量（怒氣事件為怒氣）
	}

	// Result 是戰鬥結果。
	Result struct {
		Winner Team
		Rounds int
		Events []Event
	}

	// unit 是戰鬥中的單位狀態。
	unit struct {
		ref           Ref
		id            string
		attack        int
		defense       int
		maxHealth     int
		health        int
		star          int
		level         int
		rage          int
		rageThreshold int // 施放主動技能所需怒氣，之後天賦可調整
		skills        cards.SkillConfig
	}
)

const (
	TeamAlly  Team = 1
	TeamEnemy Team = 2
)

const (
	EventBattleStart EventType = iota + 1
	EventRoundStart
	EventNormalAtk
	EventActiveSkill
	EventPassive
	EventDamage
	EventHeal
	EventRage
	EventDeath
	EventRoundEnd
	EventBattleEnd
)

const (
	maxPosition   = 6 // 每隊最多 6 個位置
	frontRowLast  = 3 // 1~3 前排，4~6 後排
	columnCount   = 3 // 位置 p 的縱排為 (p-1)%3
	random3Number = 3
)

// inner function

func (u *unit) alive() bool {
	return u.health > 0
}

func (u *unit) column() int {
	return (u.ref.Position - 1) % columnCount
}

func (u *unit) frontRow() bool {
	return u.ref.Position <= frontRowLast
}

func (u *unit) stat(s constants.Stat) int {
	switch s {
	case constants.StatHealth:
		return u.maxHealth
	case constants.StatDefense:
		return u.defense
	default:
		return u.attack
	}
}

// passives 回傳已解鎖的被動與天賦。
func (u *unit) passives() []cards.SkillDefinition {
	list := []cards.SkillDefinition{u.skills.PassiveSkill}
	list = append(list, u.skills.TalentSkill...)
	out := list[:0:0]
	for _, s := range list {
		if s.Unlock == nil || (u.star >= s.Unlock.Star && u.level >= s.Unlock.Level) {
			out = append(out, s)
		}
	}
	return out
}

// outer function

func (t Team) String() string {
	switch t {
	case TeamAlly:
		return "Ally"
	case TeamEnemy:
		return "Enemy"
	}
	return "Team(?)"
}

func (e EventType) String() string {
	names := [...]string{"", "BattleStart", "RoundStart", "NormalAtk", "ActiveSkill", "Passive", "Damage", "Heal", "Rage", "Death", "RoundEnd", "BattleEnd"}
	if e > 0 && int(e) < len(names) {
		return names[e]
	}
	return "EventType(?)"
}
