# Random Events technology gate - direct MOO2 1.31 evidence (2026-08-28)

## Result

Original global byte `0x21CAF` is the persisted New Game **Random Events enabled** flag.

The earlier unresolved Technology 52 gate can therefore be named semantically:

```text
if Random Events are disabled:
    Dimensional Portal (Technology 52) is not available
```

MOOX no longer needs an opaque `0x21CAF` policy placeholder for this rule.

## Direct setup evidence

`Set_Default_Game_Settings_` at VA `0x127E1` writes:

```text
0x21CAF = 1
```

so Random Events are enabled by the original default settings.

The New Game screen copies this byte through its temporary setup value at `0x2935A`:

```text
Globals_To_Temps_: 0x21CAF -> 0x2935A
Temps_To_Globals_: 0x2935A -> 0x21CAF
```

The same setting is persisted/restored by:

- `Load_Game_`;
- `Copy_Over_Game_Universe_Settings_`;
- network game-data broadcast/decode paths.

Original UI resources independently identify the option:

- `HELP.LBX`: `Random Events Button`;
- help text: `Clicking here allows you to turn on and off the Random Events`;
- `ESTRINGS.LBX`: `No Random Events`.

## Event-system cross-check

`Next_Turn_Calc_` at VA `0x136B3` checks `0x21CAF` before calling:

```text
Antaran_Invasion_Check_ VA 0x63D92
```

`Determine_Lucky_Players_Events_` at VA `0x24511` also branches on `0x21CAF` while establishing event eligibility/counts.

This confirms that the setup flag is the Random Events switch rather than a Research-specific setting.

## Technology 52 cross-check

Technology 52 is normalized as:

```text
dimensional_portal
```

Three independent original technology paths gate it on the same byte:

1. `At_Least_One_App_Researchable_` at VA `0x5E3FF`;
2. `Init_Player_Tech_` at VA `0x5E55F`;
3. `Ensure_Uncreative_Field_OK_` at VA `0xE408F`.

`Player_May_Get_Tech_App_` at VA `0xE412B` provides the clearest direct mapping. For Technology ID `0x34` / decimal `52`, it returns the value of `0x21CAF` directly.

Therefore Technology 52 is globally ineligible when Random Events are disabled, including starting-technology selection and later external acquisition/repair paths.

## MOOX semantic mapping

MOOX should use semantic configuration rather than persist the original byte address.

For technology-generation helpers whose zero-value options historically represented the original default, the current integration uses:

```go
RandomEventsDisabled bool
```

with zero meaning Random Events remain enabled, matching original default behavior.

The Uncreative external-repair options use the same semantic gate. Technology 52 is excluded when `RandomEventsDisabled` is true.

A future top-level New Game settings model may expose the positive UI concept (`RandomEventsEnabled`) and translate it once at the domain boundary. No Wails/network/UI adapter should reimplement Technology eligibility.

## Remaining boundary

The general `Player_May_Get_Tech_App_` rule proves this gate applies beyond Uncreative repair. Future trade, espionage, conquest and scripted Technology-grant operations must therefore invoke the same server-authoritative Random Events eligibility policy before granting Technology 52.

Advanced-start generation is currently being implemented separately. Its default-game path remains Random-Events-enabled unless/until the full New Game settings object is threaded through that initializer.
