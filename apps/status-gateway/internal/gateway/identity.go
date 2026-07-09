package gateway

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	mev1 "realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1"
)

// EnrolledDevice is the gateway-owned identity minted for a device. The uid is
// opaque and system-assigned; callers persist it and echo it back on every
// report but must never construct it themselves.
type EnrolledDevice struct {
	UID         string          `json:"uid"`
	Kind        mev1.DeviceKind `json:"kind"`
	Role        mev1.DeviceRole `json:"role"`
	DisplayName string          `json:"display_name,omitempty"`
	Model       string          `json:"model,omitempty"`
	CreateTime  string          `json:"create_time"`
}

// IdentityStore owns the durable mapping of device uid to enrolled identity.
type IdentityStore struct {
	stateFile string
	mutex     sync.Mutex
	devices   map[string]*EnrolledDevice
}

func NewIdentityStore(stateFile string) *IdentityStore {
	return &IdentityStore{
		stateFile: stateFile,
		devices:   map[string]*EnrolledDevice{},
	}
}

func (store *IdentityStore) Load() error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	data, err := os.ReadFile(store.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var devices map[string]*EnrolledDevice
	if err := json.Unmarshal(data, &devices); err != nil {
		return err
	}
	for uid, device := range devices {
		store.devices[uid] = device
	}
	return nil
}

// Enroll mints a fresh opaque uid, records the device identity, and returns it.
func (store *IdentityStore) Enroll(kind mev1.DeviceKind, role mev1.DeviceRole, displayName string, model string, now time.Time) (*EnrolledDevice, error) {
	uid, err := mintDeviceUID()
	if err != nil {
		return nil, err
	}
	device := &EnrolledDevice{
		UID:         uid,
		Kind:        kind,
		Role:        role,
		DisplayName: displayName,
		Model:       model,
		CreateTime:  now.UTC().Format(time.RFC3339),
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.devices[device.UID] = device
	return device, store.saveLocked()
}

// Lookup returns the identity for a uid, if it is enrolled.
func (store *IdentityStore) Lookup(uid string) (*EnrolledDevice, bool) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	device, ok := store.devices[uid]
	return device, ok
}

// saveLocked persists the identity map atomically (temp file + fsync + rename).
func (store *IdentityStore) saveLocked() error {
	data, err := json.Marshal(store.devices)
	if err != nil {
		return err
	}
	directory := filepath.Dir(store.stateFile)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	temp, err := os.CreateTemp(directory, ".identity-state-*.tmp")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)

	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tempName, 0o600); err != nil {
		return err
	}
	return os.Rename(tempName, store.stateFile)
}

// mintDeviceUID fails loudly rather than minting a predictable identifier: a
// zeroed buffer would give every device the same uid.
func mintDeviceUID() (string, error) {
	buffer := make([]byte, 12)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("mint device uid: %w", err)
	}
	return "dev_" + hex.EncodeToString(buffer), nil
}
