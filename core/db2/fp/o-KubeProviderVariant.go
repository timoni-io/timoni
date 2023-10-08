package fp

// This file is automatically generated, manual editing is not recommended.

type KubeProviderVariantT uint16

const (
	KubeProviderVariant_ExistingCluster KubeProviderVariantT = 1
)

var translationMapEN_KubeProviderVariant = map[KubeProviderVariantT]string{
	1: "ExistingCluster",
}

func (o KubeProviderVariantT) EN() string { return translationMapEN_KubeProviderVariant[o] }

var translationMapPL_KubeProviderVariant = map[KubeProviderVariantT]string{
	1: "ExistingCluster",
}

func (o KubeProviderVariantT) PL() string { return translationMapPL_KubeProviderVariant[o] }
