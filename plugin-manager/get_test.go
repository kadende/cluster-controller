package plugin_manager


import (
	"testing"
	"os"
	"github.com/stretchr/testify/assert"
	"path"
	"github.com/hashicorp/go-getter"
	"strconv"
	"path/filepath"
)

func TestGettingPluginPath(t *testing.T) {
	originalPluginPath, customPluginPathExists := os.LookupEnv(customPluginPathOsName)
	if customPluginPathExists{
		defer func() {os.Setenv(customPluginPathOsName, originalPluginPath)}()
	}else{
		defer func() {os.Unsetenv(customPluginPathOsName)}()
	}


	os.Setenv(customPluginPathOsName, "/test/path")

	assert.Equal(t, "/test/path", getPluginPath())

	os.Unsetenv(customPluginPathOsName)

	assert.Equal(t, getDefaultPluginPath(), getPluginPath())
}

func TestDownloadingProvider(t *testing.T) {

	pluginPath := getPluginPath()
	// delete all downloaded plugins after running the tests
	defer os.RemoveAll(pluginPath)


	workingDir, err  := os.Getwd()
	if err != nil{
		panic(err)
	}


	client := getter.Client{
		Src:  "https://github.com/mwaaas/kadende-provider-file/releases/download/0.0.1/plugin.so?filename=file-provider.so",
		Dst:  "./testDataPlugin",
		Pwd:  workingDir,
		Mode: getter.ClientModeAny,
	}
	err = client.Get()
	defer os.RemoveAll("./testDataPlugin")

	type expectedResults struct {
		path string
		errorMessage string
	}

	type scenario struct {
		plugin loadPlugin
		expectedResults expectedResults
	}

	testScenarios := [11]scenario{
		// testing happy case
		{plugin: loadPlugin{pluginName: "file", pluginType: "provider", version: "0.0.1"},
			expectedResults: expectedResults{path: "provider/file_0.0.1.so"}},

		// testing path already exists
		{plugin: loadPlugin{pluginName: "file", pluginType: "provider", version: "0.0.1"},
			expectedResults: expectedResults{errorMessage: "plugin already exists"}},


		// test downloading flavor plugin
		{plugin: loadPlugin{pluginName: "file", pluginType: "flavour", version: "0.0.1"},
			expectedResults: expectedResults{path: "flavour/file_0.0.1.so"}},

		// test with a different version
		{plugin: loadPlugin{pluginName: "file", pluginType: "provider", version: "0.0.0"},
			expectedResults: expectedResults{path: "provider/file_0.0.0.so"}},

		// test if version not specified downloads the latest version
		{plugin: loadPlugin{pluginName: "file", pluginType: "provider"},
			expectedResults: expectedResults{path: "provider/file_latest.so"}},

		// test downloading via full github url
		{plugin: loadPlugin{url: "https://github.com/mwaaas/kadende-provider-file/releases/download/0.0.1/plugin.so?pluginName=my-custom-plugin.so",
			pluginType: "provider"},
			expectedResults: expectedResults{path: "provider/my-custom-plugin_latest.so"}},


		// downloading file that already exists
		{plugin: loadPlugin{url: "https://github.com/mwaaas/kadende-provider-file/releases/download/0.0.1/plugin.so?pluginName=my-custom-plugin.so",
			pluginType: "provider"},
			expectedResults: expectedResults{errorMessage: "plugin already exists"}},


		// testing with url that does not exist
		{plugin: loadPlugin{url: "https://github.com/mwaside/file-provider-sample/archive/file-provider-sample_v1.26.0.so",
			pluginType: "provider"},
			expectedResults: expectedResults{errorMessage: "bad response code: 404"}},

		// test using file path
		{plugin: loadPlugin{url: "file://" + filepath.Join(workingDir, "testDataPlugin", "file-provider.so"),
			pluginType: "provider"},
			expectedResults: expectedResults{path: "provider/file-provider_latest.so"}},

		// test with a file that does not exist
		{plugin: loadPlugin{url: "file:///path/does/not/exist", pluginType: "provider"},
			expectedResults: expectedResults{errorMessage: "no such file"}},

		// testing with invalid plugin type
		// test using file path
		{plugin: loadPlugin{url: "file://" + filepath.Join(workingDir, "testDataPlugin", "file-provider_v0.0.1"),
			pluginType: "abcdefga"}, expectedResults: expectedResults{errorMessage: "invalid plugin type"}},
	}

	for index, scenario := range testScenarios{
		plugin, err := NewLoadPlugin(&scenario.plugin)

		if err != nil {
			assert.Contains(t, err.Error(), scenario.expectedResults.errorMessage,
				scenario.plugin.url, scenario.plugin.pluginName, index)
			continue
		}

		err = plugin.downloadPlugin()

		// if path is not empty
		if scenario.expectedResults.path != "" {
			assert.FileExists(t, path.Join(pluginPath, scenario.expectedResults.path), strconv.Itoa(index))
		}
		if err != nil && scenario.expectedResults.errorMessage == ""{
			assert.Fail(t, err.Error())
		}

		var actualErrorMessage string
		if err != nil {
			actualErrorMessage = err.Error()
		} else {
			actualErrorMessage = ""
		}

		assert.Contains(t, actualErrorMessage, scenario.expectedResults.errorMessage,
			scenario.plugin.url, scenario.plugin.pluginName, index)

	}

}
