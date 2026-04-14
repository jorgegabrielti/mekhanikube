package diagnosis

import (
	"embed"
	"fmt"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed knowledge/*.yaml
var knowledgeFS embed.FS

// KnowledgeEntry represents a single entry in the knowledge base.
type KnowledgeEntry struct {
	Key             string   `yaml:"key"`
	Title           string   `yaml:"title"`
	Explanation     string   `yaml:"explanation"`
	Commands        []string `yaml:"commands"`
	WindowsCommands []string `yaml:"windows_commands"`
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
// For cluster-scoped resources (empty namespace), "-n {namespace}" is removed from commands.
func (kb *KnowledgeBase) Enrich(p *Problem) {
	entry, ok := kb.entries[p.RemediationKey]
	if !ok {
		return
	}

	rawCommands := entry.Commands
	if runtime.GOOS == "windows" && len(entry.WindowsCommands) > 0 {
		rawCommands = entry.WindowsCommands
	}

	var commands []string
	for _, cmd := range rawCommands {
		cmd = strings.ReplaceAll(cmd, "{pod}", p.Name)
		cmd = strings.ReplaceAll(cmd, "{name}", p.Name)
		cmd = strings.ReplaceAll(cmd, "{kind}", strings.ToLower(p.Resource))

		if p.Namespace != "" {
			cmd = strings.ReplaceAll(cmd, "{namespace}", p.Namespace)
		} else {
			// Cluster-scoped resources: remove namespace flags entirely
			cmd = strings.ReplaceAll(cmd, " -n {namespace}", "")
			cmd = strings.ReplaceAll(cmd, " --namespace={namespace}", "")
			cmd = strings.ReplaceAll(cmd, " --namespace {namespace}", "")
			cmd = strings.ReplaceAll(cmd, "{namespace}", "")
		}

		cmd = strings.TrimSpace(cmd)
		if cmd != "" {
			commands = append(commands, cmd)
		}
	}
	// For Pod resources with Deployment-style names (name-rsHash-podHash),
	// prepend a command to find current pods matching the base deployment name.
	// This helps when the scanned pod was replaced before the user runs remediation.
	if strings.EqualFold(p.Resource, "Pod") && p.Namespace != "" {
		podBase := derivePodBase(p.Name)
		if podBase != p.Name {
			findCmd := fmt.Sprintf("kubectl get pods -n %s | grep %s # Find current pod(s)", p.Namespace, podBase)
			commands = append([]string{findCmd}, commands...)
		}
	}

	p.Remediation = commands
	p.Explanation = strings.TrimSpace(entry.Explanation)
}

// derivePodBase strips Deployment-style hash suffixes from a pod name.
// "my-app-7b8f6c9d5-x9k2m" → "my-app"
// "my-app-0" (StatefulSet) → "my-app-0" (unchanged)
func derivePodBase(podName string) string {
	parts := strings.Split(podName, "-")
	if len(parts) < 3 {
		return podName
	}
	last := parts[len(parts)-1]
	secondLast := parts[len(parts)-2]
	if isAlphanumHash(last, 4, 6) && isAlphanumHash(secondLast, 6, 12) {
		return strings.Join(parts[:len(parts)-2], "-")
	}
	return podName
}

func isAlphanumHash(s string, minLen, maxLen int) bool {
	if len(s) < minLen || len(s) > maxLen {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}
