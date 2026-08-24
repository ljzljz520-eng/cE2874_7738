package model

import "sort"

func SortPatientsByRisk(patients []PatientSnapshot) {
	sort.SliceStable(patients, func(i, j int) bool {
		rank := func(r RiskLevel) int {
			switch r {
			case RiskCritical:
				return 4
			case RiskHigh:
				return 3
			case RiskWatch:
				return 2
			default:
				return 1
			}
		}
		if rank(patients[i].Batch.Risk) != rank(patients[j].Batch.Risk) {
			return rank(patients[i].Batch.Risk) > rank(patients[j].Batch.Risk)
		}
		return patients[i].Patient.Name < patients[j].Patient.Name
	})
}

func PainTrend(records []Record) []int {
	values := make([]int, 0, len(records))
	for _, record := range records {
		values = append(values, record.PainScore)
	}
	return values
}

func AveragePain(records []Record) float64 {
	if len(records) == 0 {
		return 0
	}
	total := 0
	for _, record := range records {
		total += record.PainScore
	}
	return float64(total) / float64(len(records))
}
