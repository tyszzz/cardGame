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
		MaxLevel int         `json:"maxLevel"`
		ExpOwned int64       `json:"expOwned"` // 目前持有的卡片經驗道具數量
		Cards    []OwnedCard `json:"cards"`
	}

	OwnedCard struct {
		CardId     string           `json:"cardId"`
		CardName   string           `json:"cardName"`
		Rarity     constants.Rarity `json:"rarity"`
		Race       constants.Race   `json:"race"`
		Level      int32            `json:"level"`
		IsLocked   bool             `json:"isLocked"`
		ObtainedAt int64            `json:"obtainedAt"` // unix 秒
		Attack     int              `json:"attack"`
		Defense    int              `json:"defense"`
		Health     int              `json:"health"`
		NextExp    int64            `json:"nextExp"`   // 升到下一級所需經驗，已達上限為 0
		NextExp10  int64            `json:"nextExp10"` // 連續升 10 級所需的總經驗，等級不足 10 級可升時為 0
		Skills     []CardSkill      `json:"skills"`
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
		Attack       int    `json:"attack"`
		Defense      int    `json:"defense"`
		Health       int    `json:"health"`
		NextExp      int64  `json:"nextExp"`      // 升到下一級所需經驗，已達上限為 0
		NextExp10    int64  `json:"nextExp10"`    // 連續升 10 級所需的總經驗，等級不足 10 級可升時為 0
		ExpRemaining int64  `json:"expRemaining"` // 升級後剩餘的經驗道具數量
	}
)
