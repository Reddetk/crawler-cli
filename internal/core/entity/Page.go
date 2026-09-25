package entity

type Page struct {
	Resource string
	Title    string
	Links    []*Page
}

func NewPage() *Page {
	return &Page{}
}
