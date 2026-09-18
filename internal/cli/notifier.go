package cli

type Notifier interface {
	Send(title, body string) error
}
