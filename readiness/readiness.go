package readiness

import (
	"slices"
	"time"
)

type Readiness struct {
	Labels []string
	Recap  Recap
	Probes map[string]Probe
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

type Report struct {
	Reason string `json:"Reason,omitempty,omitzero"`
}

type Recap struct {
	TS    time.Time         `json:"TS"`
	Ready bool              `json:"Ready,omitempty,omitzero"`
	Ok    []string          `json:"Ok,omitempty,omitzero"`
	NOk   map[string]Report `json:"NOk,omitempty,omitzero"`
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
