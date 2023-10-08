package db

import (
	"fmt"
	"lib/tlog"
)

type elementConfigS struct {
	elementS
}

func (element *elementConfigS) RebuildImage(imageID string, user *UserS) *tlog.RecordS {
	return nil
}

func (element *elementConfigS) GetImage() *ImageS {
	return nil
}

func (element *elementConfigS) RestartAllPods(user *UserS) *tlog.RecordS {
	return nil
}

func (element *elementConfigS) check(user *UserS) *tlog.RecordS {
	if element.ToDelete {
		return nil
	}
	return element.elementS.check(user)
}

func (element *elementConfigS) Save(user *UserS) *tlog.RecordS {
	if err := element.check(user); err != nil {
		return err
	}
	return elementSave(element, user)
}

func (element *elementConfigS) Merge(el EnvElementS) *tlog.RecordS {
	e, ok := el.(*elementConfigS)
	if !ok {
		return tlog.Error(fmt.Sprintln("wrong type, expected", element.GetType(), "got", el.GetType()))
	}
	return element.elementS.Merge(&e.elementS)
}

func (element *elementConfigS) CopySecrets(el EnvElementS) *tlog.RecordS {
	e, ok := el.(*elementConfigS)
	if !ok {
		return tlog.Error(fmt.Sprintln("wrong type, expected", element.GetType(), "got", el.GetType()))
	}
	return element.elementS.CopySecrets(&e.elementS)
}

func (element *elementConfigS) DeleteFromKube() *tlog.RecordS {
	return nil
}

func (element *elementConfigS) GetScale() *ElementScaleS {
	return &ElementScaleS{}
}