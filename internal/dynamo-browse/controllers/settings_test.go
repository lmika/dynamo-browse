package controllers_test

import (
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/controllers"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSettingsController_SetSetting(t *testing.T) {
	t.Run("read-only setting", func(t *testing.T) {
		srv := newService(t, serviceConfig{})

		msg := invokeCommand(t, srv.settingsController.SetSetting("ro", ""))

		assert.True(t, srv.settingsController.IsReadOnly())
		assert.IsType(t, events.WrappedStatusMsg{}, msg)
		assert.IsType(t, controllers.SettingsUpdated{}, msg.(events.WrappedStatusMsg).Next)
	})

	t.Run("read-write setting", func(t *testing.T) {
		srv := newService(t, serviceConfig{isReadOnly: true})

		msg := invokeCommand(t, srv.settingsController.SetSetting("rw", ""))

		assert.False(t, srv.settingsController.IsReadOnly())
		assert.IsType(t, events.WrappedStatusMsg{}, msg)
		assert.IsType(t, controllers.SettingsUpdated{}, msg.(events.WrappedStatusMsg).Next)
	})

	t.Run("set default limit", func(t *testing.T) {
		srv := newService(t, serviceConfig{})

		assert.Equal(t, 1000, srv.settingProvider.DefaultLimit())
		invokeCommand(t, srv.settingsController.SetSetting("default-limit", "20"))

		assert.Equal(t, 20, srv.settingProvider.DefaultLimit())
	})
}
