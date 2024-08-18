package tray

import (
	"fmt"
	"sync"

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
	Intervals menuItem
	UseList   menuItem
	AddToList menuItem
	Quit      *systray.MenuItem
}

type menuItem struct {
	parent *systray.MenuItem
	subs   []*systray.MenuItem
}

// OnReady is the entry point when the systray is ready.
func OnReady() {
	cfg, err := config.Load("")
	if err != nil {
		log.Fatal(err)
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		log.Fatal(err)
	}

	initializeSystray()
	setupMenuItems(sess)
}

// initializeSystray sets up the systray icon, title, and tooltip.
func initializeSystray() {
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTitle("Walsh")
	systray.SetTooltip("Walsh")
}

// setupMenuItems sets up the systray menu items and handles click events.
func setupMenuItems(sess *session.Session) {
	m := menu{}
	// Synchronized access to adding submenus
	var mu sync.Mutex

	// Lock while adding menu items to prevent race conditions
	mu.Lock()
	m.createChangeMenu(sess)
	m.createViewMenu(sess)
	m.createBlacklistMenu(sess)
	m.createIntervalMenu(sess)
	m.createUseListMenu(sess)
	mu.Unlock()

	systray.AddSeparator()
	m.Quit = systray.AddMenuItem("Quit Walsh", "Quit Walsh")

	// Handle menu item click events in separate goroutines
	go handleMenuEvents(sess, m)
	go handleDisplayEvents(sess, m)
}

func (m *menu) createIntervalMenu(sess *session.Session) {
	mRotateInterval := systray.AddMenuItem("Rotate Interval…", "Set the interval to rotate wallpapers")

	menuHasInterval := false
	set := []config.RotateInterval{}

	// Prepend a '0' to the interval list
	menuIntervals := sess.Config().MenuIntervals
	menuIntervals = append([]config.RotateInterval{0}, menuIntervals...)
	for _, interval := range menuIntervals {
		if interval.InList(set) {
			continue
		}
		set = append(set, interval)

		check := " "
		// Check the current interval
		if int(interval) == sess.Config().Interval {
			menuHasInterval = true
			check = "✔"
		}

		mRotateInterval.AddSubMenuItemCheckbox(fmt.Sprintf("%s %s", check, interval.String()),
			fmt.Sprintf("Rotate wallpapers every %s", interval.String()), menuHasInterval)

	}

	if !menuHasInterval {
		mRotateInterval.AddSubMenuItemCheckbox(fmt.Sprintf("✔ %d", sess.Config().Interval),
			fmt.Sprintf("Rotate wallpapers every %d", sess.Config().Interval), true)
	}

	m.Intervals = menuItem{parent: mRotateInterval}
}

func (m *menu) createViewMenu(sess *session.Session) {
	var viewSubs []*systray.MenuItem
	mView := systray.AddMenuItem("View Wallpaper…", "View the current wallpaper on all displays")
	for i, d := range sess.Displays() {
		viewSubs = append(viewSubs, mView.AddSubMenuItem(
			fmt.Sprintf("%d: %s", i, d.Name),
			fmt.Sprintf("View the current wallpaper on %d: %s", i, d.Name),
		))
	}

	m.View = menuItem{parent: mView, subs: viewSubs}
}

func (m *menu) createChangeMenu(sess *session.Session) {
	var changeSubs []*systray.MenuItem
	mChange := systray.AddMenuItem("Change Wallpaper…", "Change the wallpaper on all displays")
	mChange.AddSubMenuItem("All", "Change the wallpaper on all displays")
	for i, d := range sess.Displays() {
		changeSubs = append(changeSubs, mChange.AddSubMenuItem(
			fmt.Sprintf("%d: %s", i, d.Name),
			fmt.Sprintf("Change the wallpaper on %d: %s", i, d.Name),
		))
	}

	m.Change = menuItem{parent: mChange, subs: changeSubs}
}

func (m *menu) createBlacklistMenu(sess *session.Session) {
	var blacklistSubs []*systray.MenuItem
	mBlacklist := systray.AddMenuItem("Blacklist…", "Blacklist the current wallpaper on all displays")
	for i, d := range sess.Displays() {
		blacklistSubs = append(blacklistSubs, mBlacklist.AddSubMenuItem(
			fmt.Sprintf("%d: %s", i, d.Name),
			fmt.Sprintf("Blacklist the current wallpaper on %d: %s", i, d.Name),
		))
	}

	m.Blacklist = menuItem{parent: mBlacklist, subs: blacklistSubs}
}

func (m *menu) createUseListMenu(sess *session.Session) {
	mUseList := systray.AddMenuItem("Use List…", "Use a list of wallpapers")
	mUseList.AddSubMenuItem("Nature", "Use wallpapers from the nature list")
	mUseList.AddSubMenuItem("Favorites", "Use wallpapers from the favorites list")
	mUseList.AddSubMenuItem("Mountains", "Use wallpapers from the mountains list")

	m.UseList = menuItem{parent: mUseList}
}

// handleDisplayEvents handles click events for the display-specific submenu items.
func handleDisplayEvents(sess *session.Session, m menu) {
	for i, d := range sess.Displays() {
		go func(i int, d session.Display) {
			disp := fmt.Sprintf("%d", i)
			for {
				select {
				case <-m.Change.subs[i].ClickedCh:
					log.Infof("Changing wallpaper on %d: %s", d.Index, d.Name)
					sess.SetWallpaper(sess.Config().Sources, d.Name)
				case <-m.View.subs[i].ClickedCh:
					log.Infof("Viewing wallpaper on %d: %s", d.Index, d.Name)
					current, err := sess.GetCurrentWallpaper(disp)
					if err != nil {
						log.Fatal(err)
					}

					if err = sess.View(current); err != nil {
						log.Fatal(err)
					}
				case <-m.Blacklist.subs[i].ClickedCh:
					log.Infof("Blacklisting wallpaper on %d: %s", d.Index, d.Name)

					err := beeep.Notify("Walsh", fmt.Sprintf("Blacklisting wallpaper on %d: %s", d.Index, d.Name), "icon/wicon.png")
					if err != nil {
						log.Fatal(err)
					}
					// Handle blacklisting wallpaper for the specific display
					if err := blacklist.Blacklist(disp, sess); err != nil {
						log.Fatal(err)
					}
					// case <-m.AddToList.subs[i].ClickedCh:
					// 	log.Infof("Adding wallpaper to list on %d: %s", d.Index, d.Name)
				}
			}
		}(i, d)
	}
}

// handleMenuEvents handles click events for the main menu items.
func handleMenuEvents(sess *session.Session, m menu) {
	go func() {
		for {
			select {
			case <-m.Change.parent.ClickedCh:
				log.Info("Changing wallpaper")
				// sess.SetWallpaper(cfg.Sources, "")
			case <-m.View.parent.ClickedCh:
				log.Info("Viewing wallpaper")
				// Handle viewing wallpaper
			case <-m.Blacklist.parent.ClickedCh:
				log.Info("Blacklisting wallpaper")
				// Handle blacklisting wallpaper
			// case <-m.AddToList.parent.ClickedCh:
			// 	log.Info("Adding wallpaper to list")
			// Handle adding wallpaper to list
			// case <-c.All.parent.ClickedCh:
			// 	log.Info("Changing wallpaper on all displays")
			// 	sess.SetWallpaper(sess.Config().Sources, "")
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
