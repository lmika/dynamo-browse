package itemrenderer

type StyleRenderer interface {
	Render(str string) string
}

type plainTextStyleRenderer struct{}

func (plainTextStyleRenderer) Render(str string) string {
	return str
}
