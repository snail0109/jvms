package cmdCli

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/baneeishaque/adoptium_jdk_go"
	"github.com/snail0109/jvms/internal/entity"
	"github.com/snail0109/jvms/utils/web"
)

// getSImilarAvailableVersions by version to support version not found error
// func getSimilarAvailableVersions(version string) {

// }

func getJavaHome(jdkTempFile string) string {
	var javaHome string
	fs.WalkDir(os.DirFS(jdkTempFile), ".", func(path string, d fs.DirEntry, err error) error {
		if filepath.Base(path) == "javac.exe" {
			temPath := strings.Replace(path, "bin/javac.exe", "", -1)
			javaHome = filepath.Join(jdkTempFile, temPath)
			return fs.SkipDir
		}
		return nil
	})
	return javaHome
}

func getJdkVersions(cfx *entity.Config) ([]entity.JdkVersion, error) {
	jsonContent, err := web.GetRemoteTextFile(cfx.Originalpath)
	if err != nil {
		return nil, err
	}
	var versions []entity.JdkVersion
	err = json.Unmarshal([]byte(jsonContent), &versions)
	if err != nil {
		return nil, err
	}
	//fmt.Println(versions)
	adoptiumJdks := strings.Split(adoptium_jdk_go.ApiListReleases(), "\n")
	for _, adoptiumJdkUrl := range adoptiumJdks {
		fileSeparatorIndex := strings.LastIndex(adoptiumJdkUrl, "/")
		fileName := adoptiumJdkUrl[fileSeparatorIndex+1:]
		fileVersion := strings.TrimSuffix(fileName, ".zip")
		//fmt.Println(fileVersion)
		versions = append(versions, entity.JdkVersion{Version: fileVersion, Url: adoptiumJdkUrl})
	}

	//Azul JDKs
	// azulJdks := jdk.AzulJDKs()
	// for _, azulJdk := range azulJdks {
	// 	fmt.Println(azulJdk.DownloadURL)
	// 	versions = append(versions, entity.JdkVersion{Version: azulJdk.ShortName, Url: azulJdk.DownloadURL})
	// }

	return versions, nil
}
