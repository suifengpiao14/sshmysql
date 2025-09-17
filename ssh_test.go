package sshmysql_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suifengpiao14/sshmysql"
)

var sshConfig = sshmysql.SSHConfig{
	Address:        "ip:port",
	User:           "root",
	PrivateKeyFile: "C:\\Users\\Admin\\.ssh\\id_rsa",
}
var dbDSN = `root:1b03f8b486908bbe34ca2f4a4b91bd1c@tcp(127.0.0.1:3306)/test?charset=utf8&timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=False&loc=Local&multiStatements=true`

func TestRegisterSSHNet(t *testing.T) {
	err := sshConfig.RegisterNetwork(dbDSN)
	require.NoError(t, err)
	db, err := sql.Open("mysql", dbDSN)
	require.NoError(t, err)

	sqlStr := "select count(*) from t_generic_language where 1=1;"
	var count int64
	sqlRaw := db.QueryRow(sqlStr)
	err = sqlRaw.Scan(&count)
	require.NoError(t, err)
	fmt.Println(count)
}
