package tray

import (
	"fmt"

	"fyne.io/systray"
	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/config"
	"github.com/joshbeard/walsh/internal/session"
)

// Setup initializes the systray menu items.
func (m *menu) Setup(sess *session.Session) {
	m.createChangeMenu(sess)
	m.createViewMenu(sess)
	m.createBlacklistMenu(sess)
	m.createIntervalMenu(sess)
	m.createUseListMenu()

	systray.AddSeparator()
	m.quit = systray.AddMenuItem("Quit Walsh", "")
}

func (m *menu) createIntervalMenu(sess *session.Session) {
	mRotateInterval := systray.AddMenuItem("Rotate Interval…", "")

	m.intervals = intervalMenu{parent: mRotateInterval}
	m.addIntervalSubMenuItems(mRotateInterval, sess)
	log.Debug("Interval menu created")
}

func (m *menu) addIntervalSubMenuItems(parent *systray.MenuItem, sess *session.Session) {
	menuHasInterval := false
	set := make(map[config.RotateInterval]bool)

	// Prepend a '0' to the interval list
	menuIntervals := append([]config.RotateInterval{0}, sess.Config().SystemTray.Intervals...)

	for _, interval := range menuIntervals {
		if set[interval] {
			continue
		}
		set[interval] = true

		check, checked := m.getIntervalCheckMark(interval, sess)
		item := parent.AddSubMenuItemCheckbox(fmt.Sprintf("%s%s", check, interval.String()), "", checked)

		m.intervals.subs = append(m.intervals.subs, intervalItem{interval: int(interval), item: item})

		if interval == 0 {
			parent.AddSeparator()
		}
	}

	if !menuHasInterval {
		parent.AddSubMenuItemCheckbox(fmt.Sprintf(" %d", sess.Interval()), "", false)
	}
}

func (m *menu) getIntervalCheckMark(interval config.RotateInterval, sess *session.Session) (string, bool) {
	if int(interval) == sess.Interval() {
		if sess.Type() == session.SessionTypeMacOS {
			return "", true
		}

		return "✔ ", true
	}
	if sess.Type() == session.SessionTypeMacOS {
		return "", false
	}
	return "  ", false
}

func (m *menu) createViewMenu(sess *session.Session) {
	displays := sess.Displays()
	m.view = m.createDisplayMenuItem("View Wallpaper", "View Wallpaper…", displays)
}

func (m *menu) createChangeMenu(sess *session.Session) {
	displays := sess.Displays()
	m.change = m.createDisplayMenuItem("Change Wallpaper", "Change Wallpaper…", displays)
}

func (m *menu) createBlacklistMenu(sess *session.Session) {
	displays := sess.Displays()
	m.blacklist = m.createDisplayMenuItem("Blacklist", "Blacklist…", displays)
}

func (m *menu) createDisplayMenuItem(singleTitle, multiTitle string, displays []session.Display) menuItem {
	if len(displays) < 2 {
		return menuItem{
			parent: systray.AddMenuItem(singleTitle, ""),
			value:  displays[0].ID,
		}
	}

	parent := systray.AddMenuItem(multiTitle, "")
	var subs []menuItem
	for _, d := range displays {
		subs = append(subs, menuItem{
			parent: parent.AddSubMenuItem(fmt.Sprintf("%s: %s", d.ID, d.Name), ""),
			value:  d.ID,
		})
	}

	return menuItem{parent: parent, subs: subs}
}

func (m *menu) createUseListMenu() {
	mUseList := systray.AddMenuItem("Use List…", "")
	useListItems := []string{"Nature", "Favorites", "Mountains"}
	for _, item := range useListItems {
		mUseList.AddSubMenuItem(item, "")
	}

	m.useList = menuItem{parent: mUseList}
}

