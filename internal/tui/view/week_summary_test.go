package view

import (
	"strings"
	"testing"
)

type testRenderer struct {
	prefix string
}

func (t testRenderer) Render(parts ...string) string {
	return t.prefix + strings.Join(parts, "")
}

func TestRenderWeekSummaryBodyWrapsAndStyles(t *testing.T) {
	lines := []WeekSummaryLine{
		{Text: "alpha beta", Style: WeekSummaryLineBody},
		{Text: "meta line", Style: WeekSummaryLineMeta},
		{Text: "section", Style: WeekSummaryLineSection},
	}
	styles := WeekSummaryStyles{
		BodyStyle:         testRenderer{prefix: "B:"},
		MetaStyle:         testRenderer{prefix: "M:"},
		SectionTitleStyle: testRenderer{prefix: "S:"},
	}

	output := RenderWeekSummaryBody(lines, styles, 6)

	if !strings.Contains(output, "B:alpha\nB:beta") {
		t.Fatalf("expected wrapped body lines, got %q", output)
	}
	if !strings.Contains(output, "M:meta\nM:line") {
		t.Fatalf("expected wrapped meta lines, got %q", output)
	}
	if !strings.Contains(output, "S:sectio\nS:n") {
		t.Fatalf("expected section line, got %q", output)
	}
}
