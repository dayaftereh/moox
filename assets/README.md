# MOOX game assets

This directory is reserved for **MOOX-owned or explicitly licensed distributable assets**.

Original Master of Orion II artwork must not be placed here. Original/reference material stays local below `reference/original/` and is ignored by Git.

Planned asset convention:

```text
assets/
  races/
  planets/
  ships/
  buildings/
  technologies/
  ui/
```

Runtime data should refer to stable semantic asset IDs rather than source filenames, for example:

```text
race.alkari.portrait
planet.terran.background
technology.research_lab.icon
ship.hull.cruiser.icon
```

A generated or hand-created asset can therefore be replaced later without touching game rules, localization or save-game data.

The intended workflow is:

1. inspect the original local reference,
2. identify its semantic role,
3. write an art brief/specification,
4. create a new MOOX visual inspired by the functional requirements rather than copying pixels,
5. review it side-by-side with the private reference,
6. commit only the new distributable MOOX asset.
