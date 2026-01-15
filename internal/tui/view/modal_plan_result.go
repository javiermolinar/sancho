// Package view provides rendering helpers for the TUI.
package view

import "strings"

// PlanResultDay represents a day's task lines in the plan result modal.
type PlanResultDay struct {
	DateLabel string
	Lines     []string
}

// PlanResultModel contains the fields needed to render the plan result body.
type PlanResultModel struct {
	IntroMessage   string
	Issues         []string
	Warnings       []string
	Days           []PlanResultDay
	NoTasks        bool
	NoTasksMessage string
	Summary        string
	AmendHint      string
}

// PlanResultStyles groups styles for the plan result body.
type PlanResultStyles struct {
	MetaStyle         stringRenderer
	SectionTitleStyle stringRenderer
	BodyStyle         stringRenderer
}

type stringRenderer interface {
	Render(...string) string
}

// RenderPlanResultBody renders the modal body for a plan result.
func RenderPlanResultBody(model PlanResultModel, styles PlanResultStyles) string {
	var body strings.Builder

	body.WriteString(styles.MetaStyle.Render(model.IntroMessage) + "\n\n")

	if len(model.Issues) > 0 {
		body.WriteString(styles.SectionTitleStyle.Render("ISSUES") + "\n")
		for _, issue := range model.Issues {
			body.WriteString(styles.BodyStyle.Render("- "+issue) + "\n")
		}
		body.WriteString("\n")
	}

	if len(model.Warnings) > 0 {
		body.WriteString(styles.SectionTitleStyle.Render("WARNINGS") + "\n")
		for _, warning := range model.Warnings {
			body.WriteString(styles.BodyStyle.Render("- "+warning) + "\n")
		}
		body.WriteString("\n")
	}

	body.WriteString(styles.SectionTitleStyle.Render("DRAFT SCHEDULE") + "\n")
	if model.NoTasks {
		body.WriteString(styles.MetaStyle.Render(model.NoTasksMessage) + "\n")
	} else {
		for _, day := range model.Days {
			if day.DateLabel == "" {
				continue
			}
			body.WriteString(styles.BodyStyle.Render(day.DateLabel) + "\n")
			for _, line := range day.Lines {
				body.WriteString(styles.BodyStyle.Render(line) + "\n")
			}
			body.WriteString("\n")
		}
	}

	body.WriteString(styles.MetaStyle.Render(model.Summary) + "\n\n")
	body.WriteString(styles.SectionTitleStyle.Render("AMEND") + "\n")
	body.WriteString(styles.MetaStyle.Render(model.AmendHint) + "\n")

	return body.String()
}
