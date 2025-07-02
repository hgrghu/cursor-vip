package machineid

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"runtime"
	"sort"
	"strings"
)

// ID returns the platform specific machine ID of the current host OS.
// Regard the returned id as "dirty", meaning the information is potentially
// unsafe for cryptographic operations unless further processed by you.
func ID() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return macOSMachineID()
	case "linux":
		return linuxMachineID()  
	case "windows":
		return windowsMachineID()
	default:
		return fallbackMachineID()
	}
}

// ProtectedID returns a hashed version of the machine ID in a cryptographically secure way,
// using a fixed, application-specific key for additional security.
func ProtectedID(appKey string) (string, error) {
	id, err := ID()
	if err != nil {
		return "", err
	}
	return protectID(id, appKey), nil
}

// protectID hashes the id with the app key
func protectID(id, appKey string) string {
	h := sha256.New()
	h.Write([]byte(id))
	h.Write([]byte(appKey))
	return hex.EncodeToString(h.Sum(nil))
}

// macOSMachineID returns the machine ID for macOS
func macOSMachineID() (string, error) {
	// Try to read hardware UUID from system
	if id, err := readFile("/sys/class/dmi/id/product_uuid"); err == nil && id != "" {
		return strings.TrimSpace(id), nil
	}
	
	// Fallback to IOPlatformUUID if available
	if id, err := readFile("/proc/sys/kernel/random/boot_id"); err == nil && id != "" {
		return strings.TrimSpace(id), nil
	}
	
	return fallbackMachineID()
}

// linuxMachineID returns the machine ID for Linux
func linuxMachineID() (string, error) {
	// Try machine-id first (systemd)
	if id, err := readFile("/etc/machine-id"); err == nil && id != "" {
		return strings.TrimSpace(id), nil
	}
	
	// Try dbus machine-id
	if id, err := readFile("/var/lib/dbus/machine-id"); err == nil && id != "" {
		return strings.TrimSpace(id), nil
	}
	
	// Try hardware UUID
	if id, err := readFile("/sys/class/dmi/id/product_uuid"); err == nil && id != "" {
		return strings.TrimSpace(id), nil
	}
	
	return fallbackMachineID()
}

// windowsMachineID returns the machine ID for Windows
func windowsMachineID() (string, error) {
	// For Windows, we'll use a combination of environment variables
	// This is a simplified implementation
	computerName := os.Getenv("COMPUTERNAME")
	username := os.Getenv("USERNAME")
	
	if computerName != "" {
		h := sha256.New()
		h.Write([]byte(computerName + username))
		return hex.EncodeToString(h.Sum(nil))[:32], nil
	}
	
	return fallbackMachineID()
}

// fallbackMachineID creates a machine ID based on network interfaces
func fallbackMachineID() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("unable to get network interfaces: %v", err)
	}
	
	var macAddresses []string
	for _, iface := range interfaces {
		// Skip virtual interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		
		// Get MAC address
		if len(iface.HardwareAddr) > 0 {
			macAddresses = append(macAddresses, iface.HardwareAddr.String())
		}
	}
	
	if len(macAddresses) == 0 {
		return "", fmt.Errorf("no network interfaces with MAC addresses found")
	}
	
	// Sort to ensure consistent ordering
	sort.Strings(macAddresses)
	
	// Hash the combined MAC addresses
	combined := strings.Join(macAddresses, ":")
	h := sha256.New()
	h.Write([]byte(combined))
	return hex.EncodeToString(h.Sum(nil))[:32], nil
}

// readFile safely reads a file and returns its contents
func readFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}