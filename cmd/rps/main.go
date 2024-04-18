package main

import (
	"github.com/pepperonirollz/cfr/pkg/rps"
)

func main() {
	trainer := rps.NewRpsTrainer()
	trainer.Train(10000)
}
