package plugin_manager

import (
	"os"
	"fmt"
	"github.com/hashicorp/go-getter"
	"net/url"
	"runtime"
	"path"
	"strings"
	"errors"
	"github.com/kadende/kadende-interfaces/plugin-interface"
)

var defaultPluginPath = getDefaultPluginPath()
const  (
	customPluginPathOsName = "PluginPath"
	//defaultPluginPath = ""

	default_release_host = "https://github.com"
	custom_default_host_os_env = "ReleaseHost"
)


type loadPlugin struct {
	pluginName string
	url string
	pluginType plugin_interface.PluginType
	version string
	filePluginExists bool
}

func NewLoadPlugin(l *loadPlugin) (*loadPlugin, error) {
	if !l.pluginType.Validate(){
		return nil, errors.New("invalid plugin type")
	}
	l.validateSetPluginName()
	if l.pluginName == ""{
		return nil, errors.New("missing plugin name")
	}
	l.setVersion()
	l.setPluginUrl()
	l.checkPluginFileExists()
	return l, nil
}

func getDefaultPluginPath() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Join(path.Dir(path.Dir(filename)), "plugins")
}

func getPluginPath() string {
	PluginPath, fileExists := os.LookupEnv(customPluginPathOsName)
	if !fileExists {
		PluginPath = defaultPluginPath
	}
	return PluginPath
}

func getReleaseHost() string {
	releaseHost, exists := os.LookupEnv(custom_default_host_os_env)
	if !exists {
		releaseHost = default_release_host
	}

	// validate the url
	_, err := url.ParseRequestURI(releaseHost)
	if err != nil{
		panic(err)
	}
	return releaseHost
}

func replacePluginExtension(name string) string {
	return strings.Replace(name, ".so", "", 1)
}

func (l *loadPlugin) setPluginUrl(){
	if l.url == ""{
		l.url = fmt.Sprintf("%s/kadende-plugins/kadende-%s-%s/releases/download/%s/plugin.so",
			getReleaseHost(), l.pluginType, l.pluginName, l.version)
	}
	urlObj, err := url.ParseRequestURI(l.url)
	if err != nil{
		panic(err)
	}
	q := urlObj.Query()
	q.Del("filename") // Its magic argument should not be defined by the user

	// define our filename
	q.Add("filename", l.pluginFileName())

	urlObj.RawQuery = q.Encode()
	l.url = urlObj.String()
}

func (l *loadPlugin) validateSetPluginName() {
	// set plugin name only if its has not been defined
	// and url has been defined so that we can deduce from it.
	if l.pluginName == "" && l.url != "" {
		urlObj, err := url.ParseRequestURI(l.url)
		if err != nil{
			panic(err)
		}

		q := urlObj.Query()
		if q.Get("pluginName") != "" {
			l.pluginName = replacePluginExtension(q.Get("pluginName"))
		} else {
			// if non of the parameters have been given
			// we can deduce it from the last name in the path
			pathNames := strings.Split(urlObj.Path, "/")
			l.pluginName = replacePluginExtension(pathNames[len(pathNames) - 1 ])
		}
	}
}

func (l *loadPlugin) pluginFilePath() string {
	return path.Join(l.pluginDir(), l.pluginFileName())
}

func (l *loadPlugin) pluginDir() string {
	return path.Join(getPluginPath(), l.pluginType.ToString())
}

func (l loadPlugin) pluginFileName() string {
	return fmt.Sprintf("%s_%s.so", l.pluginName, l.version)
}

func (l *loadPlugin) setVersion() {
	if l.version == ""{
		l.version = "latest"
	}
}

func (l *loadPlugin) checkPluginFileExists(){
	if _, err := os.Stat(l.pluginFilePath()); os.IsNotExist(err) {
		l.filePluginExists = false
	}else{
		l.filePluginExists = true
	}
}

func (l loadPlugin) downloadPlugin() (error) {
	//If file already exists throw an error
	//This is to avoid overriding existing files
	if l.filePluginExists{
		return errors.New("plugin already exists")
	}

	// Get the pwd
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	client := getter.Client{
		Src:  l.url,
		Dst:  l.pluginDir(),
		Pwd:  pwd,
		Mode: getter.ClientModeAny,
	}

	err = client.Get()
	return err
}
