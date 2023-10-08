package kube

import (
	"core/db2"
	"encoding/base64"
	"fmt"
	log "lib/tlog"
	"strings"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AccountsNamespace  = "users-viewers"
	kubeConfigTemplate = `
apiVersion: v1
kind: Config
clusters:
- name: default
  cluster:
    certificate-authority-data: {{ca}}
    server: https://{{server}}:6443
contexts:
- name: default
  context:
    cluster: default
    user: {{username}}
current-context: default
users:
- name: {{username}}
  user:
    token: {{token}}
`
)

// CreateServiceAccount only if account not exists
func CreateServiceAccount(kubeUsername string) {

	k := GetKube()

	if n := k.NamespaceGet(AccountsNamespace); n == nil {
		k.NamespaceCreate(AccountsNamespace)
	}

	sa := k.API.CoreV1().ServiceAccounts(AccountsNamespace)

	_, err := sa.Get(k.CTX, kubeUsername, metav1.GetOptions{})
	if err == nil {
		// user already exists
		return
	}

	_, err = sa.Create(
		k.CTX,
		&v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      kubeUsername,
				Namespace: AccountsNamespace,
			},
		},
		metav1.CreateOptions{},
	)
	if err != nil {
		panic(err)
	}
}

func GetKubeConfig(kubeUsername string) string {

	k := GetKube()

	sa := k.API.CoreV1().ServiceAccounts(AccountsNamespace)

	kubeUser, err := sa.Get(k.CTX, kubeUsername, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	var caCrt, token []byte
	for _, v := range kubeUser.Secrets {
		if strings.Contains(v.Name, kubeUsername+"-token") {
			secret, _ := k.API.CoreV1().Secrets(AccountsNamespace).Get(k.CTX, v.Name, metav1.GetOptions{})
			caCrt = secret.Data["ca.crt"]
			token = secret.Data["token"]
			break
		}
	}
	caCrtEncoded := base64.StdEncoding.EncodeToString(caCrt)

	return strings.NewReplacer(
		"{{server}}", db2.TheDomain.Name(),
		"{{ca}}", caCrtEncoded,
		"{{username}}", kubeUsername,
		"{{token}}", string(token),
	).Replace(kubeConfigTemplate)
}

func BindRole(kubeUsername, appID string) {

	k := GetKube()

	rb := k.API.RbacV1().RoleBindings(appID)
	rbName := fmt.Sprintf("%s-%s", appID, kubeUsername)

	_, err := rb.Get(k.CTX, rbName, metav1.GetOptions{})
	if err == nil {
		// bind already exists
		return
	}

	_, err = rb.Create(
		k.CTX,
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rbName,
				Namespace: appID,
			},
			Subjects: []rbacv1.Subject{{
				Kind:      "ServiceAccount",
				Name:      kubeUsername,
				Namespace: AccountsNamespace,
			}},
			RoleRef: rbacv1.RoleRef{
				Kind: "ClusterRole",
				Name: "timoni-viewer",
			},
		},
		metav1.CreateOptions{},
	)

	if err != nil {
		log.Error("Kube role bind", log.Vars{
			"kubeUser": kubeUsername,
			"env":      appID,
			"roleName": rbName,
			"error":    err,
		})
	} else {
		log.Info("Kube role bind", log.Vars{
			"kubeUser": kubeUsername,
			"env":      appID,
			"roleName": rbName,
		})
	}
}

func UnbindRole(kubeUsername, appID string) {

	k := GetKube()

	rb := k.API.RbacV1().RoleBindings(appID)
	rbName := fmt.Sprintf("%s-%s", appID, kubeUsername)

	_, err := rb.Get(k.CTX, rbName, metav1.GetOptions{})
	if err != nil {
		// bind already not exists
		return
	}

	err = rb.Delete(k.CTX, rbName, metav1.DeleteOptions{})

	if err != nil {
		log.Error("Kube role unbind", log.Vars{
			"kubeUser": kubeUsername,
			"env":      appID,
			"roleName": rbName,
			"error":    err,
		})
	} else {
		log.Info("Kube role unbind", log.Vars{
			"kubeUser": kubeUsername,
			"env":      appID,
			"roleName": rbName,
		})
	}
}
