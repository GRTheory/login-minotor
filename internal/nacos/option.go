package nacos

type Options struct {
	Url         string
	Port        uint64
	NamespaceId string
	GroupName   string
	Username    string
	Password    string
	DataId      string
}

// InitConfig reads nacos-related configuration from environment variables.
func InitConfig(port *uint64) (*Options, error) {
	op := &Options{}
	if port != nil {
		op.Port = *port
	} else {
		// default port of nacos
		op.Port = uint64(8848)
	}

	return op, nil
}
