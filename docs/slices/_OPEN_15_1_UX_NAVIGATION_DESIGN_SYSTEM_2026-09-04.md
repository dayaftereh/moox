# OPEN - Slice 15.1 UX architecture / navigation / design system

Opened: **2026-09-04**

Status: **Gates 1-3 complete; Gate 4 independent QA pending**.

Primary direction: **mobile-first**, with desktop as a required adaptive layout mode. **Multi-language is mandatory from this slice onward: German + English with runtime switching and persistent client-side preference.**

Permanent evidence:
- Gate 1: `docs/research/SLICE_15_1_MOBILE_FIRST_UX_GATE1_2026-09-04.md`
- Gate 2: `docs/research/SLICE_15_1_MOBILE_FIRST_UX_GATE2_2026-09-04.md`
- Gate 3: `docs/research/SLICE_15_1_MOBILE_FIRST_UX_GATE3_2026-09-04.md`

- [x] Gate 1: UX/navigation/design-system audit.
- [x] Gate 2: freeze interaction/design-system + German/English i18n contract.
- [x] Gate 3: implement mobile-first shell/navigation/primitives + DE/EN i18n.
- [ ] Gate 4: independent QA, live NetBird phone preview, commit and close.

Post-implementation preview requirement:
- build and serve the production UI from `moox-server`;
- bind the development server to `0.0.0.0:7171` with the explicit non-loopback development flag;
- use port `7171` as the stable review port;
- keep an agent-managed Chrome `moox-ui-review` session attached to the live UI for desktop testing;
- report the exact `http://<netbird-ip>:7171/` URL for phone review;
- if remote access is blocked, add a targeted Windows Firewall inbound rule for TCP 7171.
