package router

import (
	"errors"
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/config"
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

func NewFileConfigStore(dir string) ConfigStore {
	return &fileConfigStore{istioDirectory: dir}
}

func (f *fileConfigStore) CreateIstioConfig(bindingID string, configuration []model.Config) error {
	ymlPath := path.Join(f.istioDirectory, bindingID) + ".yml"
	log.Printf("PATH to istio config: %v\n", ymlPath)
	file, err := os.Create(ymlPath)
	if nil != err {
		return fmt.Errorf("Unable to write istio configuration to file '%s': %s", bindingID, err.Error())
	}
	defer file.Close()

	fileContent, err := config.ToYamlDocuments(configuration)
	if nil != err {
		return err
	}
	_, err = file.Write([]byte(fileContent))
	if nil != err {
		return err
	}
	return nil
}

func (f *fileConfigStore) DeleteBinding(bindingID string) error {
	fileName := path.Join(f.istioDirectory, bindingID) + ".yml"
	err := os.Remove(fileName)
	if err != nil {
		log.Printf("Ignoring error during removal of file %s: %v\n", fileName, err)
	}
	return nil
}

func (f *fileConfigStore) CreateService(bindingID string, service *v1.Service) (*v1.Service, error) {
	return nil, errors.New("CreateService is not available for file system")
}
