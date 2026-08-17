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
	Login(ctx context.Context, data *packet.Login) (bool, string, error)
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
) (*packet.CreateAccountResult, error) {
	createResult, err := h.as.CreateNewAccount(ctx, data)
	if !createResult || err != nil {
		utils.Log().Warnf("create new account failed, accountId: %s, err: %v", data.AccountId, err)
		return &packet.CreateAccountResult{
			BaseResult: packet.BaseResult{
				ResultCode: constants.CreateAccountFailed,
			},
			AccountId: data.AccountId,
		}, nil
	}
	return &packet.CreateAccountResult{
		BaseResult: packet.BaseResult{
			ResultCode: constants.Success,
		},
		AccountId: data.AccountId,
	}, nil
}

func (h *AccountHandler) Login(
	ctx context.Context,
	data *packet.Login,
) (*packet.LoginResult, error) {

	session := utils.GetSessionFromCtx(ctx)
	loginResult, userUid, err := h.as.Login(ctx, data)

	if !loginResult || err != nil {
		utils.Log().Warnf("login failed, accountId: %s, err: %v", data.AccountId, err)
		return &packet.LoginResult{
			BaseResult: packet.BaseResult{
				ResultCode: constants.LoginFailed,
			},
			AccountId: data.AccountId,
		}, nil
	}

	// 登入成功，將 session 與 驗證結果 綁定到帳號 ID
	session.Set("auth", true)
	session.Bind(ctx, userUid)
	utils.Log().Infof("login success, accountId: %s, UID: %s", data.AccountId, session.UID())
	return &packet.LoginResult{
		BaseResult: packet.BaseResult{
			ResultCode: constants.Success,
		},
		AccountId: data.AccountId,
	}, nil
}
