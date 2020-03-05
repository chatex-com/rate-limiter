package bandwish_limiter

const (
	StrategyImmediately Strategy = iota
	StrategyEvenly
)

type Strategy uint32

func (s Strategy) String() string {
	return [...]string{"ASAP", "Evenly"}[s]
}
