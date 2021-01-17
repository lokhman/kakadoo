package app

import (
	"sort"
	"sync"
	"time"
)

type gpState int

const (
	gpsReady = iota
	gpsStarted
	gpsAccepting
	gpsFinished
)

type gameplay struct {
	tasks              []Task
	currentTaskIdx     int
	currentAnswerStats map[int]int
	scores             gpScores
	state              gpState
	deadline           time.Time
	mu                 sync.Mutex
}

func (gp *gameplay) Init(player *Player) {
	gp.mu.Lock()
	defer gp.mu.Unlock()

	scores := make([]int64, len(gp.tasks))
	gp.scores[player] = scores
	for i := range scores {
		scores[i] = -1
	}
}

func (gp *gameplay) Start() int {
	gp.mu.Lock()
	defer gp.mu.Unlock()

	gp.state = gpsStarted
	gp.currentTaskIdx = 0
	return len(gp.tasks)
}

func (gp *gameplay) NextTask(tick func(int), callback func(gp *gameplay)) Task {
	gp.mu.Lock()
	defer gp.mu.Unlock()

	if gp.state != gpsStarted {
		return nil
	}
	if gp.currentTaskIdx >= len(gp.tasks) {
		gp.state = gpsFinished
		return nil
	}

	task := gp.tasks[gp.currentTaskIdx]

	gp.currentAnswerStats = make(map[int]int)
	for idx := range task.answers() {
		gp.currentAnswerStats[idx] = 0
	}

	gp.deadline = time.Now().Add(task.timeToAnswer())
	gp.state = gpsAccepting

	go func() {
		for timer := int(task.timeToAnswer()/time.Second) - 1; timer >= 0; timer-- {
			time.Sleep(time.Second)
			tick(timer)
		}

		gp.mu.Lock()
		defer gp.mu.Unlock()

		gp.state = gpsStarted
		gp.currentTaskIdx++
		callback(gp)
	}()

	return task
}

func (gp *gameplay) Answer(player *Player, idx int) int64 {
	gp.mu.Lock()
	defer gp.mu.Unlock()

	scores := gp.scores[player]
	if gp.state != gpsAccepting || scores[gp.currentTaskIdx] != -1 {
		return -1
	}

	task := gp.tasks[gp.currentTaskIdx]
	if idx >= len(task.answers()) {
		return -1
	}

	var score int64 = 0
	if task.correctAnswerIdx() == idx {
		score = gp.deadline.Sub(time.Now()).Milliseconds()
	}
	scores[gp.currentTaskIdx] = score
	gp.currentAnswerStats[idx]++
	return score
}

func (gp *gameplay) Finish() gpScores {
	gp.mu.Lock()
	defer gp.mu.Unlock()

	gp.state = gpsFinished
	return gp.scores
}

func newGameplay(game *Game) *gameplay {
	return &gameplay{
		tasks:  game.GetTasks(),
		scores: make(gpScores),
		state:  gpsReady,
	}
}

type gpScores map[*Player][]int64

func (s gpScores) Leaderboard() []lbScore {
	board := make([]lbScore, len(s))
	idx := 0
	for player, scores := range s {
		var total int64 = 0
		for _, score := range scores {
			if score > 0 {
				total += score
			}
		}
		board[idx] = lbScore{
			Player: player.Name,
			Score:  total,
		}
		idx++
	}
	sort.Slice(board, func(i, j int) bool {
		return board[i].Score > board[j].Score
	})
	return board
}

type lbScore struct {
	Player string `json:"player"`
	Score  int64  `json:"score"`
}
