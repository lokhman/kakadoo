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
	GameTypeQuiz    = "quiz"
	GameTypeWoC     = "woc"
	GameTypeFindCat = "find_cat"
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
	ID            int
	Type          string
	Title         string
	Author        string
	IsStarted     bool
	LastStartedAt *time.Time
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
	q := QB.Select("id", "type", "title", "author").From("games").OrderBy("created_at")
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
	q := QB.Select("type", "title", "author", "last_started_at").From("games").Where("id = ?", id)
	if err := q.Scan(&game.Type, &game.Title, &game.Author, &game.LastStartedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return game
}

func UpdateGameStartedAt(game *Game, startedAt time.Time) {
	q := QB.Update("games").Set("last_started_at", startedAt).Where("id = ?", game.ID)
	if _, err := q.Exec(); err != nil {
		panic(err)
	}
}

type Score struct {
	Game      *Game
	Task      *Task
	Player    string
	PlayerKey string
	Question  string
	Answer    string
	Score     float64
	CreatedAt time.Time
}

func InsertScores(scores ...*Score) {
	q := QB.Insert("scores").
		Columns("game_id", "task_id", "player", "player_key", "question", "answer", "score", "created_at")
	for _, sc := range scores {
		q = q.Values(sc.Game.ID, sc.Task.ID, sc.Player, sc.PlayerKey, sc.Question, sc.Answer, sc.Score, sc.CreatedAt)
	}
	if _, err := q.Exec(); err != nil {
		panic(err)
	}
}

type _dbScore struct {
	Player    string  `json:"player"`
	Score     float64 `json:"score"`
	Completed int     `json:"completed"`
}

func GetScores(game *Game) []_dbScore {
	scores := make([]_dbScore, 0)
	q := QB.Select("s.player", "SUM(s.score)",
		"COUNT(s.id) * 100 / (SELECT COUNT(t.*) FROM tasks t WHERE t.game_id = g.id) completed").
		From("scores s").Join("games g ON s.game_id = g.id").
		Where("s.game_id = ? AND s.created_at >= g.last_started_at", game.ID).
		GroupBy("g.id", "s.player").OrderBy("SUM(s.score) DESC", "completed", "MAX(s.created_at)")
	rows, err := q.Query()
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		score := _dbScore{}
		err = rows.Scan(&score.Player, &score.Score, &score.Completed)
		if err != nil {
			panic(err)
		}
		scores = append(scores, score)
	}
	return scores
}

func HasPlayerInScores(game *Game, player string, playerKey string) bool {
	var result bool
	q := QB.Select("s.id").Prefix("SELECT EXISTS(").From("scores s").Join("games g ON s.game_id = g.id").
		Where("s.game_id = ? AND s.player = ? AND s.player_key <> ?", game.ID, player, playerKey).
		Where("s.created_at >= g.last_started_at").Suffix(")")
	if err := q.Scan(&result); err != nil {
		panic(err)
	}
	return result
}
