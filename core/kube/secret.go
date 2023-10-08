package kube

import (
	"encoding/json"
	"errors"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type SecretS struct {
	KubeClient  *ClientS
	Namespace   string
	Name        string
	Type        corev1.SecretType
	Data        map[string][]byte
	Obj         *corev1.Secret
	Labels      map[string]string
	Annotations map[string]string
}

func (s *SecretS) CreateOrUpdate() (diff string, err error) {

	if s.KubeClient == nil {
		return "", errors.New("KubeClient cant be empty")
	}
	if s.Name == "" {
		return "", errors.New("Name cant be empty")
	}
	if s.Namespace == "" {
		return "", errors.New("Namespace cant be empty")
	}

	var secretOld []byte
	secret, err := s.KubeClient.API.CoreV1().Secrets(s.Namespace).Get(s.KubeClient.CTX, s.Name, metav1.GetOptions{})
	if err == nil {
		secretOld, err = json.Marshal(secret)
		if err != nil {
			panic(err)
		}

	} else {
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: s.Name,
			},
		}
	}

	secret.Type = s.Type
	secret.Data = s.Data

	secret.ObjectMeta.Labels = s.Labels
	secret.ObjectMeta.Annotations = s.Annotations

	// ---

	if len(secretOld) == 0 {
		s.Obj, err = s.KubeClient.API.CoreV1().Secrets(s.Namespace).Create(s.KubeClient.CTX, secret, metav1.CreateOptions{})
		time.Sleep(1 * time.Second)
		return "creating new obj", err
	}

	secretNew, err := json.Marshal(secret)
	if err != nil {
		panic(err)
	}

	patch, err := jsonpatch.CreateMergePatch(secretOld, secretNew)
	if err != nil {
		panic(err)
	}

	if len(patch) == 2 {
		return "", nil
	}

	s.Obj, err = s.KubeClient.API.CoreV1().Secrets(s.Namespace).Patch(s.KubeClient.CTX, s.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	return string(patch), err
}

func (s *SecretS) GetObj() *corev1.Secret {

	var err error
	s.Obj, err = s.KubeClient.API.CoreV1().Secrets(s.Namespace).Get(s.KubeClient.CTX, s.Name, metav1.GetOptions{})
	if err != nil {
		return nil
	}

	return s.Obj
}

func (s *SecretS) Exist() bool {
	return s.GetObj() != nil
}

func (s *SecretS) Delete() error {
	return s.KubeClient.API.CoreV1().Secrets(s.Namespace).Delete(s.KubeClient.CTX, s.Name, metav1.DeleteOptions{})
}
