package battle

import "cardGame/model/constants"

// fire 在全場觸發某個時機：依行動順序，讓每個存活單位發動符合 trigger 的被動與天賦。
// BattleStart、BeforeRoundStart、RoundEnd 由引擎在對應時機呼叫。
func (b *battle) fire(trigger constants.Trigger) {
	for _, u := range b.actionOrder() {
		b.fireUnit(trigger, u)
	}
}

// fireUnit 只觸發單一單位的被動與天賦（BeforeAction、AfterAction 使用）。
func (b *battle) fireUnit(trigger constants.Trigger, u *unit) {
	for _, s := range u.passives() {
		if !u.alive() {
			return
		}
		if s.Trigger == trigger {
			b.cast(u, s, EventPassive)
		}
	}
}
