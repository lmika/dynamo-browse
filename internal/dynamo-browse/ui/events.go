package ui

import "github.com/lmika/awstools/internal/dynamo-browse/models"

type newResultSet struct {
	ResultSet *models.ResultSet
}

type setStatusMessage string
type errorRaised error