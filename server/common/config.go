package common

import (
	"encoding/json"
	"github.com/go-ini/ini"
	"io/ioutil"
	"strings"
)

type Config struct {
	WorkerId    string   `json:"worker_id"`
	WorkerType  string   `json:"worker_type"`
	WorkerCount int      `json:"worker_count"`
	Service     []string `json:"service"`
	Host        string   `json:"host"`
	Port        int      `json:"port"`
	Msr         *MSR     `json:"msr"`
	Timeout     int      `json:"timeout"`
	IsTesting   bool     `json:"is_testing"`
	//use token if debug is true
	Debug bool   `json:"debug"`
	Token string `json:"token"`
}
type MSR struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Check   *Check `json:"check"`
}
type Check struct {
	Enabled    bool `json:"enabled"`
	Timeout    int  `json:"timeout"`
	Interval   int  `json:"interval"`
	RemoveTime int  `json:"remove_time"`
}

func NewConfig() *Config {
	return &Config{
		WorkerId:    "node",
		WorkerType:  "XE",
		Service:     []string{},
		WorkerCount: 2,
		Host:        "0.0.0.0",
		Port:        9090,
		Timeout:     30,
		Debug:       false,
		IsTesting:   false,
		Token:       "test",
		Msr: &MSR{
			Enabled: false,
			Host:    "127.0.0.1",
			Port:    8500,
			Check: &Check{
				Enabled:    true,
				Timeout:    1,
				Interval:   1,
				RemoveTime: 60,
			},
		},
	}
}
func (c *Config) Load(file string) error {
	cfg, err := ini.InsensitiveLoad(file)
	if err != nil {
		return err
	}
	return c.load(cfg)
}
func (c *Config) load(cfg *ini.File) error {
	err := cfg.MapTo(c)
	if err != nil {
		return err
	}
	err = cfg.Section("msr").MapTo(c.Msr)
	if err != nil {
		return err
	}
	err = cfg.Section("check").MapTo(c.Msr.Check)
	if err != nil {
		return err
	}
	return nil
}
func (c *Config) LoadArgs(args []string) error {
	cfg, err := ini.LoadSources(ini.LoadOptions{Insensitive: true}, []byte{})
	if err != nil {
		return err
	}
	for _, s := range args {
		if strings.HasPrefix(s, "-") {
			as := strings.Split(s, "=")
			if len(as) == 2 && as[1] != "" {
				if strings.HasPrefix(s, "-msr.check") {
					cfg.Section("check").Key(string(as[0][11:])).SetValue(as[1])
				} else if strings.HasPrefix(s, "-msr") {
					cfg.Section("msr").Key(string(as[0][5:])).SetValue(as[1])
				} else {
					cfg.Section("").Key(string(as[0][1:])).SetValue(as[1])
				}
			}
		}
	}
	return c.load(cfg)
}
func (c *Config) LoadJson(file string) error {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bs, c); err != nil {
		return err
	}
	return nil
}
func (c *Config) SaveJson(file string) error {
	bs, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, bs, 0764)
}
func (c *Config) PrintJson() error {
	bs, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	println(string(bs))
	return nil
}
