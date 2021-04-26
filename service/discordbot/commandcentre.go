package discordbot

import (
	"encoding/csv"
	"errors"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"strings"
)

type Command struct {
	opts    []string
	current int
	Message *discordgo.MessageCreate
}

func NewCommand(
	cmdString string,
	m *discordgo.MessageCreate) (*Command, error) {

	var command = &Command{
		Message: m,
	}

	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	fields, err := r.Read()
	if err != nil {
		return nil, err
	}

	for _, field := range fields {
		command.opts = append(command.opts, field)
	}

	return command, nil

}

func (c *Command) HasNext() bool {
	return c.current < len(c.opts)
}

func (c *Command) Next() string {
	var result string
	if c.current < len(c.opts) {
		result = c.opts[c.current]
		c.current++
	}
	return result
}

type CommandHandler func(Command) error

type CommandCenter struct {
	session        *discordgo.Session
	commandHandler map[string]CommandHandler
	botPrefix      string
}

func (c *CommandCenter) MustRegister(prefix string, handler CommandHandler) {

	if _, ok := c.commandHandler[prefix]; ok {
		panic("command already registered")
	}

	c.commandHandler[prefix] = handler
}

func (c *CommandCenter) Start() {

	c.session.AddHandler(c.botStarted)
	c.session.AddHandler(c.eventLoop)
	c.session.Open()

}

func (c *CommandCenter) eventLoop(s *discordgo.Session,
	m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	fullCommand, err := NewCommand(m.Content, m)
	if err != nil {
		log.Error(err)
		return
	}

	if !fullCommand.HasNext() || fullCommand.Next() != c.botPrefix {
		return
	}

	if !fullCommand.HasNext() {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Uh, You did not specify any command.")
		return
	}

	handler, ok := c.commandHandler[fullCommand.Next()]
	if !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Sorry, I do not understand that command. 🤔")
		return
	}

	if err := handler(*fullCommand); err != nil {

		if errors.As(err, &CommandError{}) {
			_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		_, _ = s.ChannelMessageSend(m.ChannelID, "Something unexpected happened while processing that command 😐")

		log.Error("error processing command", err)
	}
}

func (c *CommandCenter) botStarted(s *discordgo.Session,
	r *discordgo.Ready) {
	log.Info("Bot is up!")
}

func NewCommandCentre(botPrefix string, session *discordgo.Session) *CommandCenter {
	return &CommandCenter{
		botPrefix:      botPrefix,
		session:        session,
		commandHandler: map[string]CommandHandler{},
	}
}
