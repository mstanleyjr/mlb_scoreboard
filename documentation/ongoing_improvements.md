# Ongoing Improvements

## Division standings stats row overflow

- **Problem:** On the standings page, worst-case combinations like a `42-120` record with `12.5` games back can overlap on the record/GB row.
- **Preferred future fix:** Keep the current team-name row as-is, but redesign the stats line so the record and GB are treated as separate zones rather than competing for the same horizontal space.
- **Recommended direction:** Option 6 from discussion — preserve the current two-line team block, but give the second line a more explicit left/right layout for `record` and `GB` so long values remain readable without immediately introducing a new font.
- **Why deferred:** It is the best practical answer, but not something we want to implement yet.
