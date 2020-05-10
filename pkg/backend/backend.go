package backend

import (
	"fmt"
	"github.com/google/uuid"
	"math"
	"sync"
	"time"
)

const (
	roundOverScore          = 10						// 结束比分
	newRoundWaitTime        = 10 * time.Second			// 新一轮比赛等待时长
	collisionCheckFrequency = 10 * time.Millisecond		// 碰撞检查频率
	moveThrottle            = 100 * time.Millisecond	// 移动速度
	laserThrottle           = 500 * time.Millisecond    // 激光发射速度
	laserSpeed              = 50						// 激光速度
)

// 游戏是游戏的后端引擎。 无论如何呈现游戏数据或是否正在使用游戏服务器，都可以使用它。
type Game struct {
	Entities		map[uuid.UUID]Identifier
	gameMap	   		[][]rune
	Mu 				sync.RWMutex
	ChangeChannel chan Change
	ActionChanel chan Action
	lastAction map[string]time.Time
	Score map[uuid.UUID]int
	NewRoundAt time.Time
	RoundWinner uuid.UUID
	WaitForRound bool
	IsAuthoritative bool
	spawnPointIndex int

}

// NewGame constructs a new Game struct.
func NewGame() *Game {
	return &Game{
		Entities: make(map[uuid.UUID]Identifier),
		ActionChanel: make(chan Action, 1),
		lastAction: make(map[string]time.Time),
		ChangeChannel: make(chan Change, 1),
		IsAuthoritative: true,
		WaitForRound: false,
		Score: make(map[uuid.UUID]int),
		gameMap: MapDefault,
		spawnPointIndex: 0,
	}
}

// 开始游戏事件循环，并等待动作更新更新游戏状态
func (game * Game) Start() {
	go game.watchActions()
}

// 等待动作更新并执行
func (game *Game) watchActions() {
	for {
		action := <-game.ActionChanel
		if game.WaitForRound {
			continue
		}
		game.Mu.Lock()
		action.Perform(game)
		game.Mu.Unlock()
	}
}

// 检查实体碰撞-我们现在关心的是激光和玩家碰撞
func (game *Game) watchCollisions() {
	for {
		game.Mu.Lock()
		spawnPoints := game
	}
}

// 将坐标映射到实体集
func (game * Game) getCollisionMap() map[Coordinate][]Identifier {
	collisionMap := map[Coordinate][]Identifier{}
	for _, entity := range game.Entities{
		positioner, ok := entity.(Positioner)
		if !ok {
			continue
		}
		position := positioner.Position()
		collisionMap[position] = append(collisionMap[position], entity)
	}
	return collisionMap
}

// 添加实体到game
func (game *Game) AddEntity(entity Identifier) {
	game.Entities[entity.ID()] = entity
}

// 更新game实体信息
func (game *Game) UpdateEntity(entity Identifier) {
	game.Entities[entity.ID()] = entity
}

// 获取实体信息
func (game *Game) GetEntity(id uuid.UUID) Identifier {
	return game.Entities[id]
}

// 从游戏中移除一个实体玩家
func (game *Game) RemoveEntity(id uuid.UUID) {
	delete(game.Entities, id)
}

// 重置游戏状态开始新一局游戏
func (game *Game) startNewRound() {
	game.WaitForRound = false
	game.Score = map[uuid.UUID]int{}
	i := 0
	spawnPoints := game.GetMapByType()[MapTypeSpawn]
	for _, entity := range game.Entities {
		player, ok := entity.(*Player)
		if !ok {
			continue
		}
		player.Move(spawnPoints[i%len(spawnPoints)])
		i++
	}
	game.sendChange(RoundStartChange{})
}

// queueNewRound queues a new round to start.
func (game *Game) queueNewRound(roundWinner uuid.UUID) {
	game.WaitForRound = true
	game.NewRoundAt = time.Now().Add(newRoundWaitTime)
	game.RoundWinner = roundWinner
	game.sendChange(RoundOverChange{})
	go func() {
		time.Sleep(newRoundWaitTime)
		game.Mu.Lock()
		game.startNewRound()
		game.Mu.Unlock()
	}()
}

// AddScore increments an entity's score.
func (game *Game) AddScore(id uuid.UUID) {
	game.Score[id]++
}

// checkLastActionTime 检查判断最后一次的操作发送的动作
func (game *Game) checkLastActionTime(actionKey string, created time.Time, throttle time.Duration) bool {
	lastAction, ok := game.lastAction[actionKey]
	if ok && lastAction.After(created.Add(-1*throttle)) {
		return false
	}
	return true
}

// updateLastActionTime sets the last action time.
// The actionKey should be unique to the action and the actor (entity).
func (game *Game) updateLastActionTime(actionKey string, created time.Time) {
	game.lastAction[actionKey] = created
}

// sendChange sends a change to the change channel.
func (game *Game) sendChange(change Change) {
	select {
	case game.ChangeChannel <- change:
	default:
	}
}

// 坐标相关变量
type Coordinate struct {
	X int
	Y int
}

// 添加两个坐标
func (c1 Coordinate) Add(c2 Coordinate) Coordinate {
	return Coordinate{
		X: c1.X + c2.X,
		Y: c1.Y + c2.Y,
	}
}

// 计算两个坐标之间的距离。
func (c1 Coordinate) Distance(c2 Coordinate) int {
	return int(math.Sqrt(math.Pow(float64(c2.X-c1.X), 2) + math.Pow(float64(c2.Y-c1.Y), 2)))
}

// 使用常数表示方向
type Direction int

// Contains direction constants - DirectionStop will take no effect.
const (
	DirectionUp Direction = iota
	DirectionDown
	DirectionLeft
	DirectionRight
	DirectionStop
)

// 身份标识接口
type IdentifierBase struct {
	UUID uuid.UUID
}

// 返回一个UUID实体
func (e IdentifierBase) ID() uuid.UUID {
	return e.UUID
}

// 根据Actions发送给game engine 做 change
type Change interface {}

// MoveChange is sent when the game engine moves an entity.
type MoveChange struct {
	Change
	Entity Identifier
	Direction Direction
	Position Coordinate
}

// RoundOverChange表示回合结束。 有关新回合的信息应从游戏实例中检索
type RoundOverChange struct {
	Change
}

// RoundStartChange 表示新一回合开始
type RoundStartChange struct {
	Change
}

// 当添加实体以响应动作时，就会发生AddEntityChange。 目前，这仅用于新的激光器和加入游戏的玩家。
type AddEntityChange struct {
	Change
	Entity Identifier
}

// RemoveEntityChange 将实体从游戏中移除
type RemoveEntityChange struct {
	Change
	Entity Identifier
}

// 提供ID方法实体
type Identifier interface {
	ID() uuid.UUID
}

// 标明实体所在位置
type Positioner interface {
	Position() Coordinate
}

// 实体移动
type Mover interface {
	Move(Coordinate)
}

// 玩家被kill后复活
type PlayerRespawnChange struct {
	Change
	Player     *Player
	KilledByID uuid.UUID
}

// ACtion 是由客户端在尝试更改游戏状态时发送的。 如果动作无效或执行得太频繁，引擎可以选择拒绝动作。
type Action interface {
	Perform(game *Game)
}

// 当用户按下方向键时，将发送MoveAction。
type MoveAction struct {
	Direction Direction
	ID uuid.UUID
	Created time.Time
}

// 玩家移动实体应该包含的的后端逻辑
func (action MoveAction) Perform(game *Game) {
	entity := game.GetEntity(action.ID)
	if entity == nil {
		return
	}
	mover, ok := entity.(Mover)
	if !ok {
		return
	}
	positioner, ok := entity.(Positioner)
	if !ok {
		return
	}
	actionKey := fmt.Sprintf("%T:%s", action, entity.ID().String())
	if !game.checkLastActionTime(actionKey, action.Created, moveThrottle) {
		return
	}
	position := positioner.Position()
	// Move the entity.
	switch action.Direction {
	case DirectionUp:
		position.Y--
	case DirectionDown:
		position.Y++
	case DirectionLeft:
		position.X--
	case DirectionRight:
		position.X++
	}
	// Check if position collides with a wall.
	for _, wall := range game.GetMapByType()[MapTypeWall] {
		if position == wall {
			return
		}
	}
	// Check if position collides with a player.
	collidingEntities, ok := game.getCollisionMap()[position]
	if ok {
		for _, entity := range collidingEntities {
			_, ok := entity.(*Player)
			if ok {
				return
			}
		}
	}
	mover.Move(position)
	// Inform the client that the entity moved.
	change := MoveChange{
		Entity:    entity,
		Direction: action.Direction,
		Position:  position,
	}
	game.sendChange(change)
	game.updateLastActionTime(actionKey, action.Created)
}