package solver

type AnubisChallenge struct {
	Challenge string `json:"challenge"`
	Rules     struct {
		Difficulty int    `json:"difficulty"`
		ReportAs   int    `json:"report_as"`
		Algorithm  string `json:"algorithm"`
	}
}
