package main

import (
	"fmt"
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
	r.HTMLRender = renderer

	r.Static("/static", "./static/")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	r.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/")
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", nil)
	})

	rp := r.Group("/play/:id", func(c *gin.Context) {
		game := app.GetGameByHash(c.Param("id"))
		if game == nil {
			c.Redirect(http.StatusTemporaryRedirect, "/")
			c.Abort()
			return
		}
		c.Set("game", game)
	})

	rp.GET("", func(c *gin.Context) {
		game := c.MustGet("game").(*app.Game)
		c.HTML(http.StatusOK, "play", gin.H{
			"title":   game.Title,
			"wireURL": c.Request.URL.Path + "/wire",
		})
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

	r.GET("/games", func(c *gin.Context) {
		type Game struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			URL   string `json:"url"`
		}
		games := app.GetGames()

		ctx := make([]Game, len(games))
		for i, game := range games {
			id := app.GameHashID.Encode(game.ID)
			ctx[i] = Game{
				ID:    id,
				Title: game.Title,
				URL:   fmt.Sprintf("/play/%s", id),
			}
		}
		c.HTML(http.StatusOK, "games", ctx)
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
