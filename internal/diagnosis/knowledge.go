package diagnosis

import (
	"embed"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed knowledge/*.yaml
var knowledgeFS embed.FS

// KnowledgeEntry represents a single entry in the knowledge base.
type KnowledgeEntry struct {
	Key         string   `yaml:"key"`
	Title       string   `yaml:"title"`
	Explanation string   `yaml:"explanation"`
	Commands    []string `yaml:"commands"`
}

// KnowledgeBase holds all known problem patterns and their remediation steps.
type KnowledgeBase struct {
	entries map[string]KnowledgeEntry
}

// NewKnowledgeBase loads all embedded knowledge entries.
func NewKnowledgeBase() (*KnowledgeBase, error) {
	kb := &KnowledgeBase{
		entries: make(map[string]KnowledgeEntry),
	}

	files, err := knowledgeFS.ReadDir("knowledge")
	if err != nil {
		return nil, fmt.Errorf("failed to read knowledge directory: %w", err)
	}

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".yaml") {
			continue
		}

		data, err := knowledgeFS.ReadFile("knowledge/" + f.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read knowledge file %s: %w", f.Name(), err)
		}

		var entry KnowledgeEntry
		if err := yaml.Unmarshal(data, &entry); err != nil {
			return nil, fmt.Errorf("failed to parse knowledge file %s: %w", f.Name(), err)
		}

		kb.entries[entry.Key] = entry
	}

	return kb, nil
}

// Lookup returns the knowledge entry for a given key, if it exists.
func (kb *KnowledgeBase) Lookup(key string) (KnowledgeEntry, bool) {
	entry, ok := kb.entries[key]
	return entry, ok
}

// Enrich populates a Problem's Remediation field from the knowledge base.
// Placeholders in commands ({pod}, {name}, {namespace}) are replaced with actual values.
func (kb *KnowledgeBase) Enrich(p *Problem) {
	entry, ok := kb.entries[p.RemediationKey]
	if !ok {
		return
	}

	commands := make([]string, len(entry.Commands))
	for i, cmd := range entry.Commands {
		cmd = strings.ReplaceAll(cmd, "{pod}", p.Name)
		cmd = strings.ReplaceAll(cmd, "{name}", p.Name)
		cmd = strings.ReplaceAll(cmd, "{namespace}", p.Namespace)
		cmd = strings.ReplaceAll(cmd, "{kind}", strings.ToLower(p.Resource))
		commands[i] = cmd
	}
	p.Remediation = commands
}
