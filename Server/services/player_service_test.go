package services

import (
	"cardGame/feature/player"
	"cardGame/model/constants"
	"context"
	"errors"
	"testing"
)

type fakePlayerRepo struct {
	player *player.Player
	err    error
}

func (f *fakePlayerRepo) GetPlayer(context.Context, string) (*player.Player, error) {
	return f.player, f.err
}

func TestPlayerServiceGetInfo(t *testing.T) {
	s := NewPlayerService(&fakePlayerRepo{player: &player.Player{
		UID: "u1", Nickname: "tester", Level: 3, VipLevel: 1,
		Currencies: map[constants.CurrencyType]int64{
			constants.CurrencyCoin:    5000,
			constants.CurrencyDiamond: 30,
		},
	}})
	got, err := s.GetInfo(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Nickname != "tester" || got.Level != 3 || got.VipLevel != 1 {
		t.Errorf("unexpected profile: %+v", got)
	}
	if got.Currencies.Coin != 5000 || got.Currencies.Diamond != 30 || got.Currencies.FriendPoint != 0 {
		t.Errorf("unexpected currencies: %+v", got.Currencies)
	}
}

func TestPlayerServiceGetInfoError(t *testing.T) {
	s := NewPlayerService(&fakePlayerRepo{err: player.ErrPlayerNotFound})
	if _, err := s.GetInfo(context.Background(), "u1"); !errors.Is(err, player.ErrPlayerNotFound) {
		t.Errorf("err = %v, want ErrPlayerNotFound", err)
	}
}
