package solver

import "testing"

func TestSolveChallengeNative(t *testing.T) {
	challenge := AnubisChallenge{
		Challenge: "9f3c9b9f4649a0344d8d99f3db64e6b61f5331e2f8d43f9c10e3fbf5b25dc38b",
		Rules: AnubisChallengeRules{
			Difficulty: 4,
			ReportAs:   4,
			Algorithm:  "fast",
		},
	}

	result, err := SolveChallengeNative(&challenge, nil)
	if err != nil {
		t.Fatal(err)
	}

	if result.Hash != "00000aa3566056ac0271355ab75b810d780258a62b2580db50b9a4a86adebee0" {
		t.Errorf("Invalid hash")
	}

	if result.Nonce != 58962 {
		t.Errorf("Invalid nonce")
	}
}
