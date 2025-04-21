package solver

import (
	"fmt"
	"os"
	"testing"
)

func TestSolveChallenge(t *testing.T) {
	sourceBytes, err := os.ReadFile("../test/data/challenge.js")
	if err != nil {
		t.Fatalf("failed to read challenge file: %v", err)
	}

	source := string(sourceBytes)
	v, err := SolveChallenge("https://anubis.techaro.lol/", source, "v1.16.0-24-g75b97eb", `{"challenge":"9f3c9b9f4649a0344d8d99f3db64e6b61f5331e2f8d43f9c10e3fbf5b25dc38b","rules":{"difficulty":4,"report_as":4,"algorithm":"fast"}}`)
	if err != nil {
		t.Fatalf("failed to solve challenge: %v", err)
	}

	if v == "" {
		t.Fatal("expected non-empty result")
	}

	fmt.Println(v)
}
