package config

import (
	"path/filepath"
	"runtime"
	"time"
)

// GetInstallDir returns the installation directory of the Packet Sentry agent
func GetInstallDir() string {
	if runtime.GOOS == "windows" {
		return `C:\Program Files\PacketSentry`
	}
	return "/opt/packet-sentry"
}

// GetLogFilePath returns the path of the file to which the structured loggers write
func GetLogFilePath() string {
	if runtime.GOOS == "windows" {
		installDir := GetInstallDir()
		return filepath.Join(installDir, "packet-sentry-agent.log")
	}
	return "/var/log/packet-sentry-agent.log"
}

// GetBootstrapFilePath returns the path of the agent bootstrap JSON file
func GetBootstrapFilePath() string {
	if runtime.GOOS == "windows" {
		installDir := GetInstallDir()
		return filepath.Join(installDir, "agentBootstrap.json")
	}
	return "/opt/packet-sentry/agentBootstrap.json"
}

// GetClientCertFilePath returns the path of the client certificate
func GetClientCertFilePath() string {
	if runtime.GOOS == "windows" {
		installDir := GetInstallDir()
		return filepath.Join(installDir, "client.crt")
	}
	return "/opt/packet-sentry/client.crt"
}

// GetPrivateKeyFilePath returns the path of the client private key
func GetPrivateKeyFilePath() string {
	if runtime.GOOS == "windows" {
		installDir := GetInstallDir()
		return filepath.Join(installDir, "client.key")
	}
	return "/opt/packet-sentry/client.key"
}

// GetCACertFilePath() returns the path of the CA cert used to issue the client cert
func GetCACertFilePath() string {
	if runtime.GOOS == "windows" {
		installDir := GetInstallDir()
		return filepath.Join(installDir, "ca.crt")
	}
	return "/opt/packet-sentry/ca.crt"
}

// GetCertCheckInterval returns the interval at which we should check whether the client cert needs to renew
func GetCertCheckInterval() time.Duration {
	return 5 * time.Minute
}

// GetPollInterval returns the poll interval
func GetPollInterval() time.Duration {
	return 1 * time.Minute
}

// GetBPFConfigFilePath returns the path of cached on-disk BPF config
func GetBPFConfigFilePath() string {
	if runtime.GOOS == "windows" {
		installDir := GetInstallDir()
		return filepath.Join(installDir, "bpfConfig.json")
	}
	return "/opt/packet-sentry/bpfConfig.json"
}
