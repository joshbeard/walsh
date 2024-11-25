package tray

import (
	"context"
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"time"

	"fyne.io/systray"
	"github.com/charmbracelet/log"
	"github.com/joshbeard/walsh/internal/session"
)

//go:embed icon/icon-dark.png
var icon []byte

var Menu *TrayMenu

type TrayMenu struct {
	quit      *systray.MenuItem
	intervals intervalMenu
	view      menuItem
	change    menuItem
	blacklist menuItem
	useList   menuItem
	addToList menuItem
	displays  []session.Display

	refreshTicker *time.Ticker
	cancel        context.CancelFunc
	ctx           context.Context
}

type intervalMenu struct {
	parent *systray.MenuItem
	subs   []intervalItem
}

type intervalItem struct {
	interval time.Duration
	item     *systray.MenuItem
	value    string
}

type menuItem struct {
	*systray.MenuItem
	value string
	subs  []menuItem
}

type displayMenu struct {
	showAllDisplays bool
	label           string
	displays        []session.Display
}

// OnReady starts the systray menu.
func OnReady() {
	systray.SetTemplateIcon(icon, icon)
	systray.SetTooltip("Walsh")

	Menu = &TrayMenu{
		displays:      session.TargetDisplays(),
		refreshTicker: time.NewTicker(30 * time.Second),
	}

	go func() {
		for range Menu.refreshTicker.C {
			Menu.Refresh()
		}
	}()

	Menu.setup()
	log.Info("systray ready")
}

// setup initializes the systray menu items.
func (m *TrayMenu) setup() {
	m.createChangeMenu()
	m.createViewMenu()
	m.createBlacklistMenu()
	m.createIntervalMenu()
	m.createUseListMenu()
	m.createAddToListMenu()

	systray.AddSeparator()
	m.quit = systray.AddMenuItem("Quit Walsh", "")

	m.ctx, m.cancel = context.WithCancel(context.Background())

	// go m.handleDisplayEvents()
	// m.startHandlers()
	go m.handleIntervals()
	m.monitorIdle()
}

func (m *TrayMenu) Refresh() {
	session.Refresh()
	changed := len(Menu.displays) != len(session.TargetDisplays())

	if changed {
		log.Info("number of displays has changed",
			"before", len(Menu.displays),
			"after", len(session.TargetDisplays()),
		)
	} else {
		for i, d := range Menu.displays {
			if d.Index != session.TargetDisplays()[i].Index {
				changed = true
				log.Info("display configuration changed",
					"index", d.Index,
					"new", session.TargetDisplays()[i].Index,
				)
				break
			}
		}
	}

	if !changed {
		return
	}

	log.Info("updating menu items")
	m.cancel()

	// Update the displays
	Menu.displays = session.TargetDisplays()
	systray.ResetMenu()
	Menu.setup()
}

func (m *TrayMenu) createChangeMenu() {
	m.change = m.createDisplayMenuItem(displayMenu{
		showAllDisplays: true,
		label:           "Change Wallpaper",
	})
}

func (m *TrayMenu) createViewMenu() {
	m.view = m.createDisplayMenuItem(displayMenu{
		showAllDisplays: false,
		label:           "View Wallpaper",
	})
}

func (m *TrayMenu) createBlacklistMenu() {
	m.blacklist = m.createDisplayMenuItem(displayMenu{
		showAllDisplays: false,
		label:           "Blacklist",
	})
}

func (m *TrayMenu) createDisplayMenuItem(d displayMenu) menuItem {
	var parent menuItem
	displays := session.TargetDisplays()

	// If there is only one display, no submenu is needed
	if len(displays) < 2 {
		parent.MenuItem = systray.AddMenuItem(d.label, "")
		parent.value = displays[0].ID

		return parent
	}

	// Create the parent menu item for "Change"
	parent.MenuItem = systray.AddMenuItem(d.label+"…", "")

	// Optionally add an "All Displays" submenu item
	if d.showAllDisplays {
		allDisplaysItem := menuItem{
			MenuItem: parent.MenuItem.AddSubMenuItem("All Displays", ""),
			value:    "", // Special value for "All Displays"
		}
		parent.subs = append(parent.subs, allDisplaysItem)
	}

	// Add individual display-specific submenu items
	for _, display := range displays {
		label := display.Label
		if label == "" {
			label = display.Name
		}
		displayItem := menuItem{
			MenuItem: parent.MenuItem.AddSubMenuItem(fmt.Sprintf("%s: %s", display.ID, label), ""),
			value:    display.ID,
		}
		parent.subs = append(parent.subs, displayItem)
	}

	return parent
}

func (m *TrayMenu) createUseListMenu() {
	useListItems, err := session.ListLists()
	if len(useListItems) == 0 {
		return
	}

	mUseList := systray.AddMenuItem("Use List…", "")
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range useListItems {
		mUseList.AddSubMenuItem(item, "")
	}
	m.useList = menuItem{MenuItem: mUseList}
}

func (m *TrayMenu) createAddToListMenu() {
	listMenu := systray.AddMenuItem("Add to List…", "")

	lists, err := session.ListLists()
	if err != nil {
		log.Fatal(err)
	}

	displays := session.TargetDisplays()
	subs := make([]menuItem, 0, len(displays)*len(lists))
	for _, d := range displays {
		displayMenu := listMenu.AddSubMenuItem(fmt.Sprintf("  %s: %s", d.ID, d.Name), "")
		displaySubs := make([]menuItem, 0, len(lists))
		for _, list := range lists {
			displayMenu.AddSubMenuItem(list, "")
			displaySubs = append(displaySubs, menuItem{
				MenuItem: displayMenu,
				value:    list,
			})
		}
		subs = append(subs, displaySubs...)
	}

	m.addToList = menuItem{MenuItem: listMenu, subs: subs}
}

func (m *TrayMenu) createIntervalMenu() {
	mRotateInterval := systray.AddMenuItem("Rotate Interval…", "")

	m.intervals = intervalMenu{parent: mRotateInterval}
	m.addIntervalSubMenuItems(mRotateInterval)
}

func (m *TrayMenu) addIntervalSubMenuItems(parent *systray.MenuItem) {
	menuIntervals := []time.Duration{0}
	menuHasInterval := false
	set := make(map[time.Duration]bool)

	for _, interval := range menuIntervals {
		if interval == session.Interval() {
			menuHasInterval = true
			break
		}
	}

	// Add the interval menu items if it's missing
	if !menuHasInterval {
		menuIntervals = append(menuIntervals, session.Interval())
	}

	menuIntervals = append(menuIntervals, session.Config().MenuIntervals...)

	sort.Slice(menuIntervals, func(i, j int) bool {
		return menuIntervals[i] < menuIntervals[j]
	})

	for _, interval := range menuIntervals {
		if set[interval] {
			continue
		}
		set[interval] = true

		value := formatDuration(interval)
		check, checked := m.getIntervalCheckMark(interval)
		item := parent.AddSubMenuItemCheckbox(fmt.Sprintf("%s%s", check, value), "", checked)

		m.intervals.subs = append(m.intervals.subs, intervalItem{
			interval: interval,
			item:     item,
			value:    value,
		})
	}
}

func formatDuration(d time.Duration) string {
	h := int64(d.Hours())
	m := int64(d.Minutes()) % 60
	s := int64(d.Seconds()) % 60

	var parts []string
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%dh", h))
	}
	if m > 0 {
		parts = append(parts, fmt.Sprintf("%dm", m))
	}
	if s > 0 {
		parts = append(parts, fmt.Sprintf("%ds", s))
	}

	if len(parts) == 0 {
		return "Pause"
	}

	return strings.Join(parts, "")
}

func (m *TrayMenu) getIntervalCheckMark(interval time.Duration) (string, bool) {
	if interval == session.Interval() {
		if session.Type() == session.SessionTypeMacOS {
			return "", true
		}

		return "✔ ", true
	}
	if session.Type() == session.SessionTypeMacOS {
		return "", false
	}
	return "  ", false
}
