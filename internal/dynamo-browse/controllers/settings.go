package controllers

import "log"

type SettingsController struct {
	settings SettingsProvider
}

func NewSettingsController(sp SettingsProvider) *SettingsController {
	return &SettingsController{
		settings: sp,
	}
}

func (sc *SettingsController) IsReadOnly() bool {
	ro, err := sc.settings.IsReadOnly()
	if err != nil {
		log.Printf("warn: unable to determine if R/O is available: %v", err)
		return false
	}
	return ro
}
