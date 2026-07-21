# Proposal: ui-life — Animations & Micro-interactions

## Intent

The Flutter frontend is functionally complete across 5 screens but feels static. Apples Fitness sets a bar of precise, subtle motion that signals state change, hierarchy, and response without decoration. This change makes gym-tracker feel alive through technical micro-interactions while honoring the system reduced-motion accessibility setting.

## Scope

### In Scope
- Scale-on-press (~0.97, spring) for cards and the primary CTA button
- Staggered entrance (fade + slide-up, 30-50ms) for lists on Routines, Exercises, Progress, Profile
- Hero/shared-element transition between routine cards and `routine_detail_screen.dart`
- Animated weekly bars (bottom-up spring) on Progress
- Crossfade tab transitions replacing `IndexedStack` in `main_shell.dart` with `AnimatedSwitcher`
- Subtle timer-digit pulse on `active_workout_screen.dart`
- Skeleton/shimmer loading pattern (centralized) usable even on current mock data
- Reduced-motion gate via `MediaQuery.disableAnimations` / `accessibleNavigation`
- Motion duration/curve tokens in `app_theme.dart` for consistency

### Out of Scope
- Haptics (deferred sprint)
- Backend integration changes
- New screens or navigation routes
- Custom illustration/lottie assets

## Capabilities

### New Capabilities
- `ui-motion`: Reusable motion system — reduced-motion-aware animation tokens, scale-on-press wrapper, staggered list helper, hero transitions, animated bars, tab crossfade, timer pulse, and skeleton/shimmer placeholders.

### Modified Capabilities
None. Existing specs are backend-domain and unaffected at the requirement level.

## Approach

Build a thin motion layer in `core/widgets/` and `core/theme/`:
1. `MotionTokens` (durations, spring curves) inside `app_theme.dart`
2. `ScaleOnPress` widget + `StaggeredEntrance` helper for reuse
3. `SkeletonBox` shimmer widget driven by data-loading state
4. `AnimatedSwitcher` crossfade in `MainShell` (keeps state via `GlobalKey` per screen)
5. Reduced-motion single gate that short-circuits animations to instantaneous builds
Hero transitions use Flutter built-in `Hero` widget; no external animation packages.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `frontend/lib/core/theme/app_theme.dart` | Modified | Add `MotionTokens` (durations, curves) |
| `frontend/lib/core/widgets/` | New | `scale_on_press.dart`, `staggered_entrance.dart`, `skeleton_box.dart` |
| `frontend/lib/features/shell/main_shell.dart` | Modified | `IndexedStack` -> `AnimatedSwitcher` crossfade |
| `frontend/lib/features/routines/*.dart` | Modified | Scale-on-press, staggered entrance, Hero tags |
| `frontend/lib/features/exercises/exercise_catalog_screen.dart` | Modified | Scale-on-press, staggered entrance |
| `frontend/lib/features/progress/progress_screen.dart` | Modified | Animated bars, staggered entrance |
| `frontend/lib/features/profile/profile_screen.dart` | Modified | Staggered entrance |
| `frontend/lib/features/workout/active_workout_screen.dart` | Modified | Timer-digit pulse |
| `frontend/lib/core/widgets/primary_cta_button.dart` | Modified | Scale-on-press wrap |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Animation jank on low-end devices | Med | Profile on release build; respect reduced-motion; keep durations short |
| `AnimatedSwitcher` loses tab state | Med | Wrap each screen in a `GlobalKey`-backed `KeepAlive` |
| Reduced-motion edge cases missed | Low | Centralize gate in one helper; assert in widget tests |
| Hero tag collisions | Low | Derive tags from stable routine ids |

## Rollback Plan

Revert the merge commit. Motion is isolated in new core/widgets files and small per-screen edits; no data/schema changes. Removing the change restores the static `IndexedStack` shell and original screens without migration.

## Dependencies

- Flutter Material 3 (already used) — no new packages

## Success Criteria

- [ ] All 5 screens apply their target animation
- [ ] Every animation is a no-op when `MediaQuery.disableAnimations` is true
- [ ] Tab state survives crossfade transitions
- [ ] No new external packages
- [ ] `flutter analyze` clean on changed files