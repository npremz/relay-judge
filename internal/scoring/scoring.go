package scoring

import (
	"fmt"

	"relay-judge/internal/engine"
)

type Suggestion struct {
	Correction      int      `json:"correction"`
	EdgeCases       int      `json:"edge_cases"`
	Complexity      int      `json:"complexity"`
	PartialTotal    int      `json:"partial_total"`
	Notes           []string `json:"notes,omitempty"`
	ManualCriteria  []string `json:"manual_criteria,omitempty"`
	DecisionSupport []string `json:"decision_support,omitempty"`
}

func Build(report engine.Report) Suggestion {
	suggestion := Suggestion{
		ManualCriteria: []string{
			"Lisibilite / proprete du code a evaluer manuellement",
			"Bonus rapidite a saisir selon le rang d'arrivee",
		},
	}

	if report.Status == "load_error" {
		suggestion.Correction = 0
		suggestion.EdgeCases = 0
		suggestion.Complexity = 0
		suggestion.PartialTotal = 0
		suggestion.Notes = append(suggestion.Notes, "Le fichier ne se charge pas correctement.")
		suggestion.DecisionSupport = append(suggestion.DecisionSupport, "Profil probable: hors sujet ou non fonctionnel.")
		return suggestion
	}

	if report.Status == "runtime_error" {
		suggestion.Correction = 0
		suggestion.EdgeCases = 0
		suggestion.Complexity = 0
		suggestion.PartialTotal = 0
		suggestion.Notes = append(suggestion.Notes, "La solution leve une exception pendant l'execution.")
		suggestion.DecisionSupport = append(suggestion.DecisionSupport, "Profil probable: bonne intention mais execution instable.")
		return suggestion
	}

	coreRatio, hasCore := ratioForGroup(report, "core")
	edgeRatio, hasEdge := ratioForGroup(report, "edge")
	perfRatio, hasPerf := ratioForGroup(report, "perf")
	antiRatio, hasAnti := ratioForGroup(report, "anti-hardcode")

	if hasCore {
		suggestion.Correction = score40(coreRatio)
		suggestion.Notes = append(suggestion.Notes, fmt.Sprintf("Signal correction: %.0f%% des tests core passes.", coreRatio*100))
	}

	if hasEdge {
		suggestion.EdgeCases = score20(edgeRatio)
		suggestion.Notes = append(suggestion.Notes, fmt.Sprintf("Signal edge cases: %.0f%% des tests edge passes.", edgeRatio*100))
	}

	complexitySignal, complexityNote := complexitySignal(report.Status, hasPerf, perfRatio, hasAnti, antiRatio)
	suggestion.Complexity = complexityScore20(complexitySignal)
	if complexityNote != "" {
		suggestion.Notes = append(suggestion.Notes, complexityNote)
	}

	suggestion.PartialTotal = suggestion.Correction + suggestion.EdgeCases + suggestion.Complexity
	suggestion.DecisionSupport = buildDecisionSupport(suggestion)

	return suggestion
}

func ratioForGroup(report engine.Report, groupName string) (float64, bool) {
	for _, group := range report.Groups {
		if group.Name != groupName {
			continue
		}
		if group.Total == 0 {
			return 0, true
		}
		return float64(group.Passed) / float64(group.Total), true
	}
	return 0, false
}

func score40(ratio float64) int {
	switch {
	case ratio <= 0:
		return 0
	case ratio < 0.4:
		return 10
	case ratio < 0.7:
		return 20
	case ratio < 1:
		return 30
	default:
		return 40
	}
}

func score20(ratio float64) int {
	switch {
	case ratio <= 0:
		return 0
	case ratio < 0.35:
		return 5
	case ratio < 0.68:
		return 10
	case ratio < 1:
		return 15
	default:
		return 20
	}
}

func complexitySignal(status string, hasPerf bool, perfRatio float64, hasAnti bool, antiRatio float64) (float64, string) {
	if status == "timeout" {
		return 0, "Signal complexite: timeout global pendant l'evaluation."
	}

	switch {
	case hasPerf && hasAnti:
		return (perfRatio * 0.7) + (antiRatio * 0.3),
			fmt.Sprintf("Signal complexite: %.0f%% perf, %.0f%% anti-hardcode.", perfRatio*100, antiRatio*100)
	case hasPerf:
		return perfRatio, fmt.Sprintf("Signal complexite: %.0f%% des tests perf passes.", perfRatio*100)
	case hasAnti:
		return antiRatio, fmt.Sprintf("Signal complexite: %.0f%% des tests anti-hardcode passes.", antiRatio*100)
	default:
		return 0.5, "Signal complexite: aucun groupe perf/anti-hardcode defini, suggestion neutre."
	}
}

func complexityScore20(signal float64) int {
	switch {
	case signal <= 0:
		return 0
	case signal < 0.35:
		return 5
	case signal < 0.7:
		return 10
	case signal < 1:
		return 15
	default:
		return 20
	}
}

func buildDecisionSupport(suggestion Suggestion) []string {
	var notes []string

	switch {
	case suggestion.Correction >= 30 && suggestion.EdgeCases >= 15 && suggestion.Complexity >= 15:
		notes = append(notes, "Profil 4: tres bonne equipe, solution solide et generalisable.")
	case suggestion.Correction >= 30 && suggestion.EdgeCases < 15:
		notes = append(notes, "Profil 3: ca marche vite mais la robustesse semble fragile.")
	case suggestion.Complexity >= 15 && suggestion.Correction < 30:
		notes = append(notes, "Profil 2: bonne idee algorithmique mais implementation inachevee ou partielle.")
	case suggestion.Correction >= 30 && suggestion.Complexity < 15:
		notes = append(notes, "Profil 1: ca marche mais l'approche semble naive ou peu generalisable.")
	default:
		notes = append(notes, "Profil mixte: verifier manuellement la lisibilite et la generalisation de la solution.")
	}

	if suggestion.EdgeCases == 20 {
		notes = append(notes, "Les cas limites definis par le juge sont tous passes.")
	}

	if suggestion.Complexity <= 5 {
		notes = append(notes, "Attention: le signal complexite est faible.")
	}

	return notes
}
