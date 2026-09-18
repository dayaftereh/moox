# Slice 17.0 Gate 3 Amendment - Construction Buyout + Reference BC Tools

Date: 2026-09-18
Slice: 17.0
Gate: 3 amendment before Gate 4
Baseline: `78960f4` (`feat(slice): implement slice 17.0 reference harness`)
Session: `ses-20260918T152048-00240c450344`

## Why this amendment exists

The first Gate-3 harness exposed normal-turn acceleration as a prominent Reference Lab surface. Review showed that fast Ship/Construction QA is better when it also exercises the original MOO2 economic path:

1. prepare a real design/construction project;
2. give the reference empire development-only BC when needed;
3. buy the current construction through the normal player-facing economy rule;
4. complete it on the next ordinary turn.

Turn acceleration remains useful for research, fleet movement, population growth and normal production QA, but it belongs under Development Tools rather than in the main strategic surface.

## Original MOO2 evidence frozen for buyout

Primary calculation reference:

- StrategyWiki, "Master of Orion II: Battle at Antares/Calculations", section "Cost of buying production points":
  https://strategywiki.org/wiki/Master_of_Orion_II%3A_Battle_at_Antares/Calculations

Let:

- `X` = total Production Point cost of the current item;
- `Y` = Production Points already produced.

Frozen cost curve:

| Completion | BC cost |
| --- | ---: |
| 0-10% | `4*X - 10*Y` |
| 10-50% | `3.5*X - 5*Y` |
| 50-100% | `2*X - 2*Y` |

Equivalent marginal interpretation:

- first 10% not yet produced: 10 BC per PP;
- production needed from 10% to 50%: 5 BC per PP;
- production needed from 50% to 100%: 2 BC per PP.

Corroborating reference:

- GameFAQs MOO2 strategy guide:
  https://gamefaqs.gamespot.com/pc/197873-master-of-orion-ii-battle-at-antares/faqs/72470

Completion timing reference:

- StrategyWiki, "Speeding up production":
  https://strategywiki.org/wiki/Master_of_Orion_II%3A_Battle_at_Antares/Speeding_up_production

That reference explicitly describes bought construction as completing when the player proceeds with the Turn. The MOOX implementation therefore does **not** instant-complete the project inside the Buy command.

## Normal gameplay construction-buyout contract

The normal gameplay command is:

`colony.buy_construction`

This is not a reference-only cheat.

Server authority:

- Planning phase only;
- authoritative seat -> empire ownership;
- current active finite construction only;
- Housing is excluded because it has no finite completion target;
- exact current construction cost comes from the existing authoritative `constructionProjectCostPP` path;
- exact progress comes from the authoritative Colony state;
- exact BC cost is calculated server-side with the frozen original curve;
- the owning Empire Treasury must be able to pay;
- BC is deducted server-side;
- the current project's `ProgressPP` is set to its full Production cost;
- the project remains active until the next ordinary strategic turn resolves it;
- normal completion events, queue advancement, built-Ship snapshotting and Fleet creation remain unchanged.

The client never sends a price.

`PlayerSnapshot.construction_buyouts` projects an authoritative quote containing project identity, Production cost/progress, remaining PP, BC cost and affordability.

The normal Colony UI exposes **Buy / Kaufen** for the current authoritative project in both the Colony summary and Build-management view.

If the current construction is already fully funded by a prior Buy, the UI shows that it is bought and awaits the next normal turn rather than offering another purchase.

A drafted construction replacement suppresses the Buy action for the stale current project until that draft is resolved/reset.

## Development Tools contract

The Reference controls now live under the existing More/Advanced development area rather than above every strategic screen.

Trusted reference games expose:

### Turn controls

- +1 Turn;
- +5 Turns;
- +N Turns, bounded by the existing 1..25 server contract;
- Advance until construction completes, with an explicit active-construction Colony selector.

### Economy controls

Fixed BC grants only:

- +100 BC;
- +1,000 BC;
- +10,000 BC.

BC grants:

- are available only when the process has Reference controls enabled;
- additionally require trusted immutable per-game Reference metadata;
- require the registered control seat;
- run at a Planning boundary;
- mutate the authoritative Empire Treasury server-side;
- publish normal snapshot invalidation;
- cannot be invoked against ordinary games;
- accept no arbitrary user-supplied amount outside the three frozen values.

The grants are development tools. The Buy command is normal gameplay.

## Shared normal gameplay path

A reference QA loop can now be:

`All-Tech Triangle -> real Ship design -> real Colony build queue -> Dev +10,000 BC -> normal Buy -> normal next Turn -> normal built Ship/Fleet snapshot`

This exercises more production/economy authority than simply skipping hundreds of turns and avoids creating a special "complete this build" cheat path.

The existing "Advance until construction completes" remains as a secondary QA tool for testing ordinary PP accumulation without Buy.

## Verification

Focused Go tests cover:

- original buyout cost curve at 0%, 10%, 30%, 50%, 75% and 100%;
- BC deduction;
- insufficient-BC rejection without mutation;
- Housing exclusion;
- bought project remains active until normal turn;
- next normal turn emits normal construction completion;
- Host immediate-command dispatch;
- authoritative quote projection;
- Reference +100/+1,000/+10,000 BC authorization;
- ordinary-game rejection;
- invalid arbitrary grant rejection;
- Reference route disabled -> 404;
- trusted route -> authoritative Treasury mutation.

Broad verification:

- `go test ./...` - PASS;
- `go vet ./...` - PASS;
- `npm run build` - PASS.

Headless Chrome smoke verifies the complete user workflow:

- Development Tools appear only under More for trusted Reference games;
- +1, +5 and +N are present;
- +100, +1,000 and +10,000 BC are present;
- +10,000 BC changes authoritative Treasury;
- a real legal building is queued through normal turn submission;
- server projects a positive buyout quote;
- normal Colony UI shows Kaufen at 390px;
- clicking Kaufen deducts exactly the quoted BC;
- project becomes fully funded but is not instant-completed;
- +1 normal turn completes the bought project;
- desktop remains responsive;
- ordinary `game-1` has no development controls.

## Gate status

This is an explicit Gate-3 amendment.

Gate 3 remains complete after this amendment and Gate 4 is still the next step.

Slice 17.1 remains closed.
