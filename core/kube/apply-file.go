package kube

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"lib/tlog"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/util/retry"
)

func (kubeClinet *ClientS) ApplyYamlFilesInDir(dirpath string, variables map[string]string) (success bool) {

	err := filepath.Walk(dirpath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
				kubeClinet.ApplyFile(path, variables)
			}

			return nil
		})

	return tlog.Error(err) == nil
}

func (kubeClinet *ClientS) ApplyFile(filepath string, variables map[string]string) {

	fmt.Println("Kube ApplyFile: " + filepath)

	b, err := os.ReadFile(filepath)
	if err != nil {
		tlog.Fatal(err)
	}
	for k, v := range variables {
		if strings.TrimSpace(v) == "" {
			tlog.Fatal("variable {{name}} has empty value", tlog.Vars{
				"name": k,
			})
		}
		b = bytes.ReplaceAll(b, []byte("{{"+k+"}}"), []byte(v))
	}

	dd, err := dynamic.NewForConfig(kubeClinet.Config)
	if err != nil {
		tlog.Fatal(err)
	}

	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(b), 100)
	for {
		var rawObj runtime.RawExtension
		if err = decoder.Decode(&rawObj); err != nil {
			break
		}

		if len(rawObj.Raw) == 0 {
			continue
		}

		obj, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			tlog.Fatal(err)
		}

		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			tlog.Fatal(err)
		}

		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

		gr, err := restmapper.GetAPIGroupResources(kubeClinet.API.Discovery())
		if err != nil {
			tlog.Fatal(err)
		}

		mapper := restmapper.NewDiscoveryRESTMapper(gr)
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			tlog.Fatal(err)
		}

		var dri dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			if unstructuredObj.GetNamespace() == "" {
				unstructuredObj.SetNamespace("default")
			}
			dri = dd.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())

		} else {
			dri = dd.Resource(mapping.Resource)
		}

		if _, err := dri.Apply(
			context.Background(),
			unstructuredObj.GetName(),
			unstructuredObj,
			metav1.ApplyOptions{
				FieldManager: "application/apply-patch",
			},
		); err != nil {
			tlog.Fatal(err, tlog.Vars{
				"file": filepath,
				"name": unstructuredObj.GetName(),
				"kind": unstructuredObj.GetObjectKind().GroupVersionKind().Kind,
			})
		}
	}
	if err != io.EOF {
		tlog.Fatal("eof ", err)
	}
}

func ApplyFromString(data string) error {
	client := GetKube()
	decoder := yamlutil.NewYAMLOrJSONDecoder(strings.NewReader(data), 4096)
	for {
		obj := &unstructured.Unstructured{}
		err := decoder.Decode(obj)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		gvk := obj.GroupVersionKind()
		gr, err := restmapper.GetAPIGroupResources(client.API.Discovery())
		if err != nil {
			return err
		}

		mapper := restmapper.NewDiscoveryRESTMapper(gr)
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}

		fmt.Println("Kube ApplyFromString", mapping.Resource, obj.GetName(), obj.GetNamespace())
		resourceClient := client.Dynamic.Resource(mapping.Resource)

		apply := func() error {
			_, err := resourceClient.Namespace(obj.GetNamespace()).
				Apply(client.CTX, obj.GetName(), obj, metav1.ApplyOptions{
					FieldManager: "application/apply-patch",
				})
			return err
		}

		err = retry.RetryOnConflict(retry.DefaultRetry, apply)
		if err != nil {
			return err
		}
	}

	return nil
}
