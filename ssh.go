package sshmysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"net"
	"os"
	"strings"

	mysqlD "github.com/go-sql-driver/mysql"
	"github.com/jfcote87/sshdb"
	"github.com/jfcote87/sshdb/mysql"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Address        string `json:"address"`
	User           string `json:"user"`
	Password       string `json:"password"`
	PrivateKeyFile string `json:"privateKeyFile"`
}

var ERROR_EMPTY_CONFIG = errors.New("empty ssh config")

// JsonToSSHConfig 将json字符串转为SSHConfig对象
func JsonToSSHConfig(s string) (sshConfig *SSHConfig, err error) {
	if strings.TrimSpace(s) == "" {
		return nil, ERROR_EMPTY_CONFIG
	}
	sshConfig = &SSHConfig{}
	err = json.Unmarshal([]byte(s), sshConfig)
	if err != nil {
		return nil, err
	}
	return sshConfig, nil
}

func (h SSHConfig) Config() (cfg *ssh.ClientConfig, err error) {
	cfg = &ssh.ClientConfig{
		User:            h.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            make([]ssh.AuthMethod, 0),
	}
	if h.Password != "" {
		cfg.Auth = append(cfg.Auth, ssh.Password(h.Password))
		return cfg, nil
	}
	if h.PrivateKeyFile == "" {
		return cfg, nil
	}
	//优先使用keyFile
	k, err := os.ReadFile(h.PrivateKeyFile)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(k)
	if err != nil {
		return nil, err
	}
	cfg.Auth = append(cfg.Auth, ssh.PublicKeys(signer))
	return cfg, nil
}

func (h SSHConfig) Tunnel(dsn string) (sqlDB *sql.DB, err error) {
	sshConfig, err := h.Config()
	if err != nil {
		return nil, err
	}
	tunnel, err := sshdb.New(sshConfig, h.Address)
	if err != nil {
		return nil, err
	}
	tunnel.IgnoreSetDeadlineRequest(true)
	connector, err := tunnel.OpenConnector(mysql.TunnelDriver, dsn)
	if err != nil {
		err = errors.WithMessagef(err, " dsn:%s", dsn)
		return nil, err
	}
	sqlDB = sql.OpenDB(connector)
	return sqlDB, err
}

// RegisterNetwork 注册自定义网络协议,比 Tunnel 更能和其它已有项目兼容
func (h SSHConfig) RegisterNetwork(dsn string) (err error) {
	sshConfig, err := h.Config()
	if err != nil {
		return err
	}
	tunnel, err := sshdb.New(sshConfig, h.Address)
	if err != nil {
		return err
	}
	tunnel.IgnoreSetDeadlineRequest(true)
	_, err = tunnel.OpenConnector(mysql.TunnelDriver, dsn)
	if err != nil {
		err = errors.WithMessagef(err, " dsn:%s", dsn)
		return err
	}

	cfg, err := mysqlD.ParseDSN(dsn)
	if err != nil {
		return err
	}
	//注册自定义网络
	mysqlD.RegisterDialContext(cfg.Net, func(ctx context.Context, addr string) (net.Conn, error) {
		return tunnel.DialContext(ctx, cfg.Net, addr)
	})

	return nil
}
