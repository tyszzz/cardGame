package handler

import (
	"cardGame/model/constants"
	"cardGame/model/packet"
	"cardGame/utils"
	"context"

	"github.com/topfreegames/pitaya/v2/component"
)

type AccountService interface {
	CreateNewAccount(ctx context.Context, data *packet.CreateNewAccount) (bool, error)
	Login(ctx context.Context, data *packet.Login) (bool, error)
}
type (
	AccountHandler struct {
		component.Base
		as AccountService
	}
)

func NewAccountHandler(as AccountService) *AccountHandler {
	return &AccountHandler{
		as: as,
	}
}

func (h *AccountHandler) CreateNewAccount(
	ctx context.Context,
	data *packet.CreateNewAccount,
) (bool, error) {
	return h.as.CreateNewAccount(ctx, data)
}

func (h *AccountHandler) Login(
	ctx context.Context,
	data *packet.Login,
) (*packet.LoginResult, error) {

	session := utils.GetSessionFromCtx(ctx)
	loginResult, err := h.as.Login(ctx, data)

	if !loginResult || err != nil {
		utils.Log().Warnf("login failed, accountId: %s, err: %v", data.AccountId, err)
		return &packet.LoginResult{
			ResultCode: constants.LoginFailed,
			AccountId:  data.AccountId,
		}, nil
	}

	// 登入成功，將 session 與 驗證結果 綁定到帳號 ID
	session.Set("auth", true)
	session.Bind(ctx, data.AccountId)
	utils.Log().Infof("login success, accountId: %s, UID: %s", data.AccountId, session.UID())
	return &packet.LoginResult{
		ResultCode: constants.Success,
		AccountId:  data.AccountId,
	}, nil
}
