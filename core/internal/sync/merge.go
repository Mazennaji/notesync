package sync

import "strings"

type MergeResult struct {
	Merged   string
	Conflict bool
}

func Merge(base, local, remote string) MergeResult {
	baseLines := splitLines(base)
	localLines := splitLines(local)
	remoteLines := splitLines(remote)

	localOps := diffOps(baseLines, localLines)
	remoteOps := diffOps(baseLines, remoteLines)

	var out []string
	conflict := false

	bi := 0
	for bi < len(baseLines) || hasInsertAt(localOps, bi) || hasInsertAt(remoteOps, bi) {
		lChanged, lRepl := changeAt(localOps, bi)
		rChanged, rRepl := changeAt(remoteOps, bi)

		switch {
		case !lChanged && !rChanged:
			if bi < len(baseLines) {
				out = append(out, baseLines[bi])
			}
			bi++
		case lChanged && !rChanged:
			out = append(out, lRepl...)
			bi++
		case !lChanged && rChanged:
			out = append(out, rRepl...)
			bi++
		default:
			if equalLines(lRepl, rRepl) {
				out = append(out, lRepl...)
			} else {
				conflict = true
				out = append(out, "<<<<<<< LOCAL")
				out = append(out, lRepl...)
				out = append(out, "=======")
				out = append(out, rRepl...)
				out = append(out, ">>>>>>> REMOTE")
			}
			bi++
		}
	}

	return MergeResult{Merged: strings.Join(out, "\n"), Conflict: conflict}
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func equalLines(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
