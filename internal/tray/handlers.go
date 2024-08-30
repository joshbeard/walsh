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
func (m menu) handleDisplayEvents(sess *session.Session) {
	if len(sess.Displays()) < 2 {
		return
	}

	for i, d := range sess.Displays() {
		go func(i int, d session.Display) {
			for {
				select {
				case <-m.change.subs[i].parent.ClickedCh:
					log.Infof("Changing wallpaper on %s: %s", m.change.subs[i].value, d.Name)
					err := sess.SetWallpaper(m.change.subs[i].value)
					if err != nil {
						log.Fatal(err)
					}
				case <-m.view.subs[i].parent.ClickedCh:
					log.Infof("Viewing wallpaper on %s: %s", m.view.subs[i].value, d.Name)
					current, err := sess.GetCurrentWallpaper(m.view.subs[i].value)
					if err != nil {
						log.Fatalf("Error getting current wallpaper: %v", err)
					}

					if err = sess.View(current); err != nil {
						log.Fatal(err)
					}
				case <-m.blacklist.subs[i].parent.ClickedCh:
					log.Infof("Blacklisting wallpaper on %s: %s", m.blacklist.subs[i].value, d.Name)

					err := beeep.Notify("Walsh", fmt.Sprintf("Blacklisting wallpaper on %s: %s", m.blacklist.subs[i].value, d.Name), "icon/wicon.png")
					if err != nil {
						log.Fatal(err)
					}
					// Handle blacklisting wallpaper for the specific display
					if err := blacklist.Blacklist(m.blacklist.subs[i].value, sess); err != nil {
						log.Fatal(err)
					}

					// TODO: Add to list
					// case <-m.AddToList.subs[i].ClickedCh:
					// 	log.Infof("Adding wallpaper to list on %d: %s", d.Index, d.Name)

					// TODO: Set from list
					// case <-m.SetFromList.subs[i].ClickedCh:
					//	log.Infof("Setting wallpaper from list on %d: %s", d.Index, d.Name)
				}
			}
		}(i, d)
	}
}

func (m menu) handleIntervalEvents(sess *session.Session) {
	intervals := m.intervals.subs
	for i, interval := range intervals {
		go func(i int, interval intervalItem) {
			for range m.intervals.subs[i].item.ClickedCh {
				log.Infof("Changing interval to %d", m.intervals.subs[i].interval)
				sess.SetInterval(interval.interval)
			}
		}(i, interval)
	}

}

// handleMenuEvents handles click events for the main menu items.
func (m menu) handleMenuEvents(sess *session.Session) {
	go func() {
		for {
			select {
			// -- Change wallpaper --
			case <-m.change.parent.ClickedCh:
				if len(m.change.subs) > 0 {
					return
				}

				log.Info("Changing wallpaper")
				err := sess.SetWallpaper(m.change.value)
				if err != nil {
					log.Error(err)
				}

			// -- View wallpaper --
			case <-m.view.parent.ClickedCh:
				if len(m.view.subs) > 0 {
					return
				}

				log.Info("Viewing wallpaper")
				current, err := sess.GetCurrentWallpaper(m.view.value)
				if err != nil {
					log.Error(err)
				}
				if err = sess.View(current); err != nil {
					log.Error(err)
				}

			// -- Blacklist wallpaper --
			case <-m.blacklist.parent.ClickedCh:
				if len(m.blacklist.subs) > 0 {
					return
				}

				log.Info("Blacklisting wallpaper")
				err := beeep.Notify("Walsh", "Blacklisting wallpaper", "icon/wicon.png")
				if err != nil {
					log.Error(err)
				}

				if err := blacklist.Blacklist(m.view.value, sess); err != nil {
					log.Error(err)
				}

			// -- TODO: Add To List --
			// case <-m.addToList.parent.ClickedCh:
			// 	log.Info("Adding wallpaper to list")

			// -- TODO: Set From List --
			// case <-m.setFromList.parent.ClickedCh:
			//	log.Info("Setting wallpaper from list")

			// -- Quit  --
			case <-m.quit.ClickedCh:
				log.Debug("received quit signal")
				systray.Quit()
				return
			}
		}
	}()
}
