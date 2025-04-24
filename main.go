package main

import (
	"anubis-solver/solver"
	"flag"
)

func main() {
	native := flag.Bool("native", false, "Solve the challenge using native Go code (very fast)")
	flag.Parse()

	s, err := solver.New("https://anubis.techaro.lol/")
	if err != nil {
		panic(err)
	}

	if err := s.Solve(*native); err != nil {
		panic(err)
	}
}
