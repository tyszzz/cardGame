package services

import (
	"cardGame/feature/card"
	"cardGame/feature/summon"
	"cardGame/model/constants"
	"cardGame/model/packet"
	"context"
	"errors"
	"testing"
)

type zeroRand struct{}

func (zeroRand) IntN(int) int { return 0 }

type fakeSummonRepo struct {
	cards     []card.Card
	listErr   error
	listCalls int
	owned     map[string]struct{}
	applyErr  error

	applyCalls int
	gotUID     string
	gotCurr    constants.CurrencyType
	gotCost    int64
}

func (f *fakeSummonRepo) ListEnabledCards(context.Context) ([]card.Card, error) {
	f.listCalls++
	return f.cards, f.listErr
}

func (f *fakeSummonRepo) ApplySummon(_ context.Context, uid string, c constants.CurrencyType, cost int64, resolve summon.ResolveFunc) (*summon.Outcome, error) {
	f.applyCalls++
	f.gotUID, f.gotCurr, f.gotCost = uid, c, cost
	if f.applyErr != nil {
		return nil, f.applyErr
	}
	out, err := resolve(f.owned)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func newTestSummonService(repo *fakeSummonRepo) *SummonService {
	s := NewSummonService(repo)
	s.rnd = zeroRand{}
	return s
}

func allRarityCards() []card.Card {
	return []card.Card{
		{CardId: "r1", CardName: "R1", Rarity: constants.RarityR},
		{CardId: "sr1", CardName: "SR1", Rarity: constants.RaritySR},
		{CardId: "ssr1", CardName: "SSR1", Rarity: constants.RaritySSR},
		{CardId: "ur1", CardName: "UR1", Rarity: constants.RarityUR},
	}
}

func TestSummonCardChargesAndReturnsCards(t *testing.T) {
	repo := &fakeSummonRepo{cards: allRarityCards(), owned: map[string]struct{}{}}
	s := newTestSummonService(repo)

	res, err := s.SummonCard(context.Background(), "u1", &packet.SummonCard{Times: 10, SummonType: constants.DiamondSummon})
	if err != nil {
		t.Fatal(err)
	}
	if repo.gotUID != "u1" || repo.gotCurr != constants.CurrencyDiamond || repo.gotCost != 1800 {
		t.Errorf("charged uid=%s currency=%v cost=%d", repo.gotUID, repo.gotCurr, repo.gotCost)
	}
	if len(res.Cards) != 10 {
		t.Fatalf("cards = %d, want 10", len(res.Cards))
	}
	// zeroRand：前 9 抽為 R，第 10 抽被保底改為 SSR；R 只有第一張是新卡
	if !res.Cards[0].IsNew || res.Cards[1].IsNew || res.Cards[1].ShardId != "shard_r" || res.Cards[1].ShardAmount != 30 {
		t.Errorf("unexpected dup handling: %+v %+v", res.Cards[0], res.Cards[1])
	}
	if last := res.Cards[9]; last.Rarity != constants.RaritySSR || !last.IsNew {
		t.Errorf("last = %+v, want new SSR", last)
	}
}

func TestSummonCardValidation(t *testing.T) {
	repo := &fakeSummonRepo{cards: allRarityCards()}
	s := newTestSummonService(repo)
	ctx := context.Background()

	if _, err := s.SummonCard(ctx, "u1", &packet.SummonCard{Times: 1, SummonType: 0}); !errors.Is(err, summon.ErrInvalidSummonType) {
		t.Errorf("err = %v, want ErrInvalidSummonType", err)
	}
	for _, times := range []int32{0, 2, 11, -1} {
		if _, err := s.SummonCard(ctx, "u1", &packet.SummonCard{Times: times, SummonType: constants.NormalSummon}); !errors.Is(err, summon.ErrInvalidTimes) {
			t.Errorf("times %d: err = %v, want ErrInvalidTimes", times, err)
		}
	}
	if repo.applyCalls != 0 || repo.listCalls != 0 {
		t.Errorf("invalid requests must not touch the repository: apply=%d list=%d", repo.applyCalls, repo.listCalls)
	}
}

func TestSummonCardNotEnoughCurrency(t *testing.T) {
	repo := &fakeSummonRepo{cards: allRarityCards(), applyErr: summon.ErrNotEnoughCurrency}
	s := newTestSummonService(repo)
	_, err := s.SummonCard(context.Background(), "u1", &packet.SummonCard{Times: 1, SummonType: constants.NormalSummon})
	if !errors.Is(err, summon.ErrNotEnoughCurrency) {
		t.Fatalf("err = %v, want ErrNotEnoughCurrency", err)
	}
}

func TestSummonCardPoolCachedAndEmptyPool(t *testing.T) {
	repo := &fakeSummonRepo{cards: allRarityCards(), owned: map[string]struct{}{}}
	s := newTestSummonService(repo)
	for i := 0; i < 3; i++ {
		if _, err := s.SummonCard(context.Background(), "u1", &packet.SummonCard{Times: 1, SummonType: constants.NormalSummon}); err != nil {
			t.Fatal(err)
		}
	}
	if repo.listCalls != 1 {
		t.Errorf("ListEnabledCards calls = %d, want 1", repo.listCalls)
	}

	empty := &fakeSummonRepo{}
	s2 := newTestSummonService(empty)
	if _, err := s2.SummonCard(context.Background(), "u1", &packet.SummonCard{Times: 1, SummonType: constants.NormalSummon}); !errors.Is(err, summon.ErrEmptyCardPool) {
		t.Errorf("err = %v, want ErrEmptyCardPool", err)
	}

	// Diamond 卡池缺 SSR/UR：不得扣款
	partial := &fakeSummonRepo{cards: allRarityCards()[:2]}
	s3 := newTestSummonService(partial)
	if _, err := s3.SummonCard(context.Background(), "u1", &packet.SummonCard{Times: 1, SummonType: constants.DiamondSummon}); !errors.Is(err, summon.ErrEmptyCardPool) {
		t.Errorf("err = %v, want ErrEmptyCardPool", err)
	}
	if partial.applyCalls != 0 {
		t.Error("must not charge when the pool is incomplete")
	}
}
