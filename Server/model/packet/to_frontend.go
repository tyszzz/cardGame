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
	GetCurrenciesResult struct {
		BaseResult
		Coin        int64 `json:"coin"`
		Diamond     int64 `json:"diamond"`
		FriendPoint int64 `json:"friendPoint"`
	}
)
