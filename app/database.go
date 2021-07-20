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

const (
	gameTypeQuiz = "quiz"
	gameTypeWoC  = "woc"
)

type Task struct {
	ID            int
	Question      string
	Answers       pq.StringArray
	CorrectAnswer string
	TimeToAnswer  int
}

func (t *Task) timeToAnswer() time.Duration {
	return time.Duration(t.TimeToAnswer) * time.Second
}

type Game struct {
	ID        int
	Type      string
	Title     string
	Author    string
	IsStarted bool
}

func (g *Game) GetTasks() []*Task {
	q := QB.Select("id", "question", "answers", "correct_answer", "time_to_answer").
		From("tasks").Where("game_id = ?", g.ID).OrderBy("id")
	rows, err := q.Query()
	defer func() { _ = rows.Close() }()

	tasks := make([]*Task, 0)
	for rows.Next() {
		task := &Task{}
		err = rows.Scan(&task.ID, &task.Question, &task.Answers, &task.CorrectAnswer, &task.TimeToAnswer)
		if err != nil {
			panic(err)
		}
		tasks = append(tasks, task)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}
	return tasks
}

func GetGames() []*Game {
	q := QB.Select("id", "type", "title", "author").From("games").OrderBy("created_at DESC")
	rows, err := q.Query()
	defer func() { _ = rows.Close() }()

	games := make([]*Game, 0)
	for rows.Next() {
		game := &Game{}
		err = rows.Scan(&game.ID, &game.Type, &game.Title, &game.Author)
		if err != nil {
			panic(err)
		}
		games = append(games, game)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}
	return games
}

func GetGameByHash(hash string) *Game {
	id := GameHashID.Decode(hash)
	if id == -1 {
		return nil
	}

	game := &Game{ID: id}
	q := QB.Select("type", "title", "author").From("games").Where("id = ?", id)
	if err := q.Scan(&game.Type, &game.Title, &game.Author); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return game
}
