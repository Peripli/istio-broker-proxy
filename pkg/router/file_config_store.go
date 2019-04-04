package router

import (
	"errors"
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/config"
	"io/ioutil"
	"istio.io/istio/pilot/pkg/model"
	"k8s.io/api/core/v1"
	"log"
	"os"
	"path"
)

type fileConfigStore struct {
	istioDirectory string
}

var _ ConfigStore = &fileConfigStore{}

// NewFileConfigStore returns a new FileConfigStore for a given directory
func NewFileConfigStore(dir string) ConfigStore {
	return &fileConfigStore{istioDirectory: dir}
}

func (f *fileConfigStore) CreateIstioConfig(bindingID string, configuration []model.Config) error {
	ymlPath := path.Join(f.istioDirectory, bindingID) + ".yml"
	log.Printf("PATH to istio config: %v\n", ymlPath)

	fileContent, err := config.ToYamlDocuments(configuration)
	if nil != err {
		return err
	}
	err = ioutil.WriteFile(ymlPath, []byte(fileContent), 0644)
	if nil != err {
		return fmt.Errorf("unable to write istio configuration to file %s: %v", ymlPath, err)
	}
	return nil
}

func (f *fileConfigStore) DeleteBinding(bindingID string) error {
	fileName := path.Join(f.istioDirectory, bindingID) + ".yml"
	err := os.Remove(fileName)
	if err != nil {
		return fmt.Errorf("Error during removal of file %s: %v", fileName, err)
	}
	return nil
}

func (f *fileConfigStore) CreateService(bindingID string, service *v1.Service) (*v1.Service, error) {
	return nil, errors.New("CreateService is not available for file system")
}
