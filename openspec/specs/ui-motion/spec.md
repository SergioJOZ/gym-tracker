# ui-motion Specification

## Purpose

Reusable, reduced-motion-aware animation system for the gym-tracker Flutter app: motion tokens, scale-on-press, staggered entrance, skeleton shimmer, hero transitions, tab crossfade, animated bars, and timer pulse.

## Requirements

### Requirement: MotionTokens

The system SHALL expose centralized motion duration and curve tokens in `app_theme.dart`. All motion components MUST consume these tokens; raw `Duration` or curve literals in feature code MUST NOT be used.

#### Scenario: Tokens are reused

- GIVEN the motion system is imported
- WHEN any motion widget reads a duration or curve
- THEN it reads the value from `MotionTokens`
- AND no hardcoded durations appear in feature code

#### Scenario: Reduced-motion override

- GIVEN `MediaQuery.disableAnimations` is true
- WHEN any token is consumed by a motion widget
- THEN the effective duration is 0ms

### Requirement: ScaleOnPress

The system SHALL provide a `ScaleOnPress` widget that wraps any child and scales to approximately 0.97 with a spring curve on tap-down, restoring to 1.0 on release or cancel.

#### Scenario: Press scales in, release restores

- GIVEN a `ScaleOnPress` wrapping a card
- WHEN the user presses down
- THEN the child scales toward 0.97 via spring curve
- AND returns to 1.0 on release or cancel

#### Scenario: Reduced-motion skips scaling

- GIVEN reduced-motion is enabled
- WHEN the user presses the child
- THEN no scale change occurs

### Requirement: StaggeredEntrance

The system SHALL provide a `StaggeredEntrance` helper applying fade plus slide-up to list children with a configurable per-item stagger delay.

#### Scenario: List children stagger in

- GIVEN a list of N children
- WHEN the list mounts
- THEN child i fades in and slides up with delay i times stagger

#### Scenario: Empty or single-child list

- GIVEN a list with 0 or 1 children
- WHEN it mounts
- THEN no throw occurs
- AND the single child (if any) animates with delay 0

### Requirement: SkeletonBox

The system SHALL provide a `SkeletonBox` shimmer widget controlled by a loading-state flag.

#### Scenario: Loading shows shimmer

- GIVEN isLoading is true
- WHEN the widget builds
- THEN a shimmer placeholder is shown

#### Scenario: Loaded hides shimmer

- GIVEN isLoading is false
- WHEN the widget builds
- THEN no shimmer animation runs

### Requirement: Hero Transition for Routine Detail

The system SHALL use Hero or shared-element transitions between routine cards and `routine_detail_screen.dart`, keyed by stable routine ids.

#### Scenario: Open detail animates the shared element

- GIVEN a routine card with a stable Hero tag derived from its id
- WHEN the user taps the card
- THEN the shared element animates to the detail screen

#### Scenario: Tag collisions avoided

- GIVEN two routines with distinct ids
- WHEN both cards render on the same screen
- THEN no duplicate Hero tags exist

### Requirement: Animated Tab Crossfade

`MainShell` SHALL crossfade between tabs via `AnimatedSwitcher` while preserving each tab's state across transitions.

#### Scenario: Tab switch crossfades

- GIVEN the user is on tab A
- WHEN they select tab B
- THEN the content crossfades from A to B

#### Scenario: Tab state preserved

- GIVEN the user scrolled tab A, then switched away and back
- THEN tab A scroll position and state are intact

### Requirement: Animated Weekly Bars

Progress weekly bars SHALL animate from bottom upward using a spring curve on first appearance.

#### Scenario: Bars rise on entry

- GIVEN the Progress screen mounts
- WHEN bars render
- THEN each bar grows from 0 to its value via spring

#### Scenario: Reduced-motion snaps to value

- GIVEN reduced-motion is enabled
- WHEN bars render
- THEN they appear at full value instantly

### Requirement: Timer Pulse

The active-workout timer digits SHALL pulse subtly once per elapsed second.

#### Scenario: Digit pulses per tick

- GIVEN the timer is running
- WHEN each second elapses
- THEN the digits pulse subtly once

#### Scenario: Paused timer stops pulse

- GIVEN the timer is paused
- WHEN time does not advance
- THEN no pulse animation runs

### Requirement: Reduced-Motion Gate

The system SHALL expose a single centralized reduced-motion check honoring `MediaQuery.disableAnimations` and `accessibleNavigation`. When active, ALL animations MUST be instantaneous (0ms).

#### Scenario: Global disable

- GIVEN `accessibleNavigation` is true
- WHEN any motion component builds
- THEN its animation duration is 0ms

#### Scenario: Per-widget consistency

- GIVEN two motion components on the same screen
- WHEN reduced-motion is enabled
- THEN both render without animation simultaneously