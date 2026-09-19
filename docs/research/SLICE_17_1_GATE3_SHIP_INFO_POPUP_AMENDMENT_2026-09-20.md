# Slice 17.1 Gate 3 - Ship Information / Popup HMI Amendment

Date: 2026-09-20
Slice: 17.1 - Hulls and Mandatory Ship Systems
Gate: 3 amendment before Gate 4
Baseline entering amendment: `990739e` (`feat(slice): clarify designer system effects`)

## Reason

Desktop review showed that the previous effect-first HMI was informative but still consumed too much permanent screen space: Computer and Shield were rendered as two large always-visible choice cards.

The user requested a more compact and game-like interaction:

- one complete Ship information card directly under the Ship preview;
- installed Drive, Computer, Armor, Shield and Fuel visible there;
- the current gameplay effect visible directly beside each installed system;
- a dedicated `?` detail affordance on every mandatory-system row;
- Computer and Shield changed through a rich popup/card selector rather than a dropdown or permanently expanded choice list.

## Final HMI contract

### Ship information card

Directly below the procedural Ship image, one `Ship information` card contains:

- hull;
- current authoritative Production Cost;
- Command Point cost;
- used / total design Space;
- Warp Drive;
- Computer;
- Armor;
- Shield;
- Fuel Cell;
- installed weapon mounts when present.

Each mandatory-system row shows:

1. system category;
2. human-readable installed component name;
3. the current verified gameplay/original effect;
4. a trailing `?` detail control.

### Computer

The installed Computer row is selectable.

Clicking the row opens a modal/card selector showing all Computer technologies with:

- component name;
- verified Beam Attack bonus;
- current hull-specific miniaturized PP cost;
- miniaturization level;
- current selected state;
- locked state where technology is unavailable.

Choosing a legal card immediately updates the design preview and closes the popup.

The row's separate `?` opens detail-only information for the installed Computer and can transition into the selector via `Change`.

### Shield

The installed Shield row is selectable.

Clicking the row opens a modal/card selector showing:

- explicit no-Shield as a first-class option;
- each known/locked Shield;
- verified per-hit damage block;
- **current actual capacity for the selected hull class** plus the original capacity-per-class factor;
- current miniaturized PP;
- current miniaturized Space;
- miniaturization level;
- selected / locked state.

Choosing a card immediately updates the preview and closes the popup.

The row's `?` opens detail-only information for the installed Shield and can transition to the selector.

### Drive / Armor / Fuel

These remain automatic per the frozen Gate-2 contract and are not falsely made user-selectable.

Their rows show:

- Drive: strategic FTL plus verified original tactical speed contribution;
- Armor: verified original durability multiplier;
- Fuel: strategic range or unlimited range.

Each has a `?` detail popup and an explicit `automatic from Empire technology` label.

### Tactical scope honesty

The detail popup retains the existing scope note:

- currently supported strategic effects are active;
- verified original Computer/Armor/Shield/broader Drive tactical values are displayed;
- unsupported Tactical behavior is not silently simulated.

## Responsive behavior

Desktop:

- one compact Ship information card remains under the preview;
- Computer/Shield popup choice cards render in two readable columns;
- the dialog is centered.

Mobile / 390 px:

- summary grid reduces to two columns;
- system rows remain compact;
- popup becomes a bottom-sheet-style dialog;
- choice cards render in one column;
- no horizontal overflow.

## Browser acceptance

The persistent browser smoke was updated to verify:

- one Ship information card exists directly in the Designer;
- exactly five mandatory-system rows are present;
- every row has a `?` help affordance;
- hull / Production / CP / Space summary is in that card;
- no mandatory-system dropdown exists;
- Drive, Computer, Armor, Shield and Fuel effect text is visible;
- Computer row opens the popup card selector;
- Electronic and Moleculartronic Computer cards expose +25 / +125 Beam Attack;
- Fuel `?` opens a detail popup and explains unlimited range in All-Tech;
- Doom Star can be edited through the Computer popup and Shield popup;
- explicit no-Shield remains saveable;
- desktop popup choice cards render in two columns;
- mobile and desktop remain overflow-free;
- the existing Development Tools / BC / buyout workflow still passes afterward.

Fresh reference server used for the amendment browser run: loopback port 7192.

## Verification

- `npm run build` - PASS
- `npm run check:reference-lab:browser -- http://127.0.0.1:7192` - PASS
- `go test ./...` - PASS
- `go vet ./...` - PASS
- `git diff --check` - PASS before staging

No backend game-rule contract was changed by this amendment.

Gate 4 remains not started. Slice 17.2+ remain closed.
