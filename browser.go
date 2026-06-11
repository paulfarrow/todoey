package main

import (
	"os/exec"
	"runtime"
)

func openTaskInBrowser(taskID string) {
	url := "https://app.todoist.com/app/task/" + taskID
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", url).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		exec.Command("xdg-open", url).Start()
	}
}
