package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const (
	version = "0.1.2"
)

var (
	reComment = regexp.MustCompile(`//.*$`)
	reLoadout = regexp.MustCompile(`(?i).*setunitloadout *(\[[^;]+);.*`)
	reQitem   = regexp.MustCompile(`.*("[^"]+").*`)

	versioninfo = fmt.Sprintf("Sharona v%s\n© 2023 Tobias Klausmann\n\nSharona converts simple assignGear loadouts to ACE limited arsenals.\nhttps://github.com/klausman/sharona", version)

	showv = flag.Bool("v", false, "Show version number and exit")
)

func main() {
	flag.Parse()

	if *showv {
		fmt.Println(versioninfo)
		os.Exit(0)
	}

	a := app.New()
	w := a.NewWindow("Sharona")
	w.Resize(fyne.NewSize(500, 500))

	ab := fyne.NewMenuItem("About", func() { dialog.ShowInformation("About", versioninfo, w) })

	fm := fyne.NewMenu("File", ab)
	mm := fyne.NewMainMenu(fm)
	w.SetMainMenu(mm)

	input := makeEntry("Paste simple assignGear loadouts here...")
	output := makeEntry("Limited Arsenal code will appear here")

	c := make(chan string)
	button := widget.NewButton("Convert", func() {
		c <- input.Text
	})
	go func() {
		for {
			select {
			case t := <-c:
				output.SetText(getLAfromLO(t))
			}
		}
	}()
	boxes := container.NewGridWithColumns(1, input, output)
	content := container.NewBorder(button, nil, nil, nil, boxes)
	w.SetContent(content)

	w.ShowAndRun()
}

func makeEntry(pht string) *widget.Entry {
	w := widget.NewEntry()
	w.MultiLine = true
	w.SetMinRowsVisible(10)
	w.TextStyle.Monospace = true
	w.SetPlaceHolder(pht)
	return w
}

func getLAfromLO(s string) string {
	items := make(map[string]bool)
	for _, line := range strings.Split(s, "\n") {
		line = strings.Trim(line, " \n\r\t")
		line = reComment.ReplaceAllString(line, "")
		lo := reLoadout.FindStringSubmatch(line)
		if len(lo) == 0 {
			continue
		}
		for _, tok := range strings.Split(lo[1], ",") {
			stripped := reQitem.FindStringSubmatch(tok)
			if len(stripped) == 0 {
				continue
			}
			items[stripped[1]] = true
		}
	}
	itemlist := make([]string, 0, len(items))
	for k := range items {
		itemlist = append(itemlist, k)
	}
	sort.Strings(itemlist)
	return fmt.Sprintf("[this, [\n    %s]\n] call ace_arsenal_fnc_initBox;\n", strings.Join(itemlist, ",\n    "))
}
