package sync

type op struct {
	repl    []string
	changed bool
}

func diffOps(base, side []string) map[int]op {
	ops := map[int]op{}

	lcs := lcsMatrix(base, side)
	i, j := 0, 0
	pending := []string{}

	for i < len(base) && j < len(side) {
		if base[i] == side[j] {
			if len(pending) > 0 {
				ops[i] = op{repl: append(pending, base[i]), changed: true}
				pending = nil
			}
			i++
			j++
		} else if lcs[i+1][j] >= lcs[i][j+1] {
			ops[i] = op{repl: pending, changed: true}
			pending = nil
			i++
		} else {
			pending = append(pending, side[j])
			j++
		}
	}
	for i < len(base) {
		ops[i] = op{repl: pending, changed: true}
		pending = nil
		i++
	}
	if j < len(side) {
		tail := append(pending, side[j:]...)
		ops[len(base)] = op{repl: appendExisting(ops[len(base)].repl, tail), changed: true}
	}

	return ops
}

func appendExisting(a, b []string) []string {
	return append(a, b...)
}

func lcsMatrix(a, b []string) [][]int {
	m := make([][]int, len(a)+1)
	for i := range m {
		m[i] = make([]int, len(b)+1)
	}
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				m[i][j] = m[i+1][j+1] + 1
			} else if m[i+1][j] >= m[i][j+1] {
				m[i][j] = m[i+1][j]
			} else {
				m[i][j] = m[i][j+1]
			}
		}
	}
	return m
}

func changeAt(ops map[int]op, i int) (bool, []string) {
	o, ok := ops[i]
	if !ok {
		return false, nil
	}
	return o.changed, o.repl
}

func hasInsertAt(ops map[int]op, i int) bool {
	o, ok := ops[i]
	return ok && len(o.repl) > 0
}
