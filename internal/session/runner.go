package session

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fyne.io/systray"
	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/logger"
	"github.com/joshbeard/walsh/internal/util"
)

func Start(trayFunc func()) error {
	go func() {
		if err := Set(This.cfg.Display); err != nil {
			log.Fatal(err)
		}
	}()

	log.Debug("starting session", "interval", Interval())
	This.Ctx, This.Cancel = context.WithCancel(context.Background())
	This.SigChan = make(chan os.Signal, 1)

	if Config().SystemTrayEnabled {
		log.Debug("starting systray")
		systray.Run(trayFunc, Quit)
	}

	// Handle OS signals for clean shutdown
	signal.Notify(This.SigChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-This.Ctx.Done():
	case <-This.SigChan:
		Quit()
	}

	return nil
}

func Set(display string) error {
	// Set the wallpaper with retries
	log.Debug("setting wallpaper...")
	err := util.Retry(Config().MaxRetries, Config().RetryInterval,
		func() error {
			return SetWallpaper(display)
		})
	if err != nil {
		log.Fatal(err)
	}

	// ResetTicker is called here to allow changes to the ticker interval
	ResetTicker()

	for {
		// Check if the ticker is nil or uninitialized, and wait until it’s properly set
		if This.Ticker == nil || This.Ticker.C == nil {
			// log.Debug("refresh ticker is not set, waiting...")
			log.Log(logger.TraceLevel, "waiting for ticker to be set")

			// Use a time.Sleep or context-aware wait here to avoid a busy loop
			time.Sleep(1 * time.Second) // or adapt this to your needs
			continue
		}

		select {
		case <-This.Ctx.Done():
			log.Info("stopping wallpaper updates")
			return nil
		case <-This.Ticker.C:
			cfg := Config()

			// Retry setting the wallpaper, with a refreshed session if necessary
			err := util.Retry(cfg.MaxRetries, cfg.RetryInterval, func() error {
				log.Info("refreshing session...")
				if err := Refresh(); err != nil {
					log.Error(err)
					return err
				}
				return SetWallpaper(display)
			})

			if err != nil {
				log.Fatal(err)
				return err
			}

			log.Debugf("next wallpaper change in %s at approx %s",
				Interval(), NextTick())
		}
	}
}

func Quit() {
	log.Info("shutting down...")
	if This.Cancel != nil {
		log.Debug("cancelling context")
		This.Cancel()
	}
	log.Debugf("exiting...")
}
