package notify

import (
	"fmt"
	"log"
	"os/exec"
	"syscall"
)

// Show Notification displays a quick non-blocking Windows toast/balloon notification
func ShowNotification(title, message string) {
	go func() {
		psScript := fmt.Sprintf(`
[void] [System.Reflection.Assembly]::LoadWithPartialName("System.Windows.Forms")
$icon = [System.Drawing.SystemIcons]::Information
$notify = New-Object System.Windows.Forms.NotifyIcon
$notify.Icon = $icon
$notify.BalloonTipTitle = "%s"
$notify.BalloonTipText = "%s"
$notify.Visible = $true
$notify.ShowBalloonTip(3000)
Start-Sleep -Seconds 3
$notify.Dispose()
`, title, message)

		cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if err := cmd.Run(); err != nil {
			log.Printf("[notify] Error showing notification: %v", err)
		}
	}()
}
