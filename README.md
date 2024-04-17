# Counterfactual Regret Minimization in Go

Originally a self-study into counterfactual regret minimization for solving imperfect-information games, specifically poker, turned into a fun little web application and trying out HTMX.
As of now, this repository contains CFR implementations for [rocker paper scissors](https://en.wikipedia.org/wiki/Rock_paper_scissors), [colonel blotto](https://en.wikipedia.org/wiki/Blotto_game), and [kuhn poker](https://en.wikipedia.org/wiki/Kuhn_poker).  Inspired by [this paper](http://modelai.gettysburg.edu/2013/cfr/cfr.pdf)

```iterations := 10000

rps := newRpsTrainer()
blotto := newBlottoTrainer(10, 4) //newBlottoTrainer(s,n int) s = soldiers, n = battlefields
kuhn := newKuhnTrainer()

rps.train(iterations) 
blotto.train(iterations) 
kurn.train(iterations)
kuhn.getAverageStrategy() 
```

rpsTrainer  by default will train against itself to find perfect 1/3 each equilibrium strategy.

blottoTrainer  searches the entire game tree, which will crash with high inputs of s,n, as there are (s + n - 1)C(n - 1) combinations to choose from and compared. Can be improved.

kuhnTrainer will display all information sets for 3 card kuhn poker  (6 for player 1 and 6 for player 2) as well as their equilibrium strategies 
in the form [0.333,0.666] where 0th element is check/pass and the 1st element is bet/call.

## ToDo
- ~~make a readme~~
- finish ui for kuhn poker to play against ai
- switch between player 1 and player 2 between hands
- display gto strategies
- make code less crappy


super modern v1 ui 
<img width="1085" alt="image" src="https://github.com/PepperoniRollz/cfr/assets/51208066/97e546f8-9179-4b4c-9d2f-429061cc259b">


