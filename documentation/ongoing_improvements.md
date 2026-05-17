# Ongoing Improvements

## Division standings stats row overflow

- **Problem:** On the standings page, worst-case combinations like a `42-120` record with `12.5` games back can overlap on the record/GB row.
- **Preferred future fix:** Keep the current team-name row as-is, then split the stats into an additional line so `GB` and `WCGB` get their own dedicated row instead of competing with the record.
- **Recommended direction:** Preserve the current team-name row, keep the record on its own line, and add a third line for `GB` / `WCGB`. That should eliminate the current overflow pressure while also making wild-card context available in the same block.
- **Why deferred:** It is the best practical answer, but not something we want to implement yet.

## Division standings scroll smoothness

- **Problem:** The vertical standings rollup can feel choppy even at the slower overall pace.
- **Likely cause:** The scroll currently advances in integer pixel steps on a `100ms` frame interval, so motion still reads as discrete jumps instead of a smoother crawl.
- **Preferred future fix:** Increase the refresh cadence and retune the scroll curve so the offset updates more frequently, likely with a gentler easing profile at the start/end.
- **Why deferred:** The current behavior is acceptable, but motion polish would benefit from a focused pass rather than piecemeal tweaks.
