package powermenu

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type PowerAction struct {
	Name    string
	Icon    string
	Command string
}

type PowerMenu struct {
	Actions []PowerAction
}

func NewPowerMenu(settings map[string]interface{}) *PowerMenu {
	pm := &PowerMenu{}

	// Get custom commands from settings or use defaults
	lockCmd := getSettingString(settings, "lock_command", getDefaultLockCommand())
	screensaverCmd := getSettingString(settings, "screensaver_command", getDefaultScreensaverCommand())
	suspendCmd := getSettingString(settings, "suspend_command", getDefaultSuspendCommand())
	restartCmd := getSettingString(settings, "restart_command", getDefaultRestartCommand())
	shutdownCmd := getSettingString(settings, "shutdown_command", getDefaultShutdownCommand())

	pm.Actions = []PowerAction{
		{Name: "Lock", Icon: "system-lock-screen", Command: lockCmd},
		{Name: "Screensaver", Icon: "preferences-desktop-screensaver", Command: screensaverCmd},
		{Name: "Suspend", Icon: "system-suspend", Command: suspendCmd},
		{Name: "Restart", Icon: "system-reboot", Command: restartCmd},
		{Name: "Shutdown", Icon: "system-shutdown", Command: shutdownCmd},
	}

	return pm
}

func (pm *PowerMenu) ExecuteAction(actionName string) error {
	for _, action := range pm.Actions {
		if action.Name == actionName {
			return executeCommand(action.Command)
		}
	}
	return fmt.Errorf("action not found: %s", actionName)
}

func executeCommand(cmdStr string) error {
	if cmdStr == "" {
		return fmt.Errorf("empty command")
	}

	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return fmt.Errorf("invalid command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	return cmd.Start()
}

func getSettingString(settings map[string]interface{}, key, defaultValue string) string {
	if val, ok := settings[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// Default command detection based on environment
func getDefaultLockCommand() string {
	if isHyprland() {
		return "hyprlock"
	}
	if hasCommand("loginctl") {
		return "loginctl lock-session"
	}
	if hasCommand("xdg-screensaver") {
		return "xdg-screensaver lock"
	}
	if hasCommand("gnome-screensaver-command") {
		return "gnome-screensaver-command -l"
	}
	return "loginctl lock-session"
}

func getDefaultScreensaverCommand() string {
	if isHyprland() {
		return "hyprlock"
	}
	if hasCommand("xdg-screensaver") {
		return "xdg-screensaver activate"
	}
	if hasCommand("gnome-screensaver-command") {
		return "gnome-screensaver-command -a"
	}
	return "xdg-screensaver activate"
}

func getDefaultSuspendCommand() string {
	if hasCommand("systemctl") {
		return "systemctl suspend"
	}
	if hasCommand("loginctl") {
		return "loginctl suspend"
	}
	return "systemctl suspend"
}

func getDefaultRestartCommand() string {
	if hasCommand("systemctl") {
		return "systemctl reboot"
	}
	if hasCommand("loginctl") {
		return "loginctl reboot"
	}
	return "systemctl reboot"
}

func getDefaultShutdownCommand() string {
	if hasCommand("systemctl") {
		return "systemctl poweroff"
	}
	if hasCommand("loginctl") {
		return "loginctl poweroff"
	}
	return "systemctl poweroff"
}

func isHyprland() bool {
	return os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != ""
}

func hasCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func GetSystemUptime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "Unknown"
	}

	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return "Unknown"
	}

	var uptimeSeconds float64
	_, err = fmt.Sscanf(fields[0], "%f", &uptimeSeconds)
	if err != nil {
		return "Unknown"
	}

	duration := time.Duration(uptimeSeconds) * time.Second

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
