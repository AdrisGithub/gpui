package main

import (
	"bytes"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"os/exec"
	"strings"
	"time"
)

const playerctl = "playerctl"
const windowTitle = "  Now Playing "
const helpText = "\nPlay/Pause: [green]p[-]  Next: [green]n[-]  Previous: [green]b[-]  Quit: [green]q[-]"
const play_pause = "play-pause"

func main() {
	player := ""

	app := tview.NewApplication()

	outerBox := tview.NewBox().
		SetBorder(false).
		SetTitle(windowTitle).
		SetBorderColor(tcell.ColorGreen).
		SetTitleColor(tcell.ColorGreen).
		SetTitleAlign(tview.AlignCenter)

	songText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	controlText := tview.NewTextView().
		SetDynamicColors(true).
		SetText(helpText).
		SetTextAlign(tview.AlignCenter)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(outerBox, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewBox().SetBorder(false), 0, 1, false).
			AddItem(songText, 52, 1, true).
			AddItem(tview.NewBox().SetBorder(false), 0, 1, false),
			0, 4, false).
		AddItem(controlText, 5, 1, false)

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				songInfo, err := getSongInfo(player)
				if err != nil {
					songInfo = fmt.Sprintf("Error: %v", err)
				}

				app.QueueUpdateDraw(func() {
					songText.SetText(songInfo)
				})
			}
		}
	}()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q': // Quit the application
			app.Stop()
		case 'p':

			err := controlPlayer(play_pause)
			if err != nil {
				fmt.Printf("Error occured %s\n", err)
			}
		case 'n':
			err := controlPlayer("next")
			if err != nil {
				fmt.Printf("Error occured %s\n", err)
			}
		case 'b':
			err := controlPlayer("previous")
			if err != nil {
				fmt.Printf("Error occured %s\n", err)
			}
		}

		return event
	})

	// Run the application
	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}

func getSongInfo(player string) (string, error) {
	var cmd *exec.Cmd

	if player != "" {
		cmd = exec.Command(playerctl, "-p", player, "metadata", "--format", "{{title}}|{{artist}}|{{album}}|{{status}}")
	} else {
		cmd = exec.Command(playerctl, "metadata", "--format", "{{title}}|{{artist}}|{{album}}|{{status}}")
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	output := strings.TrimSpace(out.String())
	if output == "" {
		return "No song is currently playing.", nil
	}

	info := strings.Split(output, "|")
	if len(info) != 4 {
		return "Unexpected output format.", nil
	}

	title := truncateText(strings.TrimSpace(info[0]), 100)
	artist := truncateText(strings.TrimSpace(info[1]), 100)
	album := truncateText(strings.TrimSpace(info[2]), 40)
	status := strings.TrimSpace(info[3])

	cmd = exec.Command(playerctl, "metadata", "mpris:length")
	out.Reset()
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	// Song length is in microseconds, so convert it to seconds
	songLengthMicroseconds := strings.TrimSpace(out.String())
	var songLengthSeconds int64
	fmt.Sscanf(songLengthMicroseconds, "%d", &songLengthSeconds)
	songLengthSeconds = songLengthSeconds / 1e6 // Convert to seconds

	// Get current position
	cmd = exec.Command(playerctl, "position")
	out.Reset()
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	var currentPosition float64
	fmt.Sscanf(strings.TrimSpace(out.String()), "%f", &currentPosition)

	// Convert song length and position from seconds to mm:ss format
	currentPositionStr := formatTime(int64(currentPosition))
	songLengthStr := formatTime(songLengthSeconds)

	// Calculate progress percentage
	progressPercentage := currentPosition / float64(songLengthSeconds) * 100

	// Set a fixed progress bar width (e.g., 50 characters)
	progressBarTotalWidth := 25

	filledLength := int(progressPercentage / 100 * float64(progressBarTotalWidth))

	// Build the progress bar (e.g., [█████-----]) with the current progress
	progressBar := "[" + strings.Repeat("[green]█", filledLength) + strings.Repeat("[white]-", progressBarTotalWidth-filledLength) + "]"

	padding := "    "

	songInfo := fmt.Sprintf(
		"\n%s[green]Title: [-] %s\n%s[green]Artist:[-] %s\n%s[green]Album: [-] %s\n%s[green]Status:[-] %s\n\n%s%s %s/%s\n",
		padding, title,
		padding, artist,
		padding, album,
		padding, status,
		padding, progressBar, currentPositionStr, songLengthStr,
	)

	return songInfo, nil
}

func formatTime(seconds int64) string {
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func controlPlayer(command string) error {
	cmd := exec.Command("playerctl", command)
	return cmd.Run()
}

func truncateText(text string, maxLength int) string {
	if len(text) > maxLength {
		return text[:maxLength-3] + "..." // Leave room for ellipsis.
	}
	return text
}
