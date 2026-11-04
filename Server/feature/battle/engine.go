package battle

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"

	"cardGame/model/constants"
)

type battle struct {
	units      map[Team][]*unit // 依位置由小到大排序
	enemyFirst bool
	round      int
	rng        *rand.Rand
	events     []Event
}

const MaxRounds = 20 // 超過回合上限判守方（敵方）勝

var ErrInvalidInput = errors.New("invalid battle input")

func newBattle(in Input) (*battle, error) {
	ally, err := buildUnits(TeamAlly, in.Ally)
	if err != nil {
		return nil, err
	}
	enemy, err := buildUnits(TeamEnemy, in.Enemy)
	if err != nil {
		return nil, err
	}
	return &battle{
		units:      map[Team][]*unit{TeamAlly: ally, TeamEnemy: enemy},
		enemyFirst: in.EnemyFirst,
		rng:        rand.New(rand.NewSource(in.Seed)),
	}, nil
}

// inner function

func buildUnits(team Team, inputs []UnitInput) ([]*unit, error) {
	if len(inputs) == 0 || len(inputs) > maxPosition {
		return nil, fmt.Errorf("%w: %s needs 1~%d units", ErrInvalidInput, team, maxPosition)
	}
	seen := make(map[int]bool, len(inputs))
	out := make([]*unit, 0, len(inputs))
	for _, in := range inputs {
		switch {
		case in.Position < 1 || in.Position > maxPosition:
			return nil, fmt.Errorf("%w: %s position %d out of range", ErrInvalidInput, team, in.Position)
		case seen[in.Position]:
			return nil, fmt.Errorf("%w: %s position %d duplicated", ErrInvalidInput, team, in.Position)
		case in.Health <= 0 || in.Attack < 0 || in.Defense < 0:
			return nil, fmt.Errorf("%w: %s position %d has invalid stats", ErrInvalidInput, team, in.Position)
		}
		seen[in.Position] = true
		out = append(out, &unit{
			ref:           Ref{Team: team, Position: in.Position},
			id:            in.ID,
			attack:        in.Attack,
			defense:       in.Defense,
			maxHealth:     in.Health,
			health:        in.Health,
			star:          in.Star,
			level:         in.Level,
			rage:          constants.RageInitial,
			rageThreshold: constants.RageCastThreshold,
			skills:        in.Skills,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ref.Position < out[j].ref.Position })
	return out, nil
}

func opponent(t Team) Team {
	if t == TeamAlly {
		return TeamEnemy
	}
	return TeamAlly
}

func (b *battle) emit(e Event) {
	e.Round = b.round
	b.events = append(b.events, e)
}

func (b *battle) living(team Team) []*unit {
	var out []*unit
	for _, u := range b.units[team] {
		if u.alive() {
			out = append(out, u)
		}
	}
	return out
}

// over 判斷戰鬥是否結束；雙方同時全滅時算守方勝。
func (b *battle) over() (Team, bool) {
	allyAlive := len(b.living(TeamAlly)) > 0
	enemyAlive := len(b.living(TeamEnemy)) > 0
	switch {
	case allyAlive && enemyAlive:
		return 0, false
	case enemyAlive || !allyAlive:
		return TeamEnemy, true
	}
	return TeamAlly, true
}

// act 執行單位的一次行動：行動前 → 普攻或主動 → 行動後怒氣 → 行動後。
func (b *battle) act(u *unit) {
	b.fireUnit(constants.TriggerBeforeAction, u)
	if !u.alive() {
		return
	}
	if u.rage >= u.rageThreshold {
		u.rage -= u.rageThreshold
		b.cast(u, u.skills.ActiveSkill, EventActiveSkill)
	} else {
		b.cast(u, u.skills.NormalAtk, EventNormalAtk)
	}
	u.rage += constants.RageGainOnAction
	b.fireUnit(constants.TriggerAfterAction, u)
}

func (b *battle) playRound() (Team, bool) {
	b.emit(Event{Type: EventRoundStart})
	b.fire(constants.TriggerBeforeRoundStart)
	if w, ok := b.over(); ok {
		return w, true
	}
	for _, u := range b.actionOrder() {
		if !u.alive() {
			continue
		}
		b.act(u)
		if w, ok := b.over(); ok {
			return w, true
		}
	}
	b.fire(constants.TriggerRoundEnd)
	b.emit(Event{Type: EventRoundEnd})
	return b.over()
}

func (b *battle) finish(winner Team, rounds int) *Result {
	b.emit(Event{Type: EventBattleEnd, Actor: Ref{Team: winner}})
	return &Result{Winner: winner, Rounds: rounds, Events: b.events}
}

// outer function

// Run 執行一場戰鬥並回傳結果與事件日誌。
func Run(in Input) (*Result, error) {
	b, err := newBattle(in)
	if err != nil {
		return nil, err
	}
	b.emit(Event{Type: EventBattleStart})
	b.fire(constants.TriggerBattleStart)
	if w, ok := b.over(); ok {
		return b.finish(w, 0), nil
	}
	for b.round = 1; b.round <= MaxRounds; b.round++ {
		if w, ok := b.playRound(); ok {
			return b.finish(w, b.round), nil
		}
	}
	b.round = MaxRounds
	return b.finish(TeamEnemy, MaxRounds), nil
}
