package view

import (
	"strings"
	"testing"
)

func TestPlanResultFooterUsesValidationBranch(t *testing.T) {
	styles := ModalStyles{}

	withErrors := PlanResultFooter(true, styles)
	if !strings.Contains(withErrors, "[m] Amend") || strings.Contains(withErrors, "[Enter/a] Apply") {
		t.Fatalf("expected amend-only footer for validation errors, got %q", withErrors)
	}

	withoutErrors := PlanResultFooter(false, styles)
	if !strings.Contains(withoutErrors, "[Enter/a] Apply") {
		t.Fatalf("expected apply footer when no errors, got %q", withoutErrors)
	}
}

func TestWeekSummaryFooterTogglesLabels(t *testing.T) {
	styles := ModalStyles{}

	tasks := WeekSummaryFooter(true, styles)
	if !strings.Contains(tasks, "[s] Summary") {
		t.Fatalf("expected summary toggle in tasks footer, got %q", tasks)
	}

	summary := WeekSummaryFooter(false, styles)
	if !strings.Contains(summary, "[w] Tasks") {
		t.Fatalf("expected tasks toggle in summary footer, got %q", summary)
	}
}
