package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

const (
	envAccountName       = "TRANSIP_ACCOUNT_NAME"
	envPrivateKey        = "TRANSIP_PRIVATE_KEY"
	envPrivateKeyFile    = "TRANSIP_PRIVATE_KEY__FILE"
	defaultPrivateKeyOut = "/tmp/transip.key"
)

func main() {
	privateKeyPath, err := resolvePrivateKeyPath()
	if err != nil {
		fail(err)
	}

	if privateKeyPath != "" {
		if err := os.Setenv(envPrivateKey, privateKeyPath); err != nil {
			fail(fmt.Errorf("set %s: %w", envPrivateKey, err))
		}
	}

	if len(os.Args) < 2 {
		fail(fmt.Errorf("missing caddy command"))
	}

	cmd, err := exec.LookPath("caddy")
	if err != nil {
		fail(fmt.Errorf("locate caddy: %w", err))
	}

	if err := execCmd(cmd, os.Args[1:]); err != nil {
		fail(err)
	}
}

func resolvePrivateKeyPath() (string, error) {
	accountName := os.Getenv(envAccountName)
	rawKey := os.Getenv(envPrivateKey)
	filePath := os.Getenv(envPrivateKeyFile)
	if accountName == "" && rawKey == "" && filePath == "" {
		return "", nil
	}

	if accountName == "" {
		return "", fmt.Errorf("%s is required when using TransIP DNS", envAccountName)
	}

	if filePath != "" {
		return filepath.Clean(filePath), nil
	}

	if rawKey != "" {
		if isFile(rawKey) {
			return filepath.Clean(rawKey), nil
		}

		path := defaultPrivateKeyOut
		if err := os.WriteFile(path, []byte(rawKey), 0o600); err != nil {
			return "", fmt.Errorf("write %s: %w", path, err)
		}
		return path, nil
	}

	return "", fmt.Errorf("set %s or %s", envPrivateKey, envPrivateKeyFile)
}

func execCmd(cmd string, args []string) error {
	return syscall.Exec(cmd, append([]string{cmd}, args...), os.Environ())
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

func fail(err error) {
	_, _ = fmt.Fprintln(os.Stderr, "entrypoint:", err)
	os.Exit(1)
}
