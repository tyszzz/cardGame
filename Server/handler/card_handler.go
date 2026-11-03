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
		resultCode := int32(constants.CardUpgradeFailed)
		switch {
		case errors.Is(err, inventory.ErrCardNotOwned), errors.Is(err, inventory.ErrInvalidUpgradeLevels):
			resultCode = constants.CardUpgradeInvalidRequest
		case errors.Is(err, inventory.ErrCardMaxLevel):
			resultCode = constants.CardUpgradeMaxLevel
		case errors.Is(err, inventory.ErrNotEnoughExp):
			resultCode = constants.CardUpgradeNotEnoughExp
		}
		return upgradeFailed(resultCode), nil
	}
	utils.Log().Infof("UserUid: %s, UpgradeCard success, card: %s, level: %d", uid, result.CardId, result.Level)
	return result, nil
}
