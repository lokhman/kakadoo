package app

import (
	"fmt"
	"math"
	"sort"
	"strconv"
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

const correctAnswerBaseScore = 15_000

type gameplay struct {
	currentTaskIndex int
	gameType         string
	tasks            []*Task
	answers          gpAnswers
	scores           gpScores
	state            gpState
	flashMessage     string
	deadline         time.Time
	mu               sync.Mutex
}

func (gp *gameplay) Init(player *Player) {
	gp.mu.Lock()
	defer gp.mu.Unlock()

	gp.scores[player] = make([]float64, len(gp.tasks))
}

func (gp *gameplay) Start() int {
	gp.mu.Lock()
	defer gp.mu.Unlock()

	gp.state = gpsStarted
	gp.currentTaskIndex = 0
	return len(gp.tasks)
}

func (gp *gameplay) NextTask(tick func(int), callback func(gp *gameplay, task *Task)) *Task {
	gp.mu.Lock()
	defer gp.mu.Unlock()

	if gp.state != gpsStarted {
		return nil
	}
	if gp.currentTaskIndex >= len(gp.tasks) {
		gp.state = gpsFinished
		return nil
	}

	task := gp.tasks[gp.currentTaskIndex]

	gp.answers[gp.currentTaskIndex] = make(map[*Player]string)
	gp.deadline = time.Now().Add(task.timeToAnswer())
	gp.state = gpsAccepting

	go func() {
		for timer := task.TimeToAnswer - 1; timer >= 0; timer-- {
			time.Sleep(time.Second)
			tick(timer)
		}

		gp.mu.Lock()
		defer gp.mu.Unlock()

		gp.calculateScores(task)

		callback(gp, task)

		gp.state = gpsStarted
		gp.currentTaskIndex++
		gp.flashMessage = ""
	}()

	return task
}

func (gp *gameplay) Answer(player *Player, answer string) {
	gp.mu.Lock()
	defer gp.mu.Unlock()

	if gp.state != gpsAccepting {
		return
	}
	answers := gp.answers[gp.currentTaskIndex]
	if _, ok := answers[player]; !ok {
		answers[player] = answer
	}
}

func (gp *gameplay) calculateScores(task *Task) {
	answers := gp.answers[gp.currentTaskIndex]
	if len(answers) == 0 {
		return
	}

	if gp.gameType == gameTypeQuiz {
		for player, answer := range answers {
			scores := gp.scores[player]
			if answer == task.CorrectAnswer {
				baseScore := correctAnswerBaseScore * (1 / math.Max(float64(task.TimeToAnswer), 1))
				scores[gp.currentTaskIndex] = baseScore + float64(gp.deadline.Sub(time.Now()).Milliseconds())
			}
		}
	} else if gp.gameType == gameTypeWoC {
		correctAnswer, err := strconv.ParseFloat(task.CorrectAnswer, 64)
		if err != nil {
			panic(err)
		}

		mean := 0.0
		median := 0.0

		type wocScore struct {
			player *Player
			value  float64
		}
		scores := make([]*wocScore, 0)

		for player, answer := range answers {
			if answer, err := strconv.ParseFloat(answer, 64); err == nil {
				scores = append(scores, &wocScore{
					player: player,
					value:  answer,
				})
				mean += answer
			}
		}

		mean /= float64(len(scores))
		sort.Slice(scores, func(i, j int) bool {
			return scores[i].value < scores[j].value
		})

		index := len(scores) / 2
		if len(scores)%2 == 0 {
			median = (scores[index-1].value + scores[index].value) / 2
		} else {
			median = scores[index].value
		}

		scores = append(scores, &wocScore{value: mean}, &wocScore{value: median})
		gp.flashMessage = fmt.Sprintf("Mean: %.2f | Median: %.2f", mean, median)

		for _, score := range scores {
			value := score.value
			score.value = math.Abs(value - correctAnswer)
		}

		sort.Slice(scores, func(i, j int) bool {
			return scores[i].value > scores[j].value
		})

		type wocGroup struct {
			indexes []int
			players []*Player
		}
		groups := make(map[string]*wocGroup)
		for index, score := range scores {
			value := fmt.Sprintf("%.2f", score.value)
			if _, ok := groups[value]; !ok {
				groups[value] = &wocGroup{
					indexes: make([]int, 0),
					players: make([]*Player, 0),
				}
			}
			groups[value].indexes = append(groups[value].indexes, index)
			groups[value].players = append(groups[value].players, score.player)
		}

		for _, group := range groups {
			for _, player := range group.players {
				if player == nil {
					continue
				}
				score := 0.0
				for _, index = range group.indexes {
					score += float64(index)
				}
				score /= float64(len(group.indexes))
				gp.scores[player][gp.currentTaskIndex] = score
			}
		}
	}
}

func (gp *gameplay) Finish() gpScores {
	gp.mu.Lock()
	defer gp.mu.Unlock()

	gp.state = gpsFinished
	return gp.scores
}

func newGameplay(game *Game) *gameplay {
	tasks := game.GetTasks()
	return &gameplay{
		gameType: game.Type,
		tasks:    tasks,
		answers:  make(gpAnswers, len(tasks)),
		scores:   make(gpScores),
		state:    gpsReady,
	}
}

type gpAnswers []map[*Player]string

type gpScores map[*Player][]float64

func (s gpScores) Leaderboard() []lbScore {
	board := make([]lbScore, len(s))
	index := 0
	for player, scores := range s {
		total := 0.0
		for _, score := range scores {
			if score > 0 {
				total += score
			}
		}
		board[index] = lbScore{
			Player: player.Name,
			Score:  total,
		}
		index++
	}
	sort.Slice(board, func(i, j int) bool {
		return board[i].Score > board[j].Score
	})
	return board
}

type lbScore struct {
	Player string  `json:"player"`
	Score  float64 `json:"score"`
}
