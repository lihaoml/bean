package influx

import (
	"bean"
	util "bean/utils"
	"errors"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/joho/godotenv"
	"os"
	"strings"
)

func ConnectService(dbkey, port, envUserKey, envPwdKey string) ([]client.Client, error) {
	err := godotenv.Load()
	if err != nil || os.Getenv(dbkey) == "" {
		err = godotenv.Load(bean.BeanexAccountPath() + "db.env")
		if err != nil {
			panic(err.Error())
		}
	}
	dbhost := os.Getenv("MDS_DB_ADDRESS")
	return ConnectTo(dbhost, port, envUserKey, envPwdKey)
}

func ConnectTo(dbhost, port string, envUserKey, envPwdKey string) (cs []client.Client, err error) {
	hosts := strings.Split(dbhost, ",")
	usrs := strings.Split(os.Getenv(envUserKey), ",")
	if len(usrs) <= 1 {
		usrs = util.ReplicateString(os.Getenv(envUserKey), len(hosts))
	}
	pwds := strings.Split(os.Getenv(envPwdKey), ",")
	if len(pwds) <= 1 {
		pwds = util.ReplicateString(os.Getenv(envPwdKey), len(hosts))
	}
	if !(len(hosts) == len(usrs) && len(hosts) == len(pwds)) {
		return nil, errors.New("number of hosts do not match with number of users & pwds")
	}
	errMsg := []string{}
	for i, host := range hosts {
		if !strings.HasPrefix(host, "http") {
			host = "http://" + host
		}
		c, err := client.NewHTTPClient(client.HTTPConfig{
			Addr:     host + ":" + port,
			Username: usrs[i],
			Password: pwds[i],
		})
		if err != nil {
			errMsg = append(errMsg, host+": "+err.Error())
		} else {
			cs = append(cs, c)
		}
	}
	if len(errMsg) > 0 {
		err = errors.New(fmt.Sprint(errMsg))
	}
	return cs, err
}
