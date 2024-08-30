package daemon

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fyne.io/systray"
	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/session"
	"github.com/joshbeard/walsh/internal/tray"
	"github.com/joshbeard/walsh/internal/util"
)

type Daemon struct {
	session *session.Session
	display string
	ctx     context.Context
	cancel  context.CancelFunc
	sigChan chan os.Signal
	ticker  *time.Ticker
}

func NewDaemon(sess *session.Session, display string) *Daemon {
	ctx, cancel := context.WithCancel(context.Background())
	return &Daemon{
		session: sess,
		display: display,
		ctx:     ctx,
		cancel:  cancel,
		sigChan: make(chan os.Signal, 1),
	}
}

func (d *Daemon) Start() error {
	log.Info("Starting daemon...")

	go d.runWallpaperUpdater()

	if d.session.Config().ShowTray {
		log.Info("Starting systray...")
		systray.Run(tray.OnReady, d.quit)
	}

	// Handle OS signals for clean shutdown
	signal.Notify(d.sigChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-d.ctx.Done():
	case <-d.sigChan:
		d.quit()
	}

	return nil
}

func (d *Daemon) runWallpaperUpdater() {
	if err := d.setWallpaperWithContext(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func (d *Daemon) quit() {
	log.Info("Shutting down...")
	if d.cancel != nil {
		log.Debug("Cancelling context")
		d.cancel()
	}
	log.Debugf("Exiting...")
}

func (d *Daemon) setWallpaperWithContext() error {
	for {
		select {
		case <-d.ctx.Done():
			log.Info("Stopping wallpaper updates")
			return nil
		default:
			if err := d.setWallpaper(); err != nil {
				return err
			}
		}
	}
}

func (d *Daemon) setWallpaper() error {
	d.ticker = time.NewTicker(time.Duration(d.session.Config().Interval) * time.Second)
	defer d.ticker.Stop()

	for {
		select {
		case <-d.ticker.C:
			err := util.Retry(d.session.Config().MaxRetries, d.session.Config().RetryInterval, func() error {
				d.ticker.Reset(time.Duration(d.session.Interval()) * time.Second)
				log.Debugf("Ticker interval set to %d seconds", d.session.Interval())

				return d.session.SetWallpaper(d.display)
			})
			if err != nil {
				log.Fatal(err)
				return err
			}
			log.Infof("Next wallpaper change in %d seconds", d.session.Config().Interval)
		case <-d.ctx.Done():
			return nil
		}
	}
}
