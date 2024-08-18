package forms

type errors map[string][]string

func (e errors) AppendError(value, message string) {
	e[value] = append(e[value], message)
}

func (e errors) GetError(value string) string {
	if len(e[value]) == 0 {
		return ""
	} else {
		return e[value][0]
	}
}
