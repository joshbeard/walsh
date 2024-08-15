package tray

import (
	"fmt"

	"fyne.io/systray"
	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/config"
	"github.com/joshbeard/walsh/internal/session"
	"github.com/joshbeard/walsh/internal/tray/icon"
)

// onReady is the entry point when the systray is ready.
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
	setupMenuItems(sess, cfg)
}

// initializeSystray sets up the systray icon, title, and tooltip.
func initializeSystray() {
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTitle("Walsh")
	systray.SetTooltip("Walsh")
}

// setupMenuItems sets up the systray menu items and handles click events.
func setupMenuItems(sess *session.Session, cfg *config.Config) {
	mChange := systray.AddMenuItem("Change Wallpaper", "Change the wallpaper on all displays")
	mView := systray.AddMenuItem("View Wallpaper", "View the current wallpaper on all displays")
	mBlacklist := systray.AddMenuItem("Blacklist", "Blacklist the current wallpaper")

	// TODO: Implement Add to list
	mAddToList := systray.AddMenuItem("Add to List", "Add the current wallpaper to a list")
	mAddToList.Hide()

	cAll := mChange.AddSubMenuItem("All", "Change the wallpaper on all displays")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Walsh", "Quit Walsh")

	// Create a submenu item for each display
	displays := sess.Displays()
	viewSubs, changeSubs, blacklistSubs, addToListSubs := createSubMenus(mView, mChange, mBlacklist, mAddToList, displays)

	// Handle menu item click events in separate goroutines
	go handleMenuEvents(sess, cfg, mChange, mView, mBlacklist, mAddToList, cAll, mQuit)
	go handleDisplayEvents(sess, cfg, displays, viewSubs, changeSubs, blacklistSubs, addToListSubs)
}

// createSubMenus creates submenu items for each display.
func createSubMenus(mView, mChange, mBlacklist, mAddToList *systray.MenuItem, displays []session.Display) ([]*systray.MenuItem, []*systray.MenuItem, []*systray.MenuItem, []*systray.MenuItem) {
	var viewSubs, changeSubs, blacklistSubs, addToListSubs []*systray.MenuItem
	for _, d := range displays {
		viewSubs = append(viewSubs, mView.AddSubMenuItem(
			fmt.Sprintf("%d: %s", d.Index, d.Name),
			fmt.Sprintf("View the current wallpaper on %d: %s", d.Index, d.Name),
		))

		changeSubs = append(changeSubs, mChange.AddSubMenuItem(
			fmt.Sprintf("%d: %s", d.Index, d.Name),
			fmt.Sprintf("Change the wallpaper on %d: %s", d.Index, d.Name),
		))

		blacklistSubs = append(blacklistSubs, mBlacklist.AddSubMenuItem(
			fmt.Sprintf("%d: %s", d.Index, d.Name),
			fmt.Sprintf("Blacklist the current wallpaper on %d: %s", d.Index, d.Name),
		))

		// TODO: Add to list
		addToListSubs = append(addToListSubs, mAddToList.AddSubMenuItem(
			fmt.Sprintf("%d: %s", d.Index, d.Name),
			fmt.Sprintf("Add the current wallpaper to a list on %d: %s", d.Index, d.Name),
		))
	}
	return viewSubs, changeSubs, blacklistSubs, addToListSubs
}

// handleDisplayEvents handles click events for the display-specific submenu items.
func handleDisplayEvents(sess *session.Session, cfg *config.Config, displays []session.Display, viewSubs, changeSubs, blacklistSubs, addToListSubs []*systray.MenuItem) {
	for i, d := range displays {
		go func(i int, d session.Display) {
			for {
				select {
				case <-changeSubs[i].ClickedCh:
					log.Infof("Changing wallpaper on %d: %s", d.Index, d.Name)
					sess.SetWallpaper(cfg.Sources, d.Name)
				case <-viewSubs[i].ClickedCh:
					log.Infof("Viewing wallpaper on %d: %s", d.Index, d.Name)
					// Handle viewing wallpaper for the specific display
				case <-blacklistSubs[i].ClickedCh:
					log.Infof("Blacklisting wallpaper on %d: %s", d.Index, d.Name)
					// Handle blacklisting wallpaper for the specific display
				case <-addToListSubs[i].ClickedCh:
					log.Infof("Adding wallpaper to list on %d: %s", d.Index, d.Name)
				}
			}
		}(i, d)
	}
}

// handleMenuEvents handles click events for the main menu items.
func handleMenuEvents(sess *session.Session, cfg *config.Config, mChange, mView, mBlacklist, mAddToList, cAll, mQuit *systray.MenuItem) {
	go func() {
		for {
			select {
			case <-mChange.ClickedCh:
				log.Info("Changing wallpaper")
				// sess.SetWallpaper(cfg.Sources, "")
			case <-mView.ClickedCh:
				log.Info("Viewing wallpaper")
				// Handle viewing wallpaper
			case <-mBlacklist.ClickedCh:
				log.Info("Blacklisting wallpaper")
				// Handle blacklisting wallpaper
			case <-mAddToList.ClickedCh:
				log.Info("Adding wallpaper to list")
				// Handle adding wallpaper to list
			case <-cAll.ClickedCh:
				log.Info("Changing wallpaper on all displays")
				sess.SetWallpaper(cfg.Sources, "")
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func OnExit() {
	log.Info("Closing Walsh tray")
}
