package app

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type findCatForm struct {
	Player      string `form:"p"`
	Key         string `form:"k"`
	Index       int    `form:"w"`
	Answer      string `form:"v"`
	ErrorPlayer bool   `form:"ep"`
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

		if HasPlayerInScores(game, player, form.Key) {
			c.Redirect(http.StatusTemporaryRedirect, c.Request.URL.Path+"?ep=1")
			c.Abort()
			return
		}
		if game.LastStartedAt == nil {
			UpdateGameStartedAt(game, time.Now())
		}

		if index = form.Index; index < 0 {
			index = 0
		} else if index >= len(tasks) {
			c.HTML(http.StatusOK, "podium", gin.H{"scores": GetScores(game), "player": player})
			return
		}
		task = tasks[index]

		if form.Answer != "" {
			form.Index++

			var score float64
			answer := strings.Split(form.Answer, ",")
			correctAnswer := strings.Split(task.CorrectAnswer, ",")
			if len(answer) == 2 && len(correctAnswer) == 4 {
				var x, y, x1, y1, x2, y2 int
				var err error
				if x, err = strconv.Atoi(answer[0]); err != nil {
					goto score
				}
				if y, err = strconv.Atoi(answer[1]); err != nil {
					goto score
				}
				x1, _ = strconv.Atoi(correctAnswer[0])
				y1, _ = strconv.Atoi(correctAnswer[1])
				x2, _ = strconv.Atoi(correctAnswer[2])
				y2, _ = strconv.Atoi(correctAnswer[3])
				if x >= x1 && x <= x2 && y >= y1 && y <= y2 {
					score = 1
				}
			}
		score:
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
