// Package yearturn builds November-to-February windows for cross-year calendar testing.
//
// A four-month window combines 30-day November, 31-day December, a year
// rollover, 31-day January, and February's leap-year variation. The same
// regression window works for any supported starting year.
//
// For visual comparison with generated grids, see the human-readable reference
// window: [November 2026], [December 2026], [January 2027], and [February 2027].
//
// [November 2026]: https://www.betacalendars.com/november-calendar.html
// [December 2026]: https://www.betacalendars.com/december-calendar.html
// [January 2027]: https://www.betacalendars.com/january-calendar.html
// [February 2027]: https://www.betacalendars.com/february-calendar.html
package yearturn
