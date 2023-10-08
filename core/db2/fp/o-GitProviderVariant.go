package fp

// This file is automatically generated, manual editing is not recommended.

type GitProviderVariantT uint16

const (
	GitProviderVariant_LoginAndPassword GitProviderVariantT = 1
	GitProviderVariant_SSH              GitProviderVariantT = 2
	GitProviderVariant_Timoni           GitProviderVariantT = 5
)

var translationMapEN_GitProviderVariant = map[GitProviderVariantT]string{
	1: "LoginAndPassword",
	2: "SSH",
	5: "Timoni",
}

func (o GitProviderVariantT) EN() string { return translationMapEN_GitProviderVariant[o] }

var translationMapPL_GitProviderVariant = map[GitProviderVariantT]string{
	1: "LoginAndPassword",
	2: "SSH",
	5: "Timoni",
}

func (o GitProviderVariantT) PL() string { return translationMapPL_GitProviderVariant[o] }
