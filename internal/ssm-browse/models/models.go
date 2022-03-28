package models

type SSMParameters struct {
	Items []SSMParameter
	NextToken string
}

type SSMParameter struct {
	Name  string
	Value string
}
