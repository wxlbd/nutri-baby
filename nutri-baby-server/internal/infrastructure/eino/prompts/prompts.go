package prompts

import _ "embed"

var (
	//go:embed daily_tips.md
	DailyTipsSystem string
	//go:embed analysis_system.md
	AnalysisSystem string
)
