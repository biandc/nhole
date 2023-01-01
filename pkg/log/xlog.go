package log

type Logger struct {
	prefixes []string

	prefixString string
}

func New() *Logger {
	return &Logger{
		prefixes: make([]string, 0),
	}
}

func (l *Logger) ResetPrefixes() (old []string) {
	old = l.prefixes
	l.prefixes = make([]string, 0)
	l.prefixString = ""
	return
}

func (l *Logger) AppendPrefix(prefix string) *Logger {
	l.prefixes = append(l.prefixes, prefix)
	l.prefixString += "[" + prefix + "] "
	return l
}

func (l *Logger) Spawn() *Logger {
	nl := New()
	for _, v := range l.prefixes {
		nl.AppendPrefix(v)
	}
	return nl
}

func (l *Logger) Error(format string, v ...interface{}) {
	Log.Error(l.prefixString+format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	Log.Warn(l.prefixString+format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
	Log.Info(l.prefixString+format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	Log.Debug(l.prefixString+format, v...)
}

func (l *Logger) Trace(format string, v ...interface{}) {
	Log.Trace(l.prefixString+format, v...)
}
