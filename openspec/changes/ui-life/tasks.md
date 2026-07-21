# Tasks: ui-life — Animations & Micro-interactions

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~550–680 (3 new + 11 modified files) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (Foundation + Shell + Routines) → PR 2 (Remaining screens) |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Focused test command | Runtime harness | Rollback boundary |
|------|------|-----------|----------------------|-----------------|-------------------|
| 1 | Motion foundation + shell crossfade + routine Hero | PR 1 | `flutter test test/core test/features/shell test/features/routines` | `flutter run` → navigate Routines → tap card → verify Hero + crossfade + reduced-motion | `core/widgets/{scale_on_press,staggered_entrance,skeleton_box}.dart`, MotionTokens class, Routine.id, main_shell + routines screens |
| 2 | Remaining screen animations + full verification | PR 2 | `flutter test && flutter analyze` | `flutter run` → exercise each screen → verify stagger/scale/pulse/bars | Per-screen edits in progress, workout, exercises, profile, primary_cta_button |

## Phase 1: Foundation (Motion System + Model)

- [x] 1.1 **RED** — `frontend/test/core/motion_tokens_test.dart`: widget test asserting `MotionTokens.disabled()` returns true when `disableAnimations` or `accessibleNavigation` is set; `resolve()` returns `Duration.zero`
- [x] 1.2 **GREEN** — `frontend/lib/core/theme/app_theme.dart`: add `abstract final class MotionTokens` with `fast/medium/slow/stagger` durations, `spring/standard` curves, `disabled(ctx)`, `resolve(ctx, dur)`
- [x] 1.3 **RED** — `frontend/test/data/routine_model_test.dart`: assert `Routine` requires non-empty `id` field
- [x] 1.4 **GREEN** — `frontend/lib/data/models/models.dart`: add `final String id` to `Routine`; update constructor/copyWith/equality
- [x] 1.5 — `frontend/lib/data/mock/mock_data.dart`: populate unique `id` on every mock `Routine`
- [x] 1.6 **RED** — `frontend/test/core/scale_on_press_test.dart`: gesture test asserting scale→0.97 on tap-down, →1.0 on release; no-op under reduced-motion
- [x] 1.7 **GREEN** — `frontend/lib/core/widgets/scale_on_press.dart`: `StatefulWidget` with `GestureDetector` + `AnimatedScale` consuming `MotionTokens`
- [x] 1.8 **RED** — `frontend/test/core/staggered_entrance_test.dart`: assert 0, 1, N children handled; item-i delay = i × stagger
- [x] 1.9 **GREEN** — `frontend/lib/core/widgets/staggered_entrance.dart`: static `wrap()` returning `TweenAnimationBuilder<double>` children with fade+slide
- [x] 1.10 **RED** — `frontend/test/core/skeleton_box_test.dart`: assert shimmer renders when `isLoading=true`, static when `false`
- [x] 1.11 **GREEN** — `frontend/lib/core/widgets/skeleton_box.dart`: `AnimationController` + linear gradient sweep; `isLoading` flag gates animation

## Phase 2: Integration — Shell + Routines (PR 1 scope)

- [x] 2.1 **RED** — `frontend/test/features/shell/main_shell_test.dart`: assert tab state (scroll offset) survives switch-away/switch-back; crossfade occurs
- [x] 2.2 **GREEN** — `frontend/lib/features/shell/main_shell.dart`: replace `IndexedStack` with `AnimatedSwitcher` + `GlobalKey` per screen; crossfade duration from `MotionTokens.medium`
- [x] 2.3 — `frontend/lib/features/routines/routines_screen.dart`: wrap `_RoutineCard` in `ScaleOnPress` + `Hero(tag: routine.id)`; apply `StaggeredEntrance.wrap()` to card list
- [x] 2.4 — `frontend/lib/features/routines/routine_detail_screen.dart`: add matching `Hero(tag: routine.id)` on header card

## Phase 3: Integration — Remaining Screens (PR 2 scope)

- [x] 3.1 — `frontend/lib/features/progress/progress_screen.dart`: convert `_WeeklyBars` to `StatefulWidget` with `AnimationController` (spring, bottom-up); apply `StaggeredEntrance` to stat cards and records
- [x] 3.2 — `frontend/lib/features/workout/active_workout_screen.dart`: add pulse `AnimationController` triggered on `_elapsedSeconds` change; wrap timer digits in `AnimatedBuilder` with scale tween
- [x] 3.3 — `frontend/lib/core/widgets/primary_cta_button.dart`: wrap `ElevatedButton` in `ScaleOnPress`
- [x] 3.4 — `frontend/lib/features/exercises/exercise_catalog_screen.dart`: apply `ScaleOnPress` + `StaggeredEntrance`
- [x] 3.5 — `frontend/lib/features/profile/profile_screen.dart`: apply `StaggeredEntrance`

## Phase 4: Verification

- [x] 4.1 — Run full widget test suite covering all spec scenarios (reduced-motion, Hero tag uniqueness, tab state, stagger edges)
- [x] 4.2 — `flutter analyze` clean on all changed files
- [ ] 4.3 — Manual runtime pass: each screen under normal + reduced-motion settings
