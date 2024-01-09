package confs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/koanfer"
)

const (
	// https://cdn.jsdelivr.net
	ToEnableJsdelivrEnvName string = "ENABLE_JS_DELIVR"
	// proxy
	ToEnableProxyEnvName string = "ENABLE_PROXY"
)

type StorageType int

const (
	StorageGithub      StorageType = 1
	StorageGitee       StorageType = 2
	ConfigFileName     string      = "config.json"
	VPNFileName        string      = "conf.txt"
	SubscriberFileName string      = "subscribers.txt"
	DomainFileName     string      = "domains.txt"
	RawDomainFileName  string      = "raw_domains.txt"
	WorkDirName        string      = ".pxycollector"
)

type CollectorConf struct {
	Type      StorageType `json,koanf:"type"`
	UserName  string      `json,koanf:"username"` // username for github or gitee.
	Token     string      `json,koanf:"token"`
	Repo      string      `json,koanf:"repo"`
	CryptoKey string      `json,koanf:"crypto_key"`
	ProxyURI  string      `json,koanf:"proxy_uri"`
	dirpath   string
	k         *koanfer.JsonKoanfer
}

func NewCollectorConf() (cc *CollectorConf) {
	homeDir, _ := os.UserHomeDir()
	cc = &CollectorConf{
		dirpath: filepath.Join(homeDir, WorkDirName),
	}
	cc.initiate()
	return
}

func (c *CollectorConf) initiate() {
	if ok, _ := gutils.PathIsExist(c.dirpath); !ok {
		os.MkdirAll(c.dirpath, os.ModePerm)
	}
	confPath := filepath.Join(c.dirpath, ConfigFileName)
	c.k, _ = koanfer.NewKoanfer(confPath)
	if ok, _ := gutils.PathIsExist(confPath); !ok {
		if err := c.Save(); err != nil {
			gprint.PrintError("%+v", err)
			return
		}
	}

	subPath := c.subPath()
	if ok, _ := gutils.PathIsExist(subPath); !ok {
		// To save default subscriber list.
		os.WriteFile(subPath, []byte(SubscribedUrls), os.ModePerm)
	}

	rawDomainPath := c.RawDomainPath()
	if ok, _ := gutils.PathIsExist(rawDomainPath); !ok {
		// To save default domain list.
		os.WriteFile(rawDomainPath, []byte(RawEdDomains), os.ModePerm)
	}

	c.Load()
	// Setup configs for collector.
	c.setup()
}

func (c *CollectorConf) subPath() string {
	return filepath.Join(c.dirpath, SubscriberFileName)
}

func (c *CollectorConf) DomainPath() string {
	return filepath.Join(c.dirpath, DomainFileName)
}

func (c *CollectorConf) RawDomainPath() string {
	return filepath.Join(c.dirpath, RawDomainFileName)
}

func (c *CollectorConf) VPNFilePath() string {
	return filepath.Join(c.dirpath, VPNFileName)
}

func (c *CollectorConf) setup() {
	if c.Token == "" || c.Repo == "" {
		fmt.Println("Please choose storage type: ")
		fmt.Println("1. Github. (default)")
		fmt.Println("2. Gitee.")
		var sType string
		fmt.Scanln(&sType)
		switch sType {
		case "2":
			c.Type = StorageGitee
		default:
			c.Type = StorageGithub
		}

		fmt.Println("Please enter your github/gitee username: ")
		var username string
		fmt.Scanln(&username)
		if username != "" {
			c.UserName = username
		}

		fmt.Println("Please enter your github/gitee access-token: ")
		var token string
		fmt.Scanln(&token)
		if token != "" {
			c.Token = token
		}
		fmt.Println("Please enter your github/gitee repo name: ")
		var repo string
		fmt.Scanln(&repo)
		if repo != "" {
			c.Repo = repo
		}
		c.Save()

		// To reset the Crypto Key or not.
		fmt.Println("Do you want to reset the Crypto Key? (y/n)")
		var ok string
		fmt.Scanln(&ok)
		switch ok {
		case "y", "Y", "yes", "Yes":
			c.ResetCryptoKey()
		default:
			gprint.PrintWarning("Invalid input.")
		}
	}
}

func (c *CollectorConf) Load() error {
	return c.k.Load(c)
}

func (c *CollectorConf) Save() error {
	return c.k.Save(c)
}

func (c *CollectorConf) ResetCryptoKey() {
	c.Load()
	c.CryptoKey = gutils.RandomString(16)
	gprint.PrintInfo("CryptoKey: %s", c.CryptoKey)
	c.Save()
}

func (c *CollectorConf) ShowCryptoKey() {
	c.Load()
	gprint.PrintInfo("CryptoKey: %s", c.CryptoKey)
}

// Domains for cloudflare edgetunnels.
func (c *CollectorConf) ShowRawDomains() {
	dPath := c.RawDomainPath()
	if ok, _ := gutils.PathIsExist(dPath); ok {
		content, _ := os.ReadFile(dPath)
		gprint.PrintInfo("RawDomain list for cloudflare edgetunnels: ")
		fmt.Println(gprint.YellowStr(string(content)))
	} else {
		gprint.PrintError("No rawDomain list for edgetunnels available.")
	}
}

func (c *CollectorConf) GetRawDomains() (r []string) {
	dPath := c.RawDomainPath()
	if ok, _ := gutils.PathIsExist(dPath); ok {
		content, _ := os.ReadFile(dPath)
		r = strings.Split(string(content), "\n")
	} else {
		os.WriteFile(dPath, []byte(RawEdDomains), os.ModePerm)
		r = strings.Split(RawEdDomains, "\n")
	}
	return r
}

func (c *CollectorConf) AddRawDomains(domains ...string) {
	dPath := c.RawDomainPath()
	if ok, _ := gutils.PathIsExist(dPath); ok {
		content, _ := os.ReadFile(dPath)
		s := string(content)
		toSaveList := []string{}
		for _, d := range domains {
			if !strings.Contains(s, d) {
				toSaveList = append(toSaveList, d)
			}
		}
		os.WriteFile(dPath, []byte(s+strings.Join(toSaveList, "\n")), os.ModePerm)
	} else {
		os.WriteFile(dPath, []byte(RawEdDomains+strings.Join(domains, "\n")), os.ModePerm)
	}
}

// Subscriber list.
func (c *CollectorConf) ShowSubs() {
	subPath := c.subPath()
	if ok, _ := gutils.PathIsExist(subPath); ok {
		content, _ := os.ReadFile(subPath)
		gprint.PrintInfo("Subscribed urls: ")
		fmt.Println(gprint.YellowStr(string(content)))
	} else {
		gprint.PrintError("No subscribed urls available.")
	}
}

func (c *CollectorConf) GetSubs() (r []string) {
	subPath := c.subPath()
	if ok, _ := gutils.PathIsExist(subPath); ok {
		content, _ := os.ReadFile(subPath)
		r = strings.Split(string(content), "\n")
	} else {
		os.WriteFile(subPath, []byte(SubscribedUrls), os.ModePerm)
		r = strings.Split(SubscribedUrls, "\n")
	}
	return r
}

func (c *CollectorConf) AddSubs(subs ...string) {
	subPath := c.DomainPath()
	newStr := strings.Join(subs, "\n")
	if ok, _ := gutils.PathIsExist(subPath); ok {
		content, _ := os.ReadFile(subPath)
		s := string(content) + newStr
		os.WriteFile(subPath, []byte(s), os.ModePerm)
	} else {
		os.WriteFile(subPath, []byte(RawEdDomains+newStr), os.ModePerm)
	}
}
