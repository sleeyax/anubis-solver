package main

import "anubis-solver/solver"

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	s, err := solver.New("https://anubis.techaro.lol/")
	if err != nil {
		panic(err)
	}

	if err := s.Solve(); err != nil {
		panic(err)
	}
}
