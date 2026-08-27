package model

import (
	"fmt"
	"strings"
)

func NormalizeDepartment(value string) (string, error) {
	department := strings.TrimSpace(value)
	if department == "" {
		return "", fmt.Errorf("department cannot be empty")
	}
	if len(department) > 80 {
		return "", fmt.Errorf("department is too long")
	}
	return department, nil
}

func NormalizeDensity(value string) (Density, error) {
	density := Density(strings.ToLower(strings.TrimSpace(value)))
	if density != DensityCompact && density != DensityComfort && density != DensityRoomy {
		return "", fmt.Errorf("density must be compact, comfort, or roomy")
	}
	return density, nil
}

func ValidatePage(page, size int) error {
	if page < 1 {
		return fmt.Errorf("page must be positive")
	}
	if size < 1 || size > 200 {
		return fmt.Errorf("page size must be between 1 and 200")
	}
	return nil
}
