package packet

import "cardGame/model/constants"

// base packet
type (
	BaseResult struct {
		ResultCode int32 `json:"resultCode"`
	}
)

// account packet
type (
	CreateAccountResult struct {
		BaseResult
		AccountId string `json:"accountId"`
	}

	LoginResult struct {
		BaseResult
		AccountId string `json:"accountId"`
	}
)

// summon packet
type (
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
)

// player packet
type (
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
)

// inventory packet
type (
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
		Cards []OwnedCard `json:"cards"`
	}

	OwnedCard struct {
		CardId     string           `json:"cardId"`
		CardName   string           `json:"cardName"`
		Rarity     constants.Rarity `json:"rarity"`
		Race       constants.Race   `json:"race"`
		Level      int32            `json:"level"`
		IsLocked   bool             `json:"isLocked"`
		ObtainedAt int64            `json:"obtainedAt"` // unix 秒
	}
)
