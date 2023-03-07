package mailing

type MonthCount struct {
	Month string
	Count int
}

type TemplateData struct {
	Total       string
	AvgCredit   string
	AvgDebit    string
	MonthCounts []MonthCount
}
