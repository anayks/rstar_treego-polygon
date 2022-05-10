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
	Tree: rtree.NewTree(3, 2, 16),
}

const (
	PlayersCount    = 1000
	MaxDown         = 2000
	MaxRight        = 2000
	MaxDepth        = 2000
	VisibilityRange = 200
)

const playerID = PlayersCount

type Pos struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

func main() {
	for iter := 0; iter < PlayersCount; iter++ {
		posX := rand.Intn(MaxRight)
		posY := rand.Intn(MaxDown)
		posZ := rand.Intn(MaxDepth)

		point := rtree.Point{float64(posX), float64(posY), float64(posZ)}

		sem := SemNew(time.Second * 3)

		player := &Player{
			PosX:  posX,
			PosY:  posY,
			PosZ:  posZ,
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

		player := updateMainPlayerPosition(t.X, t.Y, t.Z)

		closePlayers := player.getNearestPlayersPos()

		res, err := json.Marshal(closePlayers)

		if err != nil {
			return
		}

		w.Write(res)
	})
	http.ListenAndServe("localhost:8050", nil)
}

func updateMainPlayerPosition(x, y, z int) *Player {
	AllPlayers.Sem.Acquire()
	defer AllPlayers.Sem.Release()

	if len(AllPlayers.PlayersArray) == PlayersCount {
		point := rtree.Point{float64(x), float64(y), float64(z)}
		sem := SemNew(time.Second * 3)

		newPlayer := &Player{
			PosX:  x,
			PosY:  y,
			PosZ:  z,
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
