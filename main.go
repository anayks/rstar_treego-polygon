package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	rtree "github.com/anayks/go-rstar-tree"
)

type allPlayers struct {
	Sem          *Semacquire
	PlayersArray []*Player
}

var AllPlayers = &allPlayers{
	Sem:          SemNew(time.Second * 3),
	PlayersArray: make([]*Player, 0),
}

var PlayersArray []*Player

type PlayerTree struct {
	Sem  *Semacquire
	Tree *rtree.Rtree
}

var PTree = &PlayerTree{
	Sem:  SemNew(time.Second * 3),
	Tree: rtree.NewTree(2, 2, 16),
}

const (
	PlayersCount    = 1000
	MaxDown         = 8000
	MaxRight        = 8000
	VisibilityRange = 500
)

const playerID = PlayersCount

type Pos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func main() {
	for iter := 0; iter < PlayersCount; iter++ {
		posX := rand.Intn(MaxRight)
		posY := rand.Intn(MaxDown)

		point := rtree.Point{float64(posX), float64(posY)}

		sem := SemNew(time.Second * 3)

		player := &Player{
			PosX:  posX,
			PosY:  posY,
			Id:    iter,
			Point: &point,
			Sem:   sem,
		}

		PTree.Sem.Acquire()
		PTree.Tree.Insert(player)
		// fmt.Printf("%v \t", iter)
		PTree.Sem.Release()
		AllPlayers.PlayersArray = append(AllPlayers.PlayersArray, player)
	}

	for _, v := range AllPlayers.PlayersArray {
		v.moveRandom()
	}

	fmt.Printf("Cores: %d", runtime.GOMAXPROCS(0))

	go moveAllRandom()

	go timerUpdatePlayers(PTree)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.WriteHeader(http.StatusOK)

		t := &Pos{}

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			return
		}

		defer r.Body.Close()

		player := updateMainPlayerPosition(t.X, t.Y)

		closePlayers := player.getNearestPlayersPos()

		res, err := json.Marshal(closePlayers)

		if err != nil {
			return
		}

		w.Write(res)
	})
	http.ListenAndServe("localhost:8050", nil)
}

func updateMainPlayerPosition(x, y int) *Player {
	AllPlayers.Sem.Acquire()
	defer AllPlayers.Sem.Release()

	if len(AllPlayers.PlayersArray) == PlayersCount {
		point := rtree.Point{float64(x), float64(y)}
		sem := SemNew(time.Second * 3)

		newPlayer := &Player{
			PosX:  x,
			PosY:  y,
			Id:    PlayersCount,
			Point: &point,
			Sem:   sem,
		}
		AllPlayers.PlayersArray = append(AllPlayers.PlayersArray, newPlayer)
		return newPlayer
	}

	player := AllPlayers.PlayersArray[playerID]

	player.Sem.Acquire()
	defer player.Sem.Release()

	player.PosX = x
	player.PosY = y
	return player
}

func timerUpdatePlayers(pTree *PlayerTree) {
	updatePlayers(pTree)
	for range time.Tick(time.Millisecond * 500) {
		updatePlayers(pTree)
	}
}
