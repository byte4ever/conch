package observer

// Observable defines required contract with observable object.
type Observable interface {
	// Register an observer.
	Register(o Observer) error

	// RegisterChan registers  a channel as observer.
	// This mean observer is using a private channel to listen to observable.
	RegisterChan(o Observer, c chan interface{}) error

	// Unregister an observer. For channel observers this close the listening channel.
	Unregister(o Observer) error

	// Notify all observers with a specific value.
	Notify(i interface{})
}

// Observer defines required contract with observer object.
type Observer interface {
	// OnNotify observer callback.
	OnNotify(i interface{})
}
