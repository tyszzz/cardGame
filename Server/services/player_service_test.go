package services

import (
	"cardGame/model/constants"
	"context"
	"errors"
	"testing"
)

type fakePlayerRepo struct {
	currencies map[constants.CurrencyType]int64
	err        error
}

func (f *fakePlayerRepo) GetCurrencies(context.Context, string) (map[constants.CurrencyType]int64, error) {
	return f.currencies, f.err
}

func TestPlayerServiceGetCurrencies(t *testing.T) {
	s := NewPlayerService(&fakePlayerRepo{currencies: map[constants.CurrencyType]int64{
		constants.CurrencyCoin:    5000,
		constants.CurrencyDiamond: 30,
	}})
	got, err := s.GetCurrencies(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Coin != 5000 || got.Diamond != 30 || got.FriendPoint != 0 {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestPlayerServiceGetCurrenciesError(t *testing.T) {
	boom := errors.New("boom")
	s := NewPlayerService(&fakePlayerRepo{err: boom})
	if _, err := s.GetCurrencies(context.Background(), "u1"); !errors.Is(err, boom) {
		t.Errorf("err = %v, want %v", err, boom)
	}
}
