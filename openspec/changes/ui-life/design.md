# Design: ui-life — Animations & Micro-interactions

## Technical Approach

Add a thin motion layer to `core/theme/` and `core/widgets/` that wraps Flutter's built-in animation primitives. Every motion widget consults a single reduced-motion gate before animating. The existing `IndexedStack` shell is replaced with `AnimatedSwitcher` + `GlobalKey` state preservation. Hero transitions use stable routine IDs (new field on the model). No external packages.

## Architecture Decisions

### Decision: MotionTokens location

**Choice**: Separate `MotionTokens` class inside `app_theme.dart` (same file, distinct class).
**Alternatives considered**: New `motion_tokens.dart` file (rejected — spec SHALL requires `app_theme.dart`; one extra file for ~20 lines is over-splitting).
**Rationale**: Follows spec requirement; keeps all theme-adjacent tokens co-located; `AppTheme` stays visual-only while `MotionTokens` owns durations/curves.

### Decision: Routine model needs an `id` field

**Choice**: Add `final String id` to `Routine` model, populate from mock data.
**Alternatives considered**: Use `routine.name` as Hero tag (rejected — names could collide, not stable across renames). Use list index (rejected — fragile on reorder).
**Rationale**: Spec requires "stable routine ids" for Hero tags. The model currently has no ID — this is a prerequisite change.

### Decision: Tab state preservation with AnimatedSwitcher

**Choice**: `GlobalKey` per screen widget + `AnimatedSwitcher` with `FadeTransition`. Only the active screen is in the widget tree; `GlobalKey` preserves its `State` across re-parenting.
**Alternatives considered**: Keep `IndexedStack` and overlay crossfade (rejected — double render cost). `PageView` (rejected — swipe gesture conflicts with scroll). `AutomaticKeepAliveClientMixin` (rejected — requires all children in tree, same as IndexedStack).
**Rationale**: `GlobalKey` re-parents the existing `State` when the widget moves in/out of the tree. Only the visible screen renders. Clean crossfade via `AnimatedSwitcher.defaultTransitionBuilder`.

### Decision: Reduced-motion gate implementation

**Choice**: `MotionTokens.disabled(BuildContext)` static getter returning `bool` — checks `MediaQuery.maybeOf(context)?.disableAnimations ?? false || accessibleNavigation`. All motion widgets call this; durations resolve to `Duration.zero` when true.
**Alternatives considered**: Per-widget `MediaQuery` checks (rejected — spec requires single centralized gate). InheritedWidget (rejected — over-engineering for a one-liner MediaQuery read).
**Rationale**: Single static method, called everywhere. Zero chance of inconsistency.

## Data Flow

```
MotionTokens (app_theme.dart)
    │
    ├── ScaleOnPress ──→ reads duration/curve, checks disabled()
    ├── StaggeredEntrance ──→ reads duration/stagger, checks disabled()
    ├── SkeletonBox ──→ shimmer AnimationController, checks disabled()
    ├── Hero tags ──→ routine.id from model
    ├── AnimatedSwitcher (MainShell) ──→ reads crossfade duration
    ├── _WeeklyBars ──→ reads spring curve, checks disabled()
    └── Timer pulse ──→ reads pulse duration, checks disabled()
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `core/theme/app_theme.dart` | Modify | Add `MotionTokens` class: durations (`fast`, `medium`, `slow`), curves (`spring`, `standard`), `disabled(context)` gate |
| `data/models/models.dart` | Modify | Add `id` field to `Routine` |
| `data/mock/mock_data.dart` | Modify | Populate `id` on each mock `Routine` |
| `core/widgets/scale_on_press.dart` | Create | `StatefulWidget` wrapping child with `AnimatedScale` (0.97 on press, 1.0 on release) |
| `core/widgets/staggered_entrance.dart` | Create | Helper returning `List<Widget>` wrapping each child in `TweenAnimationBuilder` with index-based delay |
| `core/widgets/skeleton_box.dart` | Create | Shimmer widget using `AnimationController` + linear gradient sweep; `isLoading` flag |
| `features/shell/main_shell.dart` | Modify | Replace `IndexedStack` with `AnimatedSwitcher`; add `GlobalKey` per screen |
| `features/routines/routines_screen.dart` | Modify | Wrap `_RoutineCard` in `ScaleOnPress` + `Hero(tag: routine.id)`; apply `StaggeredEntrance` to card list |
| `features/routines/routine_detail_screen.dart` | Modify | Add matching `Hero(tag: routine.id)` on header card |
| `features/progress/progress_screen.dart` | Modify | Convert `_WeeklyBars` to `StatefulWidget` with `AnimationController`; apply `StaggeredEntrance` to stat cards and records |
| `features/workout/active_workout_screen.dart` | Modify | Add pulse `AnimationController` triggered on `_elapsedSeconds` change; wrap timer digits in `AnimatedBuilder` with scale tween |
| `core/widgets/primary_cta_button.dart` | Modify | Wrap `ElevatedButton` in `ScaleOnPress` |
| `features/exercises/exercise_catalog_screen.dart` | Modify | Apply `ScaleOnPress` + `StaggeredEntrance` |
| `features/profile/profile_screen.dart` | Modify | Apply `StaggeredEntrance` |

## Interfaces / Contracts

```dart
// MotionTokens — added to app_theme.dart
abstract final class MotionTokens {
  static const fast = Duration(milliseconds: 150);
  static const medium = Duration(milliseconds: 250);
  static const slow = Duration(milliseconds: 400);
  static const stagger = Duration(milliseconds: 40);

  static const spring = Curve.springOut;  // Flutter built-in
  static const standard = Curves.easeOutCubic;

  /// Single reduced-motion gate. All motion widgets MUST call this.
  static bool disabled(BuildContext context) {
    final mq = MediaQuery.maybeOf(context);
    return mq?.disableAnimations == true || mq?.accessibleNavigation == true;
  }

  /// Returns [duration] or [Duration.zero] based on reduced-motion gate.
  static Duration resolve(BuildContext context, Duration duration) =>
      disabled(context) ? Duration.zero : duration;
}
```

```dart
// ScaleOnPress — core/widgets/scale_on_press.dart
class ScaleOnPress extends StatefulWidget {
  final Widget child;
  const ScaleOnPress({super.key, required this.child});
  // Uses GestureDetector + AnimatedScale (0.97 on tap-down, 1.0 on release/cancel)
}
```

```dart
// StaggeredEntrance — core/widgets/staggered_entrance.dart
class StaggeredEntrance {
  static List<Widget> wrap(List<Widget> children, {Duration stagger = MotionTokens.stagger});
  // Returns children wrapped in TweenAnimationBuilder< double > with fade+slide
}
```

```dart
// SkeletonBox — core/widgets/skeleton_box.dart
class SkeletonBox extends StatefulWidget {
  final bool isLoading;
  final double width, height;
  // Shimmer gradient animation when isLoading, static color otherwise
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `MotionTokens.disabled()` returns true for both flags | Widget test with `MediaQuery` override |
| Unit | `MotionTokens.resolve()` returns zero when disabled | Widget test |
| Widget | `ScaleOnPress` scales to 0.97 on tap-down | `flutter_test` gesture simulation |
| Widget | `StaggeredEntrance` handles 0, 1, N children | Widget test with empty/single/multi lists |
| Widget | `SkeletonBox` shows shimmer when loading, static when not | Pump widget with both flags |
| Widget | Tab state survives `AnimatedSwitcher` transition | Pump `MainShell`, scroll tab A, switch, switch back, verify scroll offset |
| Widget | Hero tags don't collide with distinct routine IDs | Pump two `_RoutineCard` widgets, verify no duplicate tag exception |
| Widget | Reduced-motion: all animations are 0ms | Wrap screen in `MediaQuery(disableAnimations: true)`, verify instant builds |

## Threat Matrix

N/A — no routing, shell, subprocess, VCS/PR automation, executable-file classification, or process-integration boundary.

## Migration / Rollout

No migration required. The `Routine.id` field is additive — mock data is the only data source. Rollback: revert merge commit; motion files are isolated in `core/widgets/` with small per-screen edits.

## Open Questions

- [ ] Should `SkeletonBox` expose a `child` slot for shaped skeletons (card vs. row), or keep it as a simple box? Leaning toward a `borderRadius` parameter for now.
- [ ] Timer pulse: scale 1.0→1.05→1.0 or opacity flash? Scale is more "alive" per Apple Fitness reference.
