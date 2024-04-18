package main

import (
	"fmt"
	"html/template"
	"io"
	"sync"

	"math/rand"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pepperonirollz/cfr/pkg/kuhn"
)

type GameLogger struct {
	Log string
}

func newGameLogger(s string) *GameLogger {
	return &GameLogger{
		Log: s + "\n",
	}
}
func (l *GameLogger) append(s string) {
	l.Log += s + "\n"
}

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
	Ai            kuhn.KuhnTrainer
	GameState     GameState
	Pot           int
	GameLog       GameLogger
	ActionHistory string
	HandNumber    int
}

func newGame() *Game {
	d := []int{1, 2, 3}
	trainer := kuhn.NewKuhnTrainer()
	trainer.Train(10000)
	kuhn.Shuffle(d)
	return &Game{
		Running:     true,
		PlayerStack: 10,
		AiStack:     10,
		Deck:        d,
		Ai:          trainer,
		Pot:         0,
		GameLog:     *newGameLogger("Starting game\n"),
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
	kuhn.Shuffle(g.Deck)
	g.PlayerCard = g.Deck[0]
	g.AiCard = g.Deck[1]
	g.PlayerStack--
	g.AiStack--
	g.Pot = 2
	g.ActionHistory = ""
	g.GameLog.append(fmt.Sprintf("Player 1 antes 1\nRoboDurrr antes 1\nYou've been dealt a %d\n", g.PlayerCard) + "...waiting for action...")
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
		g.GameLog.append("You have checked")
		aiAction := g.aiAction()
		if aiAction == Pass {
			g.ActionHistory = g.ActionHistory + "p"
			g.GameLog.append("RoboDurrr checked behind")
			g.GameState = Showdown
			resolveRound(g)
		} else {
			g.ActionHistory = g.ActionHistory + "b"
			g.GameLog.append("RoboDurrr has bet 1 currency\n...Waiting for your action...")

			g.Pot++
			g.AiStack--
			g.GameState = SecondAction
		}
	} else if g.GameState == SecondAction {
		g.GameLog.append("You have folded")
		g.GameState = PlayerFolded
		resolveRound(g)
	}
}
func (g *Game) bet() {
	g.ActionHistory = g.ActionHistory + "b"

	if g.GameState == FirstAction {
		g.GameLog.append("You have bet 1 currency")
		g.PlayerStack--
		g.Pot++
		aiAction := g.aiAction()
		if aiAction == Pass {
			g.GameLog.append("RoboDurrr has folded")
			g.GameState = AiFolded
			resolveRound(g)

		} else {
			g.Pot++
			g.AiStack--
			g.GameLog.append("RoboDurrr has called")
			g.GameState = Showdown
			resolveRound(g)
		}
	} else if g.GameState == SecondAction {
		g.GameLog.append("You have called for 1 currency")
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
			game.GameLog.append(fmt.Sprintf("You showdown a %d\nRoboDurrr shows down a %d", game.PlayerCard, game.AiCard))
			if game.PlayerCard > game.AiCard {
				game.PlayerStack += game.Pot
				game.GameLog.append("You have won!")
			} else {
				game.AiStack += game.Pot
				game.GameLog.append("RoboDurrr has won!")
			}
		case PlayerFolded:
			game.AiStack += game.Pot
			game.GameLog.append("RoboDurrr has won!")
		case AiFolded:
			game.PlayerStack += game.Pot
			game.GameLog.append("You have won!")
		}
		game.GameLog.append("\n******* New Hand *******\n")
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
