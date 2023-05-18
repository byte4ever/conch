package conch

type (
	Logger interface {
		Trace(
			msg string,
			fields ...map[string]interface{},
		)
		Debug(
			msg string,
			fields ...map[string]interface{},
		)
		Info(
			msg string,
			fields ...map[string]interface{},
		)
		Warn(
			msg string,
			fields ...map[string]interface{},
		)
		Error(
			msg string,
			fields ...map[string]interface{},
		)
	}

	NopeLogger struct{}
)

var logger Logger = NopeLogger{}

func (n NopeLogger) Trace(string, ...map[string]interface{}) {
}

func (n NopeLogger) Debug(string, ...map[string]interface{}) {
}

func (n NopeLogger) Info(string, ...map[string]interface{}) {
}

func (n NopeLogger) Warn(string, ...map[string]interface{}) {
}

func (n NopeLogger) Error(string, ...map[string]interface{}) {
}

func ReplaceLogger(with Logger) {
	logger = with
}
