// Package profile persists the user's saved IP configurations in a small JSON
// file in the home directory. Every exported function is bound to Wails and can
// therefore be called concurrently from the frontend, so the file is guarded by
// a package mutex and always written atomically.
package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/yakutozcan/fast-ip-changer/pkg/sysexec"
)

// profileFileName is the per-user config file, kept at 0600 because it records
// the local network layout.
const profileFileName = ".ip_changer_profiles.json"

const filePerm = 0o600

type IPProfile struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	IP      string `json:"ip"`
	Subnet  string `json:"subnet"`
	Gateway string `json:"gateway"`
	DNS     string `json:"dns"`
}

var (
	// mu serialises the read-modify-write cycles below.
	mu sync.Mutex
	// pathOverride lets tests point the package at a temp directory instead of
	// the real ~/.ip_changer_profiles.json.
	pathOverride string
)

func profilePath() string {
	if pathOverride != "" {
		return pathOverride
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return profileFileName
	}
	return filepath.Join(home, profileFileName)
}

func defaultProfiles() []IPProfile {
	return []IPProfile{
		{ID: "1787251936547", Name: "DF_192.168.1.100", IP: "192.168.1.100", Subnet: "255.255.255.0", Gateway: "192.168.1.1", DNS: "1.1.1.1"},
		{ID: "1787251957412", Name: "DF_192.168.0.100", IP: "192.168.0.100", Subnet: "255.255.255.0", Gateway: "192.168.0.1", DNS: "1.1.1.1"},
		{ID: "1787251936549", Name: "DF_192.168.2.100", IP: "192.168.2.100", Subnet: "255.255.255.0", Gateway: "192.168.2.1", DNS: "1.1.1.1"},
	}
}

// GetProfiles returns the saved profiles, seeding the file with defaults the
// first time the app runs.
func GetProfiles() ([]IPProfile, error) {
	mu.Lock()
	defer mu.Unlock()
	return load()
}

// SaveProfiles replaces the stored list with profiles.
func SaveProfiles(profiles []IPProfile) error {
	mu.Lock()
	defer mu.Unlock()
	return save(profiles)
}

// AddProfile appends p to the stored list.
func AddProfile(p IPProfile) error {
	mu.Lock()
	defer mu.Unlock()

	profiles, err := load()
	if err != nil {
		return err
	}
	return save(append(profiles, p))
}

// UpdateProfile replaces the stored profile carrying the same ID as p.
func UpdateProfile(p IPProfile) error {
	mu.Lock()
	defer mu.Unlock()

	profiles, err := load()
	if err != nil {
		return err
	}

	for i := range profiles {
		if profiles[i].ID == p.ID {
			profiles[i] = p
			return save(profiles)
		}
	}
	return fmt.Errorf("güncellenecek profil bulunamadı: %s", p.ID)
}

// DeleteProfile removes the profile with the given ID.
func DeleteProfile(id string) error {
	mu.Lock()
	defer mu.Unlock()

	profiles, err := load()
	if err != nil {
		return err
	}

	// Pre-allocated and non-nil, so deleting the last profile still marshals to
	// "[]" instead of the literal "null".
	kept := make([]IPProfile, 0, len(profiles))
	for _, p := range profiles {
		if p.ID != id {
			kept = append(kept, p)
		}
	}
	return save(kept)
}

// load reads the profile file. Callers must hold mu.
func load() ([]IPProfile, error) {
	data, err := os.ReadFile(profilePath())
	if err != nil {
		if os.IsNotExist(err) {
			// First run: seed the file, but still hand back the defaults so the
			// app stays usable even when the write fails.
			defaults := defaultProfiles()
			if err := save(defaults); err != nil {
				return defaults, fmt.Errorf("varsayılan profiller kaydedilemedi: %w", err)
			}
			return defaults, nil
		}
		return nil, fmt.Errorf("profiller okunamadı: %w", err)
	}

	// An empty file, or one left holding "null" by an older build, is an empty
	// list — not a corrupt file and not a nil slice.
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		return []IPProfile{}, nil
	}

	var profiles []IPProfile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("profil dosyası okunamadı (bozuk JSON): %w", err)
	}
	if profiles == nil {
		profiles = []IPProfile{}
	}
	return profiles, nil
}

// save writes profiles atomically: a temp file in the same directory is filled,
// flushed and only then renamed over the target, so an interrupted write can
// never truncate the user's profiles. Callers must hold mu.
func save(profiles []IPProfile) error {
	if profiles == nil {
		profiles = []IPProfile{}
	}

	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return fmt.Errorf("profiller kodlanamadı: %w", err)
	}
	data = append(data, '\n')

	path := profilePath()
	dir := filepath.Dir(path)

	tmp, err := os.CreateTemp(dir, ".ip_changer_profiles-*.tmp")
	if err != nil {
		return fmt.Errorf("profiller kaydedilemedi: %w", err)
	}
	tmpName := tmp.Name()
	// No-op once the rename below has succeeded; cleans up on every error path.
	defer os.Remove(tmpName)

	if err := writeTemp(tmp, data); err != nil {
		return fmt.Errorf("profiller kaydedilemedi: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("profiller kaydedilemedi: %w", err)
	}
	return nil
}

func writeTemp(tmp *os.File, data []byte) error {
	defer tmp.Close()

	if err := tmp.Chmod(filePerm); err != nil && runtime.GOOS != "windows" {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	return tmp.Close()
}

// OpenProfileFolder reveals the profile file in the platform's file manager.
func OpenProfileFolder() error {
	path := profilePath()
	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}

	ctx, cancel := sysexec.WithTimeout(context.Background())
	defer cancel()

	switch runtime.GOOS {
	case "windows":
		if _, err := exec.LookPath("explorer"); err != nil {
			return fmt.Errorf("dosya yöneticisi açılamadı: %w", err)
		}
		// explorer.exe reports exit status 1 even when it opens the window, so
		// its exit code is deliberately ignored.
		_ = sysexec.Run(ctx, "explorer", "/select,"+path)
		return nil
	case "darwin":
		return sysexec.Run(ctx, "open", "-R", path)
	default:
		return sysexec.Run(ctx, "xdg-open", filepath.Dir(path))
	}
}
