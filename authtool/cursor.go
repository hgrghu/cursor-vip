package authtool

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// CursorVersion represents the version information of Cursor IDE
type CursorVersion struct {
	Version     string `json:"version"`
	Build       string `json:"build"`
	Product     string `json:"product"`
	Architecture string `json:"arch"`
}

// GetCursorVersion returns the version of Cursor IDE installed on the system
func GetCursorVersion() string {
	version := detectCursorVersion()
	if version != "" {
		return version
	}
	
	// Default fallback version
	return "0.42.0"
}

// detectCursorVersion detects the Cursor version based on the operating system
func detectCursorVersion() string {
	switch runtime.GOOS {
	case "darwin":
		return detectCursorVersionMacOS()
	case "linux":
		return detectCursorVersionLinux()
	case "windows":
		return detectCursorVersionWindows()
	default:
		return ""
	}
}

// detectCursorVersionMacOS detects Cursor version on macOS
func detectCursorVersionMacOS() string {
	// Common paths for Cursor on macOS
	paths := []string{
		"/Applications/Cursor.app/Contents/Resources/app/package.json",
		"/Applications/Cursor.45.app/Contents/Resources/app/package.json",
		"~/Applications/Cursor.app/Contents/Resources/app/package.json",
	}
	
	for _, path := range paths {
		if strings.HasPrefix(path, "~/") {
			home, err := os.UserHomeDir()
			if err == nil {
				path = filepath.Join(home, path[2:])
			}
		}
		
		if version := readVersionFromPackageJSON(path); version != "" {
			return version
		}
	}
	
	// Check for binary version
	if version := getVersionFromBinary("/Applications/Cursor.app/Contents/MacOS/Cursor"); version != "" {
		return version
	}
	
	return ""
}

// detectCursorVersionLinux detects Cursor version on Linux
func detectCursorVersionLinux() string {
	// Common paths for Cursor on Linux
	paths := []string{
		"/opt/cursor/resources/app/package.json",
		"/usr/local/bin/cursor/resources/app/package.json",
		"~/.local/share/cursor/resources/app/package.json",
	}
	
	for _, path := range paths {
		if strings.HasPrefix(path, "~/") {
			home, err := os.UserHomeDir()
			if err == nil {
				path = filepath.Join(home, path[2:])
			}
		}
		
		if version := readVersionFromPackageJSON(path); version != "" {
			return version
		}
	}
	
	// Check AppImage locations
	if version := checkAppImageVersion(); version != "" {
		return version
	}
	
	return ""
}

// detectCursorVersionWindows detects Cursor version on Windows
func detectCursorVersionWindows() string {
	// Common paths for Cursor on Windows
	appData := os.Getenv("LOCALAPPDATA")
	if appData != "" {
		paths := []string{
			filepath.Join(appData, "Programs", "cursor", "resources", "app", "package.json"),
			filepath.Join(appData, "cursor", "app-*", "resources", "app", "package.json"),
		}
		
		for _, path := range paths {
			if version := readVersionFromPackageJSON(path); version != "" {
				return version
			}
		}
	}
	
	// Check Program Files
	programFiles := os.Getenv("PROGRAMFILES")
	if programFiles != "" {
		path := filepath.Join(programFiles, "Cursor", "resources", "app", "package.json")
		if version := readVersionFromPackageJSON(path); version != "" {
			return version
		}
	}
	
	return ""
}

// readVersionFromPackageJSON reads version from package.json file
func readVersionFromPackageJSON(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	
	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return ""
	}
	
	if version, ok := pkg["version"].(string); ok {
		return version
	}
	
	return ""
}

// getVersionFromBinary attempts to get version from binary file
func getVersionFromBinary(binaryPath string) string {
	// This is a simplified implementation
	// In a real scenario, you might parse the binary or use system commands
	if _, err := os.Stat(binaryPath); err == nil {
		// If binary exists, assume it's a recent version
		return "0.42.0"
	}
	return ""
}

// checkAppImageVersion checks for AppImage installations on Linux
func checkAppImageVersion() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	
	// Common AppImage locations
	appImagePaths := []string{
		filepath.Join(home, "Downloads", "cursor-*.AppImage"),
		filepath.Join(home, "Applications", "cursor-*.AppImage"),
		"/opt/cursor-*.AppImage",
	}
	
	for _, pattern := range appImagePaths {
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			// Extract version from filename if possible
			for _, match := range matches {
				filename := filepath.Base(match)
				if strings.Contains(filename, "cursor-") {
					parts := strings.Split(filename, "-")
					if len(parts) >= 2 {
						version := strings.TrimSuffix(parts[1], ".AppImage")
						if version != "" {
							return version
						}
					}
				}
			}
		}
	}
	
	return ""
}

// IsCursorInstalled checks if Cursor IDE is installed on the system
func IsCursorInstalled() bool {
	version := GetCursorVersion()
	return version != ""
}

// GetCursorInstallPath returns the installation path of Cursor IDE
func GetCursorInstallPath() string {
	switch runtime.GOOS {
	case "darwin":
		paths := []string{
			"/Applications/Cursor.app",
			"/Applications/Cursor.45.app",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	case "linux":
		paths := []string{
			"/opt/cursor",
			"/usr/local/bin/cursor",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	case "windows":
		appData := os.Getenv("LOCALAPPDATA")
		if appData != "" {
			path := filepath.Join(appData, "Programs", "cursor")
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	return ""
}

// GetCursorExecutablePath returns the path to the Cursor executable
func GetCursorExecutablePath() string {
	switch runtime.GOOS {
	case "darwin":
		paths := []string{
			"/Applications/Cursor.app/Contents/MacOS/Cursor",
			"/Applications/Cursor.45.app/Contents/MacOS/Cursor",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	case "linux":
		paths := []string{
			"/opt/cursor/cursor",
			"/usr/local/bin/cursor",
			"/usr/bin/cursor",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	case "windows":
		paths := []string{
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "cursor", "Cursor.exe"),
			filepath.Join(os.Getenv("PROGRAMFILES"), "Cursor", "Cursor.exe"),
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	return ""
}