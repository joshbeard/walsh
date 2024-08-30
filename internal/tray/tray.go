package tray

import (
	_ "embed"

	"fyne.io/systray"
	"github.com/joshbeard/walsh/internal/session"
)

//go:embed icon/icon-dark.png
var icon []byte

type menu struct {
	quit      *systray.MenuItem
	intervals intervalMenu
	view      menuItem
	change    menuItem
	blacklist menuItem
	useList   menuItem
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

func Run() {
	systray.SetTemplateIcon(icon, icon)
	systray.SetTooltip("Walsh")

	sess := session.Current
	m := &menu{}
	m.Setup(sess)

	go m.handleMenuEvents(sess)
	go m.handleDisplayEvents(sess)
	go m.handleIntervalEvents(sess)
}
