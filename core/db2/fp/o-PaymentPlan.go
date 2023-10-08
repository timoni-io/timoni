package fp

// This file is automatically generated, manual editing is not recommended.

type PaymentPlanT uint16

const (
	PaymentPlan_free    PaymentPlanT = 1
	PaymentPlan_basic   PaymentPlanT = 2
	PaymentPlan_premium PaymentPlanT = 3
)

var translationMapEN_PaymentPlan = map[PaymentPlanT]string{
	1: "free",
	2: "basic",
	3: "premium",
}

func (o PaymentPlanT) EN() string { return translationMapEN_PaymentPlan[o] }

var translationMapPL_PaymentPlan = map[PaymentPlanT]string{
	1: "darmowy",
	2: "podstawowy",
	3: "premium",
}

func (o PaymentPlanT) PL() string { return translationMapPL_PaymentPlan[o] }
