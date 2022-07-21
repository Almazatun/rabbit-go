package domain

type Boo struct{}

func (b *Boo) Create(mess string) string {
	return "Boo created"
}

func (b *Boo) Update(mess string) string {
	return "Boo updated"
}
