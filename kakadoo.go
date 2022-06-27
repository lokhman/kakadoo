package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/lokhman/kakadoo/app"
)

func getPool() *app.Pool {
	pool := app.NewPool()
	go pool.Run()
	return pool
}

func getRouter(pool *app.Pool) *gin.Engine {
	r := gin.Default()
	r.HandleMethodNotAllowed = true

	renderer := multitemplate.NewRenderer()
	renderer.AddFromFiles("index", "templates/index.html")
	renderer.AddFromFiles("games", "templates/index.html", "templates/games.html")
	renderer.AddFromFiles("play", "templates/index.html", "templates/play.html")
	renderer.AddFromFiles("find_cat", "templates/index.html", "templates/find_cat.html")
	renderer.AddFromFilesFuncs("podium", template.FuncMap{
		"json": func(v interface{}) template.JS {
			bytes, _ := json.Marshal(v)
			return template.JS(bytes)
		},
	}, "templates/index.html", "templates/podium.html")
	renderer.AddFromFilesFuncs("scores", template.FuncMap{
		"add": func(v int, i int) int {
			return v + i
		},
	}, "templates/index.html", "templates/scores.html")
	r.HTMLRender = renderer

	r.Static("/static", "./static/")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	r.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/")
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", nil)
	})
	r.GET("/games", func(c *gin.Context) {
		type Game struct {
			ID    string `json:"id"`
			Type  string `json:"type"`
			Title string `json:"title"`
			URL   string `json:"url"`
		}
		games := app.GetGames()

		ctx := make([]Game, len(games))
		for i, game := range games {
			id := app.GameHashID.Encode(game.ID)
			ctx[i] = Game{
				ID:    id,
				Type:  game.Type,
				Title: game.Title,
				URL:   fmt.Sprintf("/play/%s", id),
			}
		}
		c.HTML(http.StatusOK, "games", ctx)
	})

	rp := r.Group("/play/:id", func(c *gin.Context) {
		if game := app.GetGameByHash(c.Param("id")); game != nil {
			c.Set("game", game)
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, "/")
		c.Abort()
	})
	rp.GET("", func(c *gin.Context) {
		game := c.MustGet("game").(*app.Game)
		switch game.Type {
		case app.GameTypeFindCat:
			app.FindCat(c)
		default:
			c.HTML(http.StatusOK, "play", gin.H{
				"title":   game.Title,
				"wireURL": c.Request.URL.Path + "/wire",
			})
		}
	})
	rp.POST("", func(c *gin.Context) {
		game := c.MustGet("game").(*app.Game)
		switch game.Type {
		case app.GameTypeFindCat:
			app.FindCat(c)
		default:
			c.AbortWithStatus(http.StatusMethodNotAllowed)
		}
	})
	rp.GET("/wire", func(c *gin.Context) {
		game := c.MustGet("game").(*app.Game)
		player := &app.Player{
			Game: game,
			Name: app.StripHtmlTags(c.Query("player")),
		}
		if player.Name == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if player.Name == game.Author {
			player.IsAuthor = true
		}
		app.WireHandler(pool, player, c.Writer, c.Request)
	})
	rp.GET("/scores", func(c *gin.Context) {
		game := c.MustGet("game").(*app.Game)
		c.HTML(http.StatusOK, "scores", gin.H{
			"game":   game,
			"scores": app.GetScores(game),
		})
	})
	return r
}

func main() {
	pool := getPool()
	router := getRouter(pool)

	err := router.Run()
	if err != nil {
		log.Fatal(err)
	}
}
