package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// useTempFile points the package at a throwaway file so no test can touch the
// real ~/.ip_changer_profiles.json.
func useTempFile(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), profileFileName)
	previous := pathOverride
	pathOverride = path
	t.Cleanup(func() { pathOverride = previous })
	return path
}

func readRaw(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return strings.TrimSpace(string(data))
}

func sample(id string) IPProfile {
	return IPProfile{
		ID:      id,
		Name:    "Ofis " + id,
		IP:      "192.168.1." + id,
		Subnet:  "255.255.255.0",
		Gateway: "192.168.1.1",
		DNS:     "1.1.1.1",
	}
}

func TestGetProfilesSeedsDefaults(t *testing.T) {
	path := useTempFile(t)

	profiles, err := GetProfiles()
	if err != nil {
		t.Fatalf("GetProfiles: %v", err)
	}
	if len(profiles) != len(defaultProfiles()) {
		t.Fatalf("got %d profiles, want %d", len(profiles), len(defaultProfiles()))
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("defaults were not persisted: %v", err)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := useTempFile(t)

	want := []IPProfile{sample("10"), sample("11")}
	if err := SaveProfiles(want); err != nil {
		t.Fatalf("SaveProfiles: %v", err)
	}

	got, err := GetProfiles()
	if err != nil {
		t.Fatalf("GetProfiles: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d profiles, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("profile %d = %+v, want %+v", i, got[i], want[i])
		}
	}

	// The frontend models depend on these json tags.
	raw := readRaw(t, path)
	for _, key := range []string{`"id"`, `"name"`, `"ip"`, `"subnet"`, `"gateway"`, `"dns"`} {
		if !strings.Contains(raw, key) {
			t.Errorf("stored JSON is missing key %s: %s", key, raw)
		}
	}
}

func TestFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not honour POSIX file modes")
	}

	path := useTempFile(t)

	if err := SaveProfiles([]IPProfile{sample("10")}); err != nil {
		t.Fatalf("SaveProfiles: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != filePerm {
		t.Errorf("file mode = %04o, want %04o", perm, filePerm)
	}
}

// TestDeleteLastProfileLeavesEmptyArray is the regression test for the bug where
// deleting every profile wrote the literal "null", after which GetProfiles could
// never recover a usable list.
func TestDeleteLastProfileLeavesEmptyArray(t *testing.T) {
	path := useTempFile(t)

	if err := SaveProfiles([]IPProfile{sample("10")}); err != nil {
		t.Fatalf("SaveProfiles: %v", err)
	}
	if err := DeleteProfile("10"); err != nil {
		t.Fatalf("DeleteProfile: %v", err)
	}

	if raw := readRaw(t, path); raw != "[]" {
		t.Errorf("stored JSON = %q, want %q", raw, "[]")
	}

	profiles, err := GetProfiles()
	if err != nil {
		t.Fatalf("GetProfiles after deleting everything: %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("got %d profiles, want 0", len(profiles))
	}
	if profiles == nil {
		t.Error("GetProfiles returned a nil slice, want an empty one")
	}

	// The list must still be usable afterwards.
	if err := AddProfile(sample("12")); err != nil {
		t.Fatalf("AddProfile after deleting everything: %v", err)
	}
	profiles, err = GetProfiles()
	if err != nil {
		t.Fatalf("GetProfiles: %v", err)
	}
	if len(profiles) != 1 || profiles[0].ID != "12" {
		t.Errorf("got %+v, want a single profile with ID 12", profiles)
	}
}

func TestGetProfilesTolerantOfNullAndEmptyFiles(t *testing.T) {
	for _, content := range []string{"null", "", "  \n", "[]"} {
		t.Run(fmt.Sprintf("%q", content), func(t *testing.T) {
			path := useTempFile(t)
			if err := os.WriteFile(path, []byte(content), filePerm); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}

			profiles, err := GetProfiles()
			if err != nil {
				t.Fatalf("GetProfiles: %v", err)
			}
			if len(profiles) != 0 {
				t.Errorf("got %d profiles, want 0", len(profiles))
			}
		})
	}
}

func TestGetProfilesRejectsCorruptFile(t *testing.T) {
	path := useTempFile(t)
	if err := os.WriteFile(path, []byte("{not json"), filePerm); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := GetProfiles(); err == nil {
		t.Fatal("expected an error for a corrupt profile file")
	}
}

func TestUpdateProfile(t *testing.T) {
	useTempFile(t)

	if err := SaveProfiles([]IPProfile{sample("10"), sample("11")}); err != nil {
		t.Fatalf("SaveProfiles: %v", err)
	}

	edited := sample("11")
	edited.Name = "Yeni Ad"
	edited.IP = "10.0.0.5"
	edited.DNS = "8.8.8.8"

	if err := UpdateProfile(edited); err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}

	profiles, err := GetProfiles()
	if err != nil {
		t.Fatalf("GetProfiles: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("got %d profiles, want 2 (update must replace, not append)", len(profiles))
	}
	if profiles[0] != sample("10") {
		t.Errorf("untouched profile changed: %+v", profiles[0])
	}
	if profiles[1] != edited {
		t.Errorf("profile 1 = %+v, want %+v", profiles[1], edited)
	}
}

func TestUpdateProfileUnknownID(t *testing.T) {
	useTempFile(t)

	if err := SaveProfiles([]IPProfile{sample("10")}); err != nil {
		t.Fatalf("SaveProfiles: %v", err)
	}

	err := UpdateProfile(sample("999"))
	if err == nil {
		t.Fatal("expected an error for an unknown ID")
	}
	if !strings.Contains(err.Error(), "999") {
		t.Errorf("error %q should name the missing ID", err)
	}

	profiles, err := GetProfiles()
	if err != nil {
		t.Fatalf("GetProfiles: %v", err)
	}
	if len(profiles) != 1 {
		t.Errorf("got %d profiles, want 1 (a failed update must not write)", len(profiles))
	}
}

func TestConcurrentAddProfile(t *testing.T) {
	path := useTempFile(t)

	if err := SaveProfiles([]IPProfile{}); err != nil {
		t.Fatalf("SaveProfiles: %v", err)
	}

	const workers = 25

	var wg sync.WaitGroup
	errs := make([]error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = AddProfile(sample(fmt.Sprintf("%d", i)))
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("AddProfile(%d): %v", i, err)
		}
	}

	// The file must still be valid JSON, not a half-overwritten mix.
	var profiles []IPProfile
	if err := json.Unmarshal([]byte(readRaw(t, path)), &profiles); err != nil {
		t.Fatalf("stored JSON is corrupt: %v", err)
	}
	if len(profiles) != workers {
		t.Fatalf("got %d profiles, want %d", len(profiles), workers)
	}

	seen := make(map[string]bool, workers)
	for _, p := range profiles {
		if seen[p.ID] {
			t.Errorf("duplicate profile ID %q", p.ID)
		}
		seen[p.ID] = true
	}
	for i := 0; i < workers; i++ {
		if id := fmt.Sprintf("%d", i); !seen[id] {
			t.Errorf("profile %q was lost", id)
		}
	}
}

// TestSaveLeavesNoTempFiles guards the atomic-write path against leaking the
// temp files it renames from.
func TestSaveLeavesNoTempFiles(t *testing.T) {
	path := useTempFile(t)

	for i := 0; i < 5; i++ {
		if err := SaveProfiles([]IPProfile{sample(fmt.Sprintf("%d", i))}); err != nil {
			t.Fatalf("SaveProfiles: %v", err)
		}
	}

	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}
	if len(entries) != 1 {
		t.Errorf("directory holds %d entries, want only the profile file", len(entries))
	}
}
