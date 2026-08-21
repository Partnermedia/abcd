package livery

// palette maps a grid pixel key to its hex color. Terminal surfaces map the
// same keys to ANSI colors (itd-112); changing a value here restyles every
// rendered artifact at once.
var palette = map[rune]string{
	'o': "#d97757", // terracotta — duckling beak, lifeboat hull
	'k': "#1c1f26", // dark — eyes and portholes; matches the panel color
	'y': "#f0c052", // yellow — duckling body, delta flag
	'w': "#e8e8e8", // white — sails, flag whites
	'r': "#cf554d", // red — bravo flag, charlie stripe
	'b': "#4a6fb5", // blue — alfa/charlie/delta flag blues
	'c': "#2b7a8c", // water
	'm': "#9aa0a8", // mast gray
}

// The canonical grids. The full-size logo carries true ICS geometry — the
// flag geometry test holds it to the specification — and the compact logo is
// declared approximate (charlie's five stripes cannot survive three rows).
var assets = []Asset{
	{
		Name: "duckling",
		Grid: []string{
			"...yyyy.....",
			"..yyyyyy....",
			"..ykyyyy....",
			".oyyyyyy....",
			"..yyyyyyyy..",
			"..yyyyyyyyy.",
			"...yyyyyyy..",
			"cccccccccccc",
		},
	},
	{
		Name: "duckling-mini",
		Grid: []string{
			".yyy....",
			".ykyy...",
			"oyyyyyy.",
			".yyyyy..",
			"cccccccc",
		},
	},
	{
		Name: "logo-flags",
		Grid: []string{
			"wwwbb.rrrrr.bbbbb.yyyyy",
			"wwwb..rrrr..wwwww.bbbbb",
			"www...rrr...rrrrr.bbbbb",
			"wwwb..rrrr..wwwww.bbbbb",
			"wwwbb.rrrrr.bbbbb.yyyyy",
		},
	},
	{
		// The icon arrangement: the same four true-geometry flags, two per
		// row (alfa bravo / charlie delta). The 11x11 grid is naturally
		// square — the app/web icon candidate.
		Name: "logo-flags-icon",
		Grid: []string{
			"wwwbb.rrrrr",
			"wwwb..rrrr.",
			"www...rrr..",
			"wwwb..rrrr.",
			"wwwbb.rrrrr",
			"...........",
			"bbbbb.yyyyy",
			"wwwww.bbbbb",
			"rrrrr.bbbbb",
			"wwwww.bbbbb",
			"bbbbb.yyyyy",
		},
	},
	{
		Name: "logo-flags-compact",
		Grid: []string{
			"wbb.rrr.bbb.yyy",
			"wb..rr..rrr.bbb",
			"wbb.rrr.bbb.yyy",
		},
		ApproximateGeometry: true,
	},
	{
		Name: "lifeboat",
		Grid: []string{
			"......m......",
			"......mww....",
			"......mwww...",
			"......mwwww..",
			"......m......",
			".oookookoooo.",
			"..ooooooooo..",
			"ccccccccccccc",
		},
	},
	{
		Name: "lifeboat-mini",
		Grid: []string{
			"....m....",
			"....mww..",
			"....mwww.",
			".ookokoo.",
			"..ooooo..",
			"ccccccccc",
		},
	},
}
