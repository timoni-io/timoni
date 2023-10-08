package kube

import (
	"encoding/json"
	"errors"

	jsonpatch "github.com/evanphx/json-patch"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type CluserRoleRuleS struct {
	ApiGroups []string
	Resources []string
	Verbs     []string
}

type ClusterRoleS struct {
	KubeClient *ClientS
	Name       string
	Rules      []CluserRoleRuleS
}

// CreateOrUpdate creates role only when it does not exists
func (cr ClusterRoleS) CreateOrUpdate() (diff string, err error) {

	if cr.KubeClient == nil {
		return "", errors.New("KubeClient cant be empty")
	}
	if cr.Name == "" {
		return "", errors.New("name cant be empty")
	}
	if len(cr.Rules) == 0 {
		return "", errors.New("rules cant be empty")
	}

	// ---

	k := cr.KubeClient
	clusterRoles := k.API.RbacV1().ClusterRoles()

	rules := []rbacv1.PolicyRule{}
	for _, rule := range cr.Rules {
		rules = append(rules, rbacv1.PolicyRule{
			APIGroups: rule.ApiGroups,
			Resources: rule.Resources,
			Verbs:     rule.Verbs,
		})
	}

	var roleOld []byte
	role, err := clusterRoles.Get(k.CTX, cr.Name, metav1.GetOptions{})
	if err != nil {
		// create role and return
		_, err = clusterRoles.Create(
			k.CTX,
			&rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: cr.Name,
				},
				Rules: rules,
			},
			metav1.CreateOptions{},
		)
		if err != nil {
			panic(err)
		}
		return "creating new object", nil
	}

	// role already created -> update

	// save old role
	roleOld, err = json.Marshal(role)
	if err != nil {
		panic(err)
	}

	// set new fields
	role.Rules = rules

	// create diff
	roleNew, err := json.Marshal(role)
	if err != nil {
		panic(err)
	}

	patch, err := jsonpatch.CreateMergePatch(roleOld, roleNew)
	if err != nil {
		panic(err)
	}

	// check diff
	if len(patch) == 2 {
		// no changes
		return "", nil
	}

	_, err = clusterRoles.Patch(k.CTX, cr.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	return string(patch), err
}
