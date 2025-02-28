package entity

type Person struct {
	ID         int
	Name       string
	PersonType PersonType
	Sources    []string
}

type PersonType string

const (
	PersonTypeUnknown   PersonType = "Unknown"
	PersonTypeActor     PersonType = "Actor"
	PersonTypeDirector  PersonType = "Director"
	PersonTypeWriter    PersonType = "Writer"
	PersonTypeGuestStar PersonType = "GuestStar"
	PersonTypeProducer  PersonType = "Producer"
	PersonTypeCreator   PersonType = "Creator"
	PersonTypeEditor    PersonType = "Editor"
)

// IsValidPersonType checks if a PersonType is valid
func IsValidPersonType(pk PersonType) bool {
	switch pk {
	case PersonTypeUnknown, PersonTypeActor, PersonTypeDirector,
		PersonTypeWriter, PersonTypeGuestStar, PersonTypeProducer,
		PersonTypeCreator, PersonTypeEditor:
		return true
	}
	return false
}
