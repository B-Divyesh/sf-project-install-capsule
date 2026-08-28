# Visual thesis: The sealed line

Project Install Capsule uses an **art-deco transit poster** language because its job is routing: a project travels through a narrow, inspected corridor instead of roaming across the operator's machine. Concentric gates, rail geometry, ticket punches, and a single lit destination make the security model legible rather than decorative.

## Palette

The site is intentionally single-mode, painted as a night-platform poster. `Ink` #091B20 is the background; `Deep rail` #102D32 and `Panel` #163B3C are surfaces; `Paper` #F6E7C8 is primary text; `Ticket` #D7C59A is muted text; `Signal amber` #F2A93B is the action/accent; `Oxide` #D85B3F is warning/danger; `Clear signal` #72C49A is success. Text and UI pairings meet WCAG AA; amber is never used as body copy on paper.

## Type and spacing

Headlines use the self-hosted **League Spartan** variable font (SIL OFL 1.1), whose geometric capitals echo 1930s transport lettering. Body and code use the system sans/monospace stacks to keep the CLI documentary and fast. The type scale is 16, 18, 24, 34, 52, and 76 px. Layout follows an 8 px base rhythm with 4 px detail spacing; readable copy stays under 68 characters.

## Composition and interaction

The repeating octagonal “seal” is the identity device. Hairline rails connect capability stops: filesystem, network, ports, receipt. Buttons resemble punched tickets with clipped corners, but remain conventional links and buttons with 44 px minimum targets. The capability composer is a working, local-only preview: editing the install command, allowed hosts, and ports redraws the review block immediately; invalid input produces specific inline guidance.

## Depth and motion

On entry, poster layers rise 12 px into register over 220–420 ms. The moving rail signal is limited to one three-second pass, never loops, and only uses opacity/transform. In `prefers-reduced-motion`, all transforms and transitions are removed and the final registered state is painted immediately.

## Responsive intent

Desktop presents the poster and CLI review side by side. At 390 px, ornament is simplified, capability stops become a vertical route, and install instructions stack; nothing essential is removed. Navigation wraps and code samples scroll without forcing page zoom.

## Asset plan and provenance

- `site/public/capsule-poster.webp`: original raster hero generated for this product with `/opt/fleet/lib/gen-image.sh` (factory-image deployment), 2026-08-28. Prompt: “Art-deco transit poster illustration for a developer security tool; an ivory project crate travels through concentric teal and brass inspection gates toward a small amber terminal light; strict geometric screenprint, subtle paper grain, deep midnight green ground, no people, no logos, no lettering, no gradients, wide landscape composition with calm negative space.” The generated source is converted locally to WebP and kept below 300 KB. No third-party artwork.
- Seal, rail, ticket-notch, and capability icons are original CSS/SVG geometry drawn in the repository; no icon library.

## Why it fits

Transit posters promise a route that is understandable at a glance. Here, the route is a threat model: project in, explicit capabilities, process out, receipt left behind. Warm paper and brass keep the product approachable; deep teal prevents the security story from becoming alarmist.
