package main

import (
	"fmt"
	"math"
	"math/rand"
)

type BlottoTrainer struct {
	S            int
	N            int
	Combinations [][]int
	NumActions   int
	Strategy     []float64
	StrategySum  []float64
	RegretSum    []float64
	OppStrategy  []float64
}

func NewBlottoTrainer(s, n int) *BlottoTrainer {
	var combos [][]int
	generateCombinations([]int{}, s, n, 0, &combos)
	fmt.Println("num combos", len(combos))
	opp := make([]float64, len(combos))
	opp[0] = 1

	return &BlottoTrainer{
		S:            s,
		N:            n,
		Combinations: combos,
		NumActions:   len(combos),
		Strategy:     make([]float64, len(combos)),
		StrategySum:  make([]float64, len(combos)),
		RegretSum:    make([]float64, len(combos)),
		OppStrategy:  opp,
	}
}

func (t *BlottoTrainer) getAction(strategy []float64) int {
	r := rand.Float64()
	a := 0
	var cumulativeProbability float64 = 0
	for a < t.NumActions-1 {
		cumulativeProbability += strategy[a]
		if r < cumulativeProbability {
			break
		}
		a++
	}
	return a
}

func (t *BlottoTrainer) getAverageStrategy() []float64 {
	avgStrategy := make([]float64, t.NumActions) // Create a slice for the average strategy.
	var normalizingSum float64                   // Accumulator for the normalization factor.

	// Sum the strategies to find the normalization factor.
	for _, sum := range t.StrategySum {
		normalizingSum += sum
	}

	// Calculate the average strategy based on the accumulated strategy sums.
	for a, sum := range t.StrategySum {
		if normalizingSum > 0 {
			avgStrategy[a] = sum / normalizingSum // Normalize if there's a non-zero sum.
		} else {
			avgStrategy[a] = 1.0 / float64(t.NumActions) // Distribute evenly if no strategies have been played.
		}
	}

	return avgStrategy
}

func (t *BlottoTrainer) train(iterations int) {
	actionUtility := make([]float64, t.NumActions)

	for i := 0; i < iterations; i++ {
		t.Strategy = t.getStrategy()
		myAction := t.getAction(t.Strategy)
		otherAction := t.getAction(t.Strategy) //Use t.Strategy to find Nash Equilibrium or use oppStrategy to find max explot against a pure strategy

		// Reset actionUtility for each iteration
		for j := range actionUtility {
			actionUtility[j] = 0
		}

		actionUtility = getActionUtility(t, otherAction)

		//winning conditions

		// Accumulate regrets
		for a := 0; a < t.NumActions; a++ {
			t.RegretSum[a] += actionUtility[a] - actionUtility[myAction]
		}
	}
}

func (t *BlottoTrainer) getStrategy() []float64 {
	normalizingSum := 0.0
	for i := 0; i < t.NumActions; i++ {
		if t.RegretSum[i] > 0 {
			t.Strategy[i] = t.RegretSum[i]
		} else {
			t.Strategy[i] = 0
		}
		normalizingSum += t.Strategy[i]
	}

	for i := 0; i < t.NumActions; i++ {
		if normalizingSum > 0 {
			t.Strategy[i] /= normalizingSum
		} else {
			t.Strategy[i] = 1.0 / float64(t.NumActions)
		}
		t.StrategySum[i] += t.Strategy[i]

	}
	return t.Strategy
}

func getActionUtility(t *BlottoTrainer, otherAction int) []float64 {
	actionUtility := make([]float64, t.NumActions)
	for i, value := range t.Combinations {
		actionUtility[i] = float64(evaluateWinner(value, t.Combinations[otherAction]))
	}
	return actionUtility
}

func evaluateWinner(myStrat, oppStrat []int) int {
	myWins, oppWins := 0, 0
	for i := 0; i < len(myStrat); i++ {
		if myStrat[i] > oppStrat[i] {
			myWins++
		} else if myStrat[i] < oppStrat[i] {
			oppWins++
		}
	}
	if myWins > oppWins {
		return 1
	} else if myWins < oppWins {
		return -1
	} else {
		return 0
	}
}

// s = soldiers, n = numBattlefields
func generateCombinations(combination []int, s, n, start int, results *[][]int) {
	if n == 1 {
		newComb := make([]int, len(combination)+1)
		copy(newComb, combination)
		newComb[len(combination)] = s
		*results = append(*results, newComb)
		return
	}

	for i := start; i <= s; i++ {
		nextCombination := append([]int(nil), combination...)
		nextCombination = append(nextCombination, i)
		generateCombinations(nextCombination, s-i, n-1, 0, results)
	}
}

func (t *BlottoTrainer) getBestStrategy() []int {
	max := math.SmallestNonzeroFloat64
	index := -1
	avgStrat := t.getAverageStrategy()
	for i, probability := range avgStrat {
		if probability > max {
			max = probability
			index = i
		}
	}
	fmt.Println(max, "---", t.Combinations[index])
	return t.Combinations[index]
}
