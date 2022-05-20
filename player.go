package main

import (
	"fmt"
	"math/rand"
	"time"

	rtree "github.com/anayks/go-rstar-tree"
)

type Player struct {
	Sem          *Semacquire
	PosX         int `json:"x"`
	PosY         int `json:"y"`
	Id           int `json:"id"`
	Point        *rtree.Point
	playersClose []*Player
}

func updatePlayers(pTree PlayerTree) {
	start := time.Now()
	defer func() {
		fmt.Printf("function complete in %v\n", time.Since(start))
	}()

	for i := 0; i < 1000; i++ {
		v1 := AllPlayers.PlayersArray[i]
		bb, _ := rtree.NewRect(rtree.Point{float64(v1.PosX) - VisibilityRange/2, float64(v1.PosY) - VisibilityRange/2}, []float64{VisibilityRange, VisibilityRange})
		pTree.Tree.SearchIntersect(*bb)
	}

}

func (pl *Player) moveRandom() {
	pl.Sem.Acquire()
	defer pl.Sem.Release()

	randX := rand.Intn(100)
	randMoveX := rand.Intn(5)
	if randX > 50 {
		pl.moveRight(randMoveX)
	} else {
		pl.moveLeft(randMoveX)
	}

	randY := rand.Intn(100)
	randMoveY := rand.Intn(5)
	if randY > 50 {
		pl.moveTop(randMoveY)
	} else {
		pl.moveDown(randMoveY)
	}
}

func (pl *Player) moveTop(y int) {
	if pl.PosY-y <= 0 {
		pl.PosY = 0
		return
	}
	pl.PosY = pl.PosY - y
}

func (pl *Player) moveDown(y int) {
	if pl.PosY+y >= MaxDown {
		pl.PosY = MaxDown
		return
	}
	pl.PosY = pl.PosY + y
}

func (pl *Player) moveRight(x int) {
	if pl.PosX+x >= MaxRight {
		pl.PosX = MaxRight
		return
	}
	pl.PosX = pl.PosX + x
}

func (pl *Player) moveLeft(x int) {
	if pl.PosX-x <= 0 {
		pl.PosX = 0
		return
	}
	pl.PosX = pl.PosX - x
}

func (pl *Player) getNearestPlayersPos() []Player {
	pl.Sem.Acquire()
	defer pl.Sem.Release()

	array := make([]Player, 0)
	for _, v := range pl.playersClose {
		array = append(array, Player{
			PosX: v.PosX,
			PosY: v.PosY,
			Id:   v.Id,
		})
	}
	return array
}

func moveAllRandom() {
	for range time.Tick(time.Millisecond * 63) {
		AllPlayers.Sem.Acquire()

		for _, v := range AllPlayers.PlayersArray {
			v.moveRandom()
		}

		AllPlayers.Sem.Release()
	}
}

func (t *Player) Bounds() *rtree.Rect {
	return t.Point.ToRect(0.01)
}
