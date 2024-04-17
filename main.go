package main

import (
	"fmt"
	"html/template"
	"io"
	"sync"

	"math/rand"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var game *Game

type GameState int
type Action int

const (
	Pass Action = iota
	Bet
)

const (
	FirstAction GameState = iota
	SecondAction
	PlayerFolded
	AiFolded
	Showdown
)

type Game struct {
	sync.Mutex
	Deck          []int
	Running       bool
	PlayerCard    int
	PlayerStack   int
	AiCard        int
	AiStack       int
	AiStrategy    []float64
	Ai            kuhnTrainer
	GameState     GameState
	Pot           int
	GameLog       string
	ActionHistory string
	HandNumber    int
}

func newGame() *Game {
	d := []int{1, 2, 3}
	trainer := newKuhnTrainer()
	trainer.train(10000)
	shuffle(d)
	return &Game{
		Running:     true,
		PlayerStack: 10,
		AiStack:     10,
		Deck:        d,
		Ai:          trainer,
		Pot:         0,
		GameLog:     "Starting Game\n",
		GameState:   FirstAction,
	}
}

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("./static/*html")),
	}
}
func (g *Game) beginRound() {
	g.GameState = FirstAction
	shuffle(g.Deck)
	g.PlayerCard = g.Deck[0]
	g.AiCard = g.Deck[1]
	g.PlayerStack--
	g.AiStack--
	g.Pot = 2
	g.ActionHistory = ""
	g.GameLog = g.GameLog + fmt.Sprintf("Player 1 antes 1\nAi antes 1\nYou've been dealt a %d\n", g.PlayerCard) + "...waiting for action...\n"
}
func (g *Game) aiAction() Action {
	node := g.Ai.nodeMap[fmt.Sprintf("1 %d%s", g.AiCard, g.ActionHistory)]
	randomNumber := rand.Float64()
	fmt.Println("Ai strategy: ", node.getAvgStrategy())
	cumulativeProbability := 0.0
	var action int
	for index, prob := range node.getAvgStrategy() {
		cumulativeProbability += prob
		if randomNumber < cumulativeProbability {
			action = index
			break
		}
	}
	if action == 0 {
		return Pass
	} else {
		return Bet
	}

}

func (g *Game) check() {
	fmt.Println("gamestate check", g.GameState)
	g.ActionHistory = g.ActionHistory + "p"

	if g.GameState == FirstAction {
		g.GameLog = g.GameLog + "You have checked\n"
		aiAction := g.aiAction()
		if aiAction == Pass {
			g.ActionHistory = g.ActionHistory + "p"
			g.GameLog = g.GameLog + "The ai has checked behind\n"
			g.GameState = Showdown
			resolveRound(g)
		} else {
			g.ActionHistory = g.ActionHistory + "b"
			g.GameLog = g.GameLog + "The ai has bet 1 unit\n"

			g.Pot++
			g.AiStack--
			g.GameState = SecondAction
		}
	} else if g.GameState == SecondAction {
		g.GameLog = g.GameLog + "You have folded\n"
		fmt.Println("I AM FOOOOOOOLDING")
		g.GameState = PlayerFolded
		resolveRound(g)
	}
}
func (g *Game) bet() {
	g.ActionHistory = g.ActionHistory + "b"

	if g.GameState == FirstAction {
		g.GameLog = g.GameLog + "You have bet 1 unit\n"
		g.PlayerStack--
		g.Pot++
		aiAction := g.aiAction()
		if aiAction == Pass {
			g.GameLog = g.GameLog + "The ai has folded\n"
			g.GameState = AiFolded
			resolveRound(g)

		} else {
			g.Pot++
			g.AiStack--
			g.GameLog = g.GameLog + "The ai has called\n"
			g.GameState = Showdown
			resolveRound(g)
		}
	} else if g.GameState == SecondAction {
		g.GameLog = g.GameLog + "You have called for 1 unit\n"
		g.PlayerStack--
		g.Pot++
		g.GameState = Showdown
		resolveRound(game)
	}

}

func resolveRound(game *Game) {

	state := game.GameState
	if state == Showdown || state == PlayerFolded || state == AiFolded {
		fmt.Println("Game is now resolving...")
		switch state {
		case Showdown:
			game.GameLog = game.GameLog + fmt.Sprintf("You showdown a %d\nRobot shows down a %d\n", game.PlayerCard, game.AiCard)
			if game.PlayerCard > game.AiCard {
				game.PlayerStack += game.Pot
				game.GameLog = game.GameLog + "You have won!\n"
			} else {
				game.AiStack += game.Pot
				game.GameLog = game.GameLog + "Robot has won!\n"
			}
		case PlayerFolded:
			game.AiStack += game.Pot
			game.GameLog = game.GameLog + "Robot has won!\n"
		case AiFolded:
			game.PlayerStack += game.Pot
			game.GameLog = game.GameLog + "You have won!\n"
		}
		game.Pot = 0
		game.HandNumber++
		game.beginRound()
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
		game = newGame()
		game.beginRound()
		return c.Render(200, "dashboard", game)
	})
	e.POST("/pass", func(c echo.Context) error {
		game.check()
		return c.Render(200, "dashboard", game)
	})
	e.POST("/bet", func(c echo.Context) error {
		game.bet()
		return c.Render(200, "dashboard", game)
	})

	e.Logger.Fatal(e.Start(":8080"))

}
