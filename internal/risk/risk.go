package risk

import (
	"fmt"
	"strings"

	"rehab-followup/internal/model"
)

type Engine struct {
	HighPainThreshold  float64
	LowCompletion      float64
	CriticalPain       float64
	CriticalCompletion float64
}

func NewEngine() *Engine {
	return &Engine{HighPainThreshold: 6, LowCompletion: 0.5, CriticalPain: 8, CriticalCompletion: 0.25}
}

func (e *Engine) Evaluate(completion float64, pain float64, overdue bool) model.RiskLevel {
	if pain >= e.CriticalPain && completion <= e.CriticalCompletion {
		return model.RiskCritical
	}
	if overdue && pain >= e.HighPainThreshold {
		return model.RiskCritical
	}
	if pain >= e.HighPainThreshold || completion < e.LowCompletion || overdue {
		return model.RiskHigh
	}
	if pain >= 4 || completion < 0.8 {
		return model.RiskWatch
	}
	return model.RiskLow
}

func Color(level model.RiskLevel) string {
	switch level {
	case model.RiskCritical:
		return "#b42318"
	case model.RiskHigh:
		return "#d97706"
	case model.RiskWatch:
		return "#ca8a04"
	default:
		return "#15803d"
	}
}

func Label(level model.RiskLevel) string {
	switch level {
	case model.RiskCritical:
		return "critical"
	case model.RiskHigh:
		return "high"
	case model.RiskWatch:
		return "watch"
	default:
		return "low"
	}
}

func ParseLevel(value string) (model.RiskLevel, error) {
	level := model.RiskLevel(strings.ToLower(strings.TrimSpace(value)))
	if level != model.RiskLow && level != model.RiskWatch && level != model.RiskHigh && level != model.RiskCritical {
		return "", fmt.Errorf("unknown risk level %q", value)
	}
	return level, nil
}
