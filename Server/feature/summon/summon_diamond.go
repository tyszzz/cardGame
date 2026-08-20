package summon

type summonDiamond struct {
}

func newSummonDiamond() SummonManager {
	return &summonDiamond{}
}

func (d *summonDiamond) SummonOnce() string {
	return "test"
}
func (d *summonDiamond) SummonTen() []string {
	return []string{"test"}
}
