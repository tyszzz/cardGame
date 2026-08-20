package summon

type summonCoin struct {
}

func newSummonCoin() SummonManager {
	return &summonCoin{}
}

func (c *summonCoin) SummonOnce() string {
	return "test"
}
func (c *summonCoin) SummonTen() []string {
	return []string{"test"}
}
