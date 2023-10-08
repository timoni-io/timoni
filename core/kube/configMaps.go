package kube

import (
	"encoding/json"
	"errors"

	jsonpatch "github.com/evanphx/json-patch"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ConfigMapsS ...
type ConfigMapsS struct {
	KubeClient *ClientS
	Namespace  string
	Name       string
	Data       map[string]string
	Obj        *corev1.ConfigMap
}

// CreateOrUpdate ...
func (s *ConfigMapsS) CreateOrUpdate() (anyChange bool, err error) {

	if s.KubeClient == nil {
		return false, errors.New("KubeClient cant be empty")
	}
	if s.Name == "" {
		return false, errors.New("Name cant be empty")
	}
	if s.Namespace == "" {
		return false, errors.New("Namespace cant be empty")
	}

	var configMapsOld []byte
	configMaps, err := s.KubeClient.API.CoreV1().ConfigMaps(s.Namespace).Get(s.KubeClient.CTX, s.Name, metav1.GetOptions{})
	if err == nil {
		configMapsOld, err = json.Marshal(configMaps)
		if err != nil {
			panic(err)
		}

	} else {
		configMaps = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: s.Name,
			},
		}
	}

	configMaps.Data = s.Data

	// ---

	if len(configMapsOld) == 0 {
		s.Obj, err = s.KubeClient.API.CoreV1().ConfigMaps(s.Namespace).Create(s.KubeClient.CTX, configMaps, metav1.CreateOptions{})
		if err != nil {
			return false, err
		}

	} else {
		configMapsNew, err := json.Marshal(configMaps)
		if err != nil {
			panic(err)
		}

		patch, err := jsonpatch.CreateMergePatch(configMapsOld, configMapsNew)
		if err != nil {
			panic(err)
		}

		if len(patch) == 2 {
			return false, nil
		}

		s.Obj, err = s.KubeClient.API.CoreV1().ConfigMaps(s.Namespace).Patch(s.KubeClient.CTX, s.Name, types.MergePatchType, patch, metav1.PatchOptions{})
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (s *ConfigMapsS) GetObj() *corev1.ConfigMap {

	var err error
	s.Obj, err = s.KubeClient.API.CoreV1().ConfigMaps(s.Namespace).Get(s.KubeClient.CTX, s.Name, metav1.GetOptions{})
	if err != nil {
		return nil
	}

	return s.Obj
}

func (s *ConfigMapsS) Exist() bool {
	if s.GetObj() == nil {
		return false
	}
	return true
}
