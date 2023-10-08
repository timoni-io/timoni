package db

func (element *elementConfigS) KubeApply() {
	defer PanicHandler()

	es := element.GetStatus()
	es.Alerts = []string{}

	res := elementCheckRequiredElements2(element)
	if res.Status == CheckStepResultSuccess {
		es.State = ElementStatusReady

	} else {
		es.State = ElementStatusFailed
		es.Alerts = append(es.Alerts, res.Msg)
	}
}

func elementCheckRequiredElements2(element *elementConfigS) CheckStepResultS {

	if element.VariablesDependence.Len() == 0 {
		return stepSuccess("skiped: no requirements")
	}

	for x := range element.VariablesDependence.Iter() {
		if !x.Value {
			return stepFail("require elements: " + x.Key)
		}
	}

	return stepSuccess("all requirements met")
}
