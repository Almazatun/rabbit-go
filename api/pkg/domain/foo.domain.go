package domain

type Foo struct{}

func (b *Foo) Create(mess string) string {
	return "Foo created"
}

func (b *Foo) Update(mess string) string {
	return "Foo updated"
}
