package nacos

import (
	"errors"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// NamingClient will keep this server alive in nacos until the server is terminated.
var NamingClient naming_client.INamingClient

// RegisterServer2Nacos registers this server to nacos.
func RegisterServer2Naocs() error {
	option, err := InitConfig(nil)
	if err != nil {
		return err
	}

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(
			option.Url, option.Port,
		),
	}

	cc := *constant.NewClientConfig(
		constant.WithUsername(option.Username),
		constant.WithPassword(option.Password),
		constant.WithNamespaceId(option.NamespaceId),
		constant.WithNotLoadCacheAtStart(true),
	)

	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig: &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return err
	}
	
	success, err := client.RegisterInstance(vo.RegisterInstanceParam{
		Ip: "",
		Port: uint64(0),
		Healthy: true,
		ServiceName: "",
		GroupName: "",
		Ephemeral: true,
		Enable: true,
	})

	if !success {
		return errors.New("failed to register server on nacos")
	}else if err != nil {
		return err
	}

	NamingClient = client
	return nil
}

