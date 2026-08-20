package summon

type summonFriend struct {
}

func newSummonFriend() SummonManager {
	return &summonFriend{}
}

func (f *summonFriend) SummonOnce() string {
	return "test"
}
func (f *summonFriend) SummonTen() []string {
	return []string{"test"}
}
