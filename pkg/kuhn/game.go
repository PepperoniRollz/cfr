package kuhn

import (
	"fmt"
	"math/rand"
	"sync"
)

type Position int

const (
	first Position = iota
	second
)

type Action int

const (
	Pass Action = iota
	Bet
)

type GameState int

const (
	FirstAction GameState = iota
	SecondAction
	ThirdAction
	PlayerFolded
	AiFolded
	Showdown
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

type Game struct {
	sync.Mutex
	Deck             []rune
	PlayerCard       rune
	PlayerStack      int
	AiCard           rune
	AiStack          int
	AiStrategy       []float64
	Ai               KuhnTrainer
	GameState        GameState
	Pot              int
	GameLog          *GameLogger
	ActionHistory    string
	HandNumber       int
	PlayerPosition   Position
	AiPosition       Position
	PlayerLastAction Action
	AiLastAction     Action
}

func NewGame() *Game {
	d := []rune{'2', '3', '4', '5', '6', '7', '8', '9', 'T', 'J', 'Q', 'K', 'A'}
	trainer := NewKuhnTrainer()
	trainer.Train(100000)
	Shuffle(d)
	return &Game{
		PlayerStack:    10,
		AiStack:        10,
		Deck:           d,
		Ai:             trainer,
		Pot:            0,
		GameLog:        newGameLogger("Starting game\n"),
		GameState:      FirstAction,
		HandNumber:     1,
		PlayerPosition: first,
		AiPosition:     second,
	}
}

func (g *Game) BeginRound() {
	g.GameState = FirstAction
	Shuffle(g.Deck)
	g.PlayerCard = g.Deck[0]
	g.AiCard = g.Deck[1]
	g.PlayerStack--
	g.AiStack--
	g.Pot = 2
	g.ActionHistory = ""
	g.GameLog.append(fmt.Sprintf("Player 1 antes 1\nRoboDurrr antes 1\nYou've been dealt a %c\n", g.PlayerCard))
	if g.PlayerPosition == 0 {
		g.GameLog.append("...waiting for action...")
	}
	if g.AiPosition == 0 {
		g.AiResponse()
	}
}
func (g *Game) getAiAction() Action {
	infoset := fmt.Sprintf("%d %c%s", g.AiPosition, g.AiCard, g.ActionHistory)
	node := g.Ai.NodeMap[infoset]
	randomNumber := rand.Float64()
	fmt.Println("Ai strategy: ", node.GetAvgStrategy())
	cumulativeProbability := 0.0
	var action int
	for index, prob := range node.GetAvgStrategy() {
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

func (g *Game) AiResponse() {
	action := g.getAiAction()

	switch g.GameState {
	case FirstAction:
		if action == Bet {
			g.ActionHistory = g.ActionHistory + "b"
			g.GameLog.append("RoboDurrr has bet 1 currency\n...Waiting for your action...")
			g.Pot++
			g.AiStack--
			g.AiLastAction = Bet
			g.GameState = SecondAction
		} else {
			g.ActionHistory = g.ActionHistory + "p"
			g.GameLog.append("RoboDurrr checked")
			g.AiLastAction = Pass
			g.GameState = SecondAction
		}
	case SecondAction:
		if action == Bet && g.PlayerLastAction == Bet {
			g.Pot++
			g.AiStack--
			g.GameLog.append("RoboDurrr has called")
			g.GameState = Showdown
			resolveRound(g)
		} else if action == Bet && g.PlayerLastAction == Pass {
			g.GameLog.append("RoboDurrr has bet 1 currency\n...Waiting for your action...")
			g.Pot++
			g.AiStack--
			g.AiLastAction = Bet
			g.GameState = ThirdAction
		} else if action == Pass && g.PlayerLastAction == Pass {
			g.GameLog.append("RoboDurrr checked behind")
			g.GameState = Showdown
			resolveRound(g)
		} else if action == Pass && g.PlayerLastAction == Bet {
			g.GameLog.append("RoboDurrr has folded")
			g.GameState = AiFolded
			resolveRound(g)
		}
	case ThirdAction:
		if action == Bet {
			g.Pot++
			g.AiStack--
			g.GameLog.append("RoboDurrr has called")
			g.GameState = Showdown
			resolveRound(g)
		} else {
			g.GameLog.append("RoboDurrr has folded")
			g.GameState = AiFolded
			resolveRound(g)
		}
	}
}

func (g *Game) Check() {
	switch g.GameState {
	case FirstAction:
		g.ActionHistory = g.ActionHistory + "p"
		g.GameLog.append("You have checked")
		g.GameState = SecondAction
		g.PlayerLastAction = Pass
		g.AiResponse()
	case SecondAction: //depends on ai action
		if g.AiLastAction == Pass {
			g.GameLog.append("You have checked behind")
			g.GameState = Showdown
			resolveRound(g)
		} else {
			g.GameLog.append("You have folded")
			g.GameState = PlayerFolded
			resolveRound(g)
		}
	case ThirdAction: //only get third action if you checked and ai bet
		g.GameLog.append("You have folded")
		g.GameState = PlayerFolded
		resolveRound(g)
	}
}
func (g *Game) Bet() {
	switch g.GameState {
	case FirstAction:
		g.ActionHistory = g.ActionHistory + "b"
		g.GameLog.append("You have bet 1 currency")
		g.PlayerStack--
		g.Pot++
		//handle ai response
		g.GameState = SecondAction
		g.PlayerLastAction = Bet
		g.AiResponse()
	case SecondAction:
		if g.AiLastAction == Bet {
			g.GameLog.append("You have called for 1 currency")
			g.PlayerStack--
			g.Pot++
			g.GameState = Showdown
			resolveRound(g)
		} else { //ai passed
			g.ActionHistory = g.ActionHistory + "b"
			g.GameLog.append("You have bet 1 currency")
			g.PlayerStack--
			g.Pot++
			g.GameState = ThirdAction
			g.AiResponse()
		}
	case ThirdAction:
		g.GameLog.append("You have called for 1 currency")
		g.PlayerStack--
		g.Pot++
		g.GameState = Showdown
		resolveRound(g)
	}
}

func resolveRound(game *Game) {
	state := game.GameState
	fmt.Println("Game is now resolving...")
	switch state {
	case Showdown:
		game.GameLog.append(fmt.Sprintf("You showdown a %c\nRoboDurrr shows down a %c", game.PlayerCard, game.AiCard))
		if GetCardRank(game.PlayerCard) > GetCardRank(game.AiCard) {
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
	game.PlayerPosition = (game.PlayerPosition + 1) % 2
	game.AiPosition = (game.AiPosition + 1) % 2
	game.HandNumber++
	game.BeginRound()

}
