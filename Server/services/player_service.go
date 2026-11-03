package services

import (
	"cardGame/feature/player"
	"cardGame/model/constants"
	"cardGame/model/packet"
	"context"
)

type PlayerService struct {
	repo player.PlayerRepository
}

func NewPlayerService(repo player.PlayerRepository) *PlayerService {
	return &PlayerService{repo: repo}
}

// outer function

func (s *PlayerService) GetInfo(ctx context.Context, uid string) (*packet.GetInfoResult, error) {
	p, err := s.repo.GetPlayer(ctx, uid)
	if err != nil {
		return nil, err
	}
	return &packet.GetInfoResult{
		Nickname: p.Nickname,
		Level:    p.Level,
		VipLevel: p.VipLevel,
		Currencies: packet.Currencies{
			Coin:        p.Currencies[constants.CurrencyCoin],
			Diamond:     p.Currencies[constants.CurrencyDiamond],
			FriendPoint: p.Currencies[constants.CurrencyFriendPoint],
		},
	}, nil
}
