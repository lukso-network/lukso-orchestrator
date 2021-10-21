package cmd

import (
	"github.com/lukso-network/lukso-orchestrator/shared/fileutil"
	"os"
	"path/filepath"
	"runtime"
)

const (
	DefaultHTTPHost             = "localhost" // Default host interface for the HTTP RPC server
	DefaultHTTPPort             = 8545        // Default TCP port for the HTTP RPC server
	DefaultWSHost               = "localhost" // Default host interface for the websocket RPC server
	DefaultWSPort               = 8546        // Default TCP port for the websocket RPC server
	DefaultIpcPath              = "orchestrator.ipc"
	DefaultVanguardRPCEndpoint  = "http://127.0.0.1:4000"
	DefaultPandoraRPCEndpoint   = "http://127.0.0.1:8545"
)

// DefaultConfigDir is the default config directory to use for the vaults and other
// persistence requirements.
func DefaultConfigDir() string {
	// Try to place the data folder in the user's home dir
	home := fileutil.HomeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "Lukso")
		} else if runtime.GOOS == "windows" {
			appdata := os.Getenv("APPDATA")
			if appdata != "" {
				return filepath.Join(appdata, "Lukso")
			}
			return filepath.Join(home, "AppData", "Roaming", "Lukso")
		}
		return filepath.Join(home, ".lukso")
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}
