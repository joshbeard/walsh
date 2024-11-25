package tray

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/systray"
	"github.com/charmbracelet/log"
	"github.com/gen2brain/beeep"
	"github.com/joshbeard/walsh/cmd/blacklist"
	"github.com/joshbeard/walsh/internal/session"
)

var handlersStarted bool
var handlerLock sync.Mutex
var idleTimeout = 1 * time.Minute // Adjust the timeout duration as needed

func (m *TrayMenu) startHandlers() {
	handlerLock.Lock()
	defer handlerLock.Unlock()

	if !handlersStarted {
		handlersStarted = true

		go func() {
			defer func() { handlersStarted = false }()
			log.Debug("Starting handlers")
			go m.handleDisplayEvents()
			go m.handleIntervals()
		}()
	}
}

func (m *TrayMenu) stopHandlersIfNeeded() {
	handlerLock.Lock()
	defer handlerLock.Unlock()

	if handlersStarted {
		log.Debug("Stopping handlers due to inactivity")
		m.cancel()
		handlersStarted = false
	}
}

func (m *TrayMenu) monitorIdle() {
	idleTimer := time.NewTimer(idleTimeout)

	log.Debug("Starting idle monitor")

	go func() {
		for {
			select {
			case <-systray.TrayOpenedCh:
				log.Debug("Tray opened")
				m.startHandlers()

				// Reset the idle timer on tray click
				if !idleTimer.Stop() {
					<-idleTimer.C // Drain the channel if necessary
				}
				idleTimer.Reset(idleTimeout)

			case <-idleTimer.C:
				// Idle timeout reached, stop the handlers
				m.stopHandlersIfNeeded()
			}
		}
	}()
}

func (m *TrayMenu) handleDisplayEvents() {
	displays := session.TargetDisplays()
	if len(displays) == 0 {
		return
	}

	// handle the change menu's "All Displays" item
	if len(displays) > 1 {
		go func() {
			for range m.change.subs[0].MenuItem.ClickedCh {
				go m.changeWallpaper(0, session.Display{})
			}
		}()
	}

	for i, d := range displays {
		go func(i int, d session.Display) {
			for {
				select {
				case <-m.ctx.Done():
					log.Debugf("context done for display %s", d.Name)
					return

				// Change wallpaper
				case <-m.getChangeClickedCh(i + 1):
					log.Infof("changing wallpaper on %s: %s", d.Name, m.change.getCmdArg(i+1))
					go m.changeWallpaper(i+1, d)

				// View wallpaper
				case <-m.getViewClickedCh(i):
					log.Infof("viewing wallpaper on %s: %s", d.Name, m.view.getCmdArg(i))
					go m.viewWallpaper(i, d)

				// Blacklist wallpaper
				case <-m.getBlacklistClickedCh(i):
					log.Infof("blacklisting wallpaper on %s: %s", d.Name, m.blacklist.getCmdArg(i))
					go m.blacklistWallpaper(i, d)

				// Quit
				case <-m.quit.ClickedCh:
					log.Debug("received quit signal")
					systray.Quit()
					return
				}
			}
		}(i, d)
	}
}

// Helper function to handle different clicked channels
func (m *TrayMenu) getChangeClickedCh(i int) <-chan struct{} {
	displays := session.TargetDisplays()

	if len(displays) == 1 {
		log.Debug("single display")
		return m.change.MenuItem.ClickedCh
	}

	return m.change.subs[i].MenuItem.ClickedCh
}

func (m *TrayMenu) getViewClickedCh(i int) <-chan struct{} {
	if len(session.TargetDisplays()) == 1 {
		return m.view.MenuItem.ClickedCh
	}
	return m.view.subs[i].MenuItem.ClickedCh
}

func (m *TrayMenu) getBlacklistClickedCh(i int) <-chan struct{} {
	if len(session.TargetDisplays()) == 1 {
		return m.blacklist.MenuItem.ClickedCh
	}
	return m.blacklist.subs[i].MenuItem.ClickedCh
}

func (m menuItem) getCmdArg(i int) string {
	if len(m.subs) > 0 {
		return m.subs[i].value
	}

	return m.value
}

func (m TrayMenu) changeWallpaper(i int, d session.Display) {
	err := session.SetWallpaper(m.change.getCmdArg(i))
	if err != nil {
		log.Fatal(err)
	}
}

func (m TrayMenu) viewWallpaper(i int, d session.Display) {
	current, err := session.GetCurrentWallpaper(m.view.getCmdArg(i))
	if err != nil {
		log.Fatalf("error getting current wallpaper: %v", err)
	}
	if err = session.View(current); err != nil {
		log.Fatal(err)
	}
}

func (m TrayMenu) blacklistWallpaper(i int, d session.Display) {
	arg := m.blacklist.getCmdArg(i)
	err := beeep.Notify("Walsh",
		fmt.Sprintf("Blacklisting wallpaper on %s: %s", arg, d.Name),
		"icon/wicon.png")
	if err != nil {
		log.Fatal(err)
	}
	if err := blacklist.Blacklist(arg); err != nil {
		log.Fatal(err)
	}
}

func (m *TrayMenu) handleIntervals() {
	for i, interval := range m.intervals.subs {
		go func(i int, interval intervalItem) {
			for range interval.item.ClickedCh {
				// Uncheck all other intervals
				for _, item := range m.intervals.subs {
					item.item.Uncheck()
					item.item.SetTitle(fmt.Sprintf("  %s", item.value))
				}

				// Set the interval and check the selected interval
				session.SetInterval(interval.interval)
				check, _ := m.getIntervalCheckMark(interval.interval)

				// Update the menu item
				interval.item.Check()
				interval.item.SetTitle(fmt.Sprintf("%s%s", check, interval.value))

				// Reset the ticker
				// session.ResetTicker()
				session.Set(session.Config().Display)
			}
		}(i, interval)
	}
}
