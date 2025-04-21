package main

import "anubis-solver/solver"

func main() {
	s, err := solver.New("https://anubis.techaro.lol/")
	if err != nil {
		panic(err)
	}

	if err := s.Solve(); err != nil {
		panic(err)
	}
}
