package def

type Gender uint32

const (
	GenderWoman Gender = 0 //女性
	GenderMan   Gender = 1 //男性
)

func (g Gender) IsValid() bool {
	switch g {
	case GenderWoman, GenderMan:
		return true
	default:
		return false
	}
}
