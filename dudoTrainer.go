package main

// // Thus, a Dudo information set consists of (1) a history of claims
// from the current round, (2) the player’s private roll information, and (3) the number
// of opponent dice

import (
	"fmt"
	"strings"
)

type dudoTrainer struct {
	numSides   int
	numActions int
	dudo       int
	claimNum   []int
	claimRank  []int
	nodeMap    map[int]dudoNode
	infoSet    string
}

type dudoNode struct {
	infoSet     string
	numActions  int
	regretSum   []float64
	strategy    []float64
	strategySum []float64
}

func newDudoTrainer(sides int) *dudoTrainer {
	return &dudoTrainer{
		numSides:   sides,
		numActions: sides*2 + 1,
		claimNum:   []int{1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2},
		claimRank:  []int{2, 3, 4, 5, 6, 1, 2, 3, 4, 5, 6, 1},
		dudo:       sides - 1,
	}
}

func newDudoNode(sides int) *dudoNode {
	return &dudoNode{
		regretSum:   make([]float64, sides),
		strategy:    make([]float64, sides),
		strategySum: make([]float64, sides),
		numActions:  sides,
	}
}

func (n *dudoNode) getStrategy(realizationWeight float64) []float64 {

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

func (n *dudoNode) getAvgStrategy() []float64 {
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

func (d *dudoTrainer) claimHistoryToString(isClaimed []bool) string {
	var sb strings.Builder
	for a := 0; a < d.numActions; a++ {
		if isClaimed[a] {
			if sb.Len() > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf("%d*%d", d.claimNum[a], d.claimRank[a]))
		}
	}
	return sb.String()
}

func (n *dudoNode) infoSetToInteger(playerRoll int, isClaimed []bool) int {
	infoSetNum := playerRoll
	for a := n.numActions - 2; a >= 0; a-- {
		if isClaimed[a] {
			infoSetNum = 2*infoSetNum + 1
		} else {
			infoSetNum = 2 * infoSetNum
		}
	}
	return infoSetNum
}

// func (d *dudoTrainer) dudoCfr(history string, player int, timeStep int, p0 float64, p1 float64) float64 {
// 	//if terminal
// 	//get payoff
// 	//else
// 	//get or crete new node
// 	//call cfr
// 	//return nodeutility
// }

func (d *dudoTrainer) isTerminal(history string) bool {
	// node * dudoNode, ok := d.nodeMap[history]
	return true
}
