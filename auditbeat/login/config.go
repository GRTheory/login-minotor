package login

// config defines the metricset's configuration options.
type config struct {
	WtmpFilePattern string `config:"login.wtmp_file_pattern"`
	BtmpFilePattern string `config:"login.btmp_file_pattern"`
}

func defaultConfig() config {
	return config{
		WtmpFilePattern: "/var/log/wtmp*",
		BtmpFilePattern: "/var/log/btmp*",
	}
}
