package discordbot

type CommandError struct {
	Message string
}

func (c CommandError) Error() string {
	return c.Message
}

func NewCommandError(message string) error {
	return CommandError{Message: message}
}
