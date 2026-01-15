# TUI Refactor Plan and Next Enhancements

## Current status
- Extracted /tui rendering composition into `internal/tui/view` with ViewState builders.
- Moved header/footer/table renderers into view package.
- Centralized footer rendering via `view.FooterModel`.
- Extracted prompt matching/autocomplete into `internal/tui/input`.
- Added view-level modal helpers for frames/buttons.
- Moved modal body rendering into view for:
  - Task detail
  - Confirm delete
  - Task form
  - Plan result

## Active plan (step-by-step migration)
1. Identify remaining modal body rendering in `internal/tui` and move into `internal/tui/view` with models + tests.
2. Move week summary modal body rendering to view.
3. Extract modal footer button selection into view helpers or a modal-specific model.
4. Continue breaking view/data boundaries: shift any remaining formatting strings in `internal/tui` into view models.
5. Validate with `make build`, `make test`, `make lint` after each step and keep plans updated in `agents-notes/PLAN.md`.

## Next enhancements
- Consolidate modal model builders in `internal/tui` (one function per modal) to keep logic discoverable.
- Expand view tests for modal bodies to cover warnings/validation cases in plan results.
- Add small helper constructors in view models to reduce call-site verbosity.
- Audit `internal/tui/week_summary.go` for view extraction and potential model reuse.
- Consider a `view.Styles` grouping to reduce long style lists passed into renderers.
