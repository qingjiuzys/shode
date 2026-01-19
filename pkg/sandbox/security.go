package sandbox

import (
	"fmt"
	"regexp"
	"strings"

	"gitee.com/com_818cloud/shode/pkg/types"
)

// SecurityChecker provides security validation for shell commands
type SecurityChecker struct {
	dangerousCommands map[string]bool
	fileBlacklist     map[string]bool
	networkBlacklist  map[string]bool
}

// NewSecurityChecker creates a new security checker with default rules
func NewSecurityChecker() *SecurityChecker {
	sc := &SecurityChecker{
		dangerousCommands: make(map[string]bool),
		fileBlacklist:     make(map[string]bool),
		networkBlacklist:  make(map[string]bool),
	}

	// Initialize default dangerous commands
	sc.initializeDangerousCommands()
	sc.initializeFileBlacklist()
	sc.initializeNetworkBlacklist()

	return sc
}

// initializeDangerousCommands sets up the default dangerous command blacklist
func (sc *SecurityChecker) initializeDangerousCommands() {
	dangerous := []string{
		"rm",       // File deletion
		"dd",       // Disk operations
		"mkfs",     // Filesystem operations
		"fdisk",    // Partition operations
		"shutdown", // System shutdown
		"reboot",   // System reboot
		"halt",     // System halt
		"poweroff", // Power off
		"chmod",    // Permission changes
		"chown",    // Ownership changes
		"useradd",  // User management
		"userdel",  // User deletion
		"groupadd", // Group management
		"groupdel", // Group deletion
		"passwd",   // Password changes
	}

	for _, cmd := range dangerous {
		sc.dangerousCommands[cmd] = true
	}
}

// initializeFileBlacklist sets up sensitive file paths to protect
func (sc *SecurityChecker) initializeFileBlacklist() {
	sensitiveFiles := []string{
		"/etc/passwd",
		"/etc/shadow",
		"/etc/sudoers",
		"/root/",
		"/boot/",
		"/dev/",
		"/proc/",
		"/sys/",
		"/var/log/",
	}

	for _, file := range sensitiveFiles {
		sc.fileBlacklist[file] = true
	}
}

// initializeNetworkBlacklist sets up dangerous network operations
func (sc *SecurityChecker) initializeNetworkBlacklist() {
	dangerousNetwork := []string{
		"iptables", // Firewall manipulation
		"ufw",      // Firewall manipulation
		"route",    // Network routing
		"ifconfig", // Network interface configuration
		"ip",       // Network configuration
		"nc",       // Netcat
		"nmap",     // Network scanning
		"tcpdump",  // Network sniffing
	}

	for _, cmd := range dangerousNetwork {
		sc.networkBlacklist[cmd] = true
	}
}

// CheckCommand validates a command for security risks
func (sc *SecurityChecker) CheckCommand(cmd *types.CommandNode) error {
	commandName := strings.ToLower(cmd.Name)

	// Check for dangerous commands
	if sc.dangerousCommands[commandName] {
		return fmt.Errorf("security violation: dangerous command '%s' is not allowed", commandName)
	}

	// Check for network-related dangerous commands
	if sc.networkBlacklist[commandName] {
		return fmt.Errorf("security violation: network command '%s' is not allowed", commandName)
	}

	// Check arguments for sensitive file access
	for _, arg := range cmd.Args {
		if sc.isSensitiveFile(arg) {
			return fmt.Errorf("security violation: access to sensitive file '%s' is not allowed", arg)
		}
	}

	// Additional pattern-based checks
	if err := sc.checkPatterns(cmd); err != nil {
		return err
	}

	return nil
}

// isSensitiveFile checks if a path matches sensitive file patterns
func (sc *SecurityChecker) isSensitiveFile(path string) bool {
	// Check exact matches
	if sc.fileBlacklist[path] {
		return true
	}

	// Check prefix matches (e.g., /root/anything)
	for blacklistedPath := range sc.fileBlacklist {
		if strings.HasPrefix(path, blacklistedPath) {
			return true
		}
	}

	return false
}

// checkPatterns performs regex-based security checks
func (sc *SecurityChecker) checkPatterns(cmd *types.CommandNode) error {
	// Combine command and arguments for pattern matching
	fullCommand := cmd.Name + " " + strings.Join(cmd.Args, " ")

	// Check for recursive deletion patterns (rm -rf /)
	recursiveDelete := regexp.MustCompile(`rm\s+.*-r.*\s+/(\s|$)`)
	if recursiveDelete.MatchString(fullCommand) {
		return fmt.Errorf("security violation: recursive deletion of root directory detected")
	}

	// Check for password in command line
	passwordPattern := regexp.MustCompile(`(-p|--password|passwd)\s+(\S+)`)
	if passwordPattern.MatchString(fullCommand) {
		return fmt.Errorf("security violation: password in command line detected")
	}

	// Check for shell injection patterns
	shellInjection := regexp.MustCompile(`[;&|` + "`" + `$()]`)
	// Exclude database functions and shode from shell injection check
	excludedCommands := map[string]bool{
		"shode":      true,
		"QueryDB":    true,
		"QueryRowDB": true,
		"ExecDB":     true,
	}
	if shellInjection.MatchString(fullCommand) && !excludedCommands[cmd.Name] {
		return fmt.Errorf("security violation: potential shell injection detected")
	}

	return nil
}

// AddDangerousCommand adds a custom dangerous command to the blacklist
func (sc *SecurityChecker) AddDangerousCommand(command string) {
	sc.dangerousCommands[strings.ToLower(command)] = true
}

// RemoveDangerousCommand removes a command from the dangerous commands blacklist
func (sc *SecurityChecker) RemoveDangerousCommand(command string) {
	delete(sc.dangerousCommands, strings.ToLower(command))
}

// AddSensitiveFile adds a custom sensitive file path to the blacklist
func (sc *SecurityChecker) AddSensitiveFile(filepath string) {
	sc.fileBlacklist[filepath] = true
}

// RemoveSensitiveFile removes a file path from the sensitive files blacklist
func (sc *SecurityChecker) RemoveSensitiveFile(filepath string) {
	delete(sc.fileBlacklist, filepath)
}

// GetSecurityReport generates a security report for a command
func (sc *SecurityChecker) GetSecurityReport(cmd *types.CommandNode) map[string]interface{} {
	report := make(map[string]interface{})
	report["command"] = cmd.Name
	report["arguments"] = cmd.Args
	report["line_number"] = cmd.Pos.Line

	// Check various security aspects
	report["is_dangerous_command"] = sc.dangerousCommands[strings.ToLower(cmd.Name)]
	report["is_network_command"] = sc.networkBlacklist[strings.ToLower(cmd.Name)]

	// Check for sensitive file access
	sensitiveFiles := []string{}
	for _, arg := range cmd.Args {
		if sc.isSensitiveFile(arg) {
			sensitiveFiles = append(sensitiveFiles, arg)
		}
	}
	report["sensitive_files"] = sensitiveFiles

	return report
}
