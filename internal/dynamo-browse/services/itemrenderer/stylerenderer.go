package itemrenderer

type StyleRenderer interface {
	Render(str string) string
}

func PlainTextRenderer() StyleRenderer {
	return plainTextStyleRenderer{}
}

type plainTextStyleRenderer struct{}

func (plainTextStyleRenderer) Render(str string) string {
	return str
}
