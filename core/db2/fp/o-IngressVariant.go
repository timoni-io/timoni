package fp

// This file is automatically generated, manual editing is not recommended.

type IngressVariantT uint16

const (
	IngressVariant_Traefik IngressVariantT = 1
)

var translationMapEN_IngressVariant = map[IngressVariantT]string{
	1: "Traefik",
}

func (o IngressVariantT) EN() string { return translationMapEN_IngressVariant[o] }

var translationMapPL_IngressVariant = map[IngressVariantT]string{
	1: "Traefik",
}

func (o IngressVariantT) PL() string { return translationMapPL_IngressVariant[o] }
