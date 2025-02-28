package team

type Person struct {
	Name    string
	Age     int
	IsAlive bool
}

func NewPerson(name string, age int) *Person {
	return &Person{
		Name:    name,
		Age:     age,
		IsAlive: true,
	}
}

func IncrementAge(p *Person) {
	p.Age++
}

type LeagueType int

const (
	League1 = iota + 1
	League2
	PremierLeague
)

type Team struct {
	Name    string
	League  LeagueType
	Captain *Person
}

func NewTeam(name string, league LeagueType, captain *Person) *Team {
	return &Team{
		Name:    name,
		League:  league,
		Captain: captain,
	}
}

func ChangeCaptain(t *Team, captain *Person) {
	t.Captain = captain
}
