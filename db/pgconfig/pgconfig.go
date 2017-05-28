package pgconfig

// Получение настроек из файла конфигураций

import (
	"encoding/json"
	"os"
)

type Configurations struct {
	Postgre_user          string
	Postgre_password      string
	Postgre_host          string
	Postgre_database      string
	Postgre_ssl           string
	Postgre_connect_count int64
}

func (c *Configurations) getConfigurations() error {

	confFile, err := os.Open("pgconfig.conf")

	if err != nil {
		return err
	}
	defer confFile.Close()

	stat, err := confFile.Stat()

	if err != nil {
		return err
	}

	bs := make([]byte, stat.Size())
	_, err = confFile.Read(bs)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bs, &c)

	if err != nil {
		return err
	}

	return nil
}

var Config Configurations

func init() {

	err := Config.getConfigurations()

	if err != nil {
		panic("Config file not found!" + err.Error())
	}

}
