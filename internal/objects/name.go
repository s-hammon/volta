package objects

type Name struct {
	Last   string
	First  string
	Middle string
	Suffix string
	Prefix string
	Degree string
}

func NewName(last, first, middle, suffix, prefix, degree string) Name {
	return Name{
		Last:   last,
		First:  first,
		Middle: middle,
		Suffix: suffix,
		Prefix: prefix,
		Degree: degree,
	}
}

func (n *Name) Full() string {
	return n.Prefix + " " + n.First + " " + n.Middle + " " + n.Last + " " + n.Suffix + " " + n.Degree
}

// Returns a string representation of the name in format "Last, First Middle"
func (n *Name) Record() string {
	return n.Last + ", " + n.First + " " + n.Middle
}
