package models

type SSMParameters struct {
	Items []SSMParameter
}

type SSMParameter struct {
	Name  string
	Value string
}
