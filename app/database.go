package app

import (
	"database/sql"
	"log"
	"time"

	"github.com/elgris/sqrl"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var DB *sql.DB
var QB sqrl.StatementBuilderType

func init() {
	var err error

	DB, err = sql.Open("postgres", DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	QB = sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar).RunWith(DB)
}

type Task interface {
	question() string
	answers() []string
	correctAnswerIdx() int
	timeToAnswer() time.Duration
}

type Game struct {
	ID        int
	Title     string
	Author    string
	IsStarted bool
}

func (g *Game) GetTasks() []Task {
	q := QB.Select("id", "question", "answers", "correct_answer_idx", "time_to_answer").
		From("quiz").Where("game_id = ?", g.ID).OrderBy("id")
	rows, err := q.Query()
	defer func() { _ = rows.Close() }()

	tasks := make([]Task, 0)
	for rows.Next() {
		quiz := &Quiz{}
		err = rows.Scan(&quiz.ID, &quiz.Question, &quiz.Answers, &quiz.CorrectAnswerIdx, &quiz.TimeToAnswer)
		if err != nil {
			panic(err)
		}
		tasks = append(tasks, quiz)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}
	return tasks
}

func GetGameByHash(hash string) *Game {
	id := GameHashID.Decode(hash)
	if id == -1 {
		return nil
	}

	game := &Game{ID: id}
	q := QB.Select("title", "author").From("games").Where("id = ?", id)
	if err := q.Scan(&game.Title, &game.Author); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return game
}

type Quiz struct {
	ID               int
	Question         string
	Answers          pq.StringArray
	CorrectAnswerIdx int
	TimeToAnswer     int
}

func (q *Quiz) question() string {
	return q.Question
}

func (q *Quiz) answers() []string {
	return q.Answers
}

func (q *Quiz) correctAnswerIdx() int {
	return q.CorrectAnswerIdx
}

func (q *Quiz) timeToAnswer() time.Duration {
	return time.Duration(q.TimeToAnswer) * time.Second
}
