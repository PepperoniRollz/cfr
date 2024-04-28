package kuhn

import (
	"fmt"
	"math/rand"
	"strconv"
)

type KuhnTrainer struct {
	numActions int
	NodeMap    map[string]*kuhnNode
}

type kuhnNode struct {
	numActions  int
	infoSet     string
	regretSum   []float64
	strategy    []float64
	strategySum []float64
}

func NewKuhnTrainer() KuhnTrainer {
	return KuhnTrainer{
		numActions: 2,
		NodeMap:    make(map[string]*kuhnNode),
	}
}

func newKuhnNode(p int) *kuhnNode {
	return &kuhnNode{
		numActions:  2,
		infoSet:     "",
		regretSum:   make([]float64, 2),
		strategy:    make([]float64, 2),
		strategySum: make([]float64, 2),
	}
}

func (n *kuhnNode) getStrategy(realizationWeight float64) []float64 {

	normalizingSum := 0.0
	for i := 0; i < n.numActions; i++ {
		if n.regretSum[i] > 0 {
			n.strategy[i] = n.regretSum[i]
		} else {
			n.strategy[i] = 0
		}
		normalizingSum += n.strategy[i]
	}

	if normalizingSum > 0 {
		for i := 0; i < n.numActions; i++ {
			n.strategy[i] /= float64(normalizingSum)
		}
	} else {
		for i := 0; i < n.numActions; i++ {
			n.strategy[i] = 1.0 / float64(n.numActions)
		}
	}

	for i := 0; i < n.numActions; i++ {
		n.strategySum[i] += realizationWeight * n.strategy[i]
	}

	return n.strategy
}

func (n *kuhnNode) GetAvgStrategy() []float64 {
	avgStrategy := make([]float64, n.numActions)
	var normalizingSum float64
	for i := 0; i < n.numActions; i++ {
		normalizingSum += n.strategySum[i]
	}
	for i := 0; i < n.numActions; i++ {
		if normalizingSum > 0 {
			avgStrategy[i] = n.strategySum[i] / normalizingSum
		} else {
			avgStrategy[i] = 1.0 / float64(n.numActions)
		}
	}
	for i, value := range avgStrategy {
		if value < 0.001 {
			avgStrategy[i] = 0
		}
	}
	return avgStrategy
}

func (n kuhnNode) String() string {
	return fmt.Sprintf("%4s: %v", n.infoSet, n.GetAvgStrategy())
}

func (k KuhnTrainer) Train(iterations int) {
	cards := []rune{'2', '3', '4', '5', '6', '7', '8', '9', 'T', 'J', 'Q', 'K', 'A'}
	util := 0.0
	for i := 0; i < iterations; i++ {
		Shuffle(cards)
		util += k.cfr(cards, "", 1, 1)
	}
	fmt.Println("Expected value: ", util/float64(iterations), "player 2: ", -1*util/float64(iterations))
	for _, node := range k.NodeMap {
		fmt.Println(node.String())
	}
	fmt.Println("Num infosets: ", len(k.NodeMap))
}

func Shuffle(cards []rune) {
	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
}

func (k *KuhnTrainer) cfr(cards []rune, history string, p0 float64, p1 float64) float64 {
	plays := len(history)
	player := plays % 2
	opponent := 1 - player
	terminalStatePayoff := terminalStatePayoff(cards, plays, history, player, opponent)
	if terminalStatePayoff != 0 {
		return float64(terminalStatePayoff)
	}
	infoSet := strconv.Itoa(player) + " " + string(cards[player]) + history
	node := k.getOrCreateKuhnNode(infoSet, player)

	var strategy []float64
	if player == 0 {
		strategy = node.getStrategy(p0)
	} else {
		strategy = node.getStrategy(p1)
	}

	util := make([]float64, node.numActions)
	nodeUtil := 0.0

	for i := 0; i < node.numActions; i++ {
		var nextHistory string
		if i == 0 {
			nextHistory = history + "p"
		} else {
			nextHistory = history + "b"
		}
		if player == 0 {
			util[i] = -k.cfr(cards, nextHistory, p0*strategy[i], p1)
		} else {
			util[i] = -k.cfr(cards, nextHistory, p0, p1*strategy[i])
		}

		nodeUtil += strategy[i] * util[i]
	}

	for i := 0; i < node.numActions; i++ {
		regret := util[i] - nodeUtil

		if player == 0 {
			node.regretSum[i] += p1 * regret
		} else {
			node.regretSum[i] += p0 * regret
		}
	}
	return nodeUtil
}

func terminalStatePayoff(cards []rune, plays int, history string, player int, opponent int) int {
	if plays > 1 {

		terminalPass := false
		doubleBet := false
		isPlayerCardHigher := false
		if history[plays-1] == 'p' {
			terminalPass = true
		}
		if history[plays-2:plays] == "bb" {
			doubleBet = true
		}
		if GetCardRank(cards[player]) > GetCardRank(cards[opponent]) {
			isPlayerCardHigher = true
		}
		if terminalPass {
			if history == "pp" {
				if isPlayerCardHigher {
					return 1
				} else {
					return -1
				}
			} else {
				return 1
			}

		} else if doubleBet {

			if isPlayerCardHigher {
				return 2
			} else {
				return -2
			}
		}
	}
	return 0
}

func (k *KuhnTrainer) getOrCreateKuhnNode(infoSet string, player int) *kuhnNode {
	node, ok := k.NodeMap[infoSet]
	if !ok {
		node = newKuhnNode(player)
		node.infoSet = infoSet
		k.NodeMap[infoSet] = node
		return node
	}
	return node
}
func GetCardRank(r rune) int {
	switch r {
	case '2':
		return 2
	case '3':
		return 3
	case '4':
		return 4
	case '5':
		return 5
	case '6':
		return 6
	case '7':
		return 7
	case '8':
		return 8
	case '9':
		return 9
	case 'T': // T stands for 10 in some card notation systems
		return 10
	case 'J': // Jack
		return 11
	case 'Q': // Queen
		return 12
	case 'K': // King
		return 13
	case 'A': // Ace
		return 14
	default:
		return 0 // for any undefined cards
	}
}
