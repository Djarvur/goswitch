// The persist writer: menu-toggle persistence into the user's YAML
// document (plan 07-03). The write is a yaml.v3 Node round-trip — the one
// scalar under the toggle changes, every comment and the document
// structure survive byte-semantically (research Q3a) — landed through a
// same-directory temp file and an atomic rename (T-07-03-01), owner-only
// 0600 (T-07-03-02, the configFilePerm precedent). The writer logs
// nothing: journal records carry paths and outcomes only, never section
// content (D-20/D-21 extended to the new surface, T-07-03-06).

package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// The section/key vocabulary of the toggle surface — the single pair per
// persisted switch (the menu of plan 07-05 is the first caller).
const (
	sectionAutocorrect = "autocorrect"
	sectionSound       = "sound"
	keyEnabled         = "enabled"
)

// configPerm is the persisted config's mode (mnd) — owner-only by the
// e2e stand's configFilePerm precedent (test/e2e/case_select.go).
const configPerm = 0o600

// kvStride is the yaml.v3 mapping-node stride: Content alternates key
// node, value node (mnd).
const kvStride = 2

// errWriterNotMapping refuses a document whose root is not a mapping —
// there is no section to toggle inside a scalar or a sequence.
var errWriterNotMapping = errors.New("config document is not a mapping")

// errWriterSectionMissing refuses a flip whose section is absent from an
// existing document — the insertion branch of Task 2 replaces this
// refusal with a whole-section insert.
var errWriterSectionMissing = errors.New("section not found")

// SetAutocorrectEnabled persists the autocorrect master switch into the
// YAML document at path — the menu toggle's storage (plan 07-05 is the
// first caller). The document's comments and structure survive the write;
// an absent file is the ensure branch (Task 2).
func SetAutocorrectEnabled(path string, on bool) error {
	return setDocumentToggle(path, sectionAutocorrect, keyEnabled, on)
}

// setDocumentToggle is the ONE flip path: read the document, parse it into
// a Node tree, flip the single bool scalar under (section, key), encode
// back with the 2-space indent and land the bytes through the atomic
// temp+rename write. Any failure leaves the original file byte-untouched.
func setDocumentToggle(path, section, key string, on bool) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}
	var root yaml.Node
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return fmt.Errorf("decode config %s: %w", path, err)
	}
	// The walk starts at the DOCUMENT's root mapping — the DocumentNode
	// itself carries no key/value pairs (Pitfall 6: a walk from the
	// document node silently changes nothing).
	if len(root.Content) == 0 || root.Content[0].Kind != yaml.MappingNode {
		return fmt.Errorf("edit config %s: %w", path, errWriterNotMapping)
	}
	if !setMappingToggle(root.Content[0], section, key, on) {
		return fmt.Errorf("edit config %s: section %s: %w", path, section, errWriterSectionMissing)
	}

	return rewriteAtomically(path, &root)
}

// setMappingToggle flips the bool key inside the named section of the
// document mapping. Reports whether the section+key pair was found and
// set — a false return leaves the tree untouched.
func setMappingToggle(doc *yaml.Node, section, key string, on bool) bool {
	for i := 0; i+1 < len(doc.Content); i += kvStride {
		if doc.Content[i].Value != section {
			continue
		}

		return setScalar(doc.Content[i+1], key, on)
	}

	return false
}

// setScalar assigns the bool scalar in place — the node keeps its
// position, its comments and the rest of the section untouched; only Tag
// and Value move (!!bool "true"/"false", plain style).
func setScalar(sectionNode *yaml.Node, key string, on bool) bool {
	if sectionNode.Kind != yaml.MappingNode {
		return false
	}
	for i := 0; i+1 < len(sectionNode.Content); i += kvStride {
		if sectionNode.Content[i].Value != key {
			continue
		}
		v := sectionNode.Content[i+1]
		v.Tag = "!!bool"
		v.Value = strconv.FormatBool(on)
		v.Style = 0

		return true
	}

	return false
}

// rewriteAtomically encodes the node tree (2-space indent) and lands the
// bytes through the temp+rename write.
func rewriteAtomically(path string, root *yaml.Node) error {
	var out bytes.Buffer
	enc := yaml.NewEncoder(&out)
	enc.SetIndent(2) //nolint:mnd // the CONFIG.md 2-space indent is the plan-pinned literal (research Q3a)
	if err := enc.Encode(root); err != nil {
		_ = enc.Close() // teardown after an encode failure: the close error carries no signal

		return fmt.Errorf("encode config %s: %w", path, err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("encode config %s: %w", path, err)
	}

	return writeAtomically(path, out.Bytes())
}

// writeAtomically lands data at path through a temp file in the SAME
// directory (a rename across filesystems is not atomic) and os.Rename —
// POSIX-atomic, the reader sees either the old or the new file
// (T-07-03-01). The temp name carries a dot prefix, so the watcher's
// base-name filter ignores its events (research Q3b), and 0600 mode
// (T-07-03-02). A failure at any stage removes the temp and leaves the
// original untouched.
func writeAtomically(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp config %s: %w", path, err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op after a successful rename

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close() // teardown after a write failure: the close error carries no signal

		return fmt.Errorf("write temp config %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp config %s: %w", path, err)
	}
	if err := os.Chmod(tmpName, configPerm); err != nil {
		return fmt.Errorf("chmod temp config %s: %w", path, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename config %s: %w", path, err)
	}

	return nil
}
