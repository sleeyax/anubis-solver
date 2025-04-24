package solver

type AnubisChallengeRules struct {
	Difficulty int    `json:"difficulty"`
	ReportAs   int    `json:"report_as"`
	Algorithm  string `json:"algorithm"`
}

type AnubisChallenge struct {
	Challenge string               `json:"challenge"`
	Rules     AnubisChallengeRules `json:"rules"`
}
