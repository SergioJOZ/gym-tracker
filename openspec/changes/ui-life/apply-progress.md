# Apply Progress: ui-life — Animations & Micro-interactions

## Scope

PR 1 of stacked-to-main chain. Foundation (motion system + Routine.id) +
Shell crossfade + Routines Hero/stagger/scale integration.

## Mode

Standard (per orchestrator; strict_tdd:true configures the Go backend, not
Flutter). Tasks.md already documents RED/GREEN ordering; tests written
first per that ordering.

## Work Unit Evidence

| Evidence | Value |
|---|---|
| Focused test command and exact result | `flutter test test/core test/data test/features/shell` → `All tests passed!` (35 tests, exit 0). Full repo `flutter test` → 36 tests (incl. pre-existing `widget_test.dart`), exit 0. |
| Runtime harness command/scenario and exact result | `N/A` — no runtime harness command invoked by apply; manual `flutter run` pass is deferred to Phase 4 (task 4.3). Unit/widget/widget-test level coverage is the runtime proof for this slice. |
| Rollback boundary | Revert PR 1 files only: `lib/core/widgets/{scale_on_press,staggered_entrance,skeleton_box}.dart`, `MotionTokens` class block in `lib/core/theme/app_theme.dart`, `id` field in `lib/data/models/models.dart`, ids in `lib/data/mock/mock_data.dart`, `lib/features/shell/main_shell.dart`, `lib/features/routines/{routines_screen,routine_detail_screen}.dart`, and the 5 corresponding test files. Phase 3 / PR 2 files are untouched. |

## Completed Tasks

### Phase 1: Foundation

- [x] 1.1 RED — `test/core/motion_tokens_test.dart`
- [x] 1.2 GREEN — MotionTokens added to `lib/core/theme/app_theme.dart`
- [x] 1.3 RED — `test/data/routine_model_test.dart`
- [x] 1.4 GREEN — `id` field added to `Routine` with `copyWith`/`==`/`hashCode`
- [x] 1.5 — `lib/data/mock/mock_data.dart` populated with unique ids (`routine-push-day`, `routine-pull-day`, `routine-legs`)
- [x] 1.6 RED — `test/core/scale_on_press_test.dart`
- [x] 1.7 GREEN — `lib/core/widgets/scale_on_press.dart`
- [x] 1.8 RED — `test/core/staggered_entrance_test.dart`
- [x] 1.9 GREEN — `lib/core/widgets/staggered_entrance.dart`
- [x] 1.10 RED — `test/core/skeleton_box_test.dart`
- [x] 1.11 GREEN — `lib/core/widgets/skeleton_box.dart`

### Phase 2: Integration — Shell + Routines

- [x] 2.1 RED — `test/features/shell/main_shell_test.dart`
- [x] 2.2 GREEN — `lib/features/shell/main_shell.dart` (AnimatedSwitcher + GlobalKey + PageStorageBucket)
- [x] 2.3 — `lib/features/routines/routines_screen.dart` (ScaleOnPress + Hero + StaggeredEntrance.wrap + PageStorageKey on ListView)
- [x] 2.4 — `lib/features/routines/routine_detail_screen.dart` (matching Hero on header card)

## Files Changed

| File | Action | Brief |
|------|--------|-------|
| `frontend/test/core/motion_tokens_test.dart` | Created | RED: 7 tests covering `disabled`/`resolve` for disableAnimations/accessibleNavigation/no-MediaQuery paths + token presence |
| `frontend/lib/core/theme/app_theme.dart` | Modified | Added `abstract final class MotionTokens` with durations, curves, `disabled(ctx)`, `resolve(ctx, dur)` |
| `frontend/test/data/routine_model_test.dart` | Created | RED: 6 tests asserting `id` non-empty requirement + `copyWith`/`==` semantics |
| `frontend/lib/data/models/models.dart` | Modified | Added `final String id` to `Routine`; `copyWith`/`operator ==`/`hashCode` |
| `frontend/lib/data/mock/mock_data.dart` | Modified | Populated unique ids on every mock `Routine` |
| `frontend/test/core/scale_on_press_test.dart` | Created | RED: 5 tests covering tap-down scale 0.97, release/cancel 1.0, reduced-motion no-op, onTap callback |
| `frontend/lib/core/widgets/scale_on_press.dart` | Created | `StatefulWidget` with `GestureDetector` + `AnimatedScale` consuming `MotionTokens`; reduced-motion guard |
| `frontend/test/core/staggered_entrance_test.dart` | Created | RED: 7 tests covering 0/1/N children, per-child delay = i × stagger, TweenAnimationBuilder mechanics |
| `frontend/lib/core/widgets/staggered_entrance.dart` | Created | `StaggeredEntrance.wrap()` returning `StaggeredEntranceItem` widgets each wrapping a `TweenAnimationBuilder<double>` with fade+slide; reduced-motion guard |
| `frontend/test/core/skeleton_box_test.dart` | Created | RED: 5 tests covering dimension, shimmer-active, static-when-not-loading, no-advance, reduced-motion |
| `frontend/lib/core/widgets/skeleton_box.dart` | Created | `AnimationController.repeat()` + linear gradient sweep via `CustomPaint`; `isLoading` gates animation; reduced-motion falls back to static `DecoratedBox` |
| `frontend/test/features/shell/main_shell_test.dart` | Created | RED: 5 tests covering initial tab, scroll-offset preservation across switches, AnimatedSwitcher present, no IndexedStack, crossfade observable |
| `frontend/lib/features/shell/main_shell.dart` | Modified | Replaced `IndexedStack` with `AnimatedSwitcher` + `FadeTransition`; `GlobalKey` per screen; persistent `PageStorageBucket` for state restoration; crossfade duration `MotionTokens.resolve(context, MotionTokens.medium)` |
| `frontend/lib/features/routines/routines_screen.dart` | Modified | Wrapped `_RoutineCard` in `ScaleOnPress` + `Hero(tag: routine.id)`; replaced `InkWell` with `ScaleOnPress.onTap`; applied `StaggeredEntrance.wrap()` to the card list; added `PageStorageKey<String>('routines-list')` for crossfade state restoration |
| `frontend/lib/features/routines/routine_detail_screen.dart` | Modified | Wrapped the exercise-scheme `Card` in `Hero(tag: routine.id)` to receive shared-element from `_RoutineCard` |

## Deviations from Design

1. **`MotionTokens.spring` choice**: design suggested `Curve.springOut` (which is not a constant — it requires physics construction). Implemented as `Curves.easeOutBack`, which provides the same subtle overshoot feel in constant-form, keeping `MotionTokens.tokenAPI` const-friendly.

2. **AnimatedSwitcher + tab state preservation**: AnimatedSwitcher disposes its outgoing child when the transition completes (after `motion.medium`), so a `GlobalKey` alone cannot preserve scroll offset across multi-second switches. To keep the spec's literal "via `AnimatedSwitcher`" wording AND satisfy "preserve each tab's state across transitions", I added a persistent `PageStorageBucket` owned by `_MainShellState` wrapping the `AnimatedSwitcher`, and tagged Routines's ListView with a stable `PageStorageKey<String>('routines-list')`. Future PR 2 screens (progress, exercises, profile) MUST add matching `PageStorageKey`s to their primary `Scrollable`s for their per-screen state to behave identically — flagging this for the PR 2 task list.

3. **`scale_on_press.dart` shape**: design interface listed `class ScaleOnPress extends StatefulWidget { const ScaleOnPress({super.key, required this.child}); ... }`. Implemented with the same shape plus an optional `onTap` / `onTapDown` so feature code can pass nav handlers (used by `_RoutineCard`). The default press-scale constant is exposed as `ScaleOnPress.kPressedScale = 0.97`.

## Open Items / Handoff to Phase 4 / PR 2

- Pre-existing analyzer warnings exist in `lib/features/progress/progress_screen.dart` and `lib/features/workout/active_workout_screen.dart` (PR 2 files) — untouched here.
- Pre-existing `frontend/pubspec.yaml` working-tree modification (`dio`, `flutter_secure_storage` deps) predates this PR slice; unrelated to ui-life — left uncommitted and untouched.
- Phase 3 tasks (3.1–3.5) and Phase 4 verification tasks remain pending for PR 2.

## Workload / PR Boundary

- Mode: stacked PR slice
- Current work unit: PR 1 — Foundation + Shell + Routines
- Boundary: Starts fresh at Phase 1 task 1.1, ends at Phase 2 task 2.4. Does NOT include Phase 3 or Phase 4 work.
- Estimated review budget impact: 15 files changed (7 modified, 6 created lib files, 5 created test files), ~316 additions + ~93 deletions (authored excluding any generated goldens) — within the 400-line authored budget.

## Status

15/15 PR 1 tasks complete. Ready for stacked-PR slice to be committed/PR'd and then for Phase 3 (PR 2) to continue.