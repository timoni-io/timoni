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

var tmpNotificationProvider = lwhelper.ID()

type NotificationProvider interface {
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
	SMTP_Host() string
	SetSMTP_Host(value string) out.Info
	SMTP_Login() string
	SetSMTP_Login(value string) out.Info
	SMTP_Password() string
	SetSMTP_Password(value string) out.Info
	SMTP_Port() int64
	SetSMTP_Port(value int64) out.Info
	SMTP_SenderEmail() string
	SetSMTP_SenderEmail(value string) out.Info
	SMTP_SenderName() string
	SetSMTP_SenderName(value string) out.Info
	Variant() NotificationProviderVariantT
	SetVariant(value NotificationProviderVariantT) out.Info
	Delete() out.Info
}
type notificationProviderS struct {
	out.DontUseMeInfoS

	IDC      string
	CreatedC int64
	UpdatedC int64

	EnabledC          bool
	NameC             string
	OrganizationC     string
	SMTP_HostC        string
	SMTP_LoginC       string
	SMTP_PasswordC    string
	SMTP_PortC        int64
	SMTP_SenderEmailC string
	SMTP_SenderNameC  string
	VariantC          NotificationProviderVariantT
}

func (o *notificationProviderS) dataMapEN() map[string]string {
	return map[string]string{
		"ID":               o.IDC,
		"Created":          fmt.Sprint(o.CreatedC),
		"Updated":          fmt.Sprint(o.UpdatedC),
		"Enabled":          fmt.Sprint(o.EnabledC),
		"Name":             o.NameC,
		"Organization":     o.OrganizationC,
		"SMTP_Host":        o.SMTP_HostC,
		"SMTP_Login":       o.SMTP_LoginC,
		"SMTP_Password":    o.SMTP_PasswordC,
		"SMTP_Port":        fmt.Sprint(o.SMTP_PortC),
		"SMTP_SenderEmail": o.SMTP_SenderEmailC,
		"SMTP_SenderName":  o.SMTP_SenderNameC,
		"Variant":          o.VariantC.EN(),
	}
}

// ---

func (o *notificationProviderS) AddListener(l binding.DataListener) {
	fmt.Println("NotificationProvider AddListener")
}

func (o *notificationProviderS) RemoveListener(l binding.DataListener) {
	fmt.Println("NotificationProvider RemoveListener")
}

// ---

func (o *notificationProviderS) InfoJSON() []byte {
	buf, _ := json.MarshalIndent(o, "", "  ")
	return buf
}

func (o *notificationProviderS) InfoPrint() {
	fmt.Println(string(o.InfoJSON()))
}

// ---
// ID

func (o *notificationProviderS) ID() string {
	if o == nil {
		return ""
	}
	return o.IDC
}

// ---
// Created

func (o *notificationProviderS) Created() int64 {
	if o == nil {
		return 0
	}
	return o.CreatedC
}

// ---
// Updated

func (o *notificationProviderS) Updated() int64 {
	if o == nil {
		return 0
	}
	return o.UpdatedC
}

// ---
// Enabled

func (o *notificationProviderS) Enabled() bool {
	return o.EnabledC
}

func (o *notificationProviderS) SetEnabled(value bool) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"NotificationProviderSetEnabled",
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

func (o *notificationProviderS) Name() string {
	return o.NameC
}

func (o *notificationProviderS) SetName(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"NotificationProviderSetName",
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

func (o *notificationProviderS) Organization() Organization {
	return OrganizationGetByID(o.OrganizationC)
}

func (o *notificationProviderS) SetOrganization(value Organization) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"NotificationProviderSetOrganization",
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
// SMTP_Host

func (o *notificationProviderS) SMTP_Host() string {
	return o.SMTP_HostC
}

func (o *notificationProviderS) SetSMTP_Host(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"NotificationProviderSetSMTP_Host",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.SMTP_HostC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// SMTP_Login

func (o *notificationProviderS) SMTP_Login() string {
	return o.SMTP_LoginC
}

func (o *notificationProviderS) SetSMTP_Login(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"NotificationProviderSetSMTP_Login",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.SMTP_LoginC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// SMTP_Password

func (o *notificationProviderS) SMTP_Password() string {
	return o.SMTP_PasswordC
}

func (o *notificationProviderS) SetSMTP_Password(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"NotificationProviderSetSMTP_Password",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.SMTP_PasswordC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// SMTP_Port

func (o *notificationProviderS) SMTP_Port() int64 {
	return o.SMTP_PortC
}

func (o *notificationProviderS) SetSMTP_Port(value int64) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"NotificationProviderSetSMTP_Port",
		req_setInt64S{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.SMTP_PortC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// SMTP_SenderEmail

func (o *notificationProviderS) SMTP_SenderEmail() string {
	return o.SMTP_SenderEmailC
}

func (o *notificationProviderS) SetSMTP_SenderEmail(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"NotificationProviderSetSMTP_SenderEmail",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.SMTP_SenderEmailC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// SMTP_SenderName

func (o *notificationProviderS) SMTP_SenderName() string {
	return o.SMTP_SenderNameC
}

func (o *notificationProviderS) SetSMTP_SenderName(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"NotificationProviderSetSMTP_SenderName",
		req_setStringS{
			ID:       o.IDC,
			NewValue: value,
		},
		response,
	)
	if response.NotValid() {
		return response
	}

	o.SMTP_SenderNameC = value
	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Variant

func (o *notificationProviderS) Variant() NotificationProviderVariantT {
	return o.VariantC
}

func (o *notificationProviderS) SetVariant(value NotificationProviderVariantT) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	response := new(out.DontUseMeInfoS)
	client.call(
		"NotificationProviderSetVariant",
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

func NotificationProviderGetByID(ID string) NotificationProvider {

	response := &notificationProviderS{}
	client.call(
		"NotificationProviderGetByID",
		req_oneS{
			ID: ID,
		},
		response,
	)

	return response
}

// ---

func (o *notificationProviderS) Delete() out.Info {

	response := new(out.DontUseMeInfoS)
	client.call(
		"NotificationProviderDelete",
		req_oneS{
			ID: o.ID(),
		},
		response,
	)

	return response
}

// ---

type req_NotificationProviderCreateS struct {
	Name         string
	Organization string
	Variant      NotificationProviderVariantT
}

func NotificationProviderCreate(
	Name string,
	Organization Organization,
	Variant NotificationProviderVariantT,

) NotificationProvider {

	response := &notificationProviderS{}
	client.call(
		"NotificationProviderCreate",
		req_NotificationProviderCreateS{
			Name:         Name,
			Organization: Organization.ID(),
			Variant:      Variant,
		},
		response,
	)
	return response
}

// ----------------------------------------------------- table list:

type notificationProviderList interface {
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
	First() NotificationProvider
	GetByID(id string) NotificationProvider
	Iter() []NotificationProvider
	Refresh() out.Info
	SetWhere(where string) out.Info
	SetOrder(order string) out.Info
	SetOffset(offset int) out.Info
	SetLimit(limit int) out.Info

	AddListener(dl binding.DataListener)
	RemoveListener(dl binding.DataListener)
	GetItem(index int) (binding.DataItem, error)
}

type notificationProviderListS struct {
	out.DontUseMeInfoS

	query   req_listQueryS
	IDs     []string
	IDtoIdx map[string]int
	M       map[string]*notificationProviderS

	dataListener map[binding.DataListener]bool
}

func (o *notificationProviderListS) InfoJSON() []byte {
	buf, _ := json.MarshalIndent(o, "", "  ")
	return buf
}

func (o *notificationProviderListS) InfoPrint() {
	fmt.Println(string(o.InfoJSON()))
}

func (o *notificationProviderListS) AddListener(dl binding.DataListener) {
	// fmt.Println("NotificationProviderList AddListener")
	if o.dataListener == nil {
		o.dataListener = map[binding.DataListener]bool{}
	}
	o.dataListener[dl] = true
}

func (o *notificationProviderListS) RemoveListener(dl binding.DataListener) {
	// fmt.Println("NotificationProviderList RemoveListener")
	delete(o.dataListener, dl)
}

func (o *notificationProviderListS) GetItem(index int) (binding.DataItem, error) {
	// fmt.Println("NotificationProviderList GetItem")
	return o.M[o.IDs[index]], nil
}

func (o *notificationProviderListS) Length() int {
	// fmt.Println("NotificationProviderList Length")
	return len(o.IDs)
}

//---

func NotificationProviderList(where, order string, offset, limit int) notificationProviderList {

	response := &notificationProviderListS{
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

func (o *notificationProviderListS) SetWhere(where string) out.Info {
	o.query = req_listQueryS{
		Where: where,
	}
	return o.Refresh()
}

func (o *notificationProviderListS) SetOrder(order string) out.Info {
	o.query = req_listQueryS{
		Order: order,
	}
	return o.Refresh()
}

func (o *notificationProviderListS) SetOffset(offset int) out.Info {
	o.query = req_listQueryS{
		Offset: offset,
	}
	return o.Refresh()
}

func (o *notificationProviderListS) SetLimit(limit int) out.Info {
	o.query = req_listQueryS{
		Limit: limit,
	}
	return o.Refresh()
}

func (o *notificationProviderListS) First() NotificationProvider {
	for _, obj := range o.M {
		return out.CatchError(obj, nil)
	}

	res := &notificationProviderS{}
	res.InfoAddTrace(out.NotFound, "", 0)
	return res
}

func (o *notificationProviderListS) Iter() []NotificationProvider {

	if o.NotValid() {
		return nil
	}

	res := []NotificationProvider{}
	for _, id := range o.IDs {
		res = append(res, out.CatchError(o.M[id], nil))
	}
	return res
}

func (o *notificationProviderListS) Refresh() out.Info {

	if o.query.Where == "nil" {
		return o
	}

	response := &notificationProviderListS{
		query:        o.query,
		IDs:          []string{},
		IDtoIdx:      map[string]int{},
		M:            map[string]*notificationProviderS{},
		dataListener: o.dataListener,
	}

	client.call(
		"NotificationProviderList",
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

func (o *notificationProviderListS) GetByID(id string) NotificationProvider {
	return out.CatchError(o.M[id], nil)
}

//---

func (o *notificationProviderListS) GetByName(name string) NotificationProvider {

	for _, obj := range o.M {
		if obj.NameC == name {
			return out.CatchError(obj, nil)
		}
	}

	res := &notificationProviderS{}
	res.InfoAddTrace(out.NotFound, "", 0)
	return res
}
