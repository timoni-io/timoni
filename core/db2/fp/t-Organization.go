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

var tmpOrganization = lwhelper.ID()

type Organization interface {
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

	Admins() userList
	AddAdmins(value User) out.Info
	RemoveAdmins(value User) out.Info
	LogoSvgBase64() string
	SetLogoSvgBase64(value string) out.Info
	Name() string
	SetName(value string) out.Info
	PaymentPlan() PaymentPlanT
	SetPaymentPlan(value PaymentPlanT) out.Info
	Delete() out.Info
}
type organizationS struct {
	out.DontUseMeInfoS

	IDC      string
	CreatedC int64
	UpdatedC int64

	AdminsC        []byte
	AdminsMap      map[string]bool
	LogoSvgBase64C string
	NameC          string
	PaymentPlanC   PaymentPlanT
}

func (o *organizationS) dataMapEN() map[string]string {
	return map[string]string{
		"ID":            o.IDC,
		"Created":       fmt.Sprint(o.CreatedC),
		"Updated":       fmt.Sprint(o.UpdatedC),
		"LogoSvgBase64": o.LogoSvgBase64C,
		"Name":          o.NameC,
		"PaymentPlan":   o.PaymentPlanC.EN(),
	}
}

// ---

func (o *organizationS) AddListener(l binding.DataListener) {
	fmt.Println("Organization AddListener")
}

func (o *organizationS) RemoveListener(l binding.DataListener) {
	fmt.Println("Organization RemoveListener")
}

// ---

func (o *organizationS) InfoJSON() []byte {
	buf, _ := json.MarshalIndent(o, "", "  ")
	return buf
}

func (o *organizationS) InfoPrint() {
	fmt.Println(string(o.InfoJSON()))
}

// ---
// ID

func (o *organizationS) ID() string {
	if o == nil {
		return ""
	}
	return o.IDC
}

// ---
// Created

func (o *organizationS) Created() int64 {
	if o == nil {
		return 0
	}
	return o.CreatedC
}

// ---
// Updated

func (o *organizationS) Updated() int64 {
	if o == nil {
		return 0
	}
	return o.UpdatedC
}

// ---
// Admins

func (o *organizationS) Admins() userList {
	return UserList(lwhelper.QueryIn("ID", lwhelper.GetKeysFromMap(o.AdminsMap)), "", 0, 30)
}

func (o *organizationS) AddAdmins(value User) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	if o.AdminsMap == nil {
		o.AdminsMap = map[string]bool{}
	}

	response := new(out.DontUseMeInfoS)
	client.call(
		"OrganizationAddAdmins",
		req_setRelationS{
			ID:     o.IDC,
			NewKey: value.ID(),
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.AdminsMap[value.ID()] = true
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

func (o *organizationS) RemoveAdmins(value User) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	if o.AdminsMap == nil {
		o.AdminsMap = map[string]bool{}
	}

	response := new(out.DontUseMeInfoS)
	client.call(
		"OrganizationRemoveAdmins",
		req_setRelationS{
			ID:     o.IDC,
			NewKey: value.ID(),
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	delete(o.AdminsMap, value.ID())
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// LogoSvgBase64

func (o *organizationS) LogoSvgBase64() string {
	return o.LogoSvgBase64C
}

func (o *organizationS) SetLogoSvgBase64(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"OrganizationSetLogoSvgBase64",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.LogoSvgBase64C = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Name

func (o *organizationS) Name() string {
	return o.NameC
}

func (o *organizationS) SetName(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"OrganizationSetName",
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
// PaymentPlan

func (o *organizationS) PaymentPlan() PaymentPlanT {
	return o.PaymentPlanC
}

func (o *organizationS) SetPaymentPlan(value PaymentPlanT) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"OrganizationSetPaymentPlan",
		req_setUInt16S{
			ID:       o.IDC,
			NewValue: uint16(value),
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.PaymentPlanC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---

func OrganizationGetByID(ID string) Organization {

	response := &organizationS{}
	client.call(
		"OrganizationGetByID",
		req_oneS{
			ID: ID,
		},
		response,
	)

	return response
}

// ---

func (o *organizationS) Delete() out.Info {

	response := new(out.DontUseMeInfoS)
	client.call(
		"OrganizationDelete",
		req_oneS{
			ID: o.ID(),
		},
		response,
	)

	return response
}

// ---

type req_OrganizationCreateS struct {
	Name        string
	PaymentPlan PaymentPlanT
}

func OrganizationCreate(
	Name string,
	PaymentPlan PaymentPlanT,

) Organization {

	response := &organizationS{}
	client.call(
		"OrganizationCreate",
		req_OrganizationCreateS{
			Name:        Name,
			PaymentPlan: PaymentPlan,
		},
		response,
	)
	return response
}

// ----------------------------------------------------- table list:

type organizationList interface {
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
	First() Organization
	GetByID(id string) Organization
	Iter() []Organization
	Refresh() out.Info
	SetWhere(where string) out.Info
	SetOrder(order string) out.Info
	SetOffset(offset int) out.Info
	SetLimit(limit int) out.Info

	AddListener(dl binding.DataListener)
	RemoveListener(dl binding.DataListener)
	GetItem(index int) (binding.DataItem, error)
}

type organizationListS struct {
	out.DontUseMeInfoS

	query   req_listQueryS
	IDs     []string
	IDtoIdx map[string]int
	M       map[string]*organizationS

	dataListener map[binding.DataListener]bool
}

func (o *organizationListS) InfoJSON() []byte {
	buf, _ := json.MarshalIndent(o, "", "  ")
	return buf
}

func (o *organizationListS) InfoPrint() {
	fmt.Println(string(o.InfoJSON()))
}

func (o *organizationListS) AddListener(dl binding.DataListener) {
	// fmt.Println("OrganizationList AddListener")
	if o.dataListener == nil {
		o.dataListener = map[binding.DataListener]bool{}
	}
	o.dataListener[dl] = true
}

func (o *organizationListS) RemoveListener(dl binding.DataListener) {
	// fmt.Println("OrganizationList RemoveListener")
	delete(o.dataListener, dl)
}

func (o *organizationListS) GetItem(index int) (binding.DataItem, error) {
	// fmt.Println("OrganizationList GetItem")
	return o.M[o.IDs[index]], nil
}

func (o *organizationListS) Length() int {
	// fmt.Println("OrganizationList Length")
	return len(o.IDs)
}

//---

func OrganizationList(where, order string, offset, limit int) organizationList {

	response := &organizationListS{
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

func (o *organizationListS) SetWhere(where string) out.Info {
	o.query = req_listQueryS{
		Where: where,
	}
	return o.Refresh()
}

func (o *organizationListS) SetOrder(order string) out.Info {
	o.query = req_listQueryS{
		Order: order,
	}
	return o.Refresh()
}

func (o *organizationListS) SetOffset(offset int) out.Info {
	o.query = req_listQueryS{
		Offset: offset,
	}
	return o.Refresh()
}

func (o *organizationListS) SetLimit(limit int) out.Info {
	o.query = req_listQueryS{
		Limit: limit,
	}
	return o.Refresh()
}

func (o *organizationListS) First() Organization {
	for _, obj := range o.M {
		return out.CatchError(obj, nil)
	}

	res := &organizationS{}
	res.InfoAddTrace(out.NotFound, "", 0)
	return res
}

func (o *organizationListS) Iter() []Organization {

	if o.NotValid() {
		return nil
	}

	res := []Organization{}
	for _, id := range o.IDs {
		res = append(res, out.CatchError(o.M[id], nil))
	}
	return res
}

func (o *organizationListS) Refresh() out.Info {

	if o.query.Where == "nil" {
		return o
	}

	response := &organizationListS{
		query:        o.query,
		IDs:          []string{},
		IDtoIdx:      map[string]int{},
		M:            map[string]*organizationS{},
		dataListener: o.dataListener,
	}

	client.call(
		"OrganizationList",
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

func (o *organizationListS) GetByID(id string) Organization {
	return out.CatchError(o.M[id], nil)
}

//---

func (o *organizationListS) GetByName(name string) Organization {

	for _, obj := range o.M {
		if obj.NameC == name {
			return out.CatchError(obj, nil)
		}
	}

	res := &organizationS{}
	res.InfoAddTrace(out.NotFound, "", 0)
	return res
}
