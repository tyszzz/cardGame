package packet

import "cardGame/model/constants"

type (
	// base packet
	BaseResult struct {
		ResultCode int32 `json:"resultCode"`
	}

	// account packet
	CreateAccountResult struct {
		BaseResult
		AccountId string `json:"accountId"`
	}

	LoginResult struct {
		BaseResult
		AccountId string `json:"accountId"`
	}

	// summon packet
	SummonCardResult struct {
		BaseResult
		PoolId string           `json:"poolId"`
		Cards  []SummonCardItem `json:"cards"`
		Pity   *SummonPityInfo  `json:"pity,omitempty"` // 只有設定保底的卡池才有
	}

	// SummonCardItem 為單次抽取結果，重複的卡片 isNew 為 false 並帶碎片獎勵
	SummonCardItem struct {
		CardId      string           `json:"cardId"`
		CardName    string           `json:"cardName"`
		Rarity      constants.Rarity `json:"rarity"`
		IsNew       bool             `json:"isNew"`
		IsUp        bool             `json:"isUp"`
		ShardId     string           `json:"shardId,omitempty"`
		ShardAmount int64            `json:"shardAmount,omitempty"`
	}

	// SummonPityInfo 為抽完後的保底進度
	SummonPityInfo struct {
		PullsSinceHigh int32 `json:"pullsSinceHigh"`
		Threshold      int32 `json:"threshold"`
		GuaranteedUp   bool  `json:"guaranteedUp"`
	}

	// player packet
	Currencies struct {
		Coin        int64 `json:"coin"`
		Diamond     int64 `json:"diamond"`
		FriendPoint int64 `json:"friendPoint"`
	}

	GetInfoResult struct {
		BaseResult
		Nickname   string     `json:"nickname"`
		Level      int32      `json:"level"`
		VipLevel   int32      `json:"vipLevel"`
		Currencies Currencies `json:"currencies"`
	}

	// inventory packet
	GetBackpackResult struct {
		BaseResult
		Capacity int            `json:"capacity"` // 背包道具種類上限
		Used     int            `json:"used"`     // 目前持有的道具種類數
		Items    []BackpackItem `json:"items"`
	}

	BackpackItem struct {
		ItemId   string             `json:"itemId"`
		ItemName string             `json:"itemName"`
		ItemType constants.ItemType `json:"itemType"`
		Rarity   constants.Rarity   `json:"rarity"`
		Quantity int64              `json:"quantity"`
	}

	GetCardsResult struct {
		BaseResult
		MaxLevel int              `json:"maxLevel"`
		ExpOwned int64            `json:"expOwned"` // 目前持有的卡片經驗道具數量
		Items    map[string]int64 `json:"items"`    // 背包道具數量（itemId -> 數量），供前端判斷進階與升星材料是否足夠
		Cards    []OwnedCard      `json:"cards"`
	}

	OwnedCard struct {
		CardId     string           `json:"cardId"`
		CardName   string           `json:"cardName"`
		Rarity     constants.Rarity `json:"rarity"`
		Race       constants.Race   `json:"race"`
		Level      int32            `json:"level"`
		Star       int              `json:"star"`
		Stage      int              `json:"stage"`
		MaxStar    int              `json:"maxStar"`
		MaxLevel   int              `json:"maxLevel"` // 目前的等級上限（階與星數取較小）
		LimitBy    string           `json:"limitBy"`  // 到達上限時的原因：stage（需進階）、star（需升星）、max（已滿）
		IsLocked   bool             `json:"isLocked"`
		ObtainedAt int64            `json:"obtainedAt"` // unix 秒
		Attack     int              `json:"attack"`
		Defense    int              `json:"defense"`
		Health     int              `json:"health"`
		NextExp    int64            `json:"nextExp"`          // 升到下一級所需經驗，已達上限為 0
		NextExp10  int64            `json:"nextExp10"`        // 連續升 Levels10 級所需的總經驗，無法升級時為 0
		Levels10   int              `json:"levels10"`         // 「升 10 級」實際會升的級數（受等級上限截斷）
		CanAscend  bool             `json:"canAscend"`        // 等級已到當前階的上限且還有下一階，可以進階
		Ascend     *AscendCost      `json:"ascend,omitempty"` // 進階所需材料，已是最高階時沒有
		StarUp     []CostItem       `json:"starUp,omitempty"` // 升星所需材料，已達星數上限時沒有
		Skills     []CardSkill      `json:"skills"`
	}

	// CostItem 為一項道具消耗
	CostItem struct {
		ItemId   string `json:"itemId"`
		ItemName string `json:"itemName"`
		Quantity int64  `json:"quantity"`
	}

	// AscendCost 為進階所需材料：經驗道具數量與進階道具；進階還需要等級已到當前階的上限
	AscendCost struct {
		Exp   int64      `json:"exp"`
		Items []CostItem `json:"items"`
	}

	// CardSkill 為卡片技能的顯示資料，type 為 normalAtk、activeSkill、passiveSkill 或 talent
	CardSkill struct {
		Type string `json:"type"`
		Name string `json:"name"`
		Desc string `json:"desc"`
		// 解鎖條件，0 代表沒有該條件
		UnlockStar  int `json:"unlockStar,omitempty"`
		UnlockLevel int `json:"unlockLevel,omitempty"`
	}

	UpgradeCardResult struct {
		BaseResult
		CardId       string `json:"cardId"`
		Level        int32  `json:"level"`
		Levels       int    `json:"levels"` // 這次實際升的級數（受等級上限截斷）
		Star         int    `json:"star"`
		Stage        int    `json:"stage"`
		MaxLevel     int    `json:"maxLevel"`
		LimitBy      string `json:"limitBy"`
		Attack       int    `json:"attack"`
		Defense      int    `json:"defense"`
		Health       int    `json:"health"`
		NextExp      int64  `json:"nextExp"`      // 升到下一級所需經驗，已達上限為 0
		NextExp10    int64  `json:"nextExp10"`    // 連續升 Levels10 級所需的總經驗，無法升級時為 0
		Levels10     int    `json:"levels10"`     // 「升 10 級」實際會升的級數
		ExpRemaining int64  `json:"expRemaining"` // 升級後剩餘的經驗道具數量
	}

	// CardProgressResult 為進階與升星的結果
	CardProgressResult struct {
		BaseResult
		CardId   string `json:"cardId"`
		Level    int    `json:"level"`
		Star     int    `json:"star"`
		Stage    int    `json:"stage"`
		MaxLevel int    `json:"maxLevel"`
		LimitBy  string `json:"limitBy"`
		Attack   int    `json:"attack"`
		Defense  int    `json:"defense"`
		Health   int    `json:"health"`
	}
)

type (
	// BattleStage 為可挑戰的關卡
	BattleStage struct {
		StageId string `json:"stageId"`
		Name    string `json:"name"`
		Desc    string `json:"desc"`
	}

	GetStagesResult struct {
		BaseResult
		Stages []BattleStage `json:"stages"`
	}

	GetLineupResult struct {
		BaseResult
		Slots []LineupSlot `json:"slots"`
	}

	SaveLineupResult struct {
		BaseResult
	}

	// BattleUnit 為開戰時的單位，前端依此建立畫面後再套用 events
	BattleUnit struct {
		Team      string           `json:"team"` // Ally 或 Enemy
		Position  int              `json:"position"`
		CardId    string           `json:"cardId"`
		Name      string           `json:"name"`
		Rarity    constants.Rarity `json:"rarity"`
		MaxHealth int              `json:"maxHealth"`
		Health    int              `json:"health"`
		Rage      int              `json:"rage"`
	}

	// BattleEvent 為事件日誌的一筆；type 為 BattleStart、RoundStart、NormalAtk、ActiveSkill、
	// Passive、Damage、Heal、Rage、Death、RoundEnd、BattleEnd、ActionRage。
	// 沒有對象時 team 為空字串；Damage/Heal 的 remain 為作用後血量，Rage/ActionRage 的 remain 為怒氣。
	BattleEvent struct {
		Round      int    `json:"round"`
		Type       string `json:"type"`
		ActorTeam  string `json:"actorTeam,omitempty"`
		ActorPos   int    `json:"actorPos,omitempty"`
		TargetTeam string `json:"targetTeam,omitempty"`
		TargetPos  int    `json:"targetPos,omitempty"`
		Skill      string `json:"skill,omitempty"`
		Value      int    `json:"value"`
		Remain     int    `json:"remain"`
	}

	StartBattleResult struct {
		BaseResult
		StageId       string        `json:"stageId"`
		Winner        string        `json:"winner"` // Ally 或 Enemy
		Rounds        int           `json:"rounds"`
		RageThreshold int           `json:"rageThreshold"` // 施放主動技能所需怒氣，前端以此畫怒氣條
		Units         []BattleUnit  `json:"units"`
		Events        []BattleEvent `json:"events"`
	}
)
