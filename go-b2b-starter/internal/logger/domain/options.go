package domain

type Option func(*Options)

type Options struct {
	Level       Level
	Output      OutputType
	FileOptions FileOptions
}

type FileOptions struct {
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

func WithLevel(level Level) Option {
	return func(o *Options) {
		o.Level = level
	}
}

func WithOutput(output OutputType) Option {
	return func(o *Options) {
		o.Output = output
	}
}

func WithFileOptions(fileOpts FileOptions) Option {
	return func(o *Options) {
		o.FileOptions = fileOpts
	}
}
