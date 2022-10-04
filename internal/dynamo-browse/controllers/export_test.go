package controllers_test

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestExportController_ExportCSV(t *testing.T) {
	t.Run("should export result set to CSV file", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "bravo-table"})

		tempFile := tempFile(t)

		invokeCommand(t, srv.readController.Init())
		invokeCommand(t, srv.exportController.ExportCSV(tempFile))

		bts, err := os.ReadFile(tempFile)
		assert.NoError(t, err)

		assert.Equal(t, string(bts), strings.Join([]string{
			"pk,sk,alpha,beta,gamma\n",
			"abc,222,This is another some value,1231,\n",
			"bbb,131,,2468,foobar\n",
			"foo,bar,This is some value,,\n",
		}, ""))
	})

	t.Run("should return error if result set is not set", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "non-existant-table"})

		tempFile := tempFile(t)

		invokeCommandExpectingError(t, srv.readController.Init())
		invokeCommandExpectingError(t, srv.exportController.ExportCSV(tempFile))
	})

	t.Run("should honour new columns in CSV file", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		tempFile := tempFile(t)

		invokeCommand(t, srv.readController.Init())

		invokeCommandWithPrompt(t, srv.columnsController.AddColumn(0), "address.no")
		invokeCommand(t, srv.columnsController.ShiftColumnLeft(1))
		invokeCommandWithPrompt(t, srv.columnsController.AddColumn(1), "address.street")
		invokeCommand(t, srv.columnsController.ShiftColumnLeft(1))

		invokeCommand(t, srv.exportController.ExportCSV(tempFile))

		bts, err := os.ReadFile(tempFile)
		assert.NoError(t, err)

		assert.Equal(t, string(bts), strings.Join([]string{
			"pk,address.no,address.street,sk,address,age,alpha,beta,gamma,useMailing\n",
			"abc,123,Fake st.,111,,23,This is some value,,,true\n",
			"abc,,,222,,,This is another some value,1231,,\n",
			"bbb,,,131,,,,2468,foobar,\n",
		}, ""))
	})

	t.Run("should honour hidden columns in CSV file", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		tempFile := tempFile(t)

		invokeCommand(t, srv.readController.Init())

		invokeCommand(t, srv.columnsController.ToggleVisible(1))
		invokeCommand(t, srv.columnsController.ToggleVisible(2))

		invokeCommand(t, srv.exportController.ExportCSV(tempFile))

		bts, err := os.ReadFile(tempFile)
		assert.NoError(t, err)

		assert.Equal(t, string(bts), strings.Join([]string{
			"pk,age,alpha,beta,gamma,useMailing\n",
			"abc,23,This is some value,,,true\n",
			"abc,,This is another some value,1231,,\n",
			"bbb,,,2468,foobar,\n",
		}, ""))
	})

	// Hidden items?
}
