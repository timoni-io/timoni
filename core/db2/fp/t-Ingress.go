package fp

// This file is automatically generated, manual editing is not recommended.

import (
	"encoding/json"
	"fmt"
	"time"

	"fyne.io/fyne/v2/data/binding"
	"github.com/lukx33/lwhelper"
	"github.com/lukx33/lwhelper/out"
)

var tmpIngress = lwhelper.ID()

type Ingress interface {
	NotValid() bool
	InfoAddTrace(result out.ResultT, msg string, skipFrames int)
	InfoAddCause(parent out.Info) out.Info
	InfoAddVar(name string, value any) out.Info
	InfoResult() out.ResultT
	InfoTraces() []out.TraceS
	InfoLastTrace() out.TraceS
	InfoJSON() []byte
	InfoPrint()

	ID() string
	Created() int64
	Updated() int64

	Enabled() bool
	SetEnabled(value bool) out.Info
	Name() string
	SetName(value string) out.Info
	Organization() Organization
	SetOrganization(value Organization) out.Info
	Traefik_Annotations() string
	SetTraefik_Annotations(value string) out.Info
	Traefik_Internal() bool
	SetTraefik_Internal(value bool) out.Info
	Variant() IngressVariantT
	SetVariant(value IngressVariantT) out.Info
	Delete() out.Info
}
type ingressS struct {
	out.DontUseMeInfoS

	IDC      string
	CreatedC int64
	UpdatedC int64

	EnabledC             bool
	NameC                string
	OrganizationC        string
	Traefik_AnnotationsC string
	Traefik_InternalC    bool
	VariantC             IngressVariantT
}

func (o *ingressS) dataMapEN() map[string]string {
	return map[string]string{
		"ID":                  o.IDC,
		"Created":             fmt.Sprint(o.CreatedC),
		"Updated":             fmt.Sprint(o.UpdatedC),
		"Enabled":             fmt.Sprint(o.EnabledC),
		"Name":                o.NameC,
		"Organization":        o.OrganizationC,
		"Traefik_Annotations": o.Traefik_AnnotationsC,
		"Traefik_Internal":    fmt.Sprint(o.Traefik_InternalC),
		"Variant":             o.VariantC.EN(),
	}
}

// ---

func (o *ingressS) AddListener(l binding.DataListener) {
	fmt.Println("Ingress AddListener")
}

func (o *ingressS) RemoveListener(l binding.DataListener) {
	fmt.Println("Ingress RemoveListener")
}

// ---

func (o *ingressS) InfoJSON() []byte {
	buf, _ := json.MarshalIndent(o, "", "  ")
	return buf
}

func (o *ingressS) InfoPrint() {
	fmt.Println(string(o.InfoJSON()))
}

// ---
// ID

func (o *ingressS) ID() string {
	if o == nil {
		return ""
	}
	return o.IDC
}

// ---
// Created

func (o *ingressS) Created() int64 {
	if o == nil {
		return 0
	}
	return o.CreatedC
}

// ---
// Updated

func (o *ingressS) Updated() int64 {
	if o == nil {
		return 0
	}
	return o.UpdatedC
}

// ---
// Enabled

func (o *ingressS) Enabled() bool {
	return o.EnabledC
}

func (o *ingressS) SetEnabled(value bool) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"IngressSetEnabled",
		req_setBoolS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.EnabledC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Name

func (o *ingressS) Name() string {
	return o.NameC
}

func (o *ingressS) SetName(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"IngressSetName",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.NameC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Organization

func (o *ingressS) Organization() Organization {
	return OrganizationGetByID(o.OrganizationC)
}

func (o *ingressS) SetOrganization(value Organization) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"IngressSetOrganization",
		req_setRelationS{
			ID:     o.IDC,
			NewKey: value.ID(),
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.OrganizationC = value.ID()
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Traefik_Annotations

func (o *ingressS) Traefik_Annotations() string {
	return o.Traefik_AnnotationsC
}

func (o *ingressS) SetTraefik_Annotations(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"IngressSetTraefik_Annotations",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.Traefik_AnnotationsC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Traefik_Internal

func (o *ingressS) Traefik_Internal() bool {
	return o.Traefik_InternalC
}

func (o *ingressS) SetTraefik_Internal(value bool) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"IngressSetTraefik_Internal",
		req_setBoolS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.Traefik_InternalC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Variant

func (o *ingressS) Variant() IngressVariantT {
	return o.VariantC
}

func (o *ingressS) SetVariant(value IngressVariantT) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"IngressSetVariant",
		req_setUInt16S{
			ID:       o.IDC,
			NewValue: uint16(value),
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.VariantC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---

func IngressGetByID(ID string) Ingress {

	response := &ingressS{}
	client.call(
		"IngressGetByID",
		req_oneS{
			ID: ID,
		},
		response,
	)

	return response
}

// ---

func (o *ingressS) Delete() out.Info {

	response := new(out.DontUseMeInfoS)
	client.call(
		"IngressDelete",
		req_oneS{
			ID: o.ID(),
		},
		response,
	)

	return response
}

// ---

type req_IngressCreateS struct {
	Name         string
	Organization string
	Variant      IngressVariantT
}

func IngressCreate(
	Name string,
	Organization Organization,
	Variant IngressVariantT,

) Ingress {

	response := &ingressS{}
	client.call(
		"IngressCreate",
		req_IngressCreateS{
			Name:         Name,
			Organization: Organization.ID(),
			Variant:      Variant,
		},
		response,
	)
	return response
}

// ----------------------------------------------------- table list:

type ingressList interface {
	NotValid() bool
	InfoAddTrace(result out.ResultT, msg string, skipFrames int)
	InfoAddCause(parent out.Info) out.Info
	InfoAddVar(name string, value any) out.Info
	InfoResult() out.ResultT
	InfoTraces() []out.TraceS
	InfoLastTrace() out.TraceS
	InfoJSON() []byte
	InfoPrint()

	Length() int
	First() Ingress
	GetByID(id string) Ingress
	Iter() []Ingress
	Refresh() out.Info
	SetWhere(where string) out.Info
	SetOrder(order string) out.Info
	SetOffset(offset int) out.Info
	SetLimit(limit int) out.Info

	AddListener(dl binding.DataListener)
	RemoveListener(dl binding.DataListener)
	GetItem(index int) (binding.DataItem, error)
}

type ingressListS struct {
	out.DontUseMeInfoS

	query   req_listQueryS
	IDs     []string
	IDtoIdx map[string]int
	M       map[string]*ingressS

	dataListener map[binding.DataListener]bool
}

func (o *ingressListS) InfoJSON() []byte {
	buf, _ := json.MarshalIndent(o, "", "  ")
	return buf
}

func (o *ingressListS) InfoPrint() {
	fmt.Println(string(o.InfoJSON()))
}

func (o *ingressListS) AddListener(dl binding.DataListener) {
	// fmt.Println("IngressList AddListener")
	if o.dataListener == nil {
		o.dataListener = map[binding.DataListener]bool{}
	}
	o.dataListener[dl] = true
}

func (o *ingressListS) RemoveListener(dl binding.DataListener) {
	// fmt.Println("IngressList RemoveListener")
	delete(o.dataListener, dl)
}

func (o *ingressListS) GetItem(index int) (binding.DataItem, error) {
	// fmt.Println("IngressList GetItem")
	return o.M[o.IDs[index]], nil
}

func (o *ingressListS) Length() int {
	// fmt.Println("IngressList Length")
	return len(o.IDs)
}

//---

func IngressList(where, order string, offset, limit int) ingressList {

	response := &ingressListS{
		query: req_listQueryS{
			Where:  where,
			Order:  order,
			Offset: offset,
			Limit:  limit,
		},
	}

	if where == "nil" {
		return response
	}

	response.Refresh()
	return response
}

func (o *ingressListS) SetWhere(where string) out.Info {
	o.query = req_listQueryS{
		Where: where,
	}
	return o.Refresh()
}

func (o *ingressListS) SetOrder(order string) out.Info {
	o.query = req_listQueryS{
		Order: order,
	}
	return o.Refresh()
}

func (o *ingressListS) SetOffset(offset int) out.Info {
	o.query = req_listQueryS{
		Offset: offset,
	}
	return o.Refresh()
}

func (o *ingressListS) SetLimit(limit int) out.Info {
	o.query = req_listQueryS{
		Limit: limit,
	}
	return o.Refresh()
}

func (o *ingressListS) First() Ingress {
	for _, obj := range o.M {
		return out.CatchError(obj, nil)
	}

	res := &ingressS{}
	res.InfoAddTrace(out.NotFound, "", 0)
	return res
}

func (o *ingressListS) Iter() []Ingress {

	if o.NotValid() {
		return nil
	}

	res := []Ingress{}
	for _, id := range o.IDs {
		res = append(res, out.CatchError(o.M[id], nil))
	}
	return res
}

func (o *ingressListS) Refresh() out.Info {

	if o.query.Where == "nil" {
		return o
	}

	response := &ingressListS{
		query:        o.query,
		IDs:          []string{},
		IDtoIdx:      map[string]int{},
		M:            map[string]*ingressS{},
		dataListener: o.dataListener,
	}

	client.call(
		"IngressList",
		o.query,
		response,
	)
	*o = *response

	for dl := range o.dataListener {
		// fmt.Println(">>>>>>>>>>>>>>>>>> dataListener", dl)
		dl.DataChanged()
	}
	return o
}

func (o *ingressListS) GetByID(id string) Ingress {
	return out.CatchError(o.M[id], nil)
}

//---

func (o *ingressListS) GetByName(name string) Ingress {

	for _, obj := range o.M {
		if obj.NameC == name {
			return out.CatchError(obj, nil)
		}
	}

	res := &ingressS{}
	res.InfoAddTrace(out.NotFound, "", 0)
	return res
}
