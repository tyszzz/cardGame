package services

import (
	"cardGame/feature/account"
	"cardGame/model/constants"
	"errors"
	"testing"
)

func TestValidateAccountID(t *testing.T) {
	cases := []struct {
		in   string
		want error
	}{
		{"abc", nil},
		{"Abc123", nil},
		{"ab", account.ErrAccountIDLength},
		{"abcdefghijklmnopqrstu", account.ErrAccountIDLength},
		{"ab c", account.ErrAccountIDFormat},
		{"abc_1", account.ErrAccountIDFormat},
	}
	for _, c := range cases {
		if got := validateAccountID(c.in); !errors.Is(got, c.want) {
			t.Errorf("validateAccountID(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		in   string
		want error
	}{
		{"Passw0rd!", nil},
		{"short1!", account.ErrPasswordLength},
		{"12345678901234567", account.ErrPasswordLength},
		{"pass word1", account.ErrPasswordFormat},
		{"密碼密碼密碼", account.ErrPasswordLength},
	}
	for _, c := range cases {
		if got := validatePassword(c.in); !errors.Is(got, c.want) {
			t.Errorf("validatePassword(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestValidateNickname(t *testing.T) {
	cases := []struct {
		in   string
		want error
	}{
		{"Hero", nil},
		{"勇者01", nil},
		{"a", account.ErrNicknameLength},
		{"一二三四五六七八九十一二三", account.ErrNicknameLength},
		{"ab cd", account.ErrNicknameFormat},
		{"hero_1", account.ErrNicknameFormat},
	}
	for _, c := range cases {
		if got := validateNickname(c.in); !errors.Is(got, c.want) {
			t.Errorf("validateNickname(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestNewPlayerDefaults(t *testing.T) {
	p := newPlayer("uid", "Hero")
	if p.Level != 1 || p.VipLevel != 0 {
		t.Errorf("unexpected defaults: level=%d vip=%d", p.Level, p.VipLevel)
	}
	if len(p.Currencies) != len(constants.AllCurrencyTypes) {
		t.Errorf("currencies = %d, want %d", len(p.Currencies), len(constants.AllCurrencyTypes))
	}
	for currencyType, amount := range p.Currencies {
		if amount != 0 {
			t.Errorf("currency %d = %d, want 0", currencyType, amount)
		}
	}
}
