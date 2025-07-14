// Package readiness provides functionality for monitoring application readiness.
// This file implements readiness checking with multiple probes and result aggregation.
package readiness

import (
	"slices"
	"time"
)

// Readiness manages multiple readiness probes and their evaluation results.
// It maintains a collection of probes and provides methods to check overall readiness.
type Readiness struct {
	Labels []string          // Ordered list of probe labels
	Recap  Recap             // Last evaluation results
	Probes map[string]Probe  // Map of probe labels to probe instances
}

func NewReadiness() *Readiness {
	return &Readiness{
		Labels: make([]string, 0, 10),
		Probes: make(map[string]Probe, 10),
		Recap: Recap{
			Ok:  make([]string, 0, 10),
			NOk: make(map[string]Report, 10),
		},
	}
}

// Report represents a failure report for a readiness probe.
// It contains the reason why a probe failed its readiness check.
type Report struct {
	Reason string `json:"Reason,omitempty,omitzero"` // The reason for probe failure
}

// Recap represents a summary of all readiness probe evaluations.
// It includes the evaluation timestamp, overall readiness status,
// and lists of successful and failed probes.
type Recap struct {
	TS    time.Time         `json:"TS"`                      // Timestamp of the evaluation
	Ready bool              `json:"Ready,omitempty,omitzero"` // Overall readiness status
	Ok    []string          `json:"Ok,omitempty,omitzero"`   // List of successful probe labels
	NOk   map[string]Report `json:"NOk,omitempty,omitzero"`  // Map of failed probe labels to reports
}

func (r *Readiness) GetRecap() (state bool, recap Recap) {
	isOk := true
	r.Recap.Ok = r.Recap.Ok[:0]

	r.Recap.TS = time.Now().UTC()

	for _, label := range r.Labels {
		state, reason := r.Probes[label].IsWorking()
		isOk = isOk && state

		if state {
			r.Recap.Ok = append(r.Recap.Ok, label)
			delete(r.Recap.NOk, label)
			continue
		}

		r.Recap.NOk[label] = Report{
			Reason: reason,
		}
	}

	return isOk, r.Recap
}

type Probe interface {
	GetLabel() string
	IsWorking() (working bool, reason string)
}

func (r *Readiness) Register(p Probe) {
	label := p.GetLabel()
	r.Labels = append(r.Labels, label)
	slices.Sort(r.Labels)
	r.Probes[label] = p
}

func (r *Readiness) IsWorking() (working bool, recap Recap) {
	var okLabels []string
	failureReports := make(map[string]Report)

	for _, label := range r.Labels {
		if report := r.evaluateProbe(label); report != nil {
			failureReports[label] = *report
			continue
		}

		okLabels = append(okLabels, label)
	}

	allProbesWorking := len(failureReports) == 0

	return allProbesWorking, Recap{
		TS:    time.Now().UTC(),
		Ready: allProbesWorking,
		Ok:    okLabels,
		NOk:   failureReports,
	}
}

func (r *Readiness) evaluateProbe(label string) *Report {
	isWorking, reason := r.Probes[label].IsWorking()
	if isWorking {
		return nil
	}

	return &Report{Reason: reason}
}
