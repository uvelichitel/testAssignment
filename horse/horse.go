package horse

import "errors"

func Horse(pos string) ([]string, error) {
	if len(pos) == 2 {
		if x := pos[0]; x > 96 && x < 109 {
			if y := pos[1]; y > 48 && y < 57 {

				var res []string
				if u1 := x + 1; u1 < 105 {
					if u2 := x + 2; u2 < 105 {
						if r1 := y + 1; r1 < 57 {
							res = append(res, string([]byte{u2, r1}))
						}
						if l1 := y - 1; l1 > 48 {
							res = append(res, string([]byte{u2, l1}))
						}
					}

					if r2 := y + 2; r2 < 57 {
						res = append(res, string([]byte{u1, r2}))
					}

					if l2 := y - 2; l2 > 48 {
						res = append(res, string([]byte{u1, l2}))
					}
				}
				if d1 := x - 1; d1 > 96 {
					if d2 := x - 2; d2 > 96 {
						if r1 := y + 1; r1 < 57 {
							res = append(res, string([]byte{d2, r1}))
						}
						if l1 := y - 1; l1 > 48 {
							res = append(res, string([]byte{d2, l1}))
						}
					}

					if r2 := y + 2; r2 < 57 {
						res = append(res, string([]byte{d1, r2}))
					}

					if l2 := y - 2; l2 > 48 {
						res = append(res, string([]byte{d1, l2}))
					}

				}

				return res, nil
			}
		}
	}
	return nil, errors.New("Таких коней не бывает")
}
