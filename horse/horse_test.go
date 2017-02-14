package horse

import "testing"

func TestCenter(t *testing.T) {
	res, err := Horse("d4")
	if (len(res) != 8) || (err != nil) {
		t.Errorf("Из центра 8 ходов а насчитали %d ", len(res))
	}
}

func TestBorder(t *testing.T) {
	res, err := Horse("a5")
	if (len(res) != 4) || (err != nil) {
		t.Errorf("От края 4 хода а насчитали %d ", len(res))
	}
}

func TestCorner(t *testing.T) {
	res, err := Horse("h1")
	if (len(res) != 2) || (err != nil) {
		t.Errorf("Из угла 2 хода а насчитали %d ", len(res))
	}
}

func TestIncorrect(t *testing.T) {
	res, err := Horse("z5")
	if (len(res) != 0) || (err == nil) {
		t.Errorf("Должен возвращать  ошибку вместо %d ", len(res))
	}
}
