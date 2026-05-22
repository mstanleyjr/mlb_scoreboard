# Ongoing Improvements

## Division standings stats row overflow

- **Problem:** On the standings page, worst-case combinations like a `42-120` record with `12.5` games back can overlap on the record/GB row.
- **Preferred future fix:** Keep the current team-name row as-is, but redesign the stats line so the record and GB are treated as separate zones rather than competing for the same horizontal space.
- **Recommended direction:** Option 6 from discussion — preserve the current two-line team block, but give the second line a more explicit left/right layout for `record` and `GB` so long values remain readable without immediately introducing a new font.
- **Why deferred:** It is the best practical answer, but not something we want to implement yet.

## Live game last-play notation

- **Problem:** The current `LAST PLAY` slide is clearer now, but it still feels visually sparse compared with the rest of the live-game page.
- **Preferred future fix:** Turn the last-play panel into a small scoreboard-notation view instead of plain text only.
- **Recommended direction:** Add a gray basepath diamond and highlight the relevant path segments from the scoring notation. Example: for a double, light up the first-base and second-base path lines while still showing the notation text.
- **Why deferred:** It needs a small design pass of its own, and it should be handled separately from the current readability/layout work.

## Last matchup record slide overflow

- **Problem:** On the last-matchup records slide, late-season records can reach three digits in wins or losses and make the left/right record layout feel cramped.
- **Preferred future fix:** Keep the slide structure, but switch to a more compact record rendering when either side gets too wide.
- **Recommended direction:** If a record reaches the current tight-fit threshold, replace `100-62` style text with a more compact format on that slide only instead of waiting for it to collide visually.
- **Why deferred:** It has not broken yet, but it is a predictable late-season edge case and should be handled before it does.
