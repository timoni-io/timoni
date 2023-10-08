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

var tmpMetrics = lwhelper.ID()

type Metrics interface {
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
	Grafana() bool
	SetGrafana(value bool) out.Info
	GrafanaAdminPassword() string
	SetGrafanaAdminPassword(value string) out.Info
	Name() string
	SetName(value string) out.Info
	Organization() Organization
	SetOrganization(value Organization) out.Info
	VictoriaMetricsClusterSize() int64
	SetVictoriaMetricsClusterSize(value int64) out.Info
	Delete() out.Info
}
type metricsS struct {
	out.DontUseMeInfoS `gorm:"-"`

	IDC      string `gorm:"column:ID;primaryKey"`
	CreatedC int64  `gorm:"column:Created;autoCreateTime"`
	UpdatedC int64  `gorm:"column:Updated;autoUpdateTime"`

	EnabledC                    bool   `gorm:"column:Enabled"`
	GrafanaC                    bool   `gorm:"column:Grafana"`
	GrafanaAdminPasswordC       string `gorm:"column:GrafanaAdminPassword"`
	NameC                       string `gorm:"column:Name"`
	OrganizationC               string `gorm:"column:Organization"`
	VictoriaMetricsClusterSizeC int64  `gorm:"column:VictoriaMetricsClusterSize"`
}

func (metricsS) TableName() string {
	return "Metrics"
}

// ---

func (o *metricsS) AfterFind(tx *gorm.DB) error {
	return nil
}

// ---

func (o *metricsS) InfoJSON() []byte {
	buf, _ := json.MarshalIndent(o, "", "  ")
	return buf
}

func (o *metricsS) InfoPrint() {
	fmt.Println(string(o.InfoJSON()))
}

func (o *metricsS) AddListener(l binding.DataListener) {
	fmt.Println("Metrics AddListener")
}

func (o *metricsS) RemoveListener(l binding.DataListener) {
	fmt.Println("Metrics RemoveListener")
}

func (o *metricsS) Delete() out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	dbLock.Lock()
	defer dbLock.Unlock()

	err := dbConnection.Where("ID = ?", o.IDC).Delete(&metricsS{}).Error
	if err != nil {
		return out.New(err)
	}

	return out.NewSuccess()
}

func metricsCreateOrUpdate(srcB []byte) {

	src := &metricsS{}
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
	if out.New(dbConnection.Model(&metricsS{}).Select("count(*) > 0").
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
	out.New(dbConnection.Model(&metricsS{}).Where("ID = ?", id).Updates(srcMap).Error)
}

// ---
// ID

func (o *metricsS) ID() string {
	if o == nil {
		return ""
	}
	return o.IDC
}

// ---
// Created

func (o *metricsS) Created() int64 {
	if o == nil {
		return 0
	}
	return o.CreatedC
}

// ---
// Updated

func (o *metricsS) Updated() int64 {
	if o == nil {
		return 0
	}
	return o.UpdatedC
}

// ---
// Enabled

func (o *metricsS) Enabled() bool {
	if o == nil {
		return false
	}
	return o.EnabledC
}

func (o *metricsS) SetEnabled(value bool) out.Info {

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

	err := dbConnection.Model(&metricsS{}).Where("ID = ?", o.IDC).Update("Enabled", value).Error
	if err != nil {
		o.EnabledC = oldValue
		return out.New(err)
	}

	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Grafana

func (o *metricsS) Grafana() bool {
	if o == nil {
		return false
	}
	return o.GrafanaC
}

func (o *metricsS) SetGrafana(value bool) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	// TODO: if current value == new value, then there is no point in changing anything

	oldValue := o.GrafanaC
	o.GrafanaC = value

	dbLock.Lock()
	defer dbLock.Unlock()

	err := dbConnection.Model(&metricsS{}).Where("ID = ?", o.IDC).Update("Grafana", value).Error
	if err != nil {
		o.GrafanaC = oldValue
		return out.New(err)
	}

	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// GrafanaAdminPassword

func (o *metricsS) GrafanaAdminPassword() string {
	if o == nil {
		return ""
	}
	return o.GrafanaAdminPasswordC
}

func (o *metricsS) SetGrafanaAdminPassword(value string) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	// TODO: if current value == new value, then there is no point in changing anything

	oldValue := o.GrafanaAdminPasswordC
	o.GrafanaAdminPasswordC = value

	dbLock.Lock()
	defer dbLock.Unlock()

	err := dbConnection.Model(&metricsS{}).Where("ID = ?", o.IDC).Update("GrafanaAdminPassword", value).Error
	if err != nil {
		o.GrafanaAdminPasswordC = oldValue
		return out.New(err)
	}

	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Name

func (o *metricsS) Name() string {
	if o == nil {
		return ""
	}
	return o.NameC
}

func (o *metricsS) SetName(value string) out.Info {

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

	err := dbConnection.Model(&metricsS{}).Where("ID = ?", o.IDC).Update("Name", value).Error
	if err != nil {
		o.NameC = oldValue
		return out.New(err)
	}

	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// Organization

func (o *metricsS) Organization() Organization {
	if o == nil {
		return nil
	}
	return OrganizationGetByID(o.OrganizationC)
}

func (o *metricsS) SetOrganization(value Organization) out.Info {

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

	err := dbConnection.Model(&metricsS{}).Where("ID = ?", o.IDC).Update("Organization", value.ID()).Error
	if err != nil {
		o.OrganizationC = oldValue
		return out.New(err)
	}

	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ---
// VictoriaMetricsClusterSize

func (o *metricsS) VictoriaMetricsClusterSize() int64 {
	if o == nil {
		return 0
	}
	return o.VictoriaMetricsClusterSizeC
}

func (o *metricsS) SetVictoriaMetricsClusterSize(value int64) out.Info {

	if o == nil {
		return out.NewErrorMsg("object is nil")
	}
	if o.NotValid() {
		return out.NewErrorMsg("object is not valid")
	}

	// TODO: validation

	// TODO: if current value == new value, then there is no point in changing anything

	oldValue := o.VictoriaMetricsClusterSizeC
	o.VictoriaMetricsClusterSizeC = value

	dbLock.Lock()
	defer dbLock.Unlock()

	err := dbConnection.Model(&metricsS{}).Where("ID = ?", o.IDC).Update("VictoriaMetricsClusterSize", value).Error
	if err != nil {
		o.VictoriaMetricsClusterSizeC = oldValue
		return out.New(err)
	}

	o.UpdatedC = time.Now().Unix()
	return out.NewSuccess()
}

// ----------------------------------------------------- table list:

type metricsList interface {
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
	First() Metrics
	GetByID(id string) Metrics
	Iter() []Metrics
	Refresh() out.Info
	SetWhere(where string) out.Info
	SetOrder(order string) out.Info
	SetOffset(offset int) out.Info
	SetLimit(limit int) out.Info

	AddListener(dl binding.DataListener)
	RemoveListener(dl binding.DataListener)
	GetItem(index int) (binding.DataItem, error)
}

type metricsListS struct {
	out.DontUseMeInfoS

	query        req_listQueryS
	M            map[string]*metricsS
	IDs          []string
	IDtoIdx      map[string]int
	dataListener map[binding.DataListener]bool
}

//---

func (o *metricsListS) InfoJSON() []byte {
	buf, _ := json.MarshalIndent(o, "", "  ")
	return buf
}

func (o *metricsListS) InfoPrint() {
	fmt.Println(string(o.InfoJSON()))
}

func (o *metricsListS) AddListener(dl binding.DataListener) {
	// fmt.Println("MetricsList AddListener")
	if o.dataListener == nil {
		o.dataListener = map[binding.DataListener]bool{}
	}
	o.dataListener[dl] = true
}

func (o *metricsListS) RemoveListener(dl binding.DataListener) {
	// fmt.Println("MetricsList RemoveListener")
	delete(o.dataListener, dl)
}

func (o *metricsListS) GetItem(index int) (binding.DataItem, error) {
	// fmt.Println("MetricsList GetItem")
	return o.M[o.IDs[index]], nil
}

func (o *metricsListS) Length() int {
	// fmt.Println("MetricsList Length")
	return len(o.IDs)
}

func (o *metricsListS) SetWhere(where string) out.Info {
	o.query = req_listQueryS{
		Where: where,
	}
	return o.Refresh()
}

func (o *metricsListS) SetOrder(order string) out.Info {
	o.query = req_listQueryS{
		Order: order,
	}
	return o.Refresh()
}

func (o *metricsListS) SetOffset(offset int) out.Info {
	o.query = req_listQueryS{
		Offset: offset,
	}
	return o.Refresh()
}

func (o *metricsListS) SetLimit(limit int) out.Info {
	o.query = req_listQueryS{
		Limit: limit,
	}
	return o.Refresh()
}

//---

func MetricsList(where, order string, offset, limit int) metricsList {

	response := &metricsListS{
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

func (o *metricsListS) Refresh() out.Info {

	if o.query.Where == "nil" {
		return o
	}

	res := &metricsListS{
		query:        o.query,
		IDs:          []string{},
		IDtoIdx:      map[string]int{},
		M:            map[string]*metricsS{},
		dataListener: o.dataListener,
	}

	dbLock.Lock()
	defer dbLock.Unlock()

	st := dbConnection.Model(&metricsS{})
	if o.query.Where != "" {
		st.Where(o.query.Where)
	}
	if o.query.Order == "" {
		o.query.Order = "Created"
	}
	if o.query.Limit == 0 {
		o.query.Limit = 30
	}

	responseList := []*metricsS{}
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

func (o *metricsListS) First() Metrics {
	for _, obj := range o.M {
		return out.CatchError(obj, nil)
	}

	res := &metricsS{}
	res.InfoAddTrace(out.NotFound, "", 0)
	return res
}

func (o *metricsListS) GetByID(id string) Metrics {
	res, exist := o.M[id]
	if !exist {
		res.InfoAddTrace(out.NotFound, "", 0)
		return res
	}
	return out.CatchError(res, nil)
}

func (o *metricsListS) Iter() []Metrics {

	if o.NotValid() {
		return nil
	}

	res := []Metrics{}
	for _, id := range o.IDs {
		res = append(res, out.CatchError(o.M[id], nil))
	}
	return res
}

// ---

func MetricsGetByID(key string) Metrics {

	res := &metricsS{}
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

func MetricsCreate(
	Name string,
	Organization Organization,
) Metrics {

	// TODO: Input data validation

	now := time.Now()
	obj := &metricsS{
		IDC:           lwhelper.ID(),
		CreatedC:      now.Unix(),
		UpdatedC:      now.Unix(),
		NameC:         Name,
		OrganizationC: Organization.ID(),
	}

	obj.EnabledC = true
	obj.GrafanaC = true
	obj.VictoriaMetricsClusterSizeC = 1

	// Saving data to the database
	dbLock.Lock()
	defer dbLock.Unlock()

	return out.CatchError(obj, dbConnection.Create(obj).Error)
}

//---
