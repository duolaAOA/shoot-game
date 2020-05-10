package backend

// 玩家包含本地及远端玩家
type Player struct {
	IdentifierBase
	Positioner
	Mover
	CurrentPosition Coordinate
	Name  			string
	Icon 			rune			// 用户界面展示标识
}

// 玩家所在位置
func (p *Player) Position() Coordinate {
	return p.CurrentPosition
}

// 玩家进行移动
func (p *Player) Move(c Coordinate) {
	p.CurrentPosition = c
}