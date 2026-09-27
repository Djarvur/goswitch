package install

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestVerifyWritten pins the mismatch branch of the ASVS V14 read-back
// guard (T-04-01-01): an exact read-back passes, a divergent file fails
// with the named tampering error.
func TestVerifyWritten(t *testing.T) {
	path := filepath.Join(t.TempDir(), "goswitchd.service")
	want := []byte("[Service]\nExecStart=/opt/goswitch/goswitchd\n")
	if err := os.WriteFile(path, want, permPublic); err != nil {
		t.Fatalf("write the file: %v", err)
	}
	if err := verifyWritten(path, want); err != nil {
		t.Errorf("verifyWritten(exact) = %v, want nil", err)
	}

	tampered := []byte("[Service]\nExecStart=/usr/bin/evil\n")
	if err := os.WriteFile(path, tampered, permPublic); err != nil {
		t.Fatalf("tamper the file: %v", err)
	}
	err := verifyWritten(path, want)
	if err == nil {
		t.Fatal("verifyWritten(divergent) = nil, want the named tampering error")
	}
	if !errors.Is(err, errWriteVerify) {
		t.Errorf("verifyWritten(divergent) err = %v, want an errWriteVerify wrap", err)
	}
	if !strings.Contains(err.Error(), "tampering") {
		t.Errorf("err %q does not name the tampering suspicion", err)
	}
}
