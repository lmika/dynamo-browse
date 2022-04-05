package models

import "github.com/aws/aws-sdk-go-v2/service/ssm/types"

type SSMParameters struct {
	Items []SSMParameter
}

type SSMParameter struct {
	Name  string
	Type types.ParameterType
	Value string
}
