package tray

import (
	"fmt"

	"fyne.io/systray"
	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/config"
	"github.com/joshbeard/walsh/internal/session"
)

// setupMenuItems sets up the systray menu items and handles click events.
func setupMenuItems(m *menu, sess *session.Session) {
	m.createChangeMenu(sess)
	m.createViewMenu(sess)
	m.createBlacklistMenu(sess)
	m.createIntervalMenu(sess)
	m.createUseListMenu(sess)

	systray.AddSeparator()
	m.Quit = systray.AddMenuItem("Quit Walsh", "")
}

func (m *menu) createIntervalMenu(sess *session.Session) {
	mRotateInterval := systray.AddMenuItem("Rotate Interval…", "")

	menuHasInterval := false
	set := []config.RotateInterval{}

	m.Intervals = intervalMenu{parent: mRotateInterval}
	// Prepend a '0' to the interval list
	menuIntervals := sess.Config().MenuIntervals
	menuIntervals = append([]config.RotateInterval{0}, menuIntervals...)

	for _, interval := range menuIntervals {
		check := "  "
		checked := false

		if interval.InList(set) {
			continue
		}

		set = append(set, interval)

		// Check the current interval
		if int(interval) == sess.Interval() {
			menuHasInterval = true
			check = "✔ "
			checked = true
		}

		if sess.Type() == session.SessionTypeMacOS {
			check = ""
		}

		s := mRotateInterval.AddSubMenuItemCheckbox(fmt.Sprintf("%s%s", check, interval.String()), "", checked)

		m.Intervals.subs = append(m.Intervals.subs, intervalItem{interval: int(interval), item: s})

		if interval == 0 {
			// Add separator after the 'Pause' interval
			mRotateInterval.AddSeparator()
		}
	}

	if !menuHasInterval {
		mRotateInterval.AddSubMenuItemCheckbox(fmt.Sprintf(" %d", sess.Interval()), "", false)
	}

	log.Debug("Interval menu created")
}

func (m *menu) createViewMenu(sess *session.Session) {
	displays := sess.Displays()
	if len(displays) < 2 {
		m.View = menuItem{
			parent: systray.AddMenuItem("View Wallpaper", ""),
			value:  displays[0].ID,
		}
		return
	}

	// var viewSubs []*systray.MenuItem
	var viewSubs []menuItem
	mView := systray.AddMenuItem("View Wallpaper…", "")
	for _, d := range displays {
		// sub := menuitem{}
		// viewSubs = append(viewSubs, mView.AddSubMenuItem(
		// 	fmt.Sprintf("%s: %s", d.ID, d.Name), ""))
		viewSubs = append(viewSubs, menuItem{
			parent: mView.AddSubMenuItem(fmt.Sprintf("%s: %s", d.ID, d.Name), ""),
			value:  d.ID,
		})
	}

	m.View = menuItem{parent: mView, subs: viewSubs}
}

func (m *menu) createChangeMenu(sess *session.Session) {
	displays := sess.Displays()
	if len(displays) < 2 {
		m.Change = menuItem{
			parent: systray.AddMenuItem("Change Wallpaper", ""),
		}
		return
	}

	var changeSubs []menuItem
	mChange := systray.AddMenuItem("Change Wallpaper…", "")

	changeSubs = append(changeSubs, menuItem{
		parent: mChange.AddSubMenuItem("All", ""),
		value:  "",
	})

	for _, d := range displays {
		// changeSubs = append(changeSubs, mChange.AddSubMenuItem(
		// 	fmt.Sprintf("%s: %s", d.ID, d.Name), ""))
		changeSubs = append(changeSubs, menuItem{
			parent: mChange.AddSubMenuItem(fmt.Sprintf("%s: %s", d.ID, d.Name), ""),
			value:  d.ID,
		})
	}

	m.Change = menuItem{parent: mChange, subs: changeSubs}
}

func (m *menu) createBlacklistMenu(sess *session.Session) {
	displays := sess.Displays()
	if len(displays) < 2 {
		m.Blacklist = menuItem{
			parent: systray.AddMenuItem("Blacklist", ""),
		}
		return
	}

	var blacklistSubs []menuItem
	mBlacklist := systray.AddMenuItem("Blacklist…", "")
	for _, d := range displays {
		// blacklistSubs = append(blacklistSubs, mBlacklist.AddSubMenuItem(
		// 	fmt.Sprintf("%s: %s", d.ID, d.Name), ""))
		blacklistSubs = append(blacklistSubs, menuItem{
			parent: mBlacklist.AddSubMenuItem(fmt.Sprintf("%s: %s", d.ID, d.Name), ""),
			value:  d.ID,
		})
	}

	m.Blacklist = menuItem{parent: mBlacklist, subs: blacklistSubs}
}

func (m *menu) createUseListMenu(sess *session.Session) {
	mUseList := systray.AddMenuItem("Use List…", "")
	mUseList.AddSubMenuItem("Nature", "")
	mUseList.AddSubMenuItem("Favorites", "")
	mUseList.AddSubMenuItem("Mountains", "")

	m.UseList = menuItem{parent: mUseList}
}
