package versions

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gvcgo/collector/pkgs/confs"
	"github.com/gvcgo/collector/pkgs/upload"
	"github.com/gvcgo/collector/pkgs/utils"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/request"
)

const (
	InstallerVersionFileNamePattern string = "%s.version.json"
)

/*
Only the latest version for:

1. android sdkmanager
https://developer.android.com/tools/sdkmanager?hl=zh-cn

	download:
	https://developer.android.google.cn/studio?hl=zh-cn
	https://developer.android.com/studio?hl=en

2. cygwin installer
https://cygwin.com/install.html

	download:
	https://cygwin.com/setup-x86_64.exe

3. msys2 installer
https://www.msys2.org/#installation
https://github.com/msys2/msys2-installer/releases

	download:
	https://github.com/msys2/msys2-installer/releases/download/2024-01-13/msys2-x86_64-20240113.exe

4. rust installer
https://forge.rust-lang.org/infra/other-installation-methods.html

	download:
	https://static.rust-lang.org/rustup/rustup-init.sh
	https://static.rust-lang.org/rustup/dist/i686-pc-windows-gnu/rustup-init.exe

5. VSCode
https://code.visualstudio.com/sha?build=stable

6. miniconda
https://anaconda.org.cn/anaconda/install/silent-mode/
https://repo.anaconda.com/miniconda/
*/
type Installer struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions map[string]Versions
	fetcher  *request.Fetcher
	homepage string
	doc      *goquery.Document
}

func NewInstaller(cnf *confs.CollectorConf) (i *Installer) {
	i = &Installer{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: make(map[string]Versions),
		fetcher:  request.NewFetcher(),
	}
	if confs.EnableProxyOrNot() {
		pxy := i.cnf.ProxyURI
		if pxy == "" {
			pxy = confs.DefaultProxy
		}
		i.fetcher.Proxy = pxy
	}
	return
}

func (i *Installer) getDoc() {
	i.fetcher.SetUrl(i.homepage)
	i.fetcher.Timeout = 30 * time.Second
	if resp, sCode := i.fetcher.GetString(); resp != "" && sCode == 200 {
		// fmt.Println(resp)
		var err error
		i.doc, err = goquery.NewDocumentFromReader(strings.NewReader(resp))
		if err != nil {
			gprint.PrintError(fmt.Sprintf("Parse page errored: %+v", err))
		}
		if i.doc == nil {
			gprint.PrintError(fmt.Sprintf("Cannot parse html for %s", i.fetcher.Url))
			os.Exit(1)
		}
	} else {
		fmt.Println(sCode)
	}
}

func (i *Installer) GetAndroidSDKManager() {
	// https://dl.google.com/android/repository/commandlinetools-win-11076708_latest.zip
	vPattern := regexp.MustCompile(`(\d+)`)
	baseUrl := "https://dl.google.com/android/repository"
	i.homepage = "https://developer.android.com/studio?hl=en"
	i.doc = nil
	i.getDoc()
	if i.doc != nil {
		// //table[@class="download"][1]/tbody/tr
		i.doc.Find("table.download").Eq(1).Find("tr").Each(func(idx int, s *goquery.Selection) {
			if idx == 0 {
				return
			}
			platform := strings.ToLower(s.Find("td").Eq(0).Text())
			fName := s.Find("td").Eq(1).Find("button").Text()
			if platform == "" || fName == "" {
				return
			}
			vName := vPattern.FindString(fName)
			u, _ := url.JoinPath(baseUrl, fName)
			sha256Str := s.Find("td").Eq(3).Text()

			ver := &VFile{}
			ver.Url = u
			ver.Arch = "all"
			ver.Os = utils.ParsePlatform(platform)
			ver.Sum = sha256Str
			if ver.Sum != "" {
				ver.SumType = "sha256"
			}
			ver.Extra = fmt.Sprintf("v%s", vName)
			name := "sdkmanager"
			if vlist, ok := i.versions[name]; !ok || vlist == nil {
				i.versions[name] = Versions{
					vName: []*VFile{},
				}
			}

			i.versions[name][vName] = append(i.versions[name][vName], ver)
			// fmt.Println(ver.Os, ver.Arch, ver.Url, ver.Sum, ver.SumType, ver.Extra)
		})
	}
}

func (i *Installer) GetCygwinInstaller() {
	// https://cygwin.com/setup-x86_64.exe
	ver := &VFile{
		Url:   "https://cygwin.com/setup-x86_64.exe",
		Arch:  "amd64",
		Os:    "windows",
		Extra: "latest",
	}
	name := "cygwin"
	i.versions[name] = Versions{
		"latest": []*VFile{ver},
	}
}

func (i *Installer) GetMsys2Installer() {
	// https://github.com/msys2/msys2-installer/releases/download/nightly-x86_64/msys2-x86_64-latest.exe
	ver := &VFile{
		Url:   "https://github.com/msys2/msys2-installer/releases/download/nightly-x86_64/msys2-x86_64-latest.exe",
		Arch:  "amd64",
		Os:    "windows",
		Extra: "latest",
	}
	name := "msys2"
	i.versions[name] = Versions{
		"latest": []*VFile{ver},
	}
}

func (i *Installer) GetRustInstaller() {
	name := "rustup"
	rVersion := "latest"
	i.versions[name] = Versions{
		rVersion: []*VFile{},
	}

	verDarwinx64 := &VFile{
		Url:   "https://static.rust-lang.org/rustup/dist/x86_64-apple-darwin/rustup-init",
		Arch:  "amd64",
		Os:    "darwin",
		Extra: rVersion,
	}
	i.versions[name][rVersion] = append(i.versions[name][rVersion], verDarwinx64)

	verDarwinArm64 := &VFile{
		Url:   "https://static.rust-lang.org/rustup/dist/aarch64-apple-darwin/rustup-init",
		Arch:  "arm64",
		Os:    "darwin",
		Extra: rVersion,
	}
	i.versions[name][rVersion] = append(i.versions[name][rVersion], verDarwinArm64)

	verLinuxx64 := &VFile{
		Url:   "https://static.rust-lang.org/rustup/dist/x86_64-unknown-linux-gnu/rustup-init",
		Arch:  "amd64",
		Os:    "linux",
		Extra: rVersion,
	}
	i.versions[name][rVersion] = append(i.versions[name][rVersion], verLinuxx64)

	verLinuxArm64 := &VFile{
		Url:   "https://static.rust-lang.org/rustup/dist/aarch64-unknown-linux-gnu/rustup-init",
		Arch:  "arm64",
		Os:    "linux",
		Extra: rVersion,
	}
	i.versions[name][rVersion] = append(i.versions[name][rVersion], verLinuxArm64)

	verWinx64 := &VFile{
		Url:   "https://static.rust-lang.org/rustup/dist/x86_64-pc-windows-msvc/rustup-init.exe",
		Arch:  "amd64",
		Os:    "windows",
		Extra: rVersion,
	}
	i.versions[name][rVersion] = append(i.versions[name][rVersion], verWinx64)

	verWinArm64 := &VFile{
		Url:   "https://static.rust-lang.org/rustup/dist/aarch64-pc-windows-msvc/rustup-init.exe",
		Arch:  "arm64",
		Os:    "windows",
		Extra: rVersion,
	}
	i.versions[name][rVersion] = append(i.versions[name][rVersion], verWinArm64)
}

type CodePlatform struct {
	Os         string `json:"os"`
	PrettyName string `json:"prettyname"`
}

type CodeItem struct {
	Url      string        `josn:"url"`
	Sum      string        `json:"sha256hash"`
	Version  string        `json:"name"`
	Build    string        `json:"build"`
	Platform *CodePlatform `json:"platform"`
}

type CodeProducts struct {
	Products []*CodeItem `json:"products"`
}

func vscodeAllowed(item *CodeItem) bool {
	// excludeList := []string{"_cli", ".deb", ".rpm"}
	excludeList := []string{"_cli", ".tar.gz", "armhf", "armv7hl"}
	for _, excludeStr := range excludeList {
		if strings.Contains(item.Url, excludeStr) {
			return false
		}
	}
	if strings.HasSuffix(item.Url, ".exe") && !strings.Contains(item.Url, "User") {
		return true
	}
	if strings.HasSuffix(item.Url, ".tar.gz") {
		return true
	}
	if strings.HasSuffix(item.Url, ".zip") && strings.Contains(item.Url, "darwin") {
		return true
	}
	if strings.HasSuffix(item.Url, ".deb") || strings.HasSuffix(item.Url, ".rpm") {
		return true
	}
	return false
}

func (i *Installer) GetVSCode() {
	// https://code.visualstudio.com/sha?build=stable
	i.fetcher.SetUrl("https://code.visualstudio.com/sha?build=stable")
	name := "vscode"
	content, _ := i.fetcher.GetString()
	i.versions[name] = Versions{}

	if content != "" {
		products := &CodeProducts{}
		if err := json.Unmarshal([]byte(content), products); err == nil {
			for _, item := range products.Products {
				if vscodeAllowed(item) {
					ver := &VFile{}
					ver.Url = item.Url
					ver.Arch = utils.ParseArch(item.Url)
					if strings.Contains(item.Url, "win32-arm64") {
						ver.Arch = "arm64"
					} else if strings.Contains(item.Url, "win32-x64") {
						ver.Arch = "amd64"
					} else if strings.Contains(item.Url, "universal") {
						ver.Arch = "any"
					} else if strings.Contains(item.Url, "VSCode-darwin.zip") {
						ver.Arch = "amd64"
					}

					ver.Os = utils.ParsePlatform(item.Platform.PrettyName)
					ver.Sum = item.Sum
					if ver.Sum != "" {
						ver.SumType = "sha256"
					}
					ver.Extra = fmt.Sprintf("v%s", item.Version)
					if len(i.versions[name]) == 0 {
						i.versions[name][item.Version] = []*VFile{}
					}
					i.versions[name][item.Version] = append(i.versions[name][item.Version], ver)
					// fmt.Println(ver.Arch, ver.Os, ver.Url, ver.Sum, ver.SumType, ver.Extra)
				}
			}
		}
	}
}

func filterMinicondaByFName(fname string) bool {
	r := false
	toBunList := []string{
		".pkg",
		"Miniconda2-latest-", // for miniconda
		"Miniconda-latest-",  // for miniconda
	}
	for _, b := range toBunList {
		if strings.Contains(fname, b) {
			return true
		}
	}
	return r
}

func (i *Installer) GetMiniconda() {
	// https://repo.anaconda.com/miniconda/
	i.homepage = "https://repo.anaconda.com/miniconda/"
	i.doc = nil
	i.fetcher.Proxy = ""
	i.getDoc()
	if i.doc != nil {
		var (
			shaStr string
			vName  string
		)
		name := "miniconda"
		i.versions[name] = Versions{}
		i.doc.Find("table").Find("tr").Each(func(ii int, s *goquery.Selection) {
			u := s.Find("td").Eq(0).Find("a").AttrOr("href", "")
			if u == "" {
				return
			}
			fName := s.Find("td").Eq(0).Find("a").Text()
			// if strings.HasSuffix(fName, ".pkg") {
			// 	return
			// }
			if filterMinicondaByFName(fName) {
				return
			}
			sha256Str := s.Find("td").Eq(3).Text()

			if strings.Contains(fName, "latest") {
				ver := &VFile{}
				if !strings.HasPrefix(u, "http") {
					u, _ = url.JoinPath(i.homepage, u)
				}
				ver.Url = u
				ver.Arch = utils.ParseArch(fName)
				ver.Os = utils.ParsePlatform(fName)
				ver.Sum = sha256Str
				if ver.Sum != "" {
					shaStr = shaStr + ";" + sha256Str
					ver.SumType = "sha256"
				}
				ver.Extra = "latest"
				if len(i.versions[name]) == 0 {
					i.versions[name][ver.Extra] = []*VFile{}
				}
				i.versions[name][ver.Extra] = append(i.versions[name][ver.Extra], ver)
			} else {
				if sha256Str != "" && strings.Contains(shaStr, sha256Str) && vName == "" {
					vName = VersionPattern.FindString(fName)
				}
			}
		})
		if vName != "" {
			i.versions[name][vName] = i.versions[name]["latest"]
			delete(i.versions[name], "latest")
		}
	}
}

func (i *Installer) FetchAll() {
	fmt.Println("android sdkmanager...")
	i.GetAndroidSDKManager()
	fmt.Println("cygwin installer...")
	i.GetCygwinInstaller()
	fmt.Println("msys2 installer...")
	i.GetMsys2Installer()
	fmt.Println("rust installer...")
	i.GetRustInstaller()
	fmt.Println("vscode...")
	i.GetVSCode()
	fmt.Println("miniconda...")
	i.GetMiniconda()
}

func (i *Installer) Upload() {
	for name, versions := range i.versions {
		if len(versions) == 0 {
			continue
		}
		fPath := filepath.Join(i.cnf.DirPath(), fmt.Sprintf(InstallerVersionFileNamePattern, name))
		if content, err := json.MarshalIndent(versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			i.uploader.Upload(fPath)
		}
	}
}
