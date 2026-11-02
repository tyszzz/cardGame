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

func (s *PlayerService) GetCurrencies(ctx context.Context, uid string) (*packet.GetCurrenciesResult, error) {
	currencies, err := s.repo.GetCurrencies(ctx, uid)
	if err != nil {
		return nil, err
	}
	return &packet.GetCurrenciesResult{
		Coin:        currencies[constants.CurrencyCoin],
		Diamond:     currencies[constants.CurrencyDiamond],
		FriendPoint: currencies[constants.CurrencyFriendPoint],
	}, nil
}
