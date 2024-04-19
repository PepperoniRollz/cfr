package main

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pepperonirollz/cfr/pkg/kuhn"
)

var game *kuhn.Game

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("../../static/*html")),
	}
}

func main() {

	e := echo.New()
	e.Use(middleware.Logger())
	e.Renderer = newTemplate()

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", nil)
	})

	e.POST("/start", func(c echo.Context) error {
		game = kuhn.NewGame()
		game.BeginRound()
		return c.Render(200, "dashboard", game)
	})
	e.POST("/pass", func(c echo.Context) error {
		game.Check()
		return c.Render(200, "dashboard", game)
	})
	e.POST("/bet", func(c echo.Context) error {
		game.Bet()
		return c.Render(200, "dashboard", game)
	})
	e.Logger.Fatal(e.Start(":8080"))

}
