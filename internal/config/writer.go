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

	keyAppsBlocklist    = "apps_blocklist"
	keyMinWordLen       = "min_word_len"
	keyTrigramMargin    = "trigram_margin"
	keyTrigramFloor     = "trigram_floor"
	keyAutocorrectEvent = "autocorrect_event"
)

// createdHeaderComment heads every document the writer CREATES (the first
// toggle on a machine without a config, or «Настройки…» on an absent
// file) — the pointer to the key reference, never section content.
const createdHeaderComment = "goswitch config — создан тумблером меню; справочник ключей: docs/CONFIG.md"

// configPerm is the persisted config's mode (mnd) — owner-only by the
// e2e stand's configFilePerm precedent (test/e2e/case_select.go).
const configPerm = 0o600

// kvStride is the yaml.v3 mapping-node stride: Content alternates key
// node, value node (mnd).
const kvStride = 2

// errWriterNotMapping refuses a document whose root is not a mapping —
// there is no section to toggle inside a scalar or a sequence.
var errWriterNotMapping = errors.New("config document is not a mapping")

// errWriterSectionUnknown refuses to build a section the writer does not
// own — the toggle surface is the two named sections above.
var errWriterSectionUnknown = errors.New("unknown section")

// SetAutocorrectEnabled persists the autocorrect master switch into the
// YAML document at path — the menu toggle's storage (plan 07-05 is the
// first caller). The document's comments and structure survive the write;
// an absent file is the ensure branch (Task 2).
func SetAutocorrectEnabled(path string, on bool) error {
	return setDocumentToggle(path, sectionAutocorrect, keyEnabled, on)
}

// SetSoundEnabled persists the sound master switch into the YAML document
// at path — the «Звук» toggle rides the SAME Node round-trip mechanism as
// the autocorrect toggle (owner decision: one persist mechanism).
func SetSoundEnabled(path string, on bool) error {
	return setDocumentToggle(path, sectionSound, keyEnabled, on)
}

// EnsureDocument creates the full defaults document when the path has no
// file — the «Настройки…» prerequisite (the editor must never open an
// empty buffer, Q7) — and is a byte-level no-op when the file exists.
func EnsureDocument(path string) error {
	_, err := os.Stat(path)
	switch {
	case err == nil:
		return nil // exists: never rewrite a live document
	case !errors.Is(err, os.ErrNotExist):
		return fmt.Errorf("stat config %s: %w", path, err)
	}
	doc, err := buildFullDocument(false, true)
	if err != nil {
		return fmt.Errorf("assemble config %s: %w", path, err)
	}
	if err := rewriteAtomically(path, doc); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}

	return nil
}

// DefaultPath resolves the canonical user config path — the user config
// dir (XDG_CONFIG_HOME, else $HOME/.config) + /goswitch/config.yaml, the
// selfcheck-pinned path of plan 04-02. Empty when the user config dir
// cannot be resolved (no HOME/XDG in the environment).
func DefaultPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}

	return filepath.Join(dir, "goswitch", "config.yaml")
}

// setDocumentToggle is the ONE persist path: read the document, parse it
// into a Node tree, land the toggle (in-place flip, key insertion, whole
// section insertion — upsertToggle), encode back with the 2-space indent
// and write through the atomic temp+rename. An absent file is the ensure
// branch: the toggle CREATES the complete defaults document with the
// switch in place — generation happens only on a user action, never at
// daemon start (the 04-02 contract, owner decision adopt+watch). Any
// failure leaves the original file byte-untouched.
func setDocumentToggle(path, section, key string, on bool) error {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return ensureFullDocument(path, section, on)
	}
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
	if err := upsertToggle(root.Content[0], section, key, on); err != nil {
		return fmt.Errorf("edit config %s: %w", path, err)
	}

	return rewriteAtomically(path, &root)
}

// ensureFullDocument creates the complete defaults document with the
// toggled section's switch at the requested value and the sibling switch
// at its default — autocorrect OFF (D-54) and sound ON (the owner's
// default-ON verdict) when the other one is the toggle.
func ensureFullDocument(path, section string, on bool) error {
	autocorrectOn, soundOn := on, true
	if section == sectionSound {
		autocorrectOn, soundOn = false, on
	}
	doc, err := buildFullDocument(autocorrectOn, soundOn)
	if err != nil {
		return fmt.Errorf("assemble config %s: %w", path, err)
	}
	if err := rewriteAtomically(path, doc); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}

	return nil
}

// buildFullDocument assembles the complete Defaults()-equivalent document
// as a Node tree with the two switches at the given values — every other
// value comes from Defaults() (one source of truth with the schema, never
// duplicated literals), the shape is guaranteed strict-decodable because
// it round-trips through the marshaled struct itself.
func buildFullDocument(autocorrectOn, soundOn bool) (*yaml.Node, error) {
	cfg := Defaults()
	cfg.Autocorrect.Enabled = autocorrectOn
	cfg.Sound.Enabled = boolPtr(soundOn)
	raw, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal defaults document: %w", err)
	}
	var root yaml.Node
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("reparse defaults document: %w", err)
	}
	if len(root.Content) == 0 || root.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("assemble defaults document: %w", errWriterNotMapping)
	}
	root.Content[0].HeadComment = createdHeaderComment

	return &root, nil
}

// upsertToggle lands the toggle in the document mapping: the section's
// key flips in place, a missing key joins the existing section, a
// null/empty section is replaced wholesale and a missing section is
// appended whole — never a duplicate section (yaml.v3 refuses duplicate
// mapping keys at decode).
func upsertToggle(doc *yaml.Node, section, key string, on bool) error {
	for i := 0; i+1 < len(doc.Content); i += kvStride {
		if doc.Content[i].Value != section {
			continue
		}
		target := doc.Content[i+1]
		if target.Kind != yaml.MappingNode {
			fresh, err := buildSectionNode(section, on)
			if err != nil {
				return err
			}
			doc.Content[i+1] = fresh

			return nil
		}
		if !setScalar(target, key, on) {
			target.Content = append(target.Content, scalarString(key), scalarBool(on))
		}

		return nil
	}
	fresh, err := buildSectionNode(section, on)
	if err != nil {
		return err
	}
	doc.Content = append(doc.Content, scalarString(section), fresh)

	return nil
}

// buildSectionNode assembles a whole missing/null section from the schema
// defaults — the values cross-reference Defaults() and the 07-02 schema
// constants (change the pair together, never literals).
func buildSectionNode(section string, on bool) (*yaml.Node, error) {
	switch section {
	case sectionAutocorrect:
		d := Defaults().Autocorrect

		return mapping(
			scalarString(keyEnabled), scalarBool(on),
			scalarString(keyAppsBlocklist), emptySeq(),
			scalarString(keyMinWordLen), scalarInt(d.MinWordLen),
			scalarString(keyTrigramMargin), scalarFloat(d.TrigramMargin),
			scalarString(keyTrigramFloor), scalarFloat(d.TrigramFloor),
		), nil
	case sectionSound:
		d := Defaults().Sound

		return mapping(
			scalarString(keyEnabled), scalarBool(on),
			scalarString(keyAutocorrectEvent), scalarString(d.AutocorrectEvent),
		), nil
	default:
		return nil, fmt.Errorf("build section %s: %w", section, errWriterSectionUnknown)
	}
}

// mapping builds a mapping node from alternating key/value children.
func mapping(pairs ...*yaml.Node) *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map", Content: pairs}
}

// emptySeq builds the empty flow sequence — `key: []` in the document.
func emptySeq() *yaml.Node {
	return &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
}

// scalarString builds a plain string scalar.
func scalarString(s string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: s}
}

// scalarBool builds a plain bool scalar.
func scalarBool(b bool) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: strconv.FormatBool(b)}
}

// scalarInt builds a plain int scalar.
func scalarInt(i int) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.Itoa(i)}
}

// scalarFloat builds a plain float scalar — 'g' formatting renders the
// Defaults' 2.0/1.0 as the documents' "2"/"1".
func scalarFloat(f float64) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!float", Value: strconv.FormatFloat(f, 'g', -1, 64)}
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
