package main

import (
	"math/rand"
	"time"
)

type RpsTrainer struct {
	Rock        int
	Paper       int
	Scissors    int
	NumActions  int
	Strategy    []float64
	StrategySum []float64
	RegretSum   []float64
	OppStrategy []float64
}

// NewRpsTrainer is a constructor function for RpsTrainer that initializes its fields.
func NewRpsTrainer() *RpsTrainer {
	numActions := 3
	return &RpsTrainer{
		Rock:        0,
		Paper:       0,
		Scissors:    0,
		NumActions:  numActions,
		Strategy:    make([]float64, numActions),
		StrategySum: make([]float64, numActions),
		RegretSum:   make([]float64, numActions),
		OppStrategy: []float64{0.4, 0.4, 0.2},
	}
}
func (t *RpsTrainer) getStrategy() []float64 {
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
func (t *RpsTrainer) getAction(strategy []float64) int {
	rand.Seed(time.Now().UnixNano())

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

func (t *RpsTrainer) train(iterations int) {
	actionUtility := make([]float64, t.NumActions)

	for i := 0; i < iterations; i++ {
		t.Strategy = t.getStrategy() // Update the strategy at the start of each iteration
		myAction := t.getAction(t.Strategy)
		otherAction := t.getAction(t.Strategy) // Assuming this simulates an opponent's action based on t.OppStrategy

		// Reset actionUtility for each iteration
		for j := range actionUtility {
			actionUtility[j] = 0
		}

		// Determine utility of playing each action against the otherAction
		if otherAction == t.NumActions-1 {
			actionUtility[0] = 1
		} else {
			actionUtility[otherAction+1] = 1
		}
		if otherAction == 0 {
			actionUtility[t.NumActions-1] = -1
		} else {
			actionUtility[otherAction-1] = -1
		}

		// Accumulate regrets
		for a := 0; a < t.NumActions; a++ {
			t.RegretSum[a] += actionUtility[a] - actionUtility[myAction]
		}
	}
}
func (t *RpsTrainer) getAverageStrategy() []float64 {
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
