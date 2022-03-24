package controllers

import "github.com/lmika/awstools/internal/dynamo-browse/models"

type NewResultSet struct {
	ResultSet *models.ResultSet
}

type SetReadWrite struct {
	NewValue bool
}
