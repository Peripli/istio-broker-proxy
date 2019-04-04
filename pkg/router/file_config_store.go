package router

import (
	"istio.io/istio/pilot/pkg/model"
	"k8s.io/api/core/v1"
)

type fileConfigStore struct {
	istioDirectory string
}

var _ ConfigStore = &fileConfigStore{}

func (f *fileConfigStore) CreateIstioConfig(bindID string, config model.Config) error {
	//var fileName string
	//
	//ymlPath := path.Join(f.istioDirectory, fileName) + ".yml"
	//log.Printf("PATH to istio config: %v\n", ymlPath)
	//file, err := os.Create(ymlPath)
	//if nil != err {
	//	return fmt.Errorf("Unable to write istio configuration to file '%s': %s", fileName, err.Error())
	//}
	//defer file.Close()
	//
	//fileContent, err := config.ToYamlDocuments([]model.Config{configuration})
	//if nil != err {
	//	return err
	//}
	//_, err = file.Write([]byte(fileContent))
	//if nil != err {
	//	return err
	//}
	return nil
}

func (f *fileConfigStore) DeleteIstioConfig(string, string) error {
	panic("implement me")
}

func (f *fileConfigStore) CreateService(*v1.Service) (*v1.Service, error) {
	panic("implement me")
}

func (f *fileConfigStore) DeleteService(string) error {
	panic("implement me")
}

func (f *fileConfigStore) getNamespace() string {
	panic("implement me")
}
