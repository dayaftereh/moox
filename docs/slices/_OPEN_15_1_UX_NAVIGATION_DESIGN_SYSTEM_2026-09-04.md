# OPEN - Slice 15.1 UX architecture / navigation / design system

Opened: **2026-09-04**

Status: **Gate 1 audit complete; Gate 2 interaction/design-system freeze pending user approval**.

Primary direction: **mobile-first**, with desktop as a required adaptive layout mode.

Permanent Gate-1 evidence:
`docs/research/SLICE_15_1_MOBILE_FIRST_UX_GATE1_2026-09-04.md`

- [x] Gate 1: UX/navigation/design-system audit.
- [ ] Gate 2: freeze interaction/design-system contract.
- [ ] Gate 3: implement mobile-first shell/navigation/primitives.
- [ ] Gate 4: independent QA, live NetBird phone preview, commit and close.

Post-implementation preview requirement:
- build and serve the production UI from `moox-server`;
- bind the development server to `0.0.0.0:7171` with the explicit non-loopback development flag;
- use port `7171` as the stable review port;
- keep an agent-managed Chrome `moox-ui-review` session attached to the live UI for desktop testing;
- report the exact `http://<netbird-ip>:7171/` URL for phone review;
- if remote access is blocked, add a targeted Windows Firewall inbound rule for TCP 7171.
