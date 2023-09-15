package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	ipurge "github.com/eoea/ipurge/src"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const (
	WIN_START_POS     = 3
	MAX_WIN_HEIGHT    = 20
	MAX_WIN_WIDTH     = 110
	MAX_CHAR_LEN      = 20
	UNCHECK_INDICATOR = " [O] "
	CHECKED_INDICATOR = " [X] "
)

type appPurge struct {
	app             string
	toDelete        *[]string
	confirmedDelete *[]string
}

func (self *appPurge) viewHomeScreen() {

	ui.Clear()

	self.viewPurgeBrand()
	self.viewNavHelperScreen("         Press <ctrl-c> or <Esc> to exit. Press <enter> to continue.")

	p := widgets.NewParagraph()
	p.Title = " Start typing the name of the program you want to uninstall. "
	p.Text = `

Created by github.com/eoea/ipurge
Created 15/09/2023

Version 0.0.1
License MIT

WARNING: This tool is still experimental. If you check a file for deletion, it will be deleted! Other than an "Are you sure you want to delete these paths?" prompt, there are no other checks that prevent you from deleting the wrong thing. So type the full program name for better results and double check your selection.

> `
	p.TextStyle = ui.NewStyle(ui.ColorRed)
	p.SetRect(0, MAX_WIN_HEIGHT, MAX_WIN_WIDTH, WIN_START_POS)

	msgLen := len(p.Text) // Used so to prevent a buffer underflow when backspacing too much.
	ui.Render(p)

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			switch e.ID {
			case "<C-c>", "<Escape>": // <ctrl-c>
				self.close()
			case "<Enter>":
				self.app = strings.TrimSpace(p.Text[msgLen:])
				if len(self.app) == 0 {
					self.viewErrScreen("You need to pass a program name.")
				}
				self.viewSelectPathScreen()
			case "<Backspace>":
				if len(p.Text) > msgLen {
					p.Text = p.Text[:len(p.Text)-1]
					ui.Render(p)
				}
			case "<Space>", "<Tab>":
				if len(p.Text)-msgLen <= MAX_CHAR_LEN {
					p.Text += " "
					ui.Render(p)
				}
			default:
				if len(p.Text)-msgLen <= MAX_CHAR_LEN {
					p.Text += e.ID
					ui.Render(p)
				}
			}
		}
	}
}

func (self *appPurge) viewErrScreen(s string) {

	ui.Clear()

	self.viewPurgeBrand()
	self.viewNavHelperScreen("     Press <ctrl-c> or <Esc> to exit. Press 'b' to go back.")

	p := widgets.NewParagraph()
	p.Text = fmt.Sprintf(" \n\n\n\n      Error: %s", s)
	p.TextStyle = ui.NewStyle(ui.ColorRed)
	p.SetRect(0, MAX_WIN_HEIGHT, MAX_WIN_WIDTH, WIN_START_POS)

	ui.Render(p)
	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			switch e.ID {
			case "<C-c>", "<Esc>":
				self.close()
			case "B", "b":
				ui.Clear()
				self.viewHomeScreen()
			}
		}
	}
}

func (self *appPurge) viewSelectPathScreen() {

	g := widgets.NewGauge()
	g.Title = " Getting paths ... "
	g.SetRect(0, 0, 23, 3)
	g.BarColor = ui.ColorRed
	g.LabelStyle = ui.NewStyle(ui.ColorWhite)
	g.BorderStyle.Fg = ui.ColorWhite

	re := regexp.MustCompile(fmt.Sprintf("(?i)%s", self.app))
	for path, v := range ipurge.PATHS {
		ui.Render(g)
		ipurge.WalkDir(path, re, self.toDelete, v)
		g.Percent += 15
	}
	g.Percent = 100
	ui.Render(g)
	ui.Clear() // Clears the loading bar from screen.

	self.viewPurgeBrand()
	self.viewNavHelperScreen("Press <ctrl-c> or <Esc> to exit. Press <enter> to check box. Press 'd' to delete.")

	// List of paths to delete.
	l := widgets.NewList()
	l.Rows = []string{}
	for _, path := range *self.toDelete {
		l.Rows = append(l.Rows, UNCHECK_INDICATOR+path)
	}
	l.TextStyle = ui.NewStyle(ui.ColorRed)
	l.WrapText = false
	l.SetRect(0, MAX_WIN_HEIGHT, MAX_WIN_WIDTH, WIN_START_POS)

	ui.Render(l)

	nList := len(l.Rows)
	pos := 0

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			switch e.ID {
			case "<C-c>", "<Escape>":
				self.close()
			case "b":
				self.toDelete = &[]string{}
				ui.Clear()
				self.viewHomeScreen()
			case "j", "<Down>":
				l.ScrollDown()
				if pos < nList-1 {
					pos += 1
				}
			case "k", "<Up>":
				l.ScrollUp()
				if pos > 0 {
					pos -= 1
				}
			case "d":
				if len(*self.confirmedDelete) != 0 {
					self.viewDeletePromptScreen()
				}
				self.viewErrScreen("You need to select at least one file path")
			case "<Enter>":
				if strings.HasPrefix(l.Rows[pos], UNCHECK_INDICATOR) {
					s, _ := strings.CutPrefix(l.Rows[pos], UNCHECK_INDICATOR)
					l.Rows[pos] = CHECKED_INDICATOR + s
					*self.confirmedDelete = append(*self.confirmedDelete, s)
				} else {
					s, _ := strings.CutPrefix(l.Rows[pos], CHECKED_INDICATOR)
					temp := *self.confirmedDelete
					*self.confirmedDelete = nil
					for _, i := range temp {
						if i == s {
							continue
						}
						*self.confirmedDelete = append(*self.confirmedDelete, i)
					}
					l.Rows[pos] = UNCHECK_INDICATOR + s
				}
			}
			ui.Render(l)
		}
	}
}

func (self *appPurge) viewDeletePromptScreen() {
	ui.Clear()

	self.viewPurgeBrand()
	self.viewNavHelperScreen("Press <ctrl-c> or <Esc> to exit. Press 'y' to delete, 'n' to cancel and close program.")

	p := widgets.NewParagraph()
	p.TextStyle = ui.NewStyle(ui.ColorRed)
	p.SetRect(0, MAX_WIN_HEIGHT, MAX_WIN_WIDTH, WIN_START_POS)
	p.Text = "\n\n\n Are you sure you want to delete these paths?\n Press 'y' for yes or 'n' for no."
	ui.Render(p)

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			switch e.ID {
			case "<C-c>", "<Escape>":
				self.close()
			case "y", "Y":
				p.Text = ""
				for _, path := range *self.confirmedDelete {
					if err := ipurge.PurgeDir(path); err != nil {
						p.Text += "Skipped: " + path + "\n"
					} else {
						p.Text += "Deleted: " + path + "\n"
					}
					ui.Render(p)
				}
				ui.Render(p)
				goto VIEW_END
			case "n", "N":
				self.close()
			}
		}
	}

VIEW_END:
	self.viewNavHelperScreen("        Done! you can close this program by pressing <ctrl-c> or <Esc>.")
	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			switch e.ID {
			case "<C-c>", "<Escape>":
				self.close()
			}
		}
	}
}

func (self *appPurge) viewNavHelperScreen(s string) {
	p := widgets.NewParagraph()
	p.Title = ""
	p.Text = fmt.Sprintf("%s", s)
	p.TextStyle = ui.NewStyle(ui.ColorWhite)
	p.SetRect(24, 0, MAX_WIN_WIDTH, 3)
	ui.Render(p)
}

func (self *appPurge) viewPurgeBrand() {
	p := widgets.NewParagraph()
	p.Title = ""
	p.Text = "      iPurge 🐙"
	p.TextStyle = ui.NewStyle(ui.ColorWhite)
	p.SetRect(0, 0, 23, 3)
	ui.Render(p)
}

// Used as a hack to exit from any screen.
func (self *appPurge) close() {
	ui.Close()
	os.Exit(0)
}

func main() {

	a := &appPurge{}

	if err := ui.Init(); err != nil {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("ERROR: %S", err))
		os.Exit(1)
	}
	defer ui.Close()

	a.toDelete = &[]string{}
	a.confirmedDelete = &[]string{}
	a.viewHomeScreen()
}
