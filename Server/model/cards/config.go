package cards

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
		Name      string        `json:"name"`                // 技能名字
		Desc      string        `json:"desc"`                // 技能描述
		Trigger   string        `json:"trigger"`             // 觸發方式
		Target    string        `json:"target"`              // 目標對象
		Condition *Condition    `json:"condition,omitempty"` // 觸發條件
		Unlock    *UnlockRule   `json:"unlock,omitempty"`    // 解鎖條件
		Effects   []SkillEffect `json:"effects"`             // 技能效果
	}

	// 觸發條件
	Condition struct {
		Value float64 `json:"value,omitempty"` // 閾值
		Once  bool    `json:"once,omitempty"`  // 是否只觸發一次
	}

	// 解鎖條件
	UnlockRule struct {
		Star  int `json:"star,omitempty"`  // 解鎖星數
		Level int `json:"level,omitempty"` // 解鎖等級
	}

	// 技能效果
	SkillEffect struct {
		Type       string `json:"type"`                 // 效果類型
		Stat       string `json:"stat,omitempty"`       // buff影響的屬性(扣攻,扣防等)
		Status     string `json:"status,omitempty"`     // buff影響狀態(中毒,燃燒等)
		DamageType string `json:"damageType,omitempty"` // 傷害類型(物理,魔法等)
		Target     string `json:"target,omitempty"`     // 作用對象
		Value      int    `json:"value,omitempty"`      // 效果數值
		Duration   int    `json:"duration,omitempty"`   // 持續回合
		MaxStack   int    `json:"maxStack,omitempty"`   // 可疊加層數
	}
)
