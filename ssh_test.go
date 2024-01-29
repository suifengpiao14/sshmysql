package sshmysql_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suifengpiao14/sshmysql"
)

func TestSshMysql(t *testing.T) {
	sshConfig := sshmysql.SSHConfig{
		Address:  "ip:port",
		User:     "username",
		Password: "",
	}
	dbDSN := "user:password@tcp(127.0.0.1:3306)/ad?charset=utf8&timeout=1s&readTimeout=5s&writeTimeout=5s&parseTime=False&loc=Local&multiStatements=true"

	db, err := sshConfig.Tunnel(dbDSN)
	require.NoError(t, err)
	sql := "select count(*) from ad.advertise;"
	var count int64
	err = db.QueryRow(sql).Scan(&count)
	require.NoError(t, err)
	fmt.Println(count)

}
