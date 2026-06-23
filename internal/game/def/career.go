package def

type Career uint32

const (
	CareerWarrior Career = 1 //战士
	CareerMage    Career = 2 //法师
	CareerArcher  Career = 3 //弓箭手
	CareerPriest  Career = 4 //牧师
)

func (c Career) IsValid() bool {
	switch c {
	case CareerWarrior, CareerMage, CareerArcher, CareerPriest:
		return true
	}
	return false
}
