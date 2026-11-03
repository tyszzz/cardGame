package handler

import (
	"cardGame/feature/inventory"
	"cardGame/model/constants"
	"cardGame/model/packet"
	"cardGame/utils"
	"context"
	"errors"

	"github.com/topfreegames/pitaya/v2/component"
)

type (
	CardService interface {
		GetCards(ctx context.Context, uid string) (*packet.GetCardsResult, error)
		UpgradeCard(ctx context.Context, uid string, data *packet.UpgradeCard) (*packet.UpgradeCardResult, error)
		AscendCard(ctx context.Context, uid string, data *packet.AscendCard) (*packet.CardProgressResult, error)
		StarUpCard(ctx context.Context, uid string, data *packet.StarUpCard) (*packet.CardProgressResult, error)
	}

	CardHandler struct {
		component.Base
		cs CardService
	}
)

func NewCardHandler(cs CardService) *CardHandler {
	return &CardHandler{cs: cs}
}

// inner function

func upgradeFailed(resultCode int32) *packet.UpgradeCardResult {
	return &packet.UpgradeCardResult{BaseResult: packet.BaseResult{ResultCode: resultCode}}
}

func progressFailed(resultCode int32) *packet.CardProgressResult {
	return &packet.CardProgressResult{BaseResult: packet.BaseResult{ResultCode: resultCode}}
}

// cardResultCode 把卡片升級、進階、升星的錯誤轉成 resultCode；notEnough 為材料不足時使用的代碼。
func cardResultCode(err error, notEnough int32) int32 {
	switch {
	case errors.Is(err, inventory.ErrCardNotOwned), errors.Is(err, inventory.ErrInvalidUpgradeLevels):
		return constants.CardUpgradeInvalidRequest
	case errors.Is(err, inventory.ErrCardMaxLevel):
		return constants.CardUpgradeMaxLevel
	case errors.Is(err, inventory.ErrCardNeedAscend):
		return constants.CardNeedAscend
	case errors.Is(err, inventory.ErrCardNeedStar):
		return constants.CardNeedStar
	case errors.Is(err, inventory.ErrAscendNotReady):
		return constants.CardAscendNotReady
	case errors.Is(err, inventory.ErrAlreadyTopStage):
		return constants.CardAlreadyTopStage
	case errors.Is(err, inventory.ErrStarMax):
		return constants.CardStarMax
	case errors.Is(err, inventory.ErrNotEnoughItems):
		return notEnough
	}
	return constants.CardUpgradeFailed
}

// outer function

// GetCards 回傳目前登入玩家持有的卡片，uid 以登入後綁定的 session 為準。
func (h *CardHandler) GetCards(ctx context.Context, _ *packet.Empty) (*packet.GetCardsResult, error) {
	uid := utils.GetSessionFromCtx(ctx).UID()
	if uid == "" {
		utils.Log().Warnf("GetCards rejected: session not logged in")
		return &packet.GetCardsResult{BaseResult: packet.BaseResult{ResultCode: constants.NotLoggedIn}, Cards: []packet.OwnedCard{}}, nil
	}
	result, err := h.cs.GetCards(ctx, uid)
	if err != nil {
		utils.Log().Warnf("UserUid: %s, GetCards failed, err: %v", uid, err)
		return &packet.GetCardsResult{BaseResult: packet.BaseResult{ResultCode: constants.PlayerQueryFailed}, Cards: []packet.OwnedCard{}}, nil
	}
	return result, nil
}

// UpgradeCard 消耗經驗道具將卡片升 1 級或 10 級，uid 以登入後綁定的 session 為準。
func (h *CardHandler) UpgradeCard(ctx context.Context, data *packet.UpgradeCard) (*packet.UpgradeCardResult, error) {
	uid := utils.GetSessionFromCtx(ctx).UID()
	if uid == "" {
		utils.Log().Warnf("UpgradeCard rejected: session not logged in")
		return upgradeFailed(constants.NotLoggedIn), nil
	}
	result, err := h.cs.UpgradeCard(ctx, uid, data)
	if err != nil {
		utils.Log().Warnf("UserUid: %s, UpgradeCard failed, err: %v", uid, err)
		return upgradeFailed(cardResultCode(err, constants.CardUpgradeNotEnoughExp)), nil
	}
	utils.Log().Infof("UserUid: %s, UpgradeCard success, card: %s, level: %d", uid, result.CardId, result.Level)
	return result, nil
}

// AscendCard 消耗進階丹與經驗將卡片進階一階，uid 以登入後綁定的 session 為準。
func (h *CardHandler) AscendCard(ctx context.Context, data *packet.AscendCard) (*packet.CardProgressResult, error) {
	uid := utils.GetSessionFromCtx(ctx).UID()
	if uid == "" {
		utils.Log().Warnf("AscendCard rejected: session not logged in")
		return progressFailed(constants.NotLoggedIn), nil
	}
	result, err := h.cs.AscendCard(ctx, uid, data)
	if err != nil {
		utils.Log().Warnf("UserUid: %s, AscendCard failed, err: %v", uid, err)
		return progressFailed(cardResultCode(err, constants.CardNotEnoughMaterial)), nil
	}
	utils.Log().Infof("UserUid: %s, AscendCard success, card: %s, stage: %d", uid, result.CardId, result.Stage)
	return result, nil
}

// StarUpCard 消耗碎片將卡片升一星，uid 以登入後綁定的 session 為準。
func (h *CardHandler) StarUpCard(ctx context.Context, data *packet.StarUpCard) (*packet.CardProgressResult, error) {
	uid := utils.GetSessionFromCtx(ctx).UID()
	if uid == "" {
		utils.Log().Warnf("StarUpCard rejected: session not logged in")
		return progressFailed(constants.NotLoggedIn), nil
	}
	result, err := h.cs.StarUpCard(ctx, uid, data)
	if err != nil {
		utils.Log().Warnf("UserUid: %s, StarUpCard failed, err: %v", uid, err)
		return progressFailed(cardResultCode(err, constants.CardNotEnoughMaterial)), nil
	}
	utils.Log().Infof("UserUid: %s, StarUpCard success, card: %s, star: %d", uid, result.CardId, result.Star)
	return result, nil
}
