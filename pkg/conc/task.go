package conc

type Task struct {
	From string
	Data string
}

func WrapTask(from string, data string) Task {
	return Task{
		From: from,
		Data: data,
	}
}
