package db2

// This file is automatically generated, manual editing is not recommended.

import (
	"encoding/json"
	"fmt"
	"time"

	"fyne.io/fyne/v2/data/binding"
	"github.com/lukx33/lwhelper"
	"github.com/lukx33/lwhelper/out"
	"gorm.io/gorm"
)

var tmpImageBuilder = lwhelper.ID()

type ImageBuilder interface {
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
	TimoniExternal_Token() string
	SetTimoniExternal_Token(value string) out.Info
	TimoniExternal_URL() string
	SetTimoniExternal_URL(value string) out.Info
	Timoni_MaxNumberOfConcurrentBuilds() int64
	SetTimoni_MaxNumberOfConcurrentBuilds(value int64) out.Info
	Timoni_Token() string
	SetTimoni_Token(value string) out.Info
	Variant() ImageBuilderVariantT
	SetVariant(value ImageBuilderVariantT) out.Info
	Delete() out.Info
}
type imageBuilderS struct {
	out.DontUseMeInfoS `gorm:"-"`

	IDC      string `gorm:"column:ID;primaryKey"`
	CreatedC int64  `gorm:"column:Created;autoCreateTime"`
	UpdatedC int64  `gorm:"column:Updated;autoUpdateTime"`

	EnabledC                            bool                 `gorm:"column:Enabled"`
	NameC                               string               `gorm:"column:Name"`
	OrganizationC                       string               `gorm:"column:Organization"`
	TimoniExternal_TokenC               string               `gorm:"column:TimoniExternal_Token"`
	TimoniExternal_URLC                 string               `gorm:"column:TimoniExternal_URL"`
	Timoni_MaxNumberOfConcurrentBuildsC int64                `gorm:"column:Timoni_MaxNumberOfConcurrentBuilds"`
	Timoni_TokenC                       string               `gorm:"column:Timoni_Token"`
	VariantC                            ImageBuilderVariantT `gorm:"column:Variant"`
}

func (imageBuilderS) TableName() string {
	return "ImageBuilder"
}

// ---

func (o *imageBuilderS) AfterFind(tx *gorm.DB) error {
	return nil
}

// ---

func (o *imageBuilderS) InfoJSON() []byte {
	buf, _ := json.MarshalIndent(o, "", "  ")
	return buf
}

func (o *imageBuilderS) InfoPrint() {
	fmt.Println(string(o.InfoJSON()))
}

func (o *imageBuilderS) AddListener(l binding.DataListener) {
	fmt.Println("ImageBuilder AddListener")
}

func (o *imageBuilderS) RemoveListener(l binding.DataListener) {
	fmt.Println("ImageBuilder RemoveListener")
}

func (o *imageBuilderS) Delete() out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	dbLock.Lock()
	defer dbLock.Unlock()

	err := dbConnection.Where("ID = ?", o.IDC).Delete(&imageBuilderS{}).Error
	if err != nil {
		return out.New(err)
	}

	return out.NewSuccess()
}

func imageBuilderCreateOrUpdate(srcB []byte) {

	src := &imageBuilderS{}
	json.Unmarshal(srcB, src)
	id := src.IDC
	// src.InfoPrint()

	if id == "" {
		panic("cos tu jest nie tak")
	}

	srcMap := map[string]any{}
	json.Unmarshal(src.InfoJSON(), &srcMap)
	delete(srcMap, "trace")
	delete(srcMap, "result")
	delete(srcMap, "vars")
	delete(srcMap, "IDC")
	// out.PrintJSON(srcMap)

	dbLock.Lock()
	defer dbLock.Unlock()

	var exists bool
	if out.New(dbConnection.Model(&imageBuilderS{}).Select("count(*) > 0").
		Where("id = ?", id).Find(&exists).Error).NotValid() {
		return
	}
	if !exists {
		// new item, creating
		// fmt.Println("new item")
		out.New(dbConnection.Create(src).Error)
		return
	}

	// looking for changes
	// fmt.Println("changes")
	out.New(dbConnection.Model(&imageBuilderS{}).Where("ID = ?", id).Updates(srcMap).Error)
}

// ---
// ID

func (o *imageBuilderS) ID() string {
	if o == nil {
		return ""
	}
	return o.IDC
}

// ---
// Created

func (o *imageBuilderS) Created() int64 {
	if o == nil {
		return 0
	}
	return o.CreatedC
}

// ---
// Updated

func (o *imageBuilderS) Updated() int64 {
	if o == nil {
		return 0
	}
	return o.UpdatedC
}

// ---
// Enabled

func (o *imageBuilderS) Enabled() bool {
	if o == nil {
		return false
	}
	return o.EnabledC
}

func (o *imageBuilderS) SetEnabled(value bool) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	// TODO: if current value == new value, then there is no point in changing anything

	oldValue := o.EnabledC
	o.EnabledC = value

	dbLock.Lock()
	defer dbLock.Unlock()

	err := dbConnection.Model(&imageBuilderS{}).Where("ID = ?", o.IDC).Update("Enabled", value).Error
	if err != nil {
		o.EnabledC = oldValue
		return out.New(err)
	}

	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Name

func (o *imageBuilderS) Name() string {
	if o == nil {
		return ""
	}
	return o.NameC
}

func (o *imageBuilderS) SetName(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	// TODO: if current value == new value, then there is no point in changing anything

	oldValue := o.NameC
	o.NameC = value

	dbLock.Lock()
	defer dbLock.Unlock()

	err := dbConnection.Model(&imageBuilderS{}).Where("ID = ?", o.IDC).Update("Name", value).Error
	if err != nil {
		o.NameC = oldValue
		return out.New(err)
	}

	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Organization

func (o *imageBuilderS) Organization() Organization {
	if o == nil {
		return nil
	}
	return OrganizationGetByID(o.OrganizationC)
}

func (o *imageBuilderS) SetOrganization(value Organization) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	// TODO: if current value == new value, then there is no point in changing anything

	oldValue := o.OrganizationC
	o.OrganizationC = value.ID()

	dbLock.Lock()
	defer dbLock.Unlock()

	err := dbConnection.Model(&imageBuilderS{}).Where("ID = ?", o.IDC).Update("Organization", value.ID()).Error
	if err != nil {
		o.OrganizationC = oldValue
		return out.New(err)
	}

	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// TimoniExternal_Token

func (o *imageBuilderS) TimoniExternal_Token() string {
	if o == nil {
		return ""
	}
	return o.TimoniExternal_TokenC
}

func (o *imageBuilderS) SetTimoniExternal_Token(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	// TODO: if current value == new value, then there is no point in changing anything

	oldValue := o.TimoniExternal_TokenC
	o.TimoniExternal_TokenC = value

	dbLock.Lock()
	defer dbLock.Unlock()

	err := dbConnection.Model(&imageBuilderS{}).Where("ID = ?", o.IDC).Update("TimoniExternal_Token", value).Error
	if err != nil {
		o.TimoniExternal_TokenC = oldValue
		return out.New(err)
	}

	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// TimoniExternal_URL

func (o *imageBuilderS) TimoniExternal_URL() string {
	if o == nil {
		return ""
	}
	return o.TimoniExternal_URLC
}

func (o *imageBuilderS) SetTimoniExternal_URL(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	// TODO: if current value == new value, then there is no point in changing anything

	oldValue := o.TimoniExternal_URLC
	o.TimoniExternal_URLC = value

	dbLock.Lock()
	defer dbLock.Unlock()

	err := dbConnection.Model(&imageBuilderS{}).Where("ID = ?", o.IDC).Update("TimoniExternal_URL", value).Error
	if err != nil {
		o.TimoniExternal_URLC = oldValue
		return out.New(err)
	}

	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Timoni_MaxNumberOfConcurrentBuilds

func (o *imageBuilderS) Timoni_MaxNumberOfConcurrentBuilds() int64 {
	if o == nil {
		return 0
	}
	return o.Timoni_MaxNumberOfConcurrentBuildsC
}

func (o *imageBuilderS) SetTimoni_MaxNumberOfConcurrentBuilds(value int64) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	// TODO: if current value == new value, then there is no point in changing anything

	oldValue := o.Timoni_MaxNumberOfConcurrentBuildsC
	o.Timoni_MaxNumberOfConcurrentBuildsC = value

	dbLock.Lock()
	defer dbLock.Unlock()

	err := dbConnection.Model(&imageBuilderS{}).Where("ID = ?", o.IDC).Update("Timoni_MaxNumberOfConcurrentBuilds", value).Error
	if err != nil {
		o.Timoni_MaxNumberOfConcurrentBuildsC = oldValue
		return out.New(err)
	}

	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Timoni_Token

func (o *imageBuilderS) Timoni_Token() string {
	if o == nil {
		return ""
	}
	return o.Timoni_TokenC
}

func (o *imageBuilderS) SetTimoni_Token(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	// TODO: if current value == new value, then there is no point in changing anything

	oldValue := o.Timoni_TokenC
	o.Timoni_TokenC = value

	dbLock.Lock()
	defer dbLock.Unlock()

	err := dbConnection.Model(&imageBuilderS{}).Where("ID = ?", o.IDC).Update("Timoni_Token", value).Error
	if err != nil {
		o.Timoni_TokenC = oldValue
		return out.New(err)
	}

	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Variant

func (o *imageBuilderS) Variant() ImageBuilderVariantT {
	if o == nil {
		return 0
	}
	return o.VariantC
}

func (o *imageBuilderS) SetVariant(value ImageBuilderVariantT) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	// TODO: if current value == new value, then there is no point in changing anything

	oldValue := o.VariantC
	o.VariantC = value

	dbLock.Lock()
	defer dbLock.Unlock()

	err := dbConnection.Model(&imageBuilderS{}).Where("ID = ?", o.IDC).Update("Variant", value).Error
	if err != nil {
		o.VariantC = oldValue
		return out.New(err)
	}

	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ----------------------------------------------------- table list:

type imageBuilderList interface {
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
	First() ImageBuilder
	GetByID(id string) ImageBuilder
	Iter() []ImageBuilder
	Refresh() out.Info
	SetWhere(where string) out.Info
	SetOrder(order string) out.Info
	SetOffset(offset int) out.Info
	SetLimit(limit int) out.Info

	AddListener(dl binding.DataListener)
	RemoveListener(dl binding.DataListener)
	GetItem(index int) (binding.DataItem, error)
}

type imageBuilderListS struct {
	out.DontUseMeInfoS

	query        req_listQueryS
	M            map[string]*imageBuilderS
	IDs          []string
	IDtoIdx      map[string]int
	dataListener map[binding.DataListener]bool
}

//---

func (o *imageBuilderListS) InfoJSON() []byte {
	buf, _ := json.MarshalIndent(o, "", "  ")
	return buf
}

func (o *imageBuilderListS) InfoPrint() {
	fmt.Println(string(o.InfoJSON()))
}

func (o *imageBuilderListS) AddListener(dl binding.DataListener) {
	// fmt.Println("ImageBuilderList AddListener")
	if o.dataListener == nil {
		o.dataListener = map[binding.DataListener]bool{}
	}
	o.dataListener[dl] = true
}

func (o *imageBuilderListS) RemoveListener(dl binding.DataListener) {
	// fmt.Println("ImageBuilderList RemoveListener")
	delete(o.dataListener, dl)
}

func (o *imageBuilderListS) GetItem(index int) (binding.DataItem, error) {
	// fmt.Println("ImageBuilderList GetItem")
	return o.M[o.IDs[index]], nil
}

func (o *imageBuilderListS) Length() int {
	// fmt.Println("ImageBuilderList Length")
	return len(o.IDs)
}

func (o *imageBuilderListS) SetWhere(where string) out.Info {
	o.query = req_listQueryS{
		Where: where,
	}
	return o.Refresh()
}

func (o *imageBuilderListS) SetOrder(order string) out.Info {
	o.query = req_listQueryS{
		Order: order,
	}
	return o.Refresh()
}

func (o *imageBuilderListS) SetOffset(offset int) out.Info {
	o.query = req_listQueryS{
		Offset: offset,
	}
	return o.Refresh()
}

func (o *imageBuilderListS) SetLimit(limit int) out.Info {
	o.query = req_listQueryS{
		Limit: limit,
	}
	return o.Refresh()
}

//---

func ImageBuilderList(where, order string, offset, limit int) imageBuilderList {

	response := &imageBuilderListS{
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

func (o *imageBuilderListS) Refresh() out.Info {

	if o.query.Where == "nil" {
		return o
	}

	res := &imageBuilderListS{
		query:        o.query,
		IDs:          []string{},
		IDtoIdx:      map[string]int{},
		M:            map[string]*imageBuilderS{},
		dataListener: o.dataListener,
	}

	dbLock.Lock()
	defer dbLock.Unlock()

	st := dbConnection.Model(&imageBuilderS{})
	if o.query.Where != "" {
		st.Where(o.query.Where)
	}
	if o.query.Order == "" {
		o.query.Order = "Created"
	}
	if o.query.Limit == 0 {
		o.query.Limit = 30
	}

	responseList := []*imageBuilderS{}
	if out.CatchError(res,
		st.Order(o.query.Order).Offset(o.query.Offset).Limit(o.query.Limit).Find(&responseList).Error,
	).NotValid() {
		return res
	}
	// out.PrintJSON(responseList)

	for idx, entry := range responseList {
		res.IDs = append(res.IDs, entry.IDC)
		res.IDtoIdx[entry.IDC] = idx
		res.M[entry.IDC] = entry
	}

	out.CatchError(res, nil)
	*o = *res

	for dl := range o.dataListener {
		// fmt.Println(">>>>>>>>>>>>>>>>>> dataListener", dl)
		dl.DataChanged()
	}
	return res
}

func (o *imageBuilderListS) First() ImageBuilder {
	for _, obj := range o.M {
		return out.CatchError(obj, nil)
	}

	res := &imageBuilderS{}
	res.InfoAddTrace(out.NotFound, "", 0)
	return res
}

func (o *imageBuilderListS) GetByID(id string) ImageBuilder {
	res, exist := o.M[id]
	if !exist {
		res.InfoAddTrace(out.NotFound, "", 0)
		return res
	}
	return out.CatchError(res, nil)
}

func (o *imageBuilderListS) Iter() []ImageBuilder {

	if o.NotValid() {
		return nil
	}

	res := []ImageBuilder{}
	for _, id := range o.IDs {
		res = append(res, out.CatchError(o.M[id], nil))
	}
	return res
}

// ---

func ImageBuilderGetByID(key string) ImageBuilder {

	res := &imageBuilderS{}
	if key == "" {
		res.InfoAddTrace(out.NotFound, "", 0)
		return res
	}

	dbLock.Lock()
	defer dbLock.Unlock()

	out.CatchError(res, dbConnection.Where("ID = ?", key).First(res).Error)

	if res.NotValid() && res.InfoLastTrace().Message == "record not found" {
		res.Result = out.NotFound
	}

	return res
}

//---

func ImageBuilderCreate(
	Name string,
	Organization Organization,
	Variant ImageBuilderVariantT,
) ImageBuilder {

	// TODO: Input data validation

	now := time.Now()
	obj := &imageBuilderS{
		IDC:           lwhelper.ID(),
		CreatedC:      now.Unix(),
		UpdatedC:      now.Unix(),
		NameC:         Name,
		OrganizationC: Organization.ID(),
		VariantC:      Variant,
	}

	obj.EnabledC = true
	obj.Timoni_MaxNumberOfConcurrentBuildsC = 1

	// Saving data to the database
	dbLock.Lock()
	defer dbLock.Unlock()

	return out.CatchError(obj, dbConnection.Create(obj).Error)
}

//---