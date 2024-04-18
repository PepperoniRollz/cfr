package main

import (
	"github.com/pepperonirollz/cfr/pkg/blotto"
)

func main() {
	trainer := blotto.NewBlottoTrainer(10, 4)
	trainer.Train(10000)
}
