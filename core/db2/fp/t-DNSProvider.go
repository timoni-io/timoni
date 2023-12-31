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

var tmpDNSProvider = lwhelper.ID()

type DNSProvider interface {
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

	Azure_ClientID() string
	SetAzure_ClientID(value string) out.Info
	Azure_Email() string
	SetAzure_Email(value string) out.Info
	Azure_HOSTED_ZONE_NAME() string
	SetAzure_HOSTED_ZONE_NAME(value string) out.Info
	Azure_Key() string
	SetAzure_Key(value string) out.Info
	Azure_RESOURCE_GROUP_NAME() string
	SetAzure_RESOURCE_GROUP_NAME(value string) out.Info
	Azure_SUBSCRIPTION_ID() string
	SetAzure_SUBSCRIPTION_ID(value string) out.Info
	Azure_TenantID() string
	SetAzure_TenantID(value string) out.Info
	Cloudflare_Email() string
	SetCloudflare_Email(value string) out.Info
	Cloudflare_Key() string
	SetCloudflare_Key(value string) out.Info
	Enabled() bool
	SetEnabled(value bool) out.Info
	Name() string
	SetName(value string) out.Info
	Organization() Organization
	SetOrganization(value Organization) out.Info
	Variant() DNSProviderVariantT
	SetVariant(value DNSProviderVariantT) out.Info
	Delete() out.Info
}
type dNSProviderS struct {
	out.DontUseMeInfoS

	IDC      string
	CreatedC int64
	UpdatedC int64

	Azure_ClientIDC            string
	Azure_EmailC               string
	Azure_HOSTED_ZONE_NAMEC    string
	Azure_KeyC                 string
	Azure_RESOURCE_GROUP_NAMEC string
	Azure_SUBSCRIPTION_IDC     string
	Azure_TenantIDC            string
	Cloudflare_EmailC          string
	Cloudflare_KeyC            string
	EnabledC                   bool
	NameC                      string
	OrganizationC              string
	VariantC                   DNSProviderVariantT
}

func (o *dNSProviderS) dataMapEN() map[string]string {
	return map[string]string{
		"ID":                        o.IDC,
		"Created":                   fmt.Sprint(o.CreatedC),
		"Updated":                   fmt.Sprint(o.UpdatedC),
		"Azure_ClientID":            o.Azure_ClientIDC,
		"Azure_Email":               o.Azure_EmailC,
		"Azure_HOSTED_ZONE_NAME":    o.Azure_HOSTED_ZONE_NAMEC,
		"Azure_Key":                 o.Azure_KeyC,
		"Azure_RESOURCE_GROUP_NAME": o.Azure_RESOURCE_GROUP_NAMEC,
		"Azure_SUBSCRIPTION_ID":     o.Azure_SUBSCRIPTION_IDC,
		"Azure_TenantID":            o.Azure_TenantIDC,
		"Cloudflare_Email":          o.Cloudflare_EmailC,
		"Cloudflare_Key":            o.Cloudflare_KeyC,
		"Enabled":                   fmt.Sprint(o.EnabledC),
		"Name":                      o.NameC,
		"Organization":              o.OrganizationC,
		"Variant":                   o.VariantC.EN(),
	}
}

// ---

func (o *dNSProviderS) AddListener(l binding.DataListener) {
	fmt.Println("DNSProvider AddListener")
}

func (o *dNSProviderS) RemoveListener(l binding.DataListener) {
	fmt.Println("DNSProvider RemoveListener")
}

// ---

func (o *dNSProviderS) InfoJSON() []byte {
	buf, _ := json.MarshalIndent(o, "", "  ")
	return buf
}

func (o *dNSProviderS) InfoPrint() {
	fmt.Println(string(o.InfoJSON()))
}

// ---
// ID

func (o *dNSProviderS) ID() string {
	if o == nil {
		return ""
	}
	return o.IDC
}

// ---
// Created

func (o *dNSProviderS) Created() int64 {
	if o == nil {
		return 0
	}
	return o.CreatedC
}

// ---
// Updated

func (o *dNSProviderS) Updated() int64 {
	if o == nil {
		return 0
	}
	return o.UpdatedC
}

// ---
// Azure_ClientID

func (o *dNSProviderS) Azure_ClientID() string {
	return o.Azure_ClientIDC
}

func (o *dNSProviderS) SetAzure_ClientID(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"DNSProviderSetAzure_ClientID",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.Azure_ClientIDC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Azure_Email

func (o *dNSProviderS) Azure_Email() string {
	return o.Azure_EmailC
}

func (o *dNSProviderS) SetAzure_Email(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"DNSProviderSetAzure_Email",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.Azure_EmailC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Azure_HOSTED_ZONE_NAME

func (o *dNSProviderS) Azure_HOSTED_ZONE_NAME() string {
	return o.Azure_HOSTED_ZONE_NAMEC
}

func (o *dNSProviderS) SetAzure_HOSTED_ZONE_NAME(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"DNSProviderSetAzure_HOSTED_ZONE_NAME",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.Azure_HOSTED_ZONE_NAMEC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Azure_Key

func (o *dNSProviderS) Azure_Key() string {
	return o.Azure_KeyC
}

func (o *dNSProviderS) SetAzure_Key(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"DNSProviderSetAzure_Key",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.Azure_KeyC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Azure_RESOURCE_GROUP_NAME

func (o *dNSProviderS) Azure_RESOURCE_GROUP_NAME() string {
	return o.Azure_RESOURCE_GROUP_NAMEC
}

func (o *dNSProviderS) SetAzure_RESOURCE_GROUP_NAME(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"DNSProviderSetAzure_RESOURCE_GROUP_NAME",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.Azure_RESOURCE_GROUP_NAMEC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Azure_SUBSCRIPTION_ID

func (o *dNSProviderS) Azure_SUBSCRIPTION_ID() string {
	return o.Azure_SUBSCRIPTION_IDC
}

func (o *dNSProviderS) SetAzure_SUBSCRIPTION_ID(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"DNSProviderSetAzure_SUBSCRIPTION_ID",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.Azure_SUBSCRIPTION_IDC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Azure_TenantID

func (o *dNSProviderS) Azure_TenantID() string {
	return o.Azure_TenantIDC
}

func (o *dNSProviderS) SetAzure_TenantID(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"DNSProviderSetAzure_TenantID",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.Azure_TenantIDC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Cloudflare_Email

func (o *dNSProviderS) Cloudflare_Email() string {
	return o.Cloudflare_EmailC
}

func (o *dNSProviderS) SetCloudflare_Email(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"DNSProviderSetCloudflare_Email",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.Cloudflare_EmailC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Cloudflare_Key

func (o *dNSProviderS) Cloudflare_Key() string {
	return o.Cloudflare_KeyC
}

func (o *dNSProviderS) SetCloudflare_Key(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"DNSProviderSetCloudflare_Key",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.Cloudflare_KeyC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Enabled

func (o *dNSProviderS) Enabled() bool {
	return o.EnabledC
}

func (o *dNSProviderS) SetEnabled(value bool) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"DNSProviderSetEnabled",
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

func (o *dNSProviderS) Name() string {
	return o.NameC
}

func (o *dNSProviderS) SetName(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"DNSProviderSetName",
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

func (o *dNSProviderS) Organization() Organization {
	return OrganizationGetByID(o.OrganizationC)
}

func (o *dNSProviderS) SetOrganization(value Organization) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"DNSProviderSetOrganization",
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
// Variant

func (o *dNSProviderS) Variant() DNSProviderVariantT {
	return o.VariantC
}

func (o *dNSProviderS) SetVariant(value DNSProviderVariantT) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"DNSProviderSetVariant",
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

func DNSProviderGetByID(ID string) DNSProvider {

	response := &dNSProviderS{}
	client.call(
		"DNSProviderGetByID",
		req_oneS{
			ID: ID,
		},
		response,
	)

	return response
}

// ---

func (o *dNSProviderS) Delete() out.Info {

	response := new(out.DontUseMeInfoS)
	client.call(
		"DNSProviderDelete",
		req_oneS{
			ID: o.ID(),
		},
		response,
	)

	return response
}

// ---

type req_DNSProviderCreateS struct {
	Name         string
	Organization string
	Variant      DNSProviderVariantT
}

func DNSProviderCreate(
	Name string,
	Organization Organization,
	Variant DNSProviderVariantT,

) DNSProvider {

	response := &dNSProviderS{}
	client.call(
		"DNSProviderCreate",
		req_DNSProviderCreateS{
			Name:         Name,
			Organization: Organization.ID(),
			Variant:      Variant,
		},
		response,
	)
	return response
}

// ----------------------------------------------------- table list:

type dNSProviderList interface {
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
	First() DNSProvider
	GetByID(id string) DNSProvider
	Iter() []DNSProvider
	Refresh() out.Info
	SetWhere(where string) out.Info
	SetOrder(order string) out.Info
	SetOffset(offset int) out.Info
	SetLimit(limit int) out.Info

	AddListener(dl binding.DataListener)
	RemoveListener(dl binding.DataListener)
	GetItem(index int) (binding.DataItem, error)
}

type dNSProviderListS struct {
	out.DontUseMeInfoS

	query   req_listQueryS
	IDs     []string
	IDtoIdx map[string]int
	M       map[string]*dNSProviderS

	dataListener map[binding.DataListener]bool
}

func (o *dNSProviderListS) InfoJSON() []byte {
	buf, _ := json.MarshalIndent(o, "", "  ")
	return buf
}

func (o *dNSProviderListS) InfoPrint() {
	fmt.Println(string(o.InfoJSON()))
}

func (o *dNSProviderListS) AddListener(dl binding.DataListener) {
	// fmt.Println("DNSProviderList AddListener")
	if o.dataListener == nil {
		o.dataListener = map[binding.DataListener]bool{}
	}
	o.dataListener[dl] = true
}

func (o *dNSProviderListS) RemoveListener(dl binding.DataListener) {
	// fmt.Println("DNSProviderList RemoveListener")
	delete(o.dataListener, dl)
}

func (o *dNSProviderListS) GetItem(index int) (binding.DataItem, error) {
	// fmt.Println("DNSProviderList GetItem")
	return o.M[o.IDs[index]], nil
}

func (o *dNSProviderListS) Length() int {
	// fmt.Println("DNSProviderList Length")
	return len(o.IDs)
}

//---

func DNSProviderList(where, order string, offset, limit int) dNSProviderList {

	response := &dNSProviderListS{
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

func (o *dNSProviderListS) SetWhere(where string) out.Info {
	o.query = req_listQueryS{
		Where: where,
	}
	return o.Refresh()
}

func (o *dNSProviderListS) SetOrder(order string) out.Info {
	o.query = req_listQueryS{
		Order: order,
	}
	return o.Refresh()
}

func (o *dNSProviderListS) SetOffset(offset int) out.Info {
	o.query = req_listQueryS{
		Offset: offset,
	}
	return o.Refresh()
}

func (o *dNSProviderListS) SetLimit(limit int) out.Info {
	o.query = req_listQueryS{
		Limit: limit,
	}
	return o.Refresh()
}

func (o *dNSProviderListS) First() DNSProvider {
	for _, obj := range o.M {
		return out.CatchError(obj, nil)
	}

	res := &dNSProviderS{}
	res.InfoAddTrace(out.NotFound, "", 0)
	return res
}

func (o *dNSProviderListS) Iter() []DNSProvider {

	if o.NotValid() {
		return nil
	}

	res := []DNSProvider{}
	for _, id := range o.IDs {
		res = append(res, out.CatchError(o.M[id], nil))
	}
	return res
}

func (o *dNSProviderListS) Refresh() out.Info {

	if o.query.Where == "nil" {
		return o
	}

	response := &dNSProviderListS{
		query:        o.query,
		IDs:          []string{},
		IDtoIdx:      map[string]int{},
		M:            map[string]*dNSProviderS{},
		dataListener: o.dataListener,
	}

	client.call(
		"DNSProviderList",
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

func (o *dNSProviderListS) GetByID(id string) DNSProvider {
	return out.CatchError(o.M[id], nil)
}

//---

func (o *dNSProviderListS) GetByName(name string) DNSProvider {

	for _, obj := range o.M {
		if obj.NameC == name {
			return out.CatchError(obj, nil)
		}
	}

	res := &dNSProviderS{}
	res.InfoAddTrace(out.NotFound, "", 0)
	return res
}
