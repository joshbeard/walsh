package tray

import (
	"fmt"

	"fyne.io/systray"
	"github.com/charmbracelet/log"
	"github.com/gen2brain/beeep"
	"github.com/joshbeard/walsh/cmd/blacklist"
	"github.com/joshbeard/walsh/internal/session"
)

// handleDisplayEvents handles click events for the display-specific submenu items.
func handleDisplayEvents(sess *session.Session, m menu) {
	if len(sess.Displays()) < 2 {
		return
	}

	for i, d := range sess.Displays() {
		go func(i int, d session.Display) {
			for {
				select {
				case <-m.Change.subs[i].parent.ClickedCh:
					log.Infof("Changing wallpaper on %s: %s", m.Change.subs[i].value, d.Name)
					err := sess.SetWallpaper(m.Change.subs[i].value)
					if err != nil {
						log.Fatal(err)
					}
				case <-m.View.subs[i].parent.ClickedCh:
					log.Infof("Viewing wallpaper on %s: %s", m.View.subs[i].value, d.Name)
					current, err := sess.GetCurrentWallpaper(m.View.subs[i].value)
					if err != nil {
						log.Fatalf("Error getting current wallpaper: %v", err)
					}

					if err = sess.View(current); err != nil {
						log.Fatal(err)
					}
				case <-m.Blacklist.subs[i].parent.ClickedCh:
					log.Infof("Blacklisting wallpaper on %s: %s", m.Blacklist.subs[i].value, d.Name)

					err := beeep.Notify("Walsh", fmt.Sprintf("Blacklisting wallpaper on %s: %s", m.Blacklist.subs[i].value, d.Name), "icon/wicon.png")
					if err != nil {
						log.Fatal(err)
					}
					// Handle blacklisting wallpaper for the specific display
					if err := blacklist.Blacklist(m.Blacklist.subs[i].value, sess); err != nil {
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
				if len(m.Change.subs) > 0 {
					return
				}

				log.Info("Changing wallpaper")
				err := sess.SetWallpaper(m.Change.value)
				if err != nil {
					log.Error(err)
				}

			case <-m.View.parent.ClickedCh:
				if len(m.View.subs) > 0 {
					return
				}

				log.Info("Viewing wallpaper")
				current, err := sess.GetCurrentWallpaper(m.View.value)
				if err != nil {
					log.Error(err)
				}
				if err = sess.View(current); err != nil {
					log.Error(err)
				}
			case <-m.Blacklist.parent.ClickedCh:
				if len(m.Blacklist.subs) > 0 {
					return
				}

				log.Info("Blacklisting wallpaper")
				err := beeep.Notify("Walsh", "Blacklisting wallpaper", "icon/wicon.png")
				if err != nil {
					log.Error(err)
				}

				if err := blacklist.Blacklist(m.View.value, sess); err != nil {
					log.Error(err)
				}

			// TODO:
			// case <-m.AddToList.parent.ClickedCh:
			// 	log.Info("Adding wallpaper to list")
			// Handle adding wallpaper to list
			case <-m.Quit.ClickedCh:
				log.Debug("received quit signal")
				systray.Quit()
				return
			}
		}
	}()
}
