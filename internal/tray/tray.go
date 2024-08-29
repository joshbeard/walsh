package tray

import (
	"fmt"

	"fyne.io/systray"
	"github.com/charmbracelet/log"
	"github.com/gen2brain/beeep"
	"github.com/joshbeard/walsh/cmd/blacklist"
	"github.com/joshbeard/walsh/internal/config"
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
	subs   []*systray.MenuItem
}

// OnReady is the entry point when the systray is ready.
func OnReady() {
	sess := session.Current
	m := &menu{}

	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTooltip("Walsh")

	// Don't set a title for now
	// systray.SetTitle("Walsh")

	setupMenuItems(m, sess)

	// Handle menu item click events in separate goroutines
	go handleMenuEvents(sess, *m)
	go handleDisplayEvents(sess, *m)
	go handleIntervalEvents(sess, *m)
}

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
		check := " "
		if interval.InList(set) {
			continue
		}
		set = append(set, interval)

		// Check the current interval
		if int(interval) == sess.Interval() {
			menuHasInterval = true
			check = "✔"
		}

		s := mRotateInterval.AddSubMenuItemCheckbox(fmt.Sprintf("%s %s", check, interval.String()), "", false)

		m.Intervals.subs = append(m.Intervals.subs, intervalItem{interval: int(interval), item: s})
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
		}
		return
	}

	var viewSubs []*systray.MenuItem
	mView := systray.AddMenuItem("View Wallpaper…", "")
	for _, d := range displays {
		viewSubs = append(viewSubs, mView.AddSubMenuItem(
			fmt.Sprintf("%s: %s", d.ID, d.Name), ""))
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

	var changeSubs []*systray.MenuItem
	mChange := systray.AddMenuItem("Change Wallpaper…", "")
	mChange.AddSubMenuItem("All", "")
	for _, d := range displays {
		changeSubs = append(changeSubs, mChange.AddSubMenuItem(
			fmt.Sprintf("%s: %s", d.ID, d.Name), ""))
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

	var blacklistSubs []*systray.MenuItem
	mBlacklist := systray.AddMenuItem("Blacklist…", "")
	for _, d := range displays {
		blacklistSubs = append(blacklistSubs, mBlacklist.AddSubMenuItem(
			fmt.Sprintf("%s: %s", d.ID, d.Name), ""))
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

// handleDisplayEvents handles click events for the display-specific submenu items.
func handleDisplayEvents(sess *session.Session, m menu) {
	if len(sess.Displays()) < 2 {
		return
	}

	for i, d := range sess.Displays() {
		go func(i int, d session.Display) {
			for {
				select {
				case <-m.Change.subs[i].ClickedCh:
					log.Infof("Changing wallpaper on %s: %s", d.ID, d.Name)
					err := sess.SetWallpaper(sess.Config().Sources, d.ID)
					if err != nil {
						log.Fatal(err)
					}
				case <-m.View.subs[i].ClickedCh:
					log.Infof("Viewing wallpaper on %s: %s", d.ID, d.Name)
					current, err := sess.GetCurrentWallpaper(d.ID)
					if err != nil {
						log.Fatalf("Error getting current wallpaper: %v", err)
					}

					if err = sess.View(current); err != nil {
						log.Fatal(err)
					}
				case <-m.Blacklist.subs[i].ClickedCh:
					log.Infof("Blacklisting wallpaper on %s: %s", d.ID, d.Name)

					err := beeep.Notify("Walsh", fmt.Sprintf("Blacklisting wallpaper on %s: %s", d.ID, d.Name), "icon/wicon.png")
					if err != nil {
						log.Fatal(err)
					}
					// Handle blacklisting wallpaper for the specific display
					if err := blacklist.Blacklist(d.ID, sess); err != nil {
						log.Fatal(err)
					}
					// case <-m.AddToList.subs[i].ClickedCh:
					// 	log.Infof("Adding wallpaper to list on %d: %s", d.Index, d.Name)
				}
			}
		}(i, d)
	}
}

func handleIntervalEvents(sess *session.Session, m menu) {
	log.Warnf("Intervals: %v", m.Intervals)
	intervals := m.Intervals.subs
	log.Debug("Handling interval events")
	for i, interval := range intervals {
		go func(i int, interval intervalItem) {
			for {
				select {
				case <-m.Intervals.subs[i].item.ClickedCh:
					// log.Infof("Changing interval to %s", sess.Config().MenuIntervals[i].String())
					log.Infof("Changing interval to %d", m.Intervals.subs[i].interval)
					// Set the interval to the selected interval
					sess.SetInterval(interval.interval)
				}
			}
		}(i, interval)
	}

}

// handleMenuEvents handles click events for the main menu items.
func handleMenuEvents(sess *session.Session, m menu) {
	go func() {
		for {
			select {
			case <-m.Change.parent.ClickedCh:
				log.Info("Changing wallpaper")
				sess.SetWallpaper(sess.Config().Sources, "0")

			case <-m.View.parent.ClickedCh:
				log.Info("Viewing wallpaper")
				current, err := sess.GetCurrentWallpaper("0")
				if err != nil {
					log.Error(err)
				}
				if err = sess.View(current); err != nil {
					log.Error(err)
				}
			case <-m.Blacklist.parent.ClickedCh:
				log.Info("Blacklisting wallpaper")
				err := beeep.Notify("Walsh", "Blacklisting wallpaper", "icon/wicon.png")
				if err != nil {
					log.Error(err)
				}

				if err := blacklist.Blacklist("0", sess); err != nil {
					log.Error(err)
				}

			// TODO:
			// case <-m.AddToList.parent.ClickedCh:
			// 	log.Info("Adding wallpaper to list")
			// Handle adding wallpaper to list
			case <-m.Quit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func OnExit() {
	log.Info("Closing Walsh tray")
}
