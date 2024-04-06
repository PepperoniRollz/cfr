package main

import (
	"fmt"
)

func main() {

	trainer := NewRpsTrainer()
	trainer.train(100)
	fmt.Println("Average Strategy:", trainer.getAverageStrategy())

	blotto := NewBlottoTrainer(100, 3)
	blotto.train(100)
	fmt.Println("Average Blotto strategy:", blotto.getAverageStrategy())
	// fmt.Println("Best: ", blotto.getBestStrategy())
}
