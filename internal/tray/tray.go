package tray

import (
	"fyne.io/systray"
	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/session"
	"github.com/joshbeard/walsh/internal/tray/icon"
)

type menu struct {
	Change    menuItem
	View      menuItem
	Blacklist menuItem
	Intervals intervalMenu
	UseList   menuItem
	AddToList menuItem
	Quit      *systray.MenuItem
}

type intervalMenu struct {
	parent *systray.MenuItem
	subs   []intervalItem
}

type intervalItem struct {
	interval int
	item     *systray.MenuItem
}

type menuItem struct {
	parent *systray.MenuItem
	value  string
	subs   []menuItem
}

// OnReady is the entry point when the systray is ready.
func OnReady() {
	sess := session.Current
	m := &menu{}

	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTooltip("Walsh")

	setupMenuItems(m, sess)

	go handleMenuEvents(sess, *m)
	go handleDisplayEvents(sess, *m)
	go handleIntervalEvents(sess, *m)
}

func OnExit() {
	log.Info("Closing Walsh tray")
}
