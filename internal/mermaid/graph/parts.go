package graph

import "slices"

// addParts splits the edges into parts, and puts the parts that run
// sideways on tracks so that they don't run along each other.
func (l *layout) addParts() {
	for ei, path := range l.paths {
		for i := 1; i < len(path); i++ {
			from, to := &l.verts[path[i-1]], &l.verts[path[i]]
			p := part{
				edge: ei, from: path[i-1], to: path[i],
				first: i == 1, last: i == len(path)-1,
				fromX: from.cx, toX: to.cx, track: -1,
			}
			switch {
			case l.back[ei]:
				if !from.isPoint() {
					p.fromX -= l.orient.backX()
				}
				if !to.isPoint() {
					p.toX -= l.orient.backX()
				}
			// A part that would only jog by a column leaves its from box
			// straight above where it ends instead.
			case !from.isPoint() && abs(p.fromX-p.toX) <= 1 && p.toX > from.x && p.toX < from.x+from.w-1:
				p.fromX = p.toX
			}
			l.parts = append(l.parts, p)
		}
	}

	for r := range len(l.ranks) - 1 {
		var tracks [][]int // the parts on each track
		for i := range l.parts {
			p := &l.parts[i]
			if l.verts[p.from].rank != r || p.fromX == p.toX {
				continue
			}
			t := slices.IndexFunc(tracks, func(track []int) bool {
				return !slices.ContainsFunc(track, func(j int) bool { return clash(p, &l.parts[j]) })
			})
			if t < 0 {
				t = len(tracks)
				tracks = append(tracks, nil)
			}
			tracks[t] = append(tracks[t], i)
			p.track = t
		}
	}
}

// clash reports whether two parts running sideways can't share a track:
// they would run along or right next to each other, and don't leave or reach
// a vertex at the same column.
func clash(a, b *part) bool {
	if a.from == b.from && a.fromX == b.fromX || a.to == b.to && a.toX == b.toX {
		return false
	}
	aLo, aHi := min(a.fromX, a.toX), max(a.fromX, a.toX)
	bLo, bHi := min(b.fromX, b.toX), max(b.fromX, b.toX)
	return aLo <= bHi+1 && bLo <= aHi+1
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
