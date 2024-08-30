// Package scheduler provides setting the wallpaper on a schedule and handling
// session events. When Walsh is ran persistently, the scheduler is evoked to
// run continuously. Optionally, a system tray/menu bar is also displayed.
package scheduler

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

type process struct {
	session *session.Session
	display string
	ctx     context.Context
	cancel  context.CancelFunc
	sigChan chan os.Signal
	ticker  *time.Ticker
}

func New(sess *session.Session, display string) *process {
	ctx, cancel := context.WithCancel(context.Background())
	return &process{
		session: sess,
		display: display,
		ctx:     ctx,
		cancel:  cancel,
		sigChan: make(chan os.Signal, 1),
	}
}

func (p *process) Start() error {
	log.Info("Starting scheduler...")

	go p.runWallpaperUpdater()

	if p.session.Config().SystemTray.Enabled {
		log.Info("Starting systray...")
		systray.Run(tray.Run, p.quit)
	}

	// Handle OS signals for clean shutdown
	signal.Notify(p.sigChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-p.ctx.Done():
	case <-p.sigChan:
		p.quit()
	}

	return nil
}

func (p *process) runWallpaperUpdater() {
	if err := p.setWallpaperWithContext(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func (p *process) quit() {
	log.Info("Shutting down...")
	if p.cancel != nil {
		log.Debug("Cancelling context")
		p.cancel()
	}
	log.Debugf("Exiting...")
}

func (p *process) setWallpaperWithContext() error {
	for {
		select {
		case <-p.ctx.Done():
			log.Info("Stopping wallpaper updates")
			return nil
		default:
			if err := p.setWallpaper(); err != nil {
				return err
			}
		}
	}
}

func (p *process) setWallpaper() error {
	p.ticker = time.NewTicker(time.Duration(p.session.Config().Interval) * time.Second)
	defer p.ticker.Stop()

	for {
		select {
		case <-p.ticker.C:
			err := util.Retry(p.session.Config().MaxRetries, p.session.Config().RetryInterval, func() error {
				p.ticker.Reset(time.Duration(p.session.Interval()) * time.Second)
				log.Debugf("Ticker interval set to %d seconds", p.session.Interval())

				return p.session.SetWallpaper(p.display)
			})
			if err != nil {
				log.Fatal(err)
				return err
			}
			log.Infof("Next wallpaper change in %d seconds", p.session.Config().Interval)
		case <-p.ctx.Done():
			return nil
		}
	}
}
