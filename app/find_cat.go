package app

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var findCatPlayers = struct {
	playerKeys map[string]string
	mu         sync.Mutex
}{
	playerKeys: make(map[string]string),
}

type findCatForm struct {
	Player  string   `form:"p"`
	Key     string   `form:"k"`
	Answer  string   `form:"v"`
	Answers []string `form:"w"`

	ErrorPlayer bool `form:"ep"`
}

func FindCat(c *gin.Context) {
	game := c.MustGet("game").(*Game)
	tasks := game.GetTasks()

	var form findCatForm
	_ = c.ShouldBind(&form)

	var task *Task
	var index = 0
	if form.Player != "" {
		player := StripHtmlTags(form.Player)
		if form.Key == "" {
			url := c.Request.URL
			query := url.Query()
			query.Add("k", RandString(4))
			url.RawQuery = query.Encode()
			c.Redirect(http.StatusTemporaryRedirect, url.String())
			c.Abort()
			return
		}

		// findCatPlayers.mu.Lock()
		// defer findCatPlayers.mu.Unlock()
		// if key, ok := findCatPlayers.playerKeys[player]; ok && key != form.Key {
		// 	c.Redirect(http.StatusTemporaryRedirect, c.Request.URL.Path+"?ep=1")
		// 	c.Abort()
		// 	return
		// } else if !ok {
		// 	findCatPlayers.playerKeys[player] = form.Key
		// 	UpdateGameStartedAt(game, time.Now(), false)
		// }

		if HasPlayerInScores(game, player, form.Key) {
			c.Redirect(http.StatusTemporaryRedirect, c.Request.URL.Path+"?ep=1")
			c.Abort()
			return
		}
		if game.LastStartedAt == nil {
			UpdateGameStartedAt(game, time.Now())
		}

		index = len(form.Answers)
		if index >= len(tasks) {
			c.HTML(http.StatusOK, "podium", gin.H{"scores": GetScores(game), "player": player})
			return
		}
		task = tasks[index]

		if form.Answer != "" {
			var score float64
			answer := strings.Split(form.Answer, ",")
			correctAnswer := strings.Split(task.CorrectAnswer, ",")
			if len(answer) == 2 && len(correctAnswer) == 4 {
				x, y := answer[0], answer[1]
				x1, y1, x2, y2 := correctAnswer[0], correctAnswer[1], correctAnswer[2], correctAnswer[3]
				if x >= x1 && x <= x2 && y >= y1 && y <= y2 {
					score = 1
				}
			}
			InsertScores(&Score{
				Game:      game,
				Task:      task,
				Player:    player,
				PlayerKey: form.Key,
				Question:  task.Question,
				Answer:    form.Answer,
				Score:     score,
				CreatedAt: time.Now(),
			})
		}
	}

	c.HTML(http.StatusOK, "find_cat", gin.H{
		"form":    form,
		"task":    task,
		"counter": fmt.Sprintf("%d / %d", index+1, len(tasks)),
	})
}