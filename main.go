package main

import "fmt"

func main() {
	iterations := 10000
	rpsTrainer := NewRpsTrainer()
	rpsTrainer.train(iterations)
	fmt.Println("Average Strategy:", rpsTrainer.getAverageStrategy())

	blotto := NewBlottoTrainer(7, 3)
	blotto.train(iterations)
	fmt.Println("Average Blotto strategy:", blotto.getAverageStrategy())

	newKuhnTrainer().train(iterations)
}
