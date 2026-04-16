package diagnosis

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jorgegabrielti/nautikube/internal/version"
)

// Severity represents the severity level of a detected problem.
type Severity string

const (
	// Critical indicates a problem directly affecting availability.
	Critical Severity = "CRITICAL"
	// High indicates a problem requiring immediate attention.
	High Severity = "HIGH"
	// Medium indicates a moderate problem that should be resolved.
	Medium Severity = "MEDIUM"
	// Low indicates a minor issue or informational warning.
	Low Severity = "LOW"
	// Info indicates an informational finding.
	Info Severity = "INFO"
)

// SeverityWeight returns a numeric weight for sorting (higher = more severe).
func (s Severity) Weight() int {
	switch s {
	case Critical:
		return 5
	case High:
		return 4
	case Medium:
		return 3
	case Low:
		return 2
	case Info:
		return 1
	default:
		return 0
	}
}

// Problem represents a detected issue in the Kubernetes cluster.
type Problem struct {
	Resource          string   `json:"resource" yaml:"resource"`
	Namespace         string   `json:"namespace" yaml:"namespace"`
	Name              string   `json:"name" yaml:"name"`
	Issue             string   `json:"issue" yaml:"issue"`
	Explanation       string   `json:"explanation,omitempty" yaml:"explanation,omitempty"`
	OffendingProperty string   `json:"offendingProperty,omitempty" yaml:"offendingProperty,omitempty"`
	Severity          Severity `json:"severity" yaml:"severity"`
	Score             int      `json:"score" yaml:"score"`
	Remediation       []string `json:"remediation" yaml:"remediation"`
	RemediationKey    string   `json:"-" yaml:"-"`
	Details           []string `json:"details,omitempty" yaml:"details,omitempty"`
}

// String returns a human-readable representation of the problem.
func (p Problem) String() string {
	return fmt.Sprintf("[%s] %s %s/%s: %s (score: %d)", p.Severity, p.Resource, p.Namespace, p.Name, p.Issue, p.Score)
}

// CalculateScore computes the numeric score based on severity and contextual adjustments.
func (p *Problem) CalculateScore() {
	base := map[Severity]int{
		Critical: 90,
		High:     70,
		Medium:   50,
		Low:      30,
		Info:     10,
	}

	score := base[p.Severity]

	// Contextual adjustments
	if p.Namespace == "kube-system" || p.Namespace == "default" {
		score += 10
	}
	if p.Resource == "Pod" && containsAny(p.Issue, "CrashLoopBackOff", "ImagePullBackOff", "OOMKilled") {
		score += 10
	}
	if p.Resource == "Service" && strings.Contains(strings.ToLower(p.Issue), "no endpoints") {
		score += 10
	}

	// Quantitative factors: extract numbers from Issue string
	score += restartCountBonus(p.Issue)
	score += eventCountBonus(p.Issue)

	// Clamp to [0, 100]
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	p.Score = score
}

var reRestarted = regexp.MustCompile(`restarted (\d+) times`)
var reSeenTimes = regexp.MustCompile(`seen (\d+) times`)

// restartCountBonus returns a score bonus based on restart count (log-scale).
//
//	>100 restarts: +10, >50: +7, >20: +5, >10: +3
func restartCountBonus(issue string) int {
	m := reRestarted.FindStringSubmatch(issue)
	if m == nil {
		return 0
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0
	}
	switch {
	case n > 100:
		return 10
	case n > 50:
		return 7
	case n > 20:
		return 5
	case n > 10:
		return 3
	default:
		return 0
	}
}

// eventCountBonus returns a score bonus for event-based problems based on
// the "seen N times" count (log-scale capped at +10).
func eventCountBonus(issue string) int {
	m := reSeenTimes.FindStringSubmatch(issue)
	if m == nil {
		return 0
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n <= 10 {
		return 0
	}
	// log2-scale: bonus = min(10, floor(log2(n/10)))
	bonus := int(math.Log2(float64(n) / 10))
	if bonus > 10 {
		bonus = 10
	}
	return bonus
}

// containsAny checks if s contains any of the substrings (case-insensitive).
func containsAny(s string, substrs ...string) bool {
	lower := strings.ToLower(s)
	for _, sub := range substrs {
		if strings.Contains(lower, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}

// Summary holds aggregate statistics about scan results.
type Summary struct {
	ScanTime time.Time `json:"scanTime" yaml:"scanTime"`
	Version  string    `json:"version" yaml:"version"`
	Total    int       `json:"total" yaml:"total"`
	Critical int       `json:"critical" yaml:"critical"`
	High     int       `json:"high" yaml:"high"`
	Medium   int       `json:"medium" yaml:"medium"`
	Low      int       `json:"low" yaml:"low"`
	Info     int       `json:"info" yaml:"info"`
}

// NewSummary computes summary statistics from a list of problems.
func NewSummary(problems []Problem) Summary {
	s := Summary{
		ScanTime: time.Now(),
		Version:  version.Version,
		Total:    len(problems),
	}
	for _, p := range problems {
		switch p.Severity {
		case Critical:
			s.Critical++
		case High:
			s.High++
		case Medium:
			s.Medium++
		case Low:
			s.Low++
		case Info:
			s.Info++
		}
	}
	return s
}
