package cards

import "cardGame/model/constants"

type (
	BaseConfig struct {
		Attack       int     `json:"attack"`
		AttackRatio  float64 `json:"attackRatio"` // 升級時攻擊力增加的比例
		Defense      int     `json:"defense"`
		DefenseRatio float64 `json:"defenseRatio"` // 升級時防禦力增加的比例
		Health       int     `json:"health"`
		HealthRatio  float64 `json:"healthRatio"` // 升級時生命值增加的比例
	}

	SkillConfig struct {
		NormalAtk    SkillDefinition   `json:"normalAtk"`
		ActiveSkill  SkillDefinition   `json:"activeSkill"`
		PassiveSkill SkillDefinition   `json:"passiveSkill"`
		TalentSkill  []SkillDefinition `json:"talentSkill"`
	}

	SkillDefinition struct {
		Name    string            `json:"name"`              // 技能名字
		Desc    string            `json:"desc"`              // 技能描述
		Trigger constants.Trigger `json:"trigger,omitempty"` // 觸發時機，僅被動與天賦使用；普攻與主動由行動流程與怒氣決定
		Target  constants.Target  `json:"target,omitempty"`  // 預設目標形狀，效果未指定時沿用
		Unlock  *UnlockRule       `json:"unlock,omitempty"`  // 解鎖條件
		Effects []SkillEffect     `json:"effects"`           // 技能效果
	}

	// 解鎖條件
	UnlockRule struct {
		Star  int `json:"star,omitempty"`  // 解鎖星數
		Level int `json:"level,omitempty"` // 解鎖等級
	}

	// 技能效果，百分比數值以整數表示(150 代表 150%)
	SkillEffect struct {
		Type       constants.EffectType `json:"type,omitempty"`       // 效果類型
		Side       constants.Side       `json:"side,omitempty"`       // 作用陣營(敵方,我方,自己)
		Target     constants.Target     `json:"target,omitempty"`     // 作用目標形狀，未填沿用技能的 Target
		DamageType constants.DamageType `json:"damageType,omitempty"` // 傷害類型(物理,魔法等)
		Stat       constants.Stat       `json:"stat,omitempty"`       // buff影響的屬性(扣攻,扣防等)
		Status     constants.Status     `json:"status,omitempty"`     // 狀態(中毒,燃燒等)
		ScaleStat  constants.Stat       `json:"scaleStat,omitempty"`  // Value 以施放者哪個屬性為基準
		Value      int                  `json:"value,omitempty"`      // 效果數值：Rage 類型為固定怒氣點數(可為負)，其餘為百分比
		Duration   int                  `json:"duration,omitempty"`   // 持續回合
		MaxStack   int                  `json:"maxStack,omitempty"`   // 可疊加層數
	}
)
