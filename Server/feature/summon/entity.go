package summon

type SummonManager interface {
	SummonOnce() string
	SummonTen() []string
}
