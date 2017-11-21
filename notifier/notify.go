package notifier

const (
	RedBackground    = "red"
	YellowBackground = "yellow"
	GreenBackground  = "green"
	GrayBackground   = "gray"
	RandomBackground = "random"
	PurpleBackground = "purple"
)

type Notifier interface {
	Notify(msg, color string) error
}
