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
		Cards []SummonCardItem `json:"cards"`
	}

	// SummonCardItem 為單次抽取結果，重複的卡片 isNew 為 false 並帶碎片獎勵
	SummonCardItem struct {
		CardId      string           `json:"cardId"`
		CardName    string           `json:"cardName"`
		Rarity      constants.Rarity `json:"rarity"`
		IsNew       bool             `json:"isNew"`
		ShardId     string           `json:"shardId,omitempty"`
		ShardAmount int64            `json:"shardAmount,omitempty"`
	}
)
