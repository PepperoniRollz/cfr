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
	Check
	Fold
)

const (
	FirstAction GameState = iota
	AiTurn
	SecondAction
	Resolving
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
	HandNumber    int
	Pot           int
	GameLog       string
	ActionHistory string
	StateChan     chan GameState
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
		StateChan:   make(chan GameState, 1), // Buffered channel
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
		templates: template.Must(template.ParseGlob("*.html")),
	}
}
func (g *Game) beginRound() {
	if g.Running {
		shuffle(g.Deck)
		g.PlayerCard = g.Deck[0]
		g.AiCard = g.Deck[1]
		g.PlayerStack--
		g.AiStack--
		g.Pot = 2
		g.GameLog = g.GameLog + fmt.Sprintf("Player 1 antes 1\nAi antes 1\nYou've been dealt a %d\n", g.PlayerCard) + "...waiting for action...\n"
		g.GameState = FirstAction
	}
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
	fmt.Println("Ai strategy: ", action)
	if action == 0 {
		return Pass
	} else {
		return Bet
	}

}
func (g *Game) isTerminalState() bool {
	length := len(g.ActionHistory)
	if length > 1 {
		if g.ActionHistory[length-1] == 'p' || g.ActionHistory[length-2:length] == "bb" {
			return true
		}
	}
	return false
}

func (g *Game) check() {

	g.Lock()
	g.ActionHistory = g.ActionHistory + "p"

	if g.GameState == FirstAction {
		g.GameState = AiTurn
		g.GameLog = g.GameLog + "You have checked\n"
		aiAction := g.aiAction()
		if aiAction == Pass {
			g.ActionHistory = g.ActionHistory + "p"
			g.GameLog = g.GameLog + "The ai has checked\n"
			//resolve hand
		} else {
			g.ActionHistory = g.ActionHistory + "b"
			g.GameLog = g.GameLog + "The ai has bet 1 unit\n"

			g.Pot++
			g.AiStack--
		}
		g.GameState = SecondAction
	} else if g.GameState == SecondAction {
		//resolve hand ai wins becasue you folded
		g.GameState = Resolving
	}
	g.StateChan <- g.GameState // Send new state to channel
	g.Unlock()
}
func (g *Game) bet() {
	g.Lock()
	g.ActionHistory = g.ActionHistory + "b"

	if g.GameState == FirstAction {
		g.GameState = AiTurn
		g.GameLog = g.GameLog + "You have bet 1 unit\n"
		g.PlayerStack--
		g.Pot++
		g.GameState = AiTurn
		aiAction := g.aiAction()
		if aiAction == Pass {
			g.GameLog = g.GameLog + "The ai has folded\n"
			//resolve hand becasue ai folded to bet
		} else {
			g.Pot++
			g.AiStack--
			g.GameLog = g.GameLog + "The ai has called\n"
			//resolve hand because ai has called a bet
		}
		g.GameState = Resolving
	} else if g.GameState == SecondAction {
		//you called
		g.GameLog = g.GameLog + "You have called for 1 unit\n"
		g.PlayerStack--
		g.Pot++
		//resolve hand
		g.GameState = Resolving
	}
	g.StateChan <- g.GameState // Send new state to channel
	g.Unlock()

}

func watchGameState(game *Game) {
	for state := range game.StateChan { // Receive from channel
		if state == Resolving {
			fmt.Println("Game is now resolving...")

			// Handle the resolving state
		}
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
		go watchGameState(game)
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
